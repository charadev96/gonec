package service

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	gatewaypb "github.com/charadev96/gonec/gen/gateway"
	client "github.com/charadev96/gonec/internal/client/domain"
	shared "github.com/charadev96/gonec/internal/shared/domain"
	pb "github.com/charadev96/gonec/internal/shared/pb"
)

type AuthServiceStatus int

const (
	AuthDisconnected AuthServiceStatus = iota
	AuthConnected
	AuthLoggedIn
)

type AuthService struct {
	pins client.ConnPinRepository

	pin     client.ConnPin
	conn    *grpc.ClientConn
	session *shared.Session
	status  AuthServiceStatus

	rand io.Reader
}

type AuthServiceOption func(*AuthService)

func AuthWithRand(r io.Reader) AuthServiceOption {
	return func(s *AuthService) {
		s.rand = r
	}
}

func NewAuthService(p client.ConnPinRepository, opts ...AuthServiceOption) *AuthService {
	s := &AuthService{
		pins: p,
		rand: rand.Reader,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *AuthService) Register(ctx context.Context, id string, t shared.InviteTicket) error {
	if s.status == AuthLoggedIn {
		return client.ErrLoggedIn
	}

	if _, err := s.pins.Get(id); err == nil {
		return fmt.Errorf("get pin: %w", shared.ErrExist)
	}

	pub, prv, err := ed25519.GenerateKey(s.rand)
	pin := client.ConnPin{
		ID:     id,
		Server: t.Server,
		User: client.UserPrivateIdentity{
			ID:         t.Credential.UserID,
			PrivateKey: prv,
		},
	}
	err = s.pins.Set(id, pin)
	if err != nil {
		return fmt.Errorf("set pin: %w", err)
	}

	err = s.connect(ctx, id)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}
	defer s.disconnect()

	cl, err := BindClient(s, gatewaypb.NewAuthServiceClient)
	if err != nil {
		return err
	}

	_, err = cl.Register(ctx, &gatewaypb.RegisterRequest{
		UserId:    t.Credential.UserID.String(),
		Token:     t.Credential.Token,
		PublicKey: pub,
	})
	if err != nil {
		return fmt.Errorf("request register: %w", err)
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, id string) error {
	if s.status == AuthLoggedIn {
		return client.ErrLoggedIn
	}

	fail := true
	err := s.connect(ctx, id)
	defer func() {
		if fail {
			s.disconnect()
		}
	}()

	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}

	cl, err := BindClient(s, gatewaypb.NewAuthServiceClient)
	if err != nil {
		return err
	}

	pin, err := s.Pin()
	if err != nil {
		return fmt.Errorf("get active pin: %w", err)
	}

	repInitiate, err := cl.InitiateLogin(ctx, &gatewaypb.InitiateLoginRequest{
		UserId: pin.User.ID.String(),
	})
	if err != nil {
		return fmt.Errorf("request login request: %w", err)
	}

	sig := ed25519.Sign(pin.User.PrivateKey, repInitiate.Nonce)
	repComplete, err := cl.CompleteLogin(ctx, &gatewaypb.CompleteLoginRequest{
		UserId:    pin.User.ID.String(),
		Signature: sig,
	})
	if err != nil {
		return fmt.Errorf("request complete login: %w", err)
	}

	fail = false
	*s.session, err = pb.SessionFromPB(repComplete.Auth)
	if err != nil {
		return fmt.Errorf("parse session: %w", err)
	}
	s.status = AuthLoggedIn

	return nil
}

func (s *AuthService) Logout(ctx context.Context) error {
	cl, err := BindClient(s, gatewaypb.NewAuthServiceClient)
	if err != nil {
		return err
	}

	session, err := s.Session()
	if err != nil {
		return fmt.Errorf("get active session: %w", err)
	}

	defer s.disconnect()
	_, err = cl.Logout(ctx, &gatewaypb.LogoutRequest{
		Auth: pb.SessionToPB(session),
	})
	if err != nil {
		return fmt.Errorf("request logout: %w", err)
	}

	s.session = nil

	return nil
}

func BindClient[T any](s *AuthService, c func(grpc.ClientConnInterface) T) (T, error) {
	cl := c(s.conn)
	if s.status == AuthDisconnected {
		return cl, client.ErrNoConn
	}
	return cl, nil
}

func (s *AuthService) Session() (shared.Session, error) {
	if s.status == AuthDisconnected {
		return shared.Session{}, client.ErrNoConn
	}
	if s.status != AuthLoggedIn {
		return shared.Session{}, client.ErrNoLoggedIn
	}
	if s.session == nil {
		panic("invariant violation: session status is active but struct is nil")
	}
	return *s.session, nil
}

func (s *AuthService) Pin() (client.ConnPin, error) {
	if s.status == AuthDisconnected {
		return client.ConnPin{}, client.ErrNoConn
	}
	return s.pin, nil
}

func (s *AuthService) Status() AuthServiceStatus {
	return s.status
}

func (s *AuthService) connect(ctx context.Context, id string) error {
	if s.status != AuthDisconnected {
		return client.ErrConn
	}

	pin, err := s.pins.Get(id)
	if err != nil {
		return fmt.Errorf("get pin %q: %w", id, err)
	}

	config := &tls.Config{
		VerifyPeerCertificate: s.verifyServerCertificate,
		InsecureSkipVerify:    true,
		NextProtos:            []string{"h2"},
	}
	creds := credentials.NewTLS(config)
	conn, err := grpc.NewClient(
		pin.Server.IPAddress,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return fmt.Errorf("establish connection: %w", err)
	}

	s.conn = conn
	s.pin = pin
	s.status = AuthConnected

	return nil
}

func (s *AuthService) disconnect() {
	s.status = AuthDisconnected
	s.conn.Close()
}

func (s *AuthService) verifyServerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	pin, err := s.pins.Get(s.pin.ID)
	if err != nil {
		return fmt.Errorf("update pin %q: %w", s.pin.ID, err)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", pin.Server.IPAddress)
	if err != nil {
		return fmt.Errorf("resolve server tcp address: %w", err)
	}

	cert, err := x509.ParseCertificate(rawCerts[0])
	if err != nil {
		return fmt.Errorf("parse certificate: %w", err)
	}

	_, ok := cert.PublicKey.(ed25519.PublicKey)
	if !ok {
		return fmt.Errorf("bad certificate key format, must be ed25519")
	}

	if err = cert.VerifyHostname(fmt.Sprintf("[%s]", tcpAddr.IP.String())); err != nil {
		return fmt.Errorf("verify certificate hostname: %w", err)
	}

	now := time.Now()
	if now.Before(cert.NotBefore) {
		return fmt.Errorf(
			"certificate expired, current time %s is before %s",
			now.Format(time.RFC3339),
			cert.NotBefore.Format(time.RFC3339),
		)
	}
	if now.After(cert.NotAfter) {
		return fmt.Errorf(
			"certificate expired, current time %s is after %s",
			now.Format(time.RFC3339),
			cert.NotAfter.Format(time.RFC3339),
		)
	}

	if ok = ed25519.Verify(pin.Server.PublicKey, cert.RawTBSCertificate, cert.Signature); !ok {
		// TODO: User verification of mismatched signature
		return fmt.Errorf("certificate signature mismatch: user verification unimplemented")
	}

	return nil
}
