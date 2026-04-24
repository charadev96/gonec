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

type GatewayConfig struct {
	Addr        string
	Certificate tls.Certificate
	Logger      *zerolog.Logger
}

type Server struct {
	admin   AdminConfig
	gateway GatewayConfig

	user *service.UserService
}

func New(adm AdminConfig, gtw GatewayConfig, user *service.UserService) *Server {
	l := zerolog.Nop()
	s := &Server{
		admin:   adm,
		gateway: gtw,

		user: user,
	}
	if s.admin.Logger == nil {
		s.admin.Logger = &l
	}
	if s.gateway.Logger == nil {
		s.gateway.Logger = &l
	}
	return s
}

func (s *Server) ServeAdmin(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.admin.Addr)
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}
	s.admin.Logger.Info().
		Str("address", s.admin.Addr).
		Msg("started server")

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	inst := grpc.NewServer(
		grpc.UnaryInterceptor(
			logging.UnaryServerInterceptor(log.NewInterceptor(*s.admin.Logger), opts...),
		),
	)
	adminpb.RegisterUserServiceServer(inst,
		admin.NewUserHandler(s.user),
	)

	reflection.Register(inst)

	go func() {
		<-ctx.Done()
		s.admin.Logger.Info().Msg("shutting down")
		inst.GracefulStop()
	}()

	return inst.Serve(ln)
}

func (s *Server) ServeMessaging(ctx context.Context) error {
	config := &tls.Config{
		Certificates: []tls.Certificate{s.gateway.Certificate},
		NextProtos:   []string{"h2"},
	}
	ln, err := tls.Listen("tcp", s.gateway.Addr, config)
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}
	defer ln.Close()
	s.gateway.Logger.Info().
		Str("address", s.gateway.Addr).
		Msg("started server")

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	inst := grpc.NewServer(
		grpc.UnaryInterceptor(
			logging.UnaryServerInterceptor(log.NewInterceptor(*s.admin.Logger), opts...),
		),
	)
	gatewaypb.RegisterAuthServiceServer(inst,
		gateway.NewAuthHandler(s.user),
	)

	reflection.Register(inst)

	go func() {
		<-ctx.Done()
		s.gateway.Logger.Info().Msg("shutting down")
		inst.GracefulStop()
	}()

	return inst.Serve(ln)
}
