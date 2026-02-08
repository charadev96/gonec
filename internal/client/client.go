package client

import (
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/rs/zerolog"

	"github.com/charadev96/gonec/internal/client/domain"
)

type Client struct {
	ConnServerID string
	RepoPins     domain.PinRepository
	Logger       *zerolog.Logger

	UserTrustCertificate func(*x509.Certificate) bool
}

func (c *Client) DialServer(id string) error {
	server, err := c.RepoPins.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get server pin '%s': %w", id, err)
	}
	c.ConnServerID = id

	config := &tls.Config{
		VerifyPeerCertificate: c.verifyServerCertificate,
		InsecureSkipVerify:    true,
	}
	conn, err := tls.Dial("tcp", server.Identity.IPAddress, config)
	if err != nil {
		return fmt.Errorf("failed to establish connection: %w", err)
	}
	defer conn.Close()

	c.Logger.Info().
		Str("address", server.Identity.IPAddress).
		Msg("connected to server")

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read response: %w", err)
	}

	c.Logger.Info().
		Str("text", string(buf[:n])).
		Msg("got response from server")

	return nil
}

func (c *Client) verifyServerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	server, err := c.RepoPins.Get(c.ConnServerID)
	if err != nil {
		return fmt.Errorf("failed to get server pin '%s': %w", c.ConnServerID, err)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", server.Identity.IPAddress)
	if err != nil {
		return fmt.Errorf("failed to resolve server tcp address: %w", err)
	}

	cert, err := x509.ParseCertificate(rawCerts[0])
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	key, ok := cert.PublicKey.(ed25519.PublicKey)
	if !ok {
		return fmt.Errorf("incorrect certifiacte public key format (must be ed25519)")
	}

	if err = cert.VerifyHostname(fmt.Sprintf("[%s]", tcpAddr.IP.String())); err != nil {
		return fmt.Errorf("failed to verify certificate hostname: %w", err)
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

	if ok = ed25519.Verify(server.Identity.PublicKey, cert.RawTBSCertificate, cert.Signature); !ok {
		c.Logger.Warn().
			Msg("certificate signature mismatch")
		if c.UserTrustCertificate == nil {
			return fmt.Errorf("failed to verify certificate: signature mismatch")
		}
		c.Logger.Info().
			Msg("awaiting user confirmation")

		if ok = c.UserTrustCertificate(cert); !ok {
			return fmt.Errorf("failed to verify certificate: signature denied by user")
		}

		server.Identity.PublicKey = key
		if err = c.RepoPins.Set(server.ID, server); err != nil {
			return fmt.Errorf("failed to save server registry: %w", err)
		}
		c.Logger.Info().
			Msg("public key updated, attempting to reverify")

		return c.verifyServerCertificate(rawCerts, verifiedChains)
	}

	c.Logger.Info().
		Msg("certificate verified successfully")

	return nil
}
