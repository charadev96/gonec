package domain

import (
	"context"

	"github.com/google/uuid"

	shared "github.com/charadev96/gonec/internal/shared/domain"
)

type InviteCredentialRepository interface {
	Save(ctx context.Context, tok shared.InviteCredential) error
	GetByUserID(ctx context.Context, id uuid.UUID) (shared.InviteCredential, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
