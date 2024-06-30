package task

import (
	"context"
	"database/sql"
	"fmt"
	"todo-app-api/entity"
	"todo-app-api/pkg/exception"

	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

type TaskRepository interface {
	BeginTx(ctx context.Context) (tx *sql.Tx, err error)
	RollbackTx(ctx context.Context, tx *sql.Tx) (err error)
	CommitTx(ctx context.Context, tx *sql.Tx) (err error)
	Save(ctx context.Context, task TaskRequest, tx *sql.Tx) (id int64, err error)
	UpdateById(ctx context.Context, id int64, task TaskRequest, tx *sql.Tx) (err error)
	FindMany(ctx context.Context) (bunchOfTasks []entity.Task, err error)
	FindOneById(ctx context.Context, id int64) (task entity.Task, err error)
}

type sqlCommand interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type taskRepository struct {
	logger      *logrus.Logger
	dbReadOnly  *sql.DB
	dbReadWrite *sql.DB
	tableName   string
}

// NewTaskRepository is a constructor
func NewTaskRepository(logger *logrus.Logger, dbReadOnly *sql.DB, dbReadWrite *sql.DB, tableName string) TaskRepository {
	return &taskRepository{
		logger:      logger,
		dbReadOnly:  dbReadOnly,
		dbReadWrite: dbReadWrite,
		tableName:   tableName,
	}
}

// BeginTx returns sql trx for global scope.
func (r *taskRepository) BeginTx(ctx context.Context) (tx *sql.Tx, err error) {
	return r.dbReadWrite.BeginTx(ctx, nil)
}

// CommitTx will commit the transaction that has began.
func (r *taskRepository) CommitTx(ctx context.Context, tx *sql.Tx) (err error) {
	return tx.Commit()
}

// RollbackTx will rollback the transaction to achieve the consistency.
func (r *taskRepository) RollbackTx(ctx context.Context, tx *sql.Tx) (err error) {
	return tx.Rollback()
}

func (r *taskRepository) FindMany(ctx context.Context) (bunchOfTasks []entity.Task, err error) {
	var cmd sqlCommand = r.dbReadOnly
	q := fmt.Sprintf(`SELECT t.id, t.name, t.description, t.status, t.attachment, t.created_at, t.updated_at FROM %s t`, r.tableName)
	bunchOfTasks, err = r.query(ctx, cmd, q)
	if err != nil {
		err = wrapError(err)
		return
	}
	return
}

func (r *taskRepository) FindOneById(ctx context.Context, id int64) (task entity.Task, err error) {
	var cmd sqlCommand = r.dbReadOnly
	q := fmt.Sprintf(`SELECT t.id, t.name, t.description, t.status, t.attachment, t.created_at, t.updated_at FROM %s t WHERE t.id = ?`, r.tableName)
	bunchOfTasks, err := r.query(ctx, cmd, q, id)
	if err != nil {
		err = wrapError(err)
		return
	}

	lengthOfTasks := len(bunchOfTasks)
	if lengthOfTasks < 1 {
		err = exception.ErrNotFound
		return
	}

	task = bunchOfTasks[lengthOfTasks-1]
	return
}

func (r *taskRepository) query(ctx context.Context, cmd sqlCommand, query string, args ...interface{}) (bunchOfTasks []entity.Task, err error) {
	var rows *sql.Rows
	if rows, err = cmd.QueryContext(ctx, query, args...); err != nil {
		r.logger.WithContext(ctx).Error(query, err)
		return
	}

	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.WithContext(ctx).Error(query, err)
		}
	}()

	for rows.Next() {
		var task entity.Task
		var updatedAt sql.NullTime
		var attachment sql.NullString

		err = rows.Scan(&task.ID, &task.Name, &task.Description, &task.Status, &attachment, &task.CreatedAt, &updatedAt)

		if err != nil {
			r.logger.WithContext(ctx).Error(query, err)
			return
		}

		if attachment.Valid {
			task.Attachment = &attachment.String
		}

		if updatedAt.Valid {
			task.UpdatedAt = &updatedAt.Time
		}

		bunchOfTasks = append(bunchOfTasks, task)
	}

	return
}

// Save will collect the order
func (r *taskRepository) Save(ctx context.Context, task TaskRequest, tx *sql.Tx) (id int64, err error) {
	var cmd sqlCommand = r.dbReadWrite
	if tx != nil {
		cmd = tx
	}

	command := fmt.Sprintf(`INSERT INTO %s SET name = ?, description = ?, status = ?, created_at = ?`, r.tableName)
	res, err := r.exec(ctx, cmd, command, task.Name, task.Description, task.Status, task.CreatedAt)
	if err != nil {
		err = wrapError(err)
		return
	}

	id, err = res.LastInsertId()
	if err != nil {
		err = wrapError(err)
		return
	}

	return
}

func (r *taskRepository) UpdateById(ctx context.Context, id int64, task TaskRequest, tx *sql.Tx) (err error) {
	var cmd sqlCommand = r.dbReadWrite
	if tx != nil {
		cmd = tx
	}
	command := `UPDATE %s SET	name = ?, description = ?, status = ?, attachment = ?, updated_at = ? WHERE id = ?`

	_, err = r.exec(ctx, cmd, fmt.Sprintf(command, r.tableName), task.Name, task.Description, task.Status, task.Attachment, task.UpdatedAt, id)
	return
}

func (r *taskRepository) exec(ctx context.Context, cmd sqlCommand, command string, args ...interface{}) (result sql.Result, err error) {
	var stmt *sql.Stmt
	if stmt, err = cmd.PrepareContext(ctx, command); err != nil {
		r.logger.WithContext(ctx).Error(command, err)
		return
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			r.logger.WithContext(ctx).Error(command, err)
		}
	}()

	if result, err = stmt.ExecContext(ctx, args...); err != nil {
		r.logger.WithContext(ctx).Error(command, err)
	}

	return
}

func wrapError(e error) (err error) {
	if e == sql.ErrNoRows {
		return exception.ErrNotFound
	}
	if driverErr, ok := e.(*mysql.MySQLError); ok {
		if driverErr.Number == 1062 {
			return exception.ErrConflict
		}
	}
	return exception.ErrInternalServer
}
