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

type BunSessionRepository struct {
	db *bun.DB
}

func NewBunSessionRepository(ctx context.Context, db *bun.DB) (*BunSessionRepository, error) {
	r := &BunSessionRepository{
		db: db,
	}
	tx := infra.ExtractTx(ctx, r.db)
	_, err := tx.NewCreateTable().
		Model((*session)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return r, fmt.Errorf("failed to create repository: %w", err)
	}
	return r, nil
}

func (r *BunSessionRepository) Save(ctx context.Context, sess server.Session) error {
	tx := infra.ExtractTx(ctx, r.db)
	s := &session{}
	copier.Copy(s, &sess)
	_, err := tx.NewInsert().
		Model(s).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

func (r *BunSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (server.Session, error) {
	tx := infra.ExtractTx(ctx, r.db)
	s := &session{}
	sess := server.Session{}
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

func (r *BunSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx := infra.ExtractTx(ctx, r.db)
	s := &session{ID: id}
	_, err := tx.NewDelete().
		Model(s).
		WherePK().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

type session struct {
	ID        uuid.UUID `bun:",pk"`
	UserID    uuid.UUID
	Token     []byte `bun:",unique,nullzero"`
	CreatedAt time.Time
}
