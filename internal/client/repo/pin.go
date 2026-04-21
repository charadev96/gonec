package repo

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/google/uuid"

	client "github.com/charadev96/gonec/internal/client/domain"
	shared "github.com/charadev96/gonec/internal/shared/domain"
)

const (
	permRepository = 0644
)

type YAMLConnPinRepository struct {
	file string

	data       schema
	modifiedAt time.Time
}

func NewYAMLConnPinRepository(f string) *YAMLConnPinRepository {
	r := &YAMLConnPinRepository{
		file: f,
		data: schema{make(map[string]*connPin)},
	}
	return r
}

func (r *YAMLConnPinRepository) Get(id string) (client.ConnPin, error) {
	modified, err := r.fileModified()
	if err != nil {
		return client.ConnPin{}, fmt.Errorf("compare timestamp: %w", err)
	}
	if modified {
		if err := r.load(); err != nil {
			return client.ConnPin{}, fmt.Errorf("load repository: %w", err)
		}
	}
	p, ok := r.data.Conns[id]
	if !ok {
		return client.ConnPin{}, fmt.Errorf("%q: %w", id, shared.ErrNotExist)
	}
	return connPinFromDB(p, id), nil
}

func (r *YAMLConnPinRepository) Set(id string, pin client.ConnPin) error {
	modified, err := r.fileModified()
	if err != nil {
		return fmt.Errorf("compare timestamp: %w", err)
	}
	if modified {
		if err := r.load(); err != nil {
			return fmt.Errorf("load repository: %w", err)
		}
	}
	r.data.Conns[id] = connPinToDB(pin)
	if err := r.save(); err != nil {
		return fmt.Errorf("save repository: %w", err)
	}
	return nil
}

func (r *YAMLConnPinRepository) Delete(id string) error {
	_, ok := r.data.Conns[id]
	if !ok {
		return fmt.Errorf("%q: %w", id, shared.ErrNotExist)
	}
	delete(r.data.Conns, id)
	if err := r.save(); err != nil {
		return fmt.Errorf("save repository: %w", err)
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
	dst := make([]byte, base64.StdEncoding.EncodedLen(size))
	if s == 0 {
		return dst, nil
	}
	if s != size {
		return dst, fmt.Errorf("base64: bad length %d, requires %d", s, size)
	}
	base64.StdEncoding.Encode(dst, src)
	return dst, nil
}

type publicKey []byte

func (p *publicKey) UnmarshalYAML(text []byte) error {
	key, err := decodeBase64(text, ed25519.PublicKeySize)
	if err == nil {
		*p = key
	}
	return err
}

func (p publicKey) MarshalYAML() ([]byte, error) {
	text, err := encodeBase64(p, ed25519.PublicKeySize)
	return text, err
}

type privateKey []byte

func (p *privateKey) UnmarshalYAML(text []byte) error {
	key, err := decodeBase64(text, ed25519.PrivateKeySize)
	if err != nil {
		fmt.Println(text)
	}
	if err == nil {
		*p = key
	}
	return err
}

func (p privateKey) MarshalYAML() ([]byte, error) {
	text, err := encodeBase64(p, ed25519.PrivateKeySize)
	return text, err
}

type connPin struct {
	User struct {
		ID         uuid.UUID  `yaml:"id"`
		Name       string     `yaml:"name,omitempty"`
		PrivateKey privateKey `yaml:"private_key,omitempty"`
	} `yaml:"user"`
	Server struct {
		IPAddress string    `yaml:"ip_address"`
		PublicKey publicKey `yaml:"public_key"`
	} `yaml:"server"`
}

func connPinFromDB(p *connPin, id string) client.ConnPin {
	return client.ConnPin{
		ID: id,
		User: client.UserPrivateIdentity{
			ID:         p.User.ID,
			PrivateKey: ed25519.PrivateKey(p.User.PrivateKey),
		},
		Server: shared.ServerIdentity{
			IPAddress: p.Server.IPAddress,
			PublicKey: ed25519.PublicKey(p.Server.PublicKey),
		},
	}
}

func connPinToDB(pin client.ConnPin) *connPin {
	p := &connPin{}
	p.User.ID = pin.User.ID
	p.User.Name = ""
	p.User.PrivateKey = privateKey(pin.User.PrivateKey)
	p.Server.IPAddress = pin.Server.IPAddress
	p.Server.PublicKey = publicKey(pin.Server.PublicKey)
	return p
}

type schema struct {
	Conns map[string]*connPin `yaml:"connections"`
}

func (r *YAMLConnPinRepository) fileModified() (bool, error) {
	info, err := os.Stat(r.file)
	if err != nil {
		return false, err
	}
	modTime := info.ModTime()
	mod := !r.modifiedAt.Equal(modTime)
	if mod {
		r.modifiedAt = modTime
	}
	return mod, nil
}

func (r *YAMLConnPinRepository) load() error {
	f, err := os.OpenFile(r.file, os.O_RDONLY, permRepository)
	if err != nil {
		return err
	}
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read file %s: %w", r.file, err)
	}

	err = yaml.Unmarshal(raw, &r.data)
	if err != nil {
		return fmt.Errorf("unmarshal yaml: %w", err)
	}
	if r.data.Conns == nil {
		r.data.Conns = make(map[string]*connPin)
	}

	return nil
}

func (r *YAMLConnPinRepository) save() error {
	raw, err := yaml.Marshal(r.data)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}

	f, err := os.OpenFile(r.file, os.O_WRONLY|os.O_TRUNC, permRepository)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(raw)
	if err != nil {
		return err
	}

	return nil
}
