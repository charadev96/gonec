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

type BunUserSessionRepository struct {
	db *bun.DB
}

func NewBunUserSessionRepository(ctx context.Context, db *bun.DB) (*BunUserSessionRepository, error) {
	r := &BunUserSessionRepository{
		db: db,
	}
	tx := infra.ExtractTx(ctx, r.db)
	_, err := tx.NewCreateTable().
		Model((*userSession)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return r, fmt.Errorf("failed to create repository: %w", err)
	}
	return r, nil
}

func (r *BunUserSessionRepository) Save(ctx context.Context, sess server.UserSession) error {
	tx := infra.ExtractTx(ctx, r.db)
	s := new(userSession)
	copier.Copy(s, &sess)
	_, err := tx.NewInsert().
		Model(s).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

func (r *BunUserSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (server.UserSession, error) {
	tx := infra.ExtractTx(ctx, r.db)
	s := new(userSession)
	sess := server.UserSession{}
	err := tx.NewSelect().
		Model(s).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = shared.ErrNotExist
		}
		return sess, fmt.Errorf("failed to get sesion: %w", err)
	}
	copier.Copy(&sess, s)
	return sess, nil
}

func (r *BunUserSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx := infra.ExtractTx(ctx, r.db)
	s := &userSession{ID: id}
	_, err := tx.NewDelete().
		Model(s).
		WherePK().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

type userSession struct {
	ID        uuid.UUID `bun:",pk"`
	UserID    uuid.UUID
	Token     []byte `bun:",unique,nullzero"`
	CreatedAt time.Time
}
