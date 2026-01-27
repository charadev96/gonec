package domain

import (
	"crypto/ed25519"
)

type ServerPin struct {
	ID        string
	IPAddress string
	PublicKey ed25519.PublicKey
}

type PinRepository interface {
	Get(id string) (ServerPin, error)
	Set(id string, pin ServerPin) error
	Delete(id string) error
}
