package domain

import (
	"crypto/ed25519"

	server "github.com/charadev96/gonec/internal/server/domain"
)

type ServerPublicIdentity struct {
	IPAddress string
	PublicKey ed25519.PublicKey
}

type UserInviteManifest struct {
	Server ServerPublicIdentity
	Invite server.UserInvite
}
