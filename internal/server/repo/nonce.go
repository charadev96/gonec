package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
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
		return r, err
	}
	return r, nil
}

func (r *BunLoginNonceRepository) Save(ctx context.Context, nonce server.LoginNonce) error {
	tx := infra.ExtractTx(ctx, r.db)
	n := nonceToDB(nonce)
	_, err := tx.NewInsert().
		Model(n).
		Replace().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *BunLoginNonceRepository) Consume(ctx context.Context, id uuid.UUID) (server.LoginNonce, error) {
	tx := infra.ExtractTx(ctx, r.db)
	n := &loginNonce{}
	err := tx.NewSelect().
		Model(n).
		Where("user_id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = shared.ErrNotExist
		}
		return server.LoginNonce{}, err
	}
	_, err = tx.NewDelete().
		Model(n).
		WherePK().
		Exec(ctx)
	if err != nil {
		return server.LoginNonce{}, err
	}
	return nonceFromDB(*n), nil
}

type loginNonce struct {
	UserID    uuid.UUID `bun:",pk"`
	Value     []byte    `bun:",unique,nullzero"`
	CreatedAt time.Time
}

func nonceFromDB(n loginNonce) server.LoginNonce {
	return server.LoginNonce{
		UserID:    n.UserID,
		Value:     n.Value,
		CreatedAt: n.CreatedAt,
	}
}

func nonceToDB(nonce server.LoginNonce) *loginNonce {
	return &loginNonce{
		UserID:    nonce.UserID,
		Value:     nonce.Value,
		CreatedAt: nonce.CreatedAt,
	}
}
