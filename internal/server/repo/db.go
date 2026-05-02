package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"

	server "github.com/charadev96/gonec/internal/server/domain"
	"github.com/charadev96/gonec/internal/shared/infra"
)

const permDB = 0644

func ProvideBunDB(ctx context.Context, path string) (server.DB, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		f, err := os.OpenFile(path, os.O_CREATE, 0644)
		if err != nil {
			return server.DB{}, err
		}
		f.Close()
	}

	sqldb, err := sql.Open(
		sqliteshim.ShimName,
		fmt.Sprintf("file:%s?cache=shared", path),
	)
	if err != nil {
		return server.DB{}, fmt.Errorf("open database: %w", err)
	}

	go func() {
		<-ctx.Done()
		sqldb.Close()
	}()

	db := bun.NewDB(sqldb, sqlitedialect.New())

	invites, err := NewBunInviteCredentialRepository(ctx, db)
	if err != nil {
		return server.DB{}, fmt.Errorf("init invites: %w", err)
	}

	nonces, err := NewBunLoginNonceRepository(ctx, db)
	if err != nil {
		return server.DB{}, fmt.Errorf("init nonces: %w", err)
	}

	sessions, err := NewBunSessionRepository(ctx, db)
	if err != nil {
		return server.DB{}, fmt.Errorf("init sessions: %w", err)
	}

	users, err := NewBunUserRepository(ctx, db)
	if err != nil {
		return server.DB{}, fmt.Errorf("init users: %w", err)
	}

	txr := infra.NewBunTransactionRunner(db)

	return server.DB{
		Invites:  invites,
		Nonces:   nonces,
		Sessions: sessions,
		Users:    users,
		TxRunner: txr,
	}, nil
}
