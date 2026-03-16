package repository

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jinzhu/copier"

	client "github.com/charadev96/gonec/internal/client/domain"
	shared "github.com/charadev96/gonec/internal/shared/domain"
)

const (
	permRepository = 0644
)

type TOMLConnPinRepository struct {
	FilePath string

	data       schema
	modifiedAt time.Time
}

func (r *TOMLConnPinRepository) Get(id string) (client.ConnPin, error) {
	modified, err := r.fileModified()
	pin := client.ConnPin{}
	if err != nil {
		return pin, err
	}
	if modified {
		if err := r.load(); err != nil {
			return pin, err
		}
	}
	p, ok := r.data.Conns[id]
	if !ok {
		return pin, shared.ErrNotExist
	}
	copier.Copy(&pin, p)
	pin.ID = id
	pin.Server.PublicKey = p.Server.PublicKey.PublicKey
	pin.User.PrivateKey = p.User.PrivateKey.PrivateKey
	return pin, nil
}

func (r *TOMLConnPinRepository) Set(id string, pin client.ConnPin) error {
	modified, err := r.fileModified()
	if err != nil {
		return err
	}
	if modified {
		if err := r.load(); err != nil {
			return err
		}
	}
	if _, ok := r.data.Conns[id]; !ok {
		r.data.Conns[id] = &connPin{}
	}
	copier.Copy(r.data.Conns[id], pin)
	r.data.Conns[id].Server.PublicKey.PublicKey = pin.Server.PublicKey
	r.data.Conns[id].User.PrivateKey.PrivateKey = pin.User.PrivateKey
	if err := r.save(); err != nil {
		return err
	}
	return nil
}

func (r *TOMLConnPinRepository) Delete(id string) error {
	_, ok := r.data.Conns[id]
	if !ok {
		return shared.ErrNotExist
	}
	delete(r.data.Conns, id)
	if err := r.save(); err != nil {
		return err
	}
	return nil
}

func decodeBase64(src []byte, size int) ([]byte, error) {
	s := base64.StdEncoding.DecodedLen(len(src))
	if s == 0 {
		return []byte{}, nil
	}
	if s < size {
		return nil, fmt.Errorf("base64: bad length %d, requires at least %d", s, size)
	}
	dst := make([]byte, size)
	_, err := base64.StdEncoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

func encodeBase64(src []byte, size int) ([]byte, error) {
	s := len(src)
	if s == 0 {
		return []byte{}, nil
	}
	if s != size {
		return nil, fmt.Errorf("base64: bad length %d, requires %d", s, size)
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(dst, src)
	return dst, nil
}

type publicKey struct {
	ed25519.PublicKey
}

func (p *publicKey) UnmarshalText(text []byte) error {
	key, err := decodeBase64(text, ed25519.PublicKeySize)
	if err == nil {
		p.PublicKey = key
	}
	return err
}

func (p *publicKey) MarshalText() ([]byte, error) {
	text, err := encodeBase64(p.PublicKey, ed25519.PublicKeySize)
	return text, err
}

type privateKey struct {
	ed25519.PrivateKey
}

func (p *privateKey) UnmarshalText(text []byte) error {
	key, err := decodeBase64(text, ed25519.PrivateKeySize)
	if err == nil {
		p.PrivateKey = key
	}
	return err
}

func (p *privateKey) MarshalText() ([]byte, error) {
	text, err := encodeBase64(p.PrivateKey, ed25519.PrivateKeySize)
	return text, err
}

type connPin struct {
	User struct {
		ID         uuid.UUID  `toml:"id"`
		Name       string     `toml:"name"`
		PrivateKey privateKey `toml:"private_key"`
	} `toml:"user"`
	Server struct {
		IPAddress string    `toml:"ip_address"`
		PublicKey publicKey `toml:"public_key"`
	} `toml:"server"`
}

type schema struct {
	Conns map[string]*connPin `toml:"connections"`
}

func (r *TOMLConnPinRepository) fileModified() (bool, error) {
	info, err := os.Stat(r.FilePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file timestamp: %w", err)
	}
	modTime := info.ModTime()
	mod := !r.modifiedAt.Equal(modTime)
	if mod {
		r.modifiedAt = modTime
	}
	return mod, nil
}

func (r *TOMLConnPinRepository) load() error {
	_, err := toml.DecodeFile(r.FilePath, &r.data)
	if err != nil {
		return fmt.Errorf("failed to load repository: %w", err)
	}
	return nil
}

func (r *TOMLConnPinRepository) save() error {
	file, err := os.OpenFile(r.FilePath, os.O_WRONLY|os.O_TRUNC, permRepository)
	if err != nil {
		return fmt.Errorf("failed to save repository: %w", err)
	}
	enc := toml.NewEncoder(file)
	enc.Indent = ""
	return enc.Encode(r.data)
}
