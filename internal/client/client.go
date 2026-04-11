package client

import (
	"context"
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	userpb "github.com/charadev96/gonec/gen/user"
	user "github.com/charadev96/gonec/internal/client/handler/user"
	"github.com/charadev96/gonec/internal/client/service"
)

type Client struct {
	Addr   string
	Logger *zerolog.Logger

	AuthService *service.AuthService
}

func (c *Client) ServeUser(ctx context.Context) error {
	ln, err := net.Listen("tcp", c.Addr)
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}
	c.Logger.Info().
		Str("address", c.Addr).
		Msg("started client")

	inst := grpc.NewServer()
	userpb.RegisterAuthServiceServer(inst, &user.AuthServiceHandler{
		Service: c.AuthService,
	})

	reflection.Register(inst)

	go func() {
		<-ctx.Done()
		c.Logger.Info().Msg("shutting down")
		inst.GracefulStop()
	}()

	return inst.Serve(ln)
}
