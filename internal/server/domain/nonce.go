package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserLoginNonce struct {
	UserID    uuid.UUID
	Nonce     []byte
	CreatedAt time.Time
}

type UserNonceRepository interface {
	Save(ctx context.Context, chal UserLoginNonce) error
	Consume(ctx context.Context, id uuid.UUID) (UserLoginNonce, error)
}
