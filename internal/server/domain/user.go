package domain

import (
	"context"
	"crypto/ed25519"

	"github.com/google/uuid"
)

type UserState int

const (
	StatePending UserState = iota
	StateRegistered
	StateActive
)

type User struct {
	ID        uuid.UUID
	Name      string
	PublicKey ed25519.PublicKey
	State     UserState
}

type UserRepository interface {
	Create(ctx context.Context) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	GetByName(ctx context.Context, name string) (User, error)
	UpdateName(ctx context.Context, id uuid.UUID, name string) error
	UpdatePublicKey(ctx context.Context, id uuid.UUID, pk ed25519.PublicKey) error
	UpdateState(ctx context.Context, id uuid.UUID, s UserState) error
	Delete(ctx context.Context, id uuid.UUID) error
}
