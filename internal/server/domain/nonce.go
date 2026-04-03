package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LoginNonce struct {
	UserID    uuid.UUID
	Value     []byte
	CreatedAt time.Time
}

type LoginNonceRepository interface {
	Save(ctx context.Context, nonce LoginNonce) error
	Consume(ctx context.Context, id uuid.UUID) (LoginNonce, error)
}
