package user

import (
	"context"
	"database/sql"
	"fmt"
	"todo-app-api/entity"
	"todo-app-api/pkg/exception"

	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

type UserRepository interface {
	BeginTx(ctx context.Context) (tx *sql.Tx, err error)
	RollbackTx(ctx context.Context, tx *sql.Tx) (err error)
	CommitTx(ctx context.Context, tx *sql.Tx) (err error)
	// SaveUser(ctx context.Context, user UserRequest, tx *sql.Tx) (id int64, err error)
	// UpdateById(ctx context.Context, id int64, user UserRequest, tx *sql.Tx) (err error)
	FindManyUser(ctx context.Context) (bunchOfUsers []entity.User, err error)
	// FindOneUserByUUID(ctx context.Context, uuid string) (user entity.User, err error)
}

type sqlCommand interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type userRepository struct {
	logger      *logrus.Logger
	dbReadOnly  *sql.DB
	dbReadWrite *sql.DB
	tableName   string
}

// NewUserRepository is a constructor
func NewUserRepository(logger *logrus.Logger, dbReadOnly *sql.DB, dbReadWrite *sql.DB, tableName string) UserRepository {
	return &userRepository{
		logger:      logger,
		dbReadOnly:  dbReadOnly,
		dbReadWrite: dbReadWrite,
		tableName:   tableName,
	}
}

// BeginTx returns sql trx for global scope.
func (r *userRepository) BeginTx(ctx context.Context) (tx *sql.Tx, err error) {
	return r.dbReadWrite.BeginTx(ctx, nil)
}

// CommitTx will commit the transaction that has began.
func (r *userRepository) CommitTx(ctx context.Context, tx *sql.Tx) (err error) {
	return tx.Commit()
}

// RollbackTx will rollback the transaction to achieve the consistency.
func (r *userRepository) RollbackTx(ctx context.Context, tx *sql.Tx) (err error) {
	return tx.Rollback()
}

func (r *userRepository) FindManyUser(ctx context.Context) (bunchOfUsers []entity.User, err error) {
	var cmd sqlCommand = r.dbReadOnly
	q := fmt.Sprintf(`SELECT u.uuid, u.name, u.email, u.created_at FROM %s u`, r.tableName)
	bunchOfUsers, err = r.query(ctx, cmd, q)
	if err != nil {
		err = wrapError(err)
		return
	}
	return
}

func (r *userRepository) query(ctx context.Context, cmd sqlCommand, query string, args ...interface{}) (bunchOfUsers []entity.User, err error) {
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
		var user entity.User

		err = rows.Scan(&user.UUID, &user.Name, &user.Email, &user.CreatedAt)

		if err != nil {
			r.logger.WithContext(ctx).Error(query, err)
			return
		}

		bunchOfUsers = append(bunchOfUsers, user)
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
