package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	adminpb "github.com/charadev96/gonec/gen/admin"
	gatewaypb "github.com/charadev96/gonec/gen/gateway"
	"github.com/charadev96/gonec/internal/server/cert"
	"github.com/charadev96/gonec/internal/server/domain"
	admin "github.com/charadev96/gonec/internal/server/handler/admin"
	gateway "github.com/charadev96/gonec/internal/server/handler/gateway"
	"github.com/charadev96/gonec/internal/server/service"
	shared "github.com/charadev96/gonec/internal/shared/domain"
	"github.com/charadev96/gonec/internal/shared/log"
)

type AdminConfig struct {
	Addr   string
	Logger *zerolog.Logger
}

type GatewayConfig struct {
	Addr         string
	CertProvider *cert.Provider
	Logger       *zerolog.Logger
}

type Server struct {
	admin   AdminConfig
	gateway GatewayConfig
	db      domain.DB

	user *service.User
	chat *service.Chat
}

func New(adm AdminConfig, gtw GatewayConfig, db domain.DB) *Server {
	l := zerolog.Nop()

	var (
		user = service.NewUser(
			shared.ServerIdentity{
				IPAddress: gtw.Addr,
				PublicKey: gtw.CertProvider.GetPublicKey(),
			},
			db.Users,
			db.Claims,
			db.Nonces,
			db.Sessions,
			db.TxRunner,
		)
		chat = service.NewChat(
			db.Users,
			user,
		)
	)

	s := &Server{
		admin:   adm,
		gateway: gtw,
		db:      db,

		user: user,
		chat: chat,
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
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			logging.UnaryServerInterceptor(log.NewInterceptor(*s.admin.Logger), opts...),
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    30 * time.Second,
			Timeout: 10 * time.Second,
		}),
	)
	adminpb.RegisterUserServiceServer(srv, admin.NewUserHandler(s.user))

	reflection.Register(srv)

	go func() {
		<-ctx.Done()
		s.admin.Logger.Info().Msg("shutting down")
		srv.GracefulStop()
	}()

	return srv.Serve(ln)
}

func (s *Server) ServeGateway(ctx context.Context) error {
	config := &tls.Config{
		GetCertificate: s.gateway.CertProvider.GetCert,
		NextProtos:     []string{"h2"},
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
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			logging.UnaryServerInterceptor(log.NewInterceptor(*s.admin.Logger), opts...),
		),
	)
	gatewaypb.RegisterAuthServiceServer(srv, gateway.NewAuthHandler(s.user))
	gatewaypb.RegisterChatServiceServer(srv, gateway.NewChatHandler(ctx, s.chat))

	reflection.Register(srv)

	go func() {
		<-ctx.Done()
		s.gateway.Logger.Info().Msg("shutting down")
		srv.GracefulStop()
	}()

	return srv.Serve(ln)
}
