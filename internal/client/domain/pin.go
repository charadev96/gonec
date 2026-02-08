package domain

import (
	shared "github.com/charadev96/gonec/internal/shared/domain"
)

type ServerPin struct {
	ID       string
	Identity shared.ServerPublicIdentity
}

type PinRepository interface {
	Get(id string) (ServerPin, error)
	Set(id string, pin ServerPin) error
	Delete(id string) error
}
