package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	shared "github.com/charadev96/gonec/internal/shared/domain"
)

type Session struct {
	shared.Session
	CreatedAt time.Time
}

type SessionRepository interface {
	Save(ctx context.Context, sess Session) error
	GetByID(ctx context.Context, id uuid.UUID) (Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
