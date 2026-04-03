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
	"github.com/jinzhu/copier"

	server "github.com/charadev96/gonec/internal/server/domain"
	shared "github.com/charadev96/gonec/internal/shared/domain"
)

type UserService struct {
	Server   shared.ServerIdentity
	Users    server.UserRepository
	Invites  server.InviteCredentialRepository
	Nonces   server.LoginNonceRepository
	Sessions server.SessionRepository
	TXRunner shared.TransactionRunner
	Rand     io.Reader
}

type CreateInviteOptions struct {
	NotBefore time.Time
	NotAfter  time.Time
}

func (s *UserService) CreateInvite(ctx context.Context, id uuid.UUID, opts CreateInviteOptions) (shared.InviteCredential, error) {
	inv := shared.InviteCredential{}
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

	inv = shared.InviteCredential{
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

func (s *UserService) CreateLoginNonce(ctx context.Context, id uuid.UUID) ([]byte, error) {
	nonce := server.LoginNonce{}
	user, err := s.Users.GetByID(ctx, id)
	if err != nil && errors.Is(err, shared.ErrNotExist) {
		return nil, err
	}
	if user.State == server.StatePending {
		return nil, fmt.Errorf("user not registered")
	}

	rnd := s.Rand
	if rnd == nil {
		rnd = rand.Reader
	}
	tok := make([]byte, 32)
	_, err = rnd.Read(tok)
	if err != nil {
		return nil, fmt.Errorf("failed to generate login nonce: %w", err)
	}

	nonce = server.LoginNonce{
		UserID:    id,
		Value:     tok,
		CreatedAt: time.Now(),
	}

	if err := s.Nonces.Save(ctx, nonce); err != nil {
		return nil, err
	}

	return tok, nil
}

func (s *UserService) ExportInvite(ctx context.Context, id uuid.UUID) (shared.InviteTicket, error) {
	mnf := shared.InviteTicket{}
	cred, err := s.Invites.GetByUserID(ctx, id)
	if err != nil && errors.Is(err, shared.ErrNotExist) {
		return mnf, err
	}

	mnf = shared.InviteTicket{
		Server:     s.Server,
		Credential: cred,
	}

	return mnf, nil
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

func (s *UserService) VerifySession(ctx context.Context, sess shared.Session) error {
	session, err := s.Sessions.GetByID(ctx, sess.ID)
	if err != nil {
		return err
	}

	if session.UserID != sess.UserID {
		return fmt.Errorf("user id mismatch")
	}
	if subtle.ConstantTimeCompare(session.Token, sess.Token) == 0 {
		return fmt.Errorf("token mismatch")
	}
	if ok := time.Now().After(session.CreatedAt.Add(time.Hour * 12)); !ok {
		return fmt.Errorf("session expired, please login again")
	}

	return nil
}

func (s *UserService) LoginUser(ctx context.Context, id uuid.UUID, sig []byte) (shared.Session, error) {
	sess := shared.Session{}
	user, err := s.Users.GetByID(ctx, id)
	if err != nil && errors.Is(err, shared.ErrNotExist) {
		return sess, err
	}
	if user.State == server.StatePending {
		return sess, fmt.Errorf("user not registered")
	}

	nonce, err := s.Nonces.Consume(ctx, id)
	if err != nil && errors.Is(err, shared.ErrNotExist) {
		return sess, err
	}

	if expired := time.Now().After(nonce.CreatedAt.Add(time.Minute)); expired {
		return sess, fmt.Errorf("challenge nonce expired, please retry")
	}
	if ok := ed25519.Verify(user.PublicKey, nonce.Value, sig); !ok {
		return sess, fmt.Errorf("signature mismatch")
	}

	rnd := s.Rand
	if rnd == nil {
		rnd = rand.Reader
	}
	tok := make([]byte, 32)
	_, err = rnd.Read(tok)
	if err != nil {
		return sess, fmt.Errorf("failed to generate session token: %w", err)
	}

	sess = shared.Session{
		ID:     uuid.New(),
		UserID: id,
		Token:  tok,
	}

	session := server.Session{}
	copier.Copy(&session, &sess)
	session.CreatedAt = time.Now()
	if err = s.Sessions.Save(ctx, session); err != nil {
		return sess, err
	}

	return sess, nil
}

func (s *UserService) LogoutUser(ctx context.Context, sess shared.Session) error {
	if err := s.VerifySession(ctx, sess); err != nil {
		return err
	}
	if err := s.Sessions.Delete(ctx, sess.ID); err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.TXRunner.Exec(ctx, func(ctx context.Context) error {
		if err := s.Users.Delete(ctx, id); err != nil {
			return err
		}
		if err := s.Invites.Delete(ctx, id); err != nil {
			return err
		}
		if err := s.Sessions.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	})
}
