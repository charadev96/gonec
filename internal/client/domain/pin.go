package domain

import (
	"crypto/ed25519"

	"github.com/google/uuid"

	shared "github.com/charadev96/gonec/internal/shared/domain"
)

type UserPrivateIdentity struct {
	ID         uuid.UUID
	PrivateKey ed25519.PrivateKey
}

type ConnPin struct {
	ID     string
	User   UserPrivateIdentity
	Server shared.ServerIdentity
}

type ConnPinRepository interface {
	Get(id string) (ConnPin, error)
	Set(id string, pin ConnPin) error
	Delete(id string) error
}
