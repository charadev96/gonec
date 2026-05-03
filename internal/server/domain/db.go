package domain

import (
	shared "github.com/charadev96/gonec/internal/shared/domain"
)

type DB struct {
	Claims   InviteClaimsRepository
	Nonces   LoginNonceRepository
	Sessions SessionRepository
	Users    UserRepository
	TxRunner shared.TransactionRunner
}
