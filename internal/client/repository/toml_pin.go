package repository

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"

	client "github.com/charadev96/gonec/internal/client/domain"
)

const (
	permRepository = 0644
)

type TOMLPinRepository struct {
	FilePath string

	data       schema
	modifiedAt time.Time
}

func (r *TOMLPinRepository) Get(id string) (client.ServerPin, error) {
	modified, err := r.fileModified()
	pin := client.ServerPin{}
	if err != nil {
		return pin, err
	}
	if modified {
		if err := r.load(); err != nil {
			return pin, err
		}
	}
	pinRepr, ok := r.data.Servers[id]
	if !ok {
		return pin, fmt.Errorf("pin does not exist")
	}
	pin = pinRepr.toDomain(id)
	return pin, nil
}

func (r *TOMLPinRepository) Set(id string, pin client.ServerPin) error {
	modified, err := r.fileModified()
	if err != nil {
		return err
	}
	if modified {
		if err := r.load(); err != nil {
			return err
		}
	}
	if _, ok := r.data.Servers[id]; !ok {
		r.data.Servers[id] = &serverPin{}
	}
	r.data.Servers[id].fromDomain(pin)
	if err := r.save(); err != nil {
		return err
	}
	return nil
}

func (r *TOMLPinRepository) Delete(id string) error {
	_, ok := r.data.Servers[id]
	if !ok {
		return fmt.Errorf("pin does not exist")
	}
	delete(r.data.Servers, id)
	if err := r.save(); err != nil {
		return err
	}
	return nil
}

type publicKey struct {
	value ed25519.PublicKey
}

func (p *publicKey) UnmarshalText(text []byte) error {
	p.value = make([]byte, ed25519.PublicKeySize)
	_, err := hex.Decode(p.value, text)
	if err != nil {
		return err
	}
	return nil
}

func (p *publicKey) MarshalText() ([]byte, error) {
	text := make([]byte, ed25519.PublicKeySize*2)
	hex.Encode(text, p.value)
	return text, nil
}

type serverPin struct {
	IPAddress string    `toml:"address"`
	PublicKey publicKey `toml:"publicKey"`
}

func (p *serverPin) toDomain(id string) client.ServerPin {
	return client.ServerPin{
		ID:        id,
		IPAddress: p.IPAddress,
		PublicKey: p.PublicKey.value,
	}
}

func (p *serverPin) fromDomain(pin client.ServerPin) {
	p.IPAddress = pin.IPAddress
	p.PublicKey = publicKey{value: pin.PublicKey}
}

type schema struct {
	Servers map[string]*serverPin `toml:"servers"`
}

func (r *TOMLPinRepository) fileModified() (bool, error) {
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

func (r *TOMLPinRepository) load() error {
	_, err := toml.DecodeFile(r.FilePath, &r.data)
	if err != nil {
		return fmt.Errorf("failed to load repository: %w", err)
	}
	return nil
}

func (r *TOMLPinRepository) save() error {
	file, err := os.OpenFile(r.FilePath, os.O_WRONLY|os.O_TRUNC, permRepository)
	if err != nil {
		return fmt.Errorf("failed to save repository: %w", err)
	}
	enc := toml.NewEncoder(file)
	enc.Indent = ""
	return enc.Encode(r.data)
}
