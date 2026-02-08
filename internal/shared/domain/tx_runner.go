package domain

import (
	"context"
)

type TransactionRunner interface {
	Exec(ctx context.Context, fn func(ctx context.Context) error) error
}
