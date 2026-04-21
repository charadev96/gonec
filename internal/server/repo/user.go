package repo

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	server "github.com/charadev96/gonec/internal/server/domain"
	shared "github.com/charadev96/gonec/internal/shared/domain"
	"github.com/charadev96/gonec/internal/shared/infra"
)

type BunUserRepository struct {
	db *bun.DB
}

func NewBunUserRepository(ctx context.Context, db *bun.DB) (*BunUserRepository, error) {
	r := &BunUserRepository{
		db: db,
	}
	tx := infra.ExtractTx(ctx, r.db)
	_, err := tx.NewCreateTable().
		Model((*user)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return r, err
	}
	return r, nil
}

func (r *BunUserRepository) Create(ctx context.Context) (uuid.UUID, error) {
	tx := infra.ExtractTx(ctx, r.db)
	id := uuid.New()
	u := &user{ID: id}
	_, err := tx.NewInsert().
		Model(u).
		Exec(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *BunUserRepository) GetByID(ctx context.Context, id uuid.UUID) (server.User, error) {
	tx := infra.ExtractTx(ctx, r.db)
	u := &user{}
	err := tx.NewSelect().
		Model(u).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = shared.ErrNotExist
		}
		return server.User{}, err
	}
	return userFromDB(*u), nil
}

func (r *BunUserRepository) GetByName(ctx context.Context, name string) (server.User, error) {
	tx := infra.ExtractTx(ctx, r.db)
	u := &user{}
	err := tx.NewSelect().
		Model(u).
		Where("name = ?", name).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = shared.ErrNotExist
		}
		return server.User{}, err
	}
	return userFromDB(*u), nil
}

func (r *BunUserRepository) List(ctx context.Context, q server.UserListQuery) (server.UserList, error) {
	var users []server.User
	if q.Limit < 1 {
		q.Limit = 50
	}
	query := r.db.NewSelect().
		Model(&users).
		Limit(q.Limit + 1).
		Order("id ASC")
	if q.Cursor != uuid.Nil {
		query = query.Where("id > ?", q.Cursor)
	}
	if err := query.Scan(ctx); err != nil {
		return server.UserList{}, err
	}

	var next uuid.UUID
	if len(users) > q.Limit {
		users = users[:q.Limit]
		next = users[len(users)-1].ID
	}
	return server.UserList{
		Users:  users,
		Cursor: next,
	}, nil
}

func (r *BunUserRepository) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	tx := infra.ExtractTx(ctx, r.db)
	u := &user{ID: id, Name: name}
	_, err := tx.NewUpdate().
		Model(u).
		Column("name").
		WherePK().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *BunUserRepository) UpdatePublicKey(ctx context.Context, id uuid.UUID, pk ed25519.PublicKey) error {
	tx := infra.ExtractTx(ctx, r.db)
	u := &user{ID: id, PublicKey: pk}
	_, err := tx.NewUpdate().
		Model(u).
		Column("public_key").
		WherePK().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *BunUserRepository) UpdateState(ctx context.Context, id uuid.UUID, s server.UserState) error {
	tx := infra.ExtractTx(ctx, r.db)
	u := &user{ID: id, State: s}
	_, err := tx.NewUpdate().
		Model(u).
		Column("state").
		WherePK().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *BunUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx := infra.ExtractTx(ctx, r.db)
	u := &user{ID: id}
	_, err := tx.NewDelete().
		Model(u).
		WherePK().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

type user struct {
	bun.BaseModel `bun:"table:users"`

	ID        uuid.UUID         `bun:",pk"`
	Name      string            `bun:",unique,nullzero"`
	PublicKey ed25519.PublicKey `bun:",unique,nullzero"`
	State     server.UserState  `bun:",notnull"`
}

func userFromDB(u user) server.User {
	return server.User{
		ID:        u.ID,
		Name:      u.Name,
		PublicKey: u.PublicKey,
		State:     u.State,
	}
}

func userToDB(usr server.User) *user {
	return &user{
		ID:        usr.ID,
		Name:      usr.Name,
		PublicKey: usr.PublicKey,
		State:     usr.State,
	}
}
