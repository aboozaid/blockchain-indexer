package database

import (
	"context"
	"database/sql"
	"errors"
)

type Transact interface {
	Begin(context.Context) (Transact, error)

	Transaction(context.Context, func(context.Context) error) error

	Rollback() error

	Commit() error

	Context() context.Context
}

// used to work with both sql.Tx and sql.DB
type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type DatabaseService struct {
	db        DB
	rootDB    *sql.DB
	TxOptions *sql.TxOptions
	ctx       context.Context
}

type txKey struct{}

func NewDatabaseService(db *sql.DB, opt *sql.TxOptions) DatabaseService {
	return DatabaseService{db, db, opt, context.Background()}
}

func (s DatabaseService) Begin(ctx context.Context) (Transact, error) {
	tx, err := s.rootDB.BeginTx(ctx, s.TxOptions)
	if err != nil {
		return nil, err
	}

	return DatabaseService{
		ctx:       context.WithValue(ctx, txKey{}, tx),
		TxOptions: s.TxOptions,
		db:        tx,
		rootDB:    s.rootDB,
	}, nil
}

func (s DatabaseService) Commit() error {
	tx, ok := s.db.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	return tx.Commit()
}

func (s DatabaseService) Rollback() error {
	tx, ok := s.db.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}

	return tx.Rollback()
}

func (s DatabaseService) Transaction(ctx context.Context, f func(context.Context) error) error {
	tx, err := s.rootDB.BeginTx(ctx, s.TxOptions)
	if err != nil {
		return err
	}
	c := context.WithValue(ctx, txKey{}, tx)
	err = f(c)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (s DatabaseService) Context() context.Context {
	return s.ctx
}

// helper method to be used inside repositories
func GetDB(ctx context.Context, fallback *sql.DB) DB {
	tx := ctx.Value(txKey{})
	if tx == nil {
		return fallback
	}
	return tx.(*sql.Tx)
}