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

	userpb "github.com/charadev96/gonec/gen/user"
	user "github.com/charadev96/gonec/internal/client/handler/user"
	"github.com/charadev96/gonec/internal/client/service"
)

type Config struct {
	Addr   string
	Logger *zerolog.Logger
}

type Client struct {
	cfg Config

	auth *service.AuthService
	chat *service.ChatService
}

func New(cfg Config, auth *service.AuthService, chat *service.ChatService) *Client {
	l := zerolog.Nop()
	s := &Client{
		cfg: cfg,

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
		return fmt.Errorf("init server: %w", err)
	}
	c.cfg.Logger.Info().
		Str("address", c.cfg.Addr).
		Msg("started client")

	inst := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    30 * time.Second,
			Timeout: 10 * time.Second,
		}),
	)
	userpb.RegisterAuthServiceServer(inst, user.NewAuthHandler(c.auth))
	userpb.RegisterChatServiceServer(inst, user.NewChatHandler(ctx, c.chat))

	reflection.Register(inst)

	go func() {
		<-ctx.Done()
		c.cfg.Logger.Info().Msg("shutting down")
		inst.GracefulStop()
	}()

	return inst.Serve(ln)
}
