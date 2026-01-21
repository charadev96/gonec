package shared

import (
	"crypto/ed25519"
	"encoding/hex"
	"os"

	"github.com/BurntSushi/toml"
)

const (
	permRegistry = 0644
)

type PublicKey struct {
	hex   []byte
	value ed25519.PublicKey
}

func (p *PublicKey) UnmarshalText(text []byte) error {
	value := make([]byte, ed25519.PublicKeySize)
	_, err := hex.Decode(value, text)
	if err != nil {
		return err
	}
	p.Update(value)
	return nil
}

func (p *PublicKey) MarshalText() (text []byte, err error) {
	text = make([]byte, ed25519.PublicKeySize*2)
	hex.Encode(text, p.Value())
	return text, nil
}

func (p *PublicKey) Value() ed25519.PublicKey {
	return p.value
}

func (p *PublicKey) Hex() []byte {
	return p.hex
}

func (p *PublicKey) Update(key ed25519.PublicKey) {
	p.value = key
	p.hex = make([]byte, ed25519.PublicKeySize*2)
	hex.Encode(p.hex, key)
}

type ServerInfo struct {
	IPAddress string    `toml:"address"`
	PublicKey PublicKey `toml:"publicKey"`
}

type serverTable struct {
	Servers map[string]*ServerInfo `toml:"servers"`
}

type ServerRegistryTOML struct {
	FilePath string

	table serverTable
}

func (r *ServerRegistryTOML) Get(id string) (*ServerInfo, bool) {
	s, ok := r.table.Servers[id]
	return s, ok
}

func (r *ServerRegistryTOML) LoadFile() error {
	_, err := toml.DecodeFile(r.FilePath, &r.table)
	return err
}

func (r *ServerRegistryTOML) SaveFile() error {
	file, err := os.OpenFile(r.FilePath, os.O_WRONLY|os.O_TRUNC, permRegistry)
	if err != nil {
		return err
	}
	enc := toml.NewEncoder(file)
	enc.Indent = ""
	return enc.Encode(r.table)
}
