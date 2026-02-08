package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/charadev96/gonec/api/admin/gen"
	"github.com/charadev96/gonec/internal/server/handler/admin"
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
	gen.RegisterUserServiceServer(inst, &admin.UserServiceHandler{
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

	go func() {
		<-ctx.Done()
		s.Messaging.Logger.Info().Msg("shutting down")
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		s.Messaging.Logger.Info().
			Str("address", conn.RemoteAddr().String()).
			Msg("established connection")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Fprintf(conn, "0\n")
}
