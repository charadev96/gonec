package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserInvite struct {
	UserID    uuid.UUID
	Token     []byte
	NotBefore time.Time
	NotAfter  time.Time
}

type UserInviteRepository interface {
	Save(ctx context.Context, inv UserInvite) error
	GetByUserID(ctx context.Context, id uuid.UUID) (UserInvite, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
