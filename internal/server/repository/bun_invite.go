package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"database/sql"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/uptrace/bun"

	server "github.com/charadev96/gonec/internal/server/domain"
	domain "github.com/charadev96/gonec/internal/shared/domain"
	"github.com/charadev96/gonec/internal/shared/infra"
)

type BunUserInviteRepository struct {
	db *bun.DB
}

func NewBunUserInviteRepository(ctx context.Context, db *bun.DB) (*BunUserInviteRepository, error) {
	r := &BunUserInviteRepository{
		db: db,
	}
	tx := infra.ExtractTx(ctx, r.db)
	_, err := tx.NewCreateTable().
		Model((*userInvite)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return r, fmt.Errorf("failed to create repository: %w", err)
	}
	return r, nil
}

func (r *BunUserInviteRepository) Save(ctx context.Context, inv server.UserInvite) error {
	tx := infra.ExtractTx(ctx, r.db)
	i := new(userInvite)
	copier.Copy(i, &inv)
	_, err := tx.NewInsert().
		Model(i).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save invite: %w", err)
	}
	return nil
}

func (r *BunUserInviteRepository) GetByUserID(ctx context.Context, id uuid.UUID) (server.UserInvite, error) {
	tx := infra.ExtractTx(ctx, r.db)
	i := new(userInvite)
	inv := server.UserInvite{}
	err := tx.NewSelect().
		Model(i).
		Where("user_id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = domain.ErrNotExist
		}
		return inv, fmt.Errorf("failed to get invite: %w", err)
	}
	copier.Copy(&inv, i)
	return inv, nil
}

func (r *BunUserInviteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx := infra.ExtractTx(ctx, r.db)
	i := &userInvite{UserID: id}
	_, err := tx.NewDelete().
		Model(i).
		WherePK().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete invite: %w", err)
	}
	return nil
}

type userInvite struct {
	UserID    uuid.UUID `bun:",pk"`
	Token     []byte    `bun:",unique,nullzero"`
	NotBefore time.Time
	NotAfter  time.Time
}
