package domain

import (
	"crypto/ed25519"
	"time"

	"github.com/google/uuid"
)

type ServerIdentity struct {
	IPAddress string
	PublicKey ed25519.PublicKey
}

type InviteCredential struct {
	UserID    uuid.UUID
	Token     []byte
	NotBefore time.Time
	NotAfter  time.Time
}

type InviteTicket struct {
	Server     ServerIdentity
	Credential InviteCredential
}

type Session struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Token  []byte
}
