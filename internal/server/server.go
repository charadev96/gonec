package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	genadm "github.com/charadev96/gonec/api/admin/gen"
	genmsg "github.com/charadev96/gonec/api/messaging/gen"
	"github.com/charadev96/gonec/internal/server/handler/admin"
	"github.com/charadev96/gonec/internal/server/handler/messaging"
	"github.com/charadev96/gonec/internal/server/service"
)

type AdminConfig struct {
	Addr   string
	Logger *zerolog.Logger
}

type MessagingConfig struct {
	Addr        string
	Certificate tls.Certificate
	Logger      *zerolog.Logger
}

type Server struct {
	Admin     AdminConfig
	Messaging MessagingConfig

	UserService *service.UserService
}

func (s *Server) ServeAdmin(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.Admin.Addr)
	if err != nil {
		return fmt.Errorf("failed to init server: %w", err)
	}
	s.Admin.Logger.Info().
		Str("address", s.Admin.Addr).
		Msg("started server")

	inst := grpc.NewServer()
	genadm.RegisterUserServiceServer(inst, &admin.UserServiceHandler{
		Service: s.UserService,
	})

	reflection.Register(inst)

	go func() {
		<-ctx.Done()
		s.Admin.Logger.Info().Msg("shutting down")
		inst.GracefulStop()
	}()

	return inst.Serve(ln)
}

func (s *Server) ServeMessaging(ctx context.Context) error {
	config := &tls.Config{
		Certificates: []tls.Certificate{s.Messaging.Certificate},
	}
	ln, err := tls.Listen("tcp", s.Messaging.Addr, config)
	if err != nil {
		return fmt.Errorf("failed to init server: %w", err)
	}
	defer ln.Close()
	s.Messaging.Logger.Info().
		Str("address", s.Messaging.Addr).
		Msg("started server")

	inst := grpc.NewServer()
	genmsg.RegisterAuthServiceServer(inst, &messaging.AuthServiceHandler{
		Service: s.UserService,
	})

	reflection.Register(inst)

	go func() {
		<-ctx.Done()
		s.Messaging.Logger.Info().Msg("shutting down")
		inst.GracefulStop()
	}()

	return inst.Serve(ln)
}
