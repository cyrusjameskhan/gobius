package models

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gobius/common"
	"gobius/config"
	"gobius/ipfs"
	"gobius/utils"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/mr-tron/base58"
	"github.com/rs/zerolog"
)

type QwenInner struct {
	Prompt string `json:"prompt"`
	Seed   uint64 `json:"seed"`
}

type QwenPrompt struct {
	Input QwenInner `json:"input"`
}

type QwenModelResponse struct {
	Input  map[string]any `json:"input"`
	Output []string       `json:"output"`
}

type QwenTestModel struct {
	Model
	timeoutDuration     time.Duration
	ipfsTimeoutDuration time.Duration
	Filters             []MiningFilter
	config              *config.AppConfig
	client              *http.Client
	logger              zerolog.Logger
	ipfs                ipfs.IPFSClient
}

// Ensure QwenTestModel implements the Model interface.
var _ ModelInterface = (*QwenTestModel)(nil)

var QwenTestModelTemplate = Model{
	ID:       "",
	Mineable: true,
	Template: map[string]any{
		"meta": map[string]any{
			"title":       "Qwen",
			"description": "Qwen Sepolia Test Model",
			"version":     1,
			"input": []map[string]any{
				{
					"variable":    "prompt",
					"type":        "string",
					"required":    true,
					"default":     "",
					"description": "Input prompt",
				},
				{
					"variable":    "seed",
					"type":        "int",
					"required":    false,
					"default":     500,
					"description": "Seed for the random number generator.",
				},
			},
			"output": []map[string]any{
				{
					"filename": "out-1.txt",
					"type":     "text",
				},
			},
		},
	},
}

func NewQwenTestModel(client ipfs.IPFSClient, appConfig *config.AppConfig, logger zerolog.Logger) *QwenTestModel {

	model, ok := appConfig.BaseConfig.Models["qwen-test"]
	if !ok {
		return nil
	}

	if model.ID == "" {
		logger.Error().Str("model", "qwen-test").Msg("qwen-test model ID is empty")
		return nil
	}

	httpClient := &http.Client{
		Transport: &http.Transport{MaxIdleConnsPerHost: 10}, // Use a dedicated transport for test model
		// Timeout is now handled per-request via context
		// Timeout: time.Second * 30,
	}

	// Use model.ID (the hex string CID) as the key for the Cog map
	cogConfig, ok := appConfig.ML.Cog[model.ID]
	// Set default timeouts first
	var timeout time.Duration = 120 * time.Second    // Default inference timeout
	var ipfsTimeout time.Duration = 30 * time.Second // Default IPFS timeout
	if ok {
		// Parse inference timeout only if the string is not empty
		if cogConfig.HttpTimeout != "" {
			parsedTimeout, err := time.ParseDuration(cogConfig.HttpTimeout)
			if err != nil {
				logger.Warn().Err(err).Str("model", model.ID).Str("config_timeout", cogConfig.HttpTimeout).Msg("failed to parse model timeout from cog config, using default 120s")
				// Keep default timeout
			} else {
				timeout = parsedTimeout
			}
		} // Else: HttpTimeout is empty, silently use the default

		// Parse IPFS timeout only if the string is not empty
		if cogConfig.IpfsTimeout != "" {
			parsedIpfsTimeout, err := time.ParseDuration(cogConfig.IpfsTimeout)
			if err != nil {
				logger.Warn().Err(err).Str("model", model.ID).Str("config_ipfs_timeout", cogConfig.IpfsTimeout).Msg("failed to parse IPFS timeout from cog config, using default 30s")
				// Keep default ipfsTimeout
			} else {
				ipfsTimeout = parsedIpfsTimeout
			}
		} // Else: IpfsTimeout is empty, silently use the default

	} else {
		logger.Error().Str("model", model.ID).Msg("model ID not found in ML.Cog map, required for QwenTestModel. Using default timeout 120s")
		// Keep default timeout, but log as Error as it's unexpected for a Cog model
	}

	m := &QwenTestModel{
		Model:               QwenTestModelTemplate,
		timeoutDuration:     timeout,
		ipfsTimeoutDuration: ipfsTimeout, // Store the IPFS timeout
		config:              appConfig,
		Filters: []MiningFilter{
			{
				MinFee:  0,
				MinTime: 0,
			},
		},
		ipfs:   client,
		client: httpClient,
		logger: logger,
	}
	// set this from config for now
	m.Model.ID = model.ID
	return m
}

func (m *QwenTestModel) HydrateInput(preprocessedInput map[string]any, seed uint64) (InputHydrationResult, error) {
	input := make(map[string]any)

	// Helper functions for type conversion
	convertToInt := func(val any) (int, error) {
		switch v := val.(type) {
		case int:
			return v, nil
		case float64:
			return int(v), nil
		default:
			return 0, fmt.Errorf("cannot convert %T to int", val)
		}
	}

	convertToFloat := func(val any) (float64, error) {
		switch v := val.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		default:
			return 0, fmt.Errorf("cannot convert %T to float64", val)
		}
	}

	// Get template metadata for input validation
	templateMeta, ok := m.Model.Template.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid template format")
	}

	meta, ok := templateMeta["meta"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid meta format in template")
	}

	inputFields, ok := meta["input"].([]map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid input format in template meta")
	}

	// Process each input field according to template rules
	for _, field := range inputFields {
		varName := field["variable"].(string)
		fieldType := field["type"].(string)
		required, _ := field["required"].(bool)

		// Get value from input or use default
		value, exists := preprocessedInput[varName]

		// Check if required field is missing
		if required && !exists {
			return nil, fmt.Errorf("input missing required field (%s)", varName)
		}

		// If value exists, validate it
		if exists {
			// Validate type
			if err := validateType(value, fieldType, varName); err != nil {
				return nil, err
			}

			// Validate range for numeric types
			switch fieldType {
			case "int":
				intVal, err := convertToInt(value)
				if err != nil {
					return nil, err
				}

				// Check min if defined
				if minVal, ok := field["min"].(int); ok {
					if intVal < minVal {
						return nil, fmt.Errorf("input below minimum (%s): value %d, minimum %d",
							varName, intVal, minVal)
					}
				}

				// Check max if defined
				if maxVal, ok := field["max"].(int); ok {
					if intVal > maxVal {
						return nil, fmt.Errorf("input above maximum (%s): value %d, maximum %d",
							varName, intVal, maxVal)
					}
				}

			case "decimal":
				floatVal, err := convertToFloat(value)
				if err != nil {
					return nil, err
				}

				// Check min if defined
				if minVal, ok := field["min"].(float64); ok {
					if floatVal < minVal {
						return nil, fmt.Errorf("input below minimum (%s): value %f, minimum %f",
							varName, floatVal, minVal)
					}
				}

				// Check max if defined
				if maxVal, ok := field["max"].(float64); ok {
					if floatVal > maxVal {
						return nil, fmt.Errorf("input above maximum (%s): value %f, maximum %f",
							varName, floatVal, maxVal)
					}
				}

			case "string_enum", "int_enum":
				if err := validateEnum(value, field, varName, convertToInt); err != nil {
					return nil, err
				}
			}

			input[varName] = value
		} else {
			// Use default value if provided
			input[varName] = field["default"]
		}
	}

	// Convert validated input to the expected QwenInner format
	var inner QwenInner
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, &inner); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to QwenInner: %w", err)
	}

	// TODO: probable a better way to handle values we need to set
	inner.Seed = seed

	return QwenPrompt{Input: inner}, nil
}

func (m *QwenTestModel) GetID() string {
	return m.Model.ID
}

func (m *QwenTestModel) GetFiles(ctx context.Context, gpu *common.GPU, taskid string, input any) ([]ipfs.IPFSFile, error) {

	// Check if context is already canceled before doing anything
	if err := ctx.Err(); err != nil {
		m.logger.Warn().Err(err).Str("task", taskid).Msg("Context canceled before GetFiles execution")
		return nil, err
	}

	marshaledInput, _ := json.Marshal(input)

	req, err := http.NewRequestWithContext(ctx, "POST", gpu.Url, bytes.NewBuffer(marshaledInput))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	postResp, err := m.client.Do(req)
	if err != nil {
		// Check if the error is context deadline exceeded
		if errors.Is(err, context.DeadlineExceeded) {
			m.logger.Error().Err(err).Str("task", taskid).Str("gpu", gpu.Url).Msg("model inference request timed out")
			return nil, fmt.Errorf("model inference timed out: %w", err)
		}
		return nil, fmt.Errorf("failed to POST to GPU: %w", err)
	}
	defer postResp.Body.Close()

	// Check for non-OK status codes
	if postResp.StatusCode != http.StatusOK {
		// Handle specific 409 Conflict (GPU busy) status
		if postResp.StatusCode == http.StatusConflict {
			bodyBytes, _ := io.ReadAll(postResp.Body)
			m.logger.Warn().Str("task", taskid).Str("gpu", gpu.Url).Int("status", postResp.StatusCode).Str("body", string(bodyBytes)).Msg("GPU busy")
			// Return the specific non-retryable error
			return nil, ErrResourceBusy
		}
		// Handle other non-200 statuses as errors
		bodyBytes, _ := io.ReadAll(postResp.Body)
		return nil, fmt.Errorf("server returned non-200 status: %d - %s", postResp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(postResp.Body)
	if err != nil {
		return nil, err
	}

	var resp QwenModelResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	if len(resp.Output) != 1 {
		return nil, err
	}

	fileName := fmt.Sprintf("%d.%s.txt", gpu.ID, uuid.New().String())
	path := filepath.Join(m.config.CachePath, fileName)
	buffer := bytes.NewBufferString(resp.Output[0])

	return []ipfs.IPFSFile{{Name: "out-1.txt", Path: path, Buffer: buffer}}, nil
}

func (m *QwenTestModel) GetCID(ctx context.Context, gpu *common.GPU, taskid string, input any) ([]byte, error) {

	// Create a new context with the stored model-specific timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, m.timeoutDuration)
	defer cancel()

	// Use ExpRetryWithContext
	paths, err := utils.ExpRetryWithContext(timeoutCtx, m.logger, func() (any, error) {
		// Pass the timeout context to GetFiles
		return m.GetFiles(timeoutCtx, gpu, taskid, input)
	}, 3, 1000)
	if err != nil {
		return nil, err
	}

	// Note: IPFS pinning might need its own context/timeout strategy if it becomes slow
	// Create a new context for IPFS pinning with its specific timeout
	ipfsCtx, ipfsCancel := context.WithTimeout(ctx, m.ipfsTimeoutDuration)
	defer ipfsCancel()

	// Use ExpRetryWithContext
	cid58, err := utils.ExpRetryWithContext(ipfsCtx, m.logger, func() (any, error) {
		// Pass the ipfsCtx to PinFilesToIPFS
		return m.ipfs.PinFilesToIPFS(ipfsCtx, taskid, paths.([]ipfs.IPFSFile))
	}, 3, 1000)

	if err != nil {
		return nil, errors.New("cannot pin files to retrieve cid")
	}

	cidBytes, err := base58.Decode(cid58.(string))
	if err != nil {
		return nil, err
	}

	return cidBytes, nil
}

func (m *QwenTestModel) Validate(gpu *common.GPU, taskid string) error {

	testPrompt := QwenPrompt{
		Input: QwenInner{
			Prompt: "why is the meaning of life 42?",
			Seed:   1337,
		},
	}

	cid, err := m.GetCID(context.Background(), gpu, "startup-test-taskid", testPrompt)
	if err != nil {
		return err
	}

	expected := "0x122041fa6dba0027cf73097b4c75cae5f16122b4f1cfe85616bdccb6eaea711e7238"
	cidStr := "0x" + hex.EncodeToString(cid)
	if cidStr == expected {
		m.logger.Info().Str("model", m.GetID()).Str("cid", cidStr).Str("expected", expected).Msg("model CID matches expected CID")
	} else {
		m.logger.Error().Str("model", m.GetID()).Str("cid", cidStr).Str("expected", expected).Msg("model CID does not match expected CID")
		return errors.New("model CID does not match expected CID")
	}

	return nil
}
