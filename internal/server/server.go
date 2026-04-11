package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	adminpb "github.com/charadev96/gonec/gen/admin"
	gatewaypb "github.com/charadev96/gonec/gen/gateway"
	admin "github.com/charadev96/gonec/internal/server/handler/admin"
	gateway "github.com/charadev96/gonec/internal/server/handler/gateway"
	"github.com/charadev96/gonec/internal/server/service"
	"github.com/charadev96/gonec/internal/shared/log"
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
		return fmt.Errorf("init server: %w", err)
	}
	s.Admin.Logger.Info().
		Str("address", s.Admin.Addr).
		Msg("started server")

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	inst := grpc.NewServer(
		grpc.UnaryInterceptor(
			logging.UnaryServerInterceptor(log.NewInterceptor(*s.Admin.Logger), opts...),
		),
	)
	adminpb.RegisterUserServiceServer(inst, &admin.UserServiceHandler{
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
		NextProtos:   []string{"h2"},
	}
	ln, err := tls.Listen("tcp", s.Messaging.Addr, config)
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}
	defer ln.Close()
	s.Messaging.Logger.Info().
		Str("address", s.Messaging.Addr).
		Msg("started server")

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	inst := grpc.NewServer(
		grpc.UnaryInterceptor(
			logging.UnaryServerInterceptor(log.NewInterceptor(*s.Admin.Logger), opts...),
		),
	)
	gatewaypb.RegisterAuthServiceServer(inst, &gateway.AuthServiceHandler{
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
