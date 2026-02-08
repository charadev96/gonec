package service

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	server "github.com/charadev96/gonec/internal/server/domain"
	shared "github.com/charadev96/gonec/internal/shared/domain"
)

type UserService struct {
	Users    server.UserRepository
	Invites  server.UserInviteRepository
	TXRunner shared.TransactionRunner
	Rand     io.Reader
}

type CreateInviteOptions struct {
	NotBefore time.Time
	NotAfter  time.Time
}

func (s *UserService) CreateInvite(ctx context.Context, id uuid.UUID, opts CreateInviteOptions) (server.UserInvite, error) {
	inv := server.UserInvite{}
	if _, err := s.Users.GetByID(ctx, id); err != nil && errors.Is(err, shared.ErrNotExist) {
		return inv, err
	}

	rnd := s.Rand
	if rnd == nil {
		rnd = rand.Reader
	}
	tok := make([]byte, 32)
	_, err := rnd.Read(tok)
	if err != nil {
		return inv, fmt.Errorf("failed to generate invite token: %w", err)
	}

	unix := time.Unix(0, 0).UTC()
	if opts.NotBefore.IsZero() || opts.NotBefore.Equal(unix) {
		opts.NotBefore = time.Now()
	}
	if opts.NotAfter.IsZero() || opts.NotAfter.Equal(unix) {
		opts.NotAfter = time.Now().AddDate(0, 0, 1)
	}
	if opts.NotAfter.Before(opts.NotBefore) {
		return inv, fmt.Errorf("invalid invite time period, NotAfter must be after NotBefore")
	}

	inv = server.UserInvite{
		UserID:    id,
		Token:     tok,
		NotBefore: opts.NotBefore,
		NotAfter:  opts.NotAfter,
	}

	if err := s.Invites.Save(ctx, inv); err != nil {
		return inv, err
	}

	return inv, nil
}

func (s *UserService) RegisterUser(ctx context.Context, id uuid.UUID, tok []byte, pk ed25519.PublicKey) error {
	if _, err := s.Users.GetByID(ctx, id); err != nil && errors.Is(err, shared.ErrNotExist) {
		return err
	}

	inv, err := s.Invites.GetByUserID(ctx, id)
	if err != nil {
		return err
	}

	if subtle.ConstantTimeCompare(inv.Token, tok) == 0 {
		return fmt.Errorf("token mismatch")
	}

	now := time.Now()
	if now.Before(inv.NotBefore) {
		return fmt.Errorf(
			"invitation not yet valid, current time %s is before %s",
			now.Format(time.RFC3339),
			inv.NotBefore.Format(time.RFC3339),
		)
	}
	if now.After(inv.NotAfter) {
		return fmt.Errorf(
			"invitation expired, current time %s is after %s",
			now.Format(time.RFC3339),
			inv.NotAfter.Format(time.RFC3339),
		)
	}

	return s.TXRunner.Exec(ctx, func(ctx context.Context) error {
		if err := s.Users.UpdateState(ctx, id, server.StateRegistered); err != nil {
			return err
		}
		if err := s.Users.UpdatePublicKey(ctx, id, pk); err != nil {
			return err
		}
		if err := s.Invites.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.TXRunner.Exec(ctx, func(ctx context.Context) error {
		if err := s.Users.Delete(ctx, id); err != nil {
			return err
		}
		if err := s.Invites.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	})
}
