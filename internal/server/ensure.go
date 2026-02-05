package server

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

const (
	permKey  = 0600
	permCert = 0644
)

func EnsureX509KeyPair(
	certPath, keyPath string,
	template x509.Certificate,
	logger *zerolog.Logger,
) ([]byte, []byte, error) {
	var (
		keyIsNew bool
		key      ed25519.PrivateKey
		keyPEM   []byte
		certPEM  []byte
	)
	if logger == nil {
		logger = zerolog.DefaultContextLogger
	}

	if _, err := os.Stat(keyPath); errors.Is(err, os.ErrNotExist) {
		logger.Warn().
			Str("file", keyPath).
			Msg("private key does not exist")
		key, keyPEM, err = generateKeyFile(keyPath)
		if err != nil {
			return nil, nil, err
		}
		keyIsNew = true
		logger.Info().
			Str("file", keyPath).
			Msg("created new private key")
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve private key: %w", err)
	} else {
		key, keyPEM, err = loadKeyFile(keyPath)
		if err != nil {
			return nil, nil, err
		}
	}

	logger.Info().
		Str("file", keyPath).
		Msg("parsed private key")

	if _, err := os.Stat(certPath); errors.Is(err, os.ErrNotExist) || keyIsNew {
		logger.Warn().
			Str("file", certPath).
			Msg("certificate does not exist or invalid")
		certPEM, err = generateCertificateFile(certPath, key, template)
		if err != nil {
			return nil, nil, err
		}
		logger.Info().
			Str("file", certPath).
			Msg("created new certficate")
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve certificate: %w", err)
	} else {
		certPEM, err = os.ReadFile(certPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read certificate: %w", err)
		}
	}

	logger.Info().
		Str("file", certPath).
		Msg("parsed certificate")

	return certPEM, keyPEM, nil
}

func generateKeyFile(keyPath string) (ed25519.PrivateKey, []byte, error) {
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE, permKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create private key file: %w", err)
	}
	defer keyFile.Close()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	keyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	}

	var keyBuf bytes.Buffer
	if err := pem.Encode(&keyBuf, keyBlock); err != nil {
		return nil, nil, fmt.Errorf("failed to encode private key: %w", err)
	}
	_, err = keyFile.Write(keyBuf.Bytes())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write key PEM file to disk: %w", err)
	}

	return key, keyBuf.Bytes(), nil
}

func loadKeyFile(keyPath string) (ed25519.PrivateKey, []byte, error) {
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	keyAny, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	key, ok := keyAny.(ed25519.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("incorrect private key format (must be ed25519)")
	}

	return key, keyPEM, nil
}

func generateCertificateFile(
	certPath string, key ed25519.PrivateKey,
	template x509.Certificate,
) ([]byte, error) {
	certFile, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE, permCert)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certFile.Close()

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, key.Public(), key)
	if err != nil {
		return nil, fmt.Errorf("failed generating certificate: %w", err)
	}

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}

	var certBuf bytes.Buffer
	if err := pem.Encode(&certBuf, certBlock); err != nil {
		return nil, fmt.Errorf("failed encoding certificate: %w", err)
	}
	_, err = certFile.Write(certBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write certificate PEM file to disk: %w", err)
	}

	return certBuf.Bytes(), nil
}
