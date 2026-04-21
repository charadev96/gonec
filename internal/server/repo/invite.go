package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	shared "github.com/charadev96/gonec/internal/shared/domain"
	"github.com/charadev96/gonec/internal/shared/infra"
)

type BunInviteCredentialRepository struct {
	db *bun.DB
}

func NewBunInviteCredentialRepository(ctx context.Context, db *bun.DB) (*BunInviteCredentialRepository, error) {
	r := &BunInviteCredentialRepository{
		db: db,
	}
	tx := infra.ExtractTx(ctx, r.db)
	_, err := tx.NewCreateTable().
		Model((*inviteCredential)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return r, err
	}
	return r, nil
}

func (r *BunInviteCredentialRepository) Save(ctx context.Context, cred shared.InviteCredential) error {
	tx := infra.ExtractTx(ctx, r.db)
	c := inviteToDB(cred)
	_, err := tx.NewInsert().
		Model(c).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *BunInviteCredentialRepository) GetByUserID(ctx context.Context, id uuid.UUID) (shared.InviteCredential, error) {
	tx := infra.ExtractTx(ctx, r.db)
	c := &inviteCredential{}
	err := tx.NewSelect().
		Model(c).
		Where("user_id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = shared.ErrNotExist
		}
		return shared.InviteCredential{}, err
	}
	return inviteFromDB(*c), nil
}

func (r *BunInviteCredentialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx := infra.ExtractTx(ctx, r.db)
	c := &inviteCredential{UserID: id}
	_, err := tx.NewDelete().
		Model(c).
		WherePK().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

type inviteCredential struct {
	UserID    uuid.UUID `bun:",pk"`
	Token     []byte    `bun:",unique,nullzero"`
	NotBefore time.Time
	NotAfter  time.Time
}

func inviteFromDB(c inviteCredential) shared.InviteCredential {
	return shared.InviteCredential{
		UserID:    c.UserID,
		Token:     c.Token,
		NotBefore: c.NotBefore,
		NotAfter:  c.NotAfter,
	}
}

func inviteToDB(cred shared.InviteCredential) *inviteCredential {
	return &inviteCredential{
		UserID:    cred.UserID,
		Token:     cred.Token,
		NotBefore: cred.NotBefore,
		NotAfter:  cred.NotAfter,
	}
}
