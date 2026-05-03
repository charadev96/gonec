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

type BunInviteClaims struct {
	db *bun.DB
}

func NewBunInviteClaims(ctx context.Context, db *bun.DB) (*BunInviteClaims, error) {
	r := &BunInviteClaims{
		db: db,
	}
	tx := infra.ExtractTx(ctx, r.db)
	_, err := tx.NewCreateTable().
		Model((*inviteClaims)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return r, err
	}
	return r, nil
}

func (r *BunInviteClaims) Save(ctx context.Context, cl shared.InviteClaims) error {
	tx := infra.ExtractTx(ctx, r.db)
	c := inviteToDB(cl)
	_, err := tx.NewInsert().
		Model(c).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *BunInviteClaims) GetByUserID(ctx context.Context, id uuid.UUID) (shared.InviteClaims, error) {
	tx := infra.ExtractTx(ctx, r.db)
	c := &inviteClaims{}
	err := tx.NewSelect().
		Model(c).
		Where("user_id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = shared.ErrNotExist
		}
		return shared.InviteClaims{}, err
	}
	return inviteFromDB(*c), nil
}

func (r *BunInviteClaims) Delete(ctx context.Context, id uuid.UUID) error {
	tx := infra.ExtractTx(ctx, r.db)
	c := &inviteClaims{UserID: id}
	_, err := tx.NewDelete().
		Model(c).
		WherePK().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

type inviteClaims struct {
	UserID    uuid.UUID `bun:",pk"`
	Token     []byte    `bun:",unique,nullzero"`
	NotBefore time.Time
	NotAfter  time.Time
}

func inviteFromDB(c inviteClaims) shared.InviteClaims {
	return shared.InviteClaims{
		UserID:    c.UserID,
		Token:     c.Token,
		NotBefore: c.NotBefore,
		NotAfter:  c.NotAfter,
	}
}

func inviteToDB(cl shared.InviteClaims) *inviteClaims {
	return &inviteClaims{
		UserID:    cl.UserID,
		Token:     cl.Token,
		NotBefore: cl.NotBefore,
		NotAfter:  cl.NotAfter,
	}
}
