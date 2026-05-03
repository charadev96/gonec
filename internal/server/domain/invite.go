package domain

import (
	"context"

	"github.com/google/uuid"

	shared "github.com/charadev96/gonec/internal/shared/domain"
)

type InviteClaimsRepository interface {
	Save(ctx context.Context, c shared.InviteClaims) error
	GetByUserID(ctx context.Context, id uuid.UUID) (shared.InviteClaims, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
