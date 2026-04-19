package repo

import (
	"context"
	"database/sql"
	"otcountingbackend/internal/service"
)

type txKey struct{}

type Transactor struct{ DB *sql.DB }

type sqlTx struct{ tx *sql.Tx }

func (t *Transactor) BeginTx(ctx context.Context) (context.Context, service.Tx, error) {
	tx, err := t.DB.BeginTx(ctx, nil)
	if err != nil {
		return ctx, nil, err
	}
	return context.WithValue(ctx, txKey{}, tx), &sqlTx{tx: tx}, nil
}

func (s *sqlTx) Commit(context.Context) error   { return s.tx.Commit() }
func (s *sqlTx) Rollback(context.Context) error { return s.tx.Rollback() }

func querier(ctx context.Context, db *sql.DB) execQuerier {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return db
}

type execQuerier interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
