package domain

import (
	"crypto/ed25519"
	"github.com/google/uuid"

	shared "github.com/charadev96/gonec/internal/shared/domain"
)

type UserIdentity struct {
	ID         uuid.UUID
	Name       string
	PrivateKey ed25519.PrivateKey
}

type ConnPin struct {
	ID     string
	User   UserIdentity
	Server shared.ServerPublicIdentity
}

type ConnPinRepository interface {
	Get(id string) (ConnPin, error)
	Set(id string, pin ConnPin) error
	Delete(id string) error
}
