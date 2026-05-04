package client

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	userpb "github.com/charadev96/gonec/gen/gonec/user/v1"
	"github.com/charadev96/gonec/internal/client/domain"
	user "github.com/charadev96/gonec/internal/client/handler/user"
	"github.com/charadev96/gonec/internal/client/service"
)

type Config struct {
	Addr   string
	Logger *zerolog.Logger
}

type Client struct {
	cfg Config
	db  domain.DB

	auth *service.Auth
	chat *service.Chat
}

func New(cfg Config, db domain.DB) *Client {
	l := zerolog.Nop()

	var (
		auth = service.NewAuth(db.Pins)
		chat = service.NewChat(auth)
	)

	s := &Client{
		cfg: cfg,
		db:  db,

		auth: auth,
		chat: chat,
	}
	if s.cfg.Logger == nil {
		s.cfg.Logger = &l
	}

	return s
}

func (c *Client) ServeUser(ctx context.Context) error {
	ln, err := net.Listen("tcp", c.cfg.Addr)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	c.cfg.Logger.Info().
		Str("address", c.cfg.Addr).
		Msg("started server")

	srv := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    30 * time.Second,
			Timeout: 10 * time.Second,
		}),
	)
	userpb.RegisterAuthServiceServer(srv, user.NewAuthHandler(c.auth))
	userpb.RegisterChatServiceServer(srv, user.NewChatHandler(ctx, c.chat))

	reflection.Register(srv)

	go func() {
		<-ctx.Done()
		c.cfg.Logger.Info().Msg("shutting down")
		srv.GracefulStop()
	}()

	return srv.Serve(ln)
}
