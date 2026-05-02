package cert

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
)

const (
	permKey  = 0600
	permCert = 0644
)

type ExpiredError struct {
	Now       time.Time
	NotBefore time.Time
	NotAfter  time.Time
}

func (e *ExpiredError) Error() string {
	if e.Now.Before(e.NotBefore) {
		return fmt.Sprintf(
			"not yet valid, current time %s is before %s",
			e.Now.Format(time.RFC3339),
			e.NotBefore.Format(time.RFC3339),
		)
	}
	return fmt.Sprintf(
		"expired, current time %s is after %s",
		e.Now.Format(time.RFC3339),
		e.NotAfter.Format(time.RFC3339),
	)
}

type Config struct {
	PathKey   string
	PathCert  string
	IPAddress string
	ValidFor  time.Duration
	Logger    *zerolog.Logger
}

type Provider struct {
	cert    tls.Certificate
	cfg     Config
	expires time.Time
}

func NewProvider(cfg Config) (*Provider, error) {
	l := zerolog.Nop()

	p := &Provider{cfg: cfg}
	if p.cfg.Logger == nil {
		p.cfg.Logger = &l
	}
	err := p.Ensure()
	if err != nil {
		return nil, err
	}

	return p, err
}

func (p *Provider) GetCert(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return &p.cert, nil
}

func (p *Provider) GetPublicKey() ed25519.PublicKey {
	return p.cert.Leaf.PublicKey.(ed25519.PublicKey)
}

func (p *Provider) Watch(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Until(p.expires)):
			p.cfg.Logger.Info().Msg("renewing certificate")
			if err := p.Ensure(); err != nil {
				return err
			}
		}
	}
}

func (p *Provider) Ensure() error {
	var (
		keyIsNew bool
		key      ed25519.PrivateKey
		pemKey   []byte
		pemCert  []byte
	)

	if _, err := os.Stat(p.cfg.PathKey); err == nil {
		key, pemKey, err = loadKey(p.cfg.PathKey)
		if err != nil {
			return fmt.Errorf("load key: %w", err)
		}
		p.cfg.Logger.Info().
			Str("path", p.cfg.PathKey).
			Msg("loaded private key")
	} else if errors.Is(err, os.ErrNotExist) {
		p.cfg.Logger.Warn().Msg("private key missing")
		key, pemKey, err = generateKey(p.cfg.PathKey)
		if err != nil {
			return fmt.Errorf("generate key: %w", err)
		}
		keyIsNew = true
		p.cfg.Logger.Info().
			Str("path", p.cfg.PathKey).
			Msg("generated private key")
	} else {
		return err
	}

	if _, err := os.Stat(p.cfg.PathCert); err == nil {
		pemCert, err = os.ReadFile(p.cfg.PathCert)
		if err != nil {
			return err
		}
		p.cfg.Logger.Info().
			Str("path", p.cfg.PathCert).
			Msg("loaded certificate")
	} else if errors.Is(err, os.ErrNotExist) || keyIsNew {
		p.cfg.Logger.Warn().Msg("certificate missing or key is new")
		temp := p.template()
		pemCert, err = generateCert(p.cfg.PathCert, key, temp)
		if err != nil {
			return fmt.Errorf("generate certificate: %w", err)
		}
		p.cfg.Logger.Info().
			Str("path", p.cfg.PathCert).
			Msg("generated certficate")
	} else {
		return err
	}

	pair, err := tls.X509KeyPair(pemCert, pemKey)
	if err != nil {
		return fmt.Errorf("load key pair: %w", err)
	}

	cert := tls.Certificate(pair)
	var expired *ExpiredError
	if err := p.verify(cert); errors.As(err, &expired) {
		p.cfg.Logger.Warn().Msg("certificate expired, regenerating")
		pemCert, err = generateCert(p.cfg.PathCert, key, p.template())
		if err != nil {
			return fmt.Errorf("regenerate certificate: %w", err)
		}
		p.cfg.Logger.Info().
			Str("path", p.cfg.PathCert).
			Msg("generated certficate")

		pair, err = tls.X509KeyPair(pemCert, pemKey)
		if err != nil {
			return fmt.Errorf("load key pair: %w", err)
		}
		cert = tls.Certificate(pair)
		if err := p.verify(cert); err != nil {
			return fmt.Errorf("verify certificate: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("verify certificate: %w", err)
	}

	p.cert = cert
	p.expires = cert.Leaf.NotAfter

	return nil
}

func (p *Provider) verify(cert tls.Certificate) error {
	var (
		leaf = cert.Leaf
		now  = time.Now()
	)

	if now.Before(leaf.NotBefore) || now.After(leaf.NotAfter) {
		return &ExpiredError{now, leaf.NotBefore, leaf.NotAfter}
	}

	if leaf.PublicKeyAlgorithm != x509.Ed25519 {
		return fmt.Errorf("bad key format, must be ed25519")
	}

	if leaf.BasicConstraintsValid == false || leaf.IsCA == false {
		return fmt.Errorf("not a CA")
	}

	if len(leaf.IPAddresses) < 1 {
		return fmt.Errorf("missing ip addresses")
	}

	if !leaf.IPAddresses[0].Equal(net.ParseIP(p.cfg.IPAddress)) {
		return fmt.Errorf("ip address mismatch")
	}

	return nil
}

func (p *Provider) template() x509.Certificate {
	return x509.Certificate{
		Subject: pkix.Name{},

		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(p.cfg.ValidFor),
		IsCA:      true,

		BasicConstraintsValid: true,

		IPAddresses: []net.IP{net.ParseIP(p.cfg.IPAddress)},
	}
}

func generateKey(path string) (ed25519.PrivateKey, []byte, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, permKey)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}
	raw, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal key: %w", err)
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: raw,
	}

	var buf bytes.Buffer
	if err := pem.Encode(&buf, block); err != nil {
		return nil, nil, fmt.Errorf("encode key: %w", err)
	}
	_, err = file.Write(buf.Bytes())
	if err != nil {
		return nil, nil, err
	}

	return key, buf.Bytes(), nil
}

func loadKey(path string) (ed25519.PrivateKey, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	block, _ := pem.Decode(raw)
	pkcs8, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse key: %w", err)
	}
	key, ok := pkcs8.(ed25519.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("bad key format, must be ed25519")
	}

	return key, raw, nil
}

func generateCert(path string, key ed25519.PrivateKey, temp x509.Certificate) ([]byte, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, permCert)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cert, err := x509.CreateCertificate(
		rand.Reader,
		&temp, &temp,
		key.Public(), key,
	)
	if err != nil {
		return nil, fmt.Errorf("create certificate: %w", err)
	}

	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}

	var buf bytes.Buffer
	if err := pem.Encode(&buf, block); err != nil {
		return nil, fmt.Errorf("encode certificate: %w", err)
	}
	_, err = file.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
