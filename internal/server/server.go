package server

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/rs/zerolog"
)

type Server struct {
	Addr        string
	Certificate tls.Certificate
	Logger      *zerolog.Logger
}

func (s *Server) ListenAndServeTLS() error {
	config := &tls.Config{
		Certificates: []tls.Certificate{s.Certificate},
	}
	ln, err := tls.Listen("tcp", s.Addr, config)
	if err != nil {
		return fmt.Errorf("failed to init server: %w", err)
	}
	defer ln.Close()

	s.Logger.Info().
		Str("address", s.Addr).
		Msg("started server")

	for {
		conn, err := ln.Accept()
		if err != nil {
			s.Logger.Error().
				Err(err).
				Msg("failed to establish connection")
			continue
		}
		s.Logger.Info().
			Str("address", conn.RemoteAddr().String()).
			Msg("established connection")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Fprintf(conn, "0\n")
}
