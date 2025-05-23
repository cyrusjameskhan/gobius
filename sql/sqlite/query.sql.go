// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package db

import (
	"context"
	"database/sql"
	"strings"

	common "github.com/ethereum/go-ethereum/common"
	task "gobius/common"
)

const addIPFSCid = `-- name: AddIPFSCid :exec
INSERT INTO ipfs_cids (
  taskid, cid
) VALUES (
  ?, ?
)
`

type AddIPFSCidParams struct {
	Taskid task.TaskId
	Cid    []byte
}

func (q *Queries) AddIPFSCid(ctx context.Context, arg AddIPFSCidParams) error {
	_, err := q.db.ExecContext(ctx, addIPFSCid, arg.Taskid, arg.Cid)
	return err
}

const addOrUpdateTaskWithStatus = `-- name: AddOrUpdateTaskWithStatus :exec
INSERT INTO tasks (taskid, txhash, status)
VALUES (?, ?, ?) 
ON CONFLICT(taskid) DO UPDATE SET
    status = excluded.status
`

type AddOrUpdateTaskWithStatusParams struct {
	Taskid task.TaskId
	Txhash common.Hash
	Status int64
}

func (q *Queries) AddOrUpdateTaskWithStatus(ctx context.Context, arg AddOrUpdateTaskWithStatusParams) error {
	_, err := q.db.ExecContext(ctx, addOrUpdateTaskWithStatus, arg.Taskid, arg.Txhash, arg.Status)
	return err
}

const addTask = `-- name: AddTask :exec
INSERT INTO tasks(
  taskid, txhash, cumulativeGas
) VALUES (
  ?,?, ?
)
`

type AddTaskParams struct {
	Taskid        task.TaskId
	Txhash        common.Hash
	Cumulativegas float64
}

func (q *Queries) AddTask(ctx context.Context, arg AddTaskParams) error {
	_, err := q.db.ExecContext(ctx, addTask, arg.Taskid, arg.Txhash, arg.Cumulativegas)
	return err
}

const addTaskWithStatus = `-- name: AddTaskWithStatus :exec
INSERT INTO tasks(
  taskid, txhash, cumulativeGas, status
) VALUES (
  ?,?, ?, ? 
)
`

type AddTaskWithStatusParams struct {
	Taskid        task.TaskId
	Txhash        common.Hash
	Cumulativegas float64
	Status        int64
}

func (q *Queries) AddTaskWithStatus(ctx context.Context, arg AddTaskWithStatusParams) error {
	_, err := q.db.ExecContext(ctx, addTaskWithStatus,
		arg.Taskid,
		arg.Txhash,
		arg.Cumulativegas,
		arg.Status,
	)
	return err
}

const addTasks = `-- name: AddTasks :exec
INSERT INTO tasks(
  taskid, txhash, cumulativeGas
) VALUES (/*SLICE:taskids*/?,?, ?)
`

type AddTasksParams struct {
	Taskids       []task.TaskId
	Txhash        common.Hash
	Cumulativegas float64
}

func (q *Queries) AddTasks(ctx context.Context, arg AddTasksParams) error {
	query := addTasks
	var queryParams []interface{}
	if len(arg.Taskids) > 0 {
		for _, v := range arg.Taskids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:taskids*/?", strings.Repeat(",?", len(arg.Taskids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:taskids*/?", "NULL", 1)
	}
	queryParams = append(queryParams, arg.Txhash)
	queryParams = append(queryParams, arg.Cumulativegas)
	_, err := q.db.ExecContext(ctx, query, queryParams...)
	return err
}

const checkCommitmentExists = `-- name: CheckCommitmentExists :one
SELECT EXISTS(SELECT 1 FROM commitments WHERE taskid = ?)
`

func (q *Queries) CheckCommitmentExists(ctx context.Context, taskid task.TaskId) (int64, error) {
	row := q.db.QueryRowContext(ctx, checkCommitmentExists, taskid)
	var column_1 int64
	err := row.Scan(&column_1)
	return column_1, err
}

const createCommitment = `-- name: CreateCommitment :exec
INSERT INTO commitments (
  taskid, commitment, validator
) VALUES (
  ?, ?, ?
)
`

type CreateCommitmentParams struct {
	Taskid     task.TaskId
	Commitment task.TaskId
	Validator  common.Address
}

func (q *Queries) CreateCommitment(ctx context.Context, arg CreateCommitmentParams) error {
	_, err := q.db.ExecContext(ctx, createCommitment, arg.Taskid, arg.Commitment, arg.Validator)
	return err
}

const createSolution = `-- name: CreateSolution :exec
INSERT INTO solutions (
  taskid, cid, validator
) VALUES (
  ?, ?, ?
)
`

type CreateSolutionParams struct {
	Taskid    task.TaskId
	Cid       []byte
	Validator common.Address
}

func (q *Queries) CreateSolution(ctx context.Context, arg CreateSolutionParams) error {
	_, err := q.db.ExecContext(ctx, createSolution, arg.Taskid, arg.Cid, arg.Validator)
	return err
}

const deleteCommitment = `-- name: DeleteCommitment :exec
DELETE FROM commitments
WHERE taskid = ?
`

func (q *Queries) DeleteCommitment(ctx context.Context, taskid task.TaskId) error {
	_, err := q.db.ExecContext(ctx, deleteCommitment, taskid)
	return err
}

const deleteSolution = `-- name: DeleteSolution :exec
DELETE FROM solutions
WHERE taskid = ?
`

func (q *Queries) DeleteSolution(ctx context.Context, taskid task.TaskId) error {
	_, err := q.db.ExecContext(ctx, deleteSolution, taskid)
	return err
}

const deletedClaimedTask = `-- name: DeletedClaimedTask :execrows
DELETE FROM tasks WHERE taskid = ? AND status = 3
`

func (q *Queries) DeletedClaimedTask(ctx context.Context, taskid task.TaskId) (int64, error) {
	result, err := q.db.ExecContext(ctx, deletedClaimedTask, taskid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deletedCommitment = `-- name: DeletedCommitment :execrows
DELETE FROM commitments WHERE taskid = ?
`

func (q *Queries) DeletedCommitment(ctx context.Context, taskid task.TaskId) (int64, error) {
	result, err := q.db.ExecContext(ctx, deletedCommitment, taskid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deletedIPFSCid = `-- name: DeletedIPFSCid :execrows
DELETE FROM ipfs_cids WHERE taskid = ?
`

func (q *Queries) DeletedIPFSCid(ctx context.Context, taskid task.TaskId) (int64, error) {
	result, err := q.db.ExecContext(ctx, deletedIPFSCid, taskid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deletedSolution = `-- name: DeletedSolution :execrows
DELETE FROM solutions WHERE taskid = ?
`

func (q *Queries) DeletedSolution(ctx context.Context, taskid task.TaskId) (int64, error) {
	result, err := q.db.ExecContext(ctx, deletedSolution, taskid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deletedTask = `-- name: DeletedTask :execrows
DELETE FROM tasks WHERE taskid = ?
`

func (q *Queries) DeletedTask(ctx context.Context, taskid task.TaskId) (int64, error) {
	result, err := q.db.ExecContext(ctx, deletedTask, taskid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getAllTasks = `-- name: GetAllTasks :many
SELECT taskid, txhash, cumulativegas, status, claimtime FROM tasks
`

func (q *Queries) GetAllTasks(ctx context.Context) ([]Task, error) {
	rows, err := q.db.QueryContext(ctx, getAllTasks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Task
	for rows.Next() {
		var i Task
		if err := rows.Scan(
			&i.Taskid,
			&i.Txhash,
			&i.Cumulativegas,
			&i.Status,
			&i.Claimtime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCommitmentBatch = `-- name: GetCommitmentBatch :many
SELECT 
commitments.taskid, commitments.commitment, commitments.validator
FROM commitments
JOIN tasks ON commitments.taskid = tasks.taskid 
WHERE tasks.status = 1
ORDER BY commitments.added ASC 
LIMIT ?
`

type GetCommitmentBatchRow struct {
	Taskid     task.TaskId
	Commitment task.TaskId
	Validator  common.Address
}

// WHERE tasks.committed = false
func (q *Queries) GetCommitmentBatch(ctx context.Context, limit int64) ([]GetCommitmentBatchRow, error) {
	rows, err := q.db.QueryContext(ctx, getCommitmentBatch, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetCommitmentBatchRow
	for rows.Next() {
		var i GetCommitmentBatchRow
		if err := rows.Scan(&i.Taskid, &i.Commitment, &i.Validator); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCommitments = `-- name: GetCommitments :many
SELECT taskid, commitment, validator, added FROM commitments
ORDER BY added ASC
`

func (q *Queries) GetCommitments(ctx context.Context) ([]Commitment, error) {
	rows, err := q.db.QueryContext(ctx, getCommitments)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Commitment
	for rows.Next() {
		var i Commitment
		if err := rows.Scan(
			&i.Taskid,
			&i.Commitment,
			&i.Validator,
			&i.Added,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getIPFSCids = `-- name: GetIPFSCids :many
SELECT 
    taskid, cid, added
FROM ipfs_cids
ORDER BY added ASC 
LIMIT ?
`

func (q *Queries) GetIPFSCids(ctx context.Context, limit int64) ([]IpfsCid, error) {
	rows, err := q.db.QueryContext(ctx, getIPFSCids, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []IpfsCid
	for rows.Next() {
		var i IpfsCid
		if err := rows.Scan(&i.Taskid, &i.Cid, &i.Added); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPendingSolutionsCountPerValidator = `-- name: GetPendingSolutionsCountPerValidator :many
SELECT 
    solutions.validator,
    COUNT(solutions.taskid) AS solution_count
FROM solutions 
JOIN tasks ON solutions.taskid = tasks.taskid 
WHERE tasks.status = 2
GROUP BY solutions.validator
ORDER BY solution_count DESC
`

type GetPendingSolutionsCountPerValidatorRow struct {
	Validator     common.Address
	SolutionCount int64
}

func (q *Queries) GetPendingSolutionsCountPerValidator(ctx context.Context) ([]GetPendingSolutionsCountPerValidatorRow, error) {
	rows, err := q.db.QueryContext(ctx, getPendingSolutionsCountPerValidator)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPendingSolutionsCountPerValidatorRow
	for rows.Next() {
		var i GetPendingSolutionsCountPerValidatorRow
		if err := rows.Scan(&i.Validator, &i.SolutionCount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getQueuedTasks = `-- name: GetQueuedTasks :many
SELECT 
taskid, txhash
FROM tasks 
WHERE status = 0
`

type GetQueuedTasksRow struct {
	Taskid task.TaskId
	Txhash common.Hash
}

func (q *Queries) GetQueuedTasks(ctx context.Context) ([]GetQueuedTasksRow, error) {
	rows, err := q.db.QueryContext(ctx, getQueuedTasks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetQueuedTasksRow
	for rows.Next() {
		var i GetQueuedTasksRow
		if err := rows.Scan(&i.Taskid, &i.Txhash); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSolutionBatch = `-- name: GetSolutionBatch :many
SELECT 
solutions.taskid, solutions.cid 
FROM solutions 
JOIN tasks ON solutions.taskid = tasks.taskid 
WHERE tasks.status = 2 AND solutions.validator = ?
ORDER BY solutions.added ASC 
LIMIT ?
`

type GetSolutionBatchParams struct {
	Validator common.Address
	Limit     int64
}

type GetSolutionBatchRow struct {
	Taskid task.TaskId
	Cid    []byte
}

func (q *Queries) GetSolutionBatch(ctx context.Context, arg GetSolutionBatchParams) ([]GetSolutionBatchRow, error) {
	rows, err := q.db.QueryContext(ctx, getSolutionBatch, arg.Validator, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetSolutionBatchRow
	for rows.Next() {
		var i GetSolutionBatchRow
		if err := rows.Scan(&i.Taskid, &i.Cid); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSolutions = `-- name: GetSolutions :many
SELECT 
solutions.taskid, solutions.cid 
FROM solutions 
JOIN tasks ON solutions.taskid = tasks.taskid
`

type GetSolutionsRow struct {
	Taskid task.TaskId
	Cid    []byte
}

func (q *Queries) GetSolutions(ctx context.Context) ([]GetSolutionsRow, error) {
	rows, err := q.db.QueryContext(ctx, getSolutions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetSolutionsRow
	for rows.Next() {
		var i GetSolutionsRow
		if err := rows.Scan(&i.Taskid, &i.Cid); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTasksByLowestCost = `-- name: GetTasksByLowestCost :many
SELECT taskid, txhash, cumulativegas, status, claimtime FROM tasks
WHERE status = 3 AND claimtime < ?
ORDER BY cumulativeGas ASC 
LIMIT ?
`

type GetTasksByLowestCostParams struct {
	Claimtime int64
	Limit     int64
}

func (q *Queries) GetTasksByLowestCost(ctx context.Context, arg GetTasksByLowestCostParams) ([]Task, error) {
	rows, err := q.db.QueryContext(ctx, getTasksByLowestCost, arg.Claimtime, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Task
	for rows.Next() {
		var i Task
		if err := rows.Scan(
			&i.Taskid,
			&i.Txhash,
			&i.Cumulativegas,
			&i.Status,
			&i.Claimtime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTotalTasksGas = `-- name: GetTotalTasksGas :one
SELECT count(*), sum(cumulativeGas) FROM tasks
`

type GetTotalTasksGasRow struct {
	Count int64
	Sum   sql.NullFloat64
}

func (q *Queries) GetTotalTasksGas(ctx context.Context) (GetTotalTasksGasRow, error) {
	row := q.db.QueryRowContext(ctx, getTotalTasksGas)
	var i GetTotalTasksGasRow
	err := row.Scan(&i.Count, &i.Sum)
	return i, err
}

const popTask = `-- name: PopTask :one
UPDATE tasks
SET status = 1
WHERE taskid = (SELECT taskid
FROM tasks
WHERE status = 0
LIMIT 1)
RETURNING taskid, txhash
`

type PopTaskRow struct {
	Taskid task.TaskId
	Txhash common.Hash
}

func (q *Queries) PopTask(ctx context.Context) (PopTaskRow, error) {
	row := q.db.QueryRowContext(ctx, popTask)
	var i PopTaskRow
	err := row.Scan(&i.Taskid, &i.Txhash)
	return i, err
}

const popTaskRandom = `-- name: PopTaskRandom :one
UPDATE tasks
SET status = 1
WHERE taskid = (SELECT taskid
FROM tasks
WHERE status = 0
ORDER BY RANDOM()
LIMIT 1)
RETURNING taskid, txhash
`

type PopTaskRandomRow struct {
	Taskid task.TaskId
	Txhash common.Hash
}

func (q *Queries) PopTaskRandom(ctx context.Context) (PopTaskRandomRow, error) {
	row := q.db.QueryRowContext(ctx, popTaskRandom)
	var i PopTaskRandomRow
	err := row.Scan(&i.Taskid, &i.Txhash)
	return i, err
}

const recoverStaleTasks = `-- name: RecoverStaleTasks :exec
UPDATE tasks
SET status = 0
WHERE status = 1
AND NOT EXISTS (
    SELECT 1
    FROM solutions
    WHERE solutions.taskid = tasks.taskid
)
`

func (q *Queries) RecoverStaleTasks(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, recoverStaleTasks)
	return err
}

const requeueTaskIfNoCommitmentOrSolution = `-- name: RequeueTaskIfNoCommitmentOrSolution :execrows
UPDATE tasks
SET status = 0 -- Set back to pending
WHERE taskid = ? -- For the specific task that failed
  AND status = 1 -- Only reset if it was in the 'processing' state (set by PopTask)
  AND NOT EXISTS (
      SELECT 1
      FROM commitments c
      WHERE c.taskid = tasks.taskid
  )
  AND NOT EXISTS (
      SELECT 1
      FROM solutions s
      WHERE s.taskid = tasks.taskid
  )
`

func (q *Queries) RequeueTaskIfNoCommitmentOrSolution(ctx context.Context, taskid task.TaskId) (int64, error) {
	result, err := q.db.ExecContext(ctx, requeueTaskIfNoCommitmentOrSolution, taskid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const setTaskQueuedStatus = `-- name: SetTaskQueuedStatus :execrows
UPDATE tasks SET status = 1 WHERE taskid = ? and status = 0
`

func (q *Queries) SetTaskQueuedStatus(ctx context.Context, taskid task.TaskId) (int64, error) {
	result, err := q.db.ExecContext(ctx, setTaskQueuedStatus, taskid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const totalCommitments = `-- name: TotalCommitments :one
SELECT 
count(commitments.taskid)
FROM commitments
JOIN tasks ON commitments.taskid = tasks.taskid 
WHERE tasks.status = 1
`

func (q *Queries) TotalCommitments(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, totalCommitments)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const totalPendingTasks = `-- name: TotalPendingTasks :one
SELECT 
count(taskid)
FROM tasks 
WHERE status = 0
`

func (q *Queries) TotalPendingTasks(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, totalPendingTasks)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const totalSolutionsAndClaims = `-- name: TotalSolutionsAndClaims :one
SELECT 
    count(CASE WHEN tasks.status = 2 AND solutions.taskid IS NOT NULL THEN 1 END) AS total_solutions,
    count(CASE WHEN tasks.status = 3 AND claimtime > 0 THEN 1 END) AS total_claims
FROM tasks 
LEFT JOIN solutions ON solutions.taskid = tasks.taskid 
WHERE tasks.status IN (2, 3)
`

type TotalSolutionsAndClaimsRow struct {
	TotalSolutions int64
	TotalClaims    int64
}

func (q *Queries) TotalSolutionsAndClaims(ctx context.Context) (TotalSolutionsAndClaimsRow, error) {
	row := q.db.QueryRowContext(ctx, totalSolutionsAndClaims)
	var i TotalSolutionsAndClaimsRow
	err := row.Scan(&i.TotalSolutions, &i.TotalClaims)
	return i, err
}

const updateTaskGas = `-- name: UpdateTaskGas :exec
UPDATE tasks
SET cumulativeGas = cumulativeGas + ?
WHERE taskid = ?
`

type UpdateTaskGasParams struct {
	Cumulativegas float64
	Taskid        task.TaskId
}

func (q *Queries) UpdateTaskGas(ctx context.Context, arg UpdateTaskGasParams) error {
	_, err := q.db.ExecContext(ctx, updateTaskGas, arg.Cumulativegas, arg.Taskid)
	return err
}

const updateTaskSolution = `-- name: UpdateTaskSolution :exec
UPDATE tasks
SET status = 3, claimtime = ?, cumulativeGas = cumulativeGas + ?
WHERE taskid = ?
`

type UpdateTaskSolutionParams struct {
	Claimtime     int64
	Cumulativegas float64
	Taskid        task.TaskId
}

func (q *Queries) UpdateTaskSolution(ctx context.Context, arg UpdateTaskSolutionParams) error {
	_, err := q.db.ExecContext(ctx, updateTaskSolution, arg.Claimtime, arg.Cumulativegas, arg.Taskid)
	return err
}

const updateTaskStatus = `-- name: UpdateTaskStatus :execrows
UPDATE tasks SET status = ? WHERE taskid = ?
`

type UpdateTaskStatusParams struct {
	Status int64
	Taskid task.TaskId
}

func (q *Queries) UpdateTaskStatus(ctx context.Context, arg UpdateTaskStatusParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateTaskStatus, arg.Status, arg.Taskid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const updateTaskStatusAndGas = `-- name: UpdateTaskStatusAndGas :execrows
UPDATE tasks
SET cumulativeGas = cumulativeGas + ?, status = ?
WHERE taskid = ?
`

type UpdateTaskStatusAndGasParams struct {
	Cumulativegas float64
	Status        int64
	Taskid        task.TaskId
}

func (q *Queries) UpdateTaskStatusAndGas(ctx context.Context, arg UpdateTaskStatusAndGasParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateTaskStatusAndGas, arg.Cumulativegas, arg.Status, arg.Taskid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const upsertTaskToClaimable = `-- name: UpsertTaskToClaimable :exec
INSERT INTO tasks (taskid, txhash, status, claimtime)
VALUES (?, ?, 3, ?)
ON CONFLICT(taskid) DO UPDATE SET
    status = 3,
    claimtime = excluded.claimtime
`

type UpsertTaskToClaimableParams struct {
	Taskid    task.TaskId
	Txhash    common.Hash
	Claimtime int64
}

func (q *Queries) UpsertTaskToClaimable(ctx context.Context, arg UpsertTaskToClaimableParams) error {
	_, err := q.db.ExecContext(ctx, upsertTaskToClaimable, arg.Taskid, arg.Txhash, arg.Claimtime)
	return err
}
