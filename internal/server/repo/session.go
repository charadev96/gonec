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

type BunSession struct {
	db *bun.DB
}

func NewBunSession(ctx context.Context, db *bun.DB) (*BunSession, error) {
	r := &BunSession{
		db: db,
	}
	tx := infra.ExtractTx(ctx, r.db)
	_, err := tx.NewCreateTable().
		Model((*session)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return r, err
	}
	return r, nil
}

func (r *BunSession) Save(ctx context.Context, ses server.Session) error {
	tx := infra.ExtractTx(ctx, r.db)
	s := sessionToDB(ses)
	_, err := tx.NewInsert().
		Model(s).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *BunSession) GetByID(ctx context.Context, id uuid.UUID) (server.Session, error) {
	tx := infra.ExtractTx(ctx, r.db)
	s := &session{}
	err := tx.NewSelect().
		Model(s).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = shared.ErrNotExist
		}
		return server.Session{}, err
	}
	return sessionFromDB(*s), nil
}

func (r *BunSession) Delete(ctx context.Context, id uuid.UUID) error {
	tx := infra.ExtractTx(ctx, r.db)
	s := &session{ID: id}
	_, err := tx.NewDelete().
		Model(s).
		WherePK().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

type session struct {
	ID        uuid.UUID `bun:",pk"`
	UserID    uuid.UUID
	Token     []byte `bun:",unique,nullzero"`
	CreatedAt time.Time
}

func sessionFromDB(s session) server.Session {
	return server.Session{
		Session: shared.Session{
			ID:     s.ID,
			UserID: s.UserID,
			Token:  s.Token,
		},
		CreatedAt: s.CreatedAt,
	}
}

func sessionToDB(ses server.Session) *session {
	return &session{
		ID:        ses.ID,
		UserID:    ses.UserID,
		Token:     ses.Token,
		CreatedAt: ses.CreatedAt,
	}
}
