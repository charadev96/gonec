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
	shared "github.com/charadev96/gonec/internal/shared/domain"
	"github.com/charadev96/gonec/internal/shared/infra"
)

type BunLoginNonceRepository struct {
	db *bun.DB
}

func NewBunLoginNonceRepository(ctx context.Context, db *bun.DB) (*BunLoginNonceRepository, error) {
	r := &BunLoginNonceRepository{
		db: db,
	}
	tx := infra.ExtractTx(ctx, r.db)
	_, err := tx.NewCreateTable().
		Model((*loginNonce)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return r, fmt.Errorf("failed to create repository: %w", err)
	}
	return r, nil
}

func (r *BunLoginNonceRepository) Save(ctx context.Context, nonce server.LoginNonce) error {
	tx := infra.ExtractTx(ctx, r.db)
	n := &loginNonce{}
	copier.Copy(n, &nonce)
	_, err := tx.NewInsert().
		Model(n).
		Replace().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save nonce: %w", err)
	}
	return nil
}

func (r *BunLoginNonceRepository) Consume(ctx context.Context, id uuid.UUID) (server.LoginNonce, error) {
	tx := infra.ExtractTx(ctx, r.db)
	n := &loginNonce{}
	nonce := server.LoginNonce{}
	err := tx.NewSelect().
		Model(n).
		Where("user_id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = shared.ErrNotExist
		}
		return nonce, fmt.Errorf("failed to get nonce: %w", err)
	}
	_, err = tx.NewDelete().
		Model(n).
		WherePK().
		Exec(ctx)
	if err != nil {
		return nonce, fmt.Errorf("failed to delete nonce: %w", err)
	}
	copier.Copy(&nonce, n)
	return nonce, nil
}

type loginNonce struct {
	UserID    uuid.UUID `bun:",pk"`
	Value     []byte    `bun:",unique,nullzero"`
	CreatedAt time.Time
}
