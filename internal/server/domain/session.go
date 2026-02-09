package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     []byte
	CreatedAt time.Time
}

type UserSessionRepository interface {
	Save(ctx context.Context, sess UserSession) error
	GetByID(ctx context.Context, id uuid.UUID) (UserSession, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
