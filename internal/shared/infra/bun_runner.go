package infra

import (
	"context"

	"github.com/uptrace/bun"
)

type contextKey struct{}

var txContextKey = &contextKey{}

func InjectTx(ctx context.Context, db bun.IDB) context.Context {
	return context.WithValue(ctx, txContextKey, db)
}

func ExtractTx(ctx context.Context, fallback bun.IDB) bun.IDB {
	if db, ok := ctx.Value(txContextKey).(bun.IDB); ok {
		return db
	}
	return fallback
}

type BunTransactionRunner struct {
	db *bun.DB
}

func NewBunTransactionRunner(db *bun.DB) *BunTransactionRunner {
	return &BunTransactionRunner{db: db}
}

func (r *BunTransactionRunner) Exec(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txContextKey).(bun.IDB); ok {
		return fn(ctx)
	}
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(InjectTx(ctx, tx))
	})
}
