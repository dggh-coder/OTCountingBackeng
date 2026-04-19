package service

import (
	"context"
	"otcountingbackend/internal/engine"
)

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Transactor interface {
	BeginTx(ctx context.Context) (context.Context, Tx, error)
}

type SessionRepo interface {
	GetByID(ctx context.Context, sessionID int64) (*Session, error)
	GetByDatePeriod(ctx context.Context, date, period string) (*Session, error)
	Create(ctx context.Context, s Session) error
}

type EntryRepo interface {
	ReplaceBySession(ctx context.Context, sessionID int64, entries []SessionEntry) error
	ListBySession(ctx context.Context, sessionID int64) ([]SessionEntry, error)
}

type ResultRepo interface {
	UpsertMany(ctx context.Context, records []ResultRecord) error
	ListBySession(ctx context.Context, sessionID int64) ([]ResultRecord, error)
}

type RenderedRepo interface {
	Upsert(ctx context.Context, row RenderedFragment) error
	Get(ctx context.Context, sessionID int64, employeeID, fragmentType string, formatVersion int) (*RenderedFragment, error)
}

type Engine interface {
	Calculate(input engine.Input) (engine.Output, error)
}
