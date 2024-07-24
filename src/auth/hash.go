package auth

import (
	cryptorand "crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Hash struct {
	Config HashConfig
}

type HashConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func NewHashConfig(memory uint32, iterations uint32, parallelism uint8, saltLength uint32, keyLength uint32) HashConfig {
	return HashConfig{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		SaltLength:  saltLength,
		KeyLength:   keyLength,
	}
}

func NewHash(config HashConfig) *Hash {
	return &Hash{
		Config: config,
	}
}

func NewDefaultHash() *Hash {
	return &Hash{
		Config: HashConfig{
			Memory:      64 * 1024,
			Iterations:  4,
			Parallelism: 4,
			SaltLength:  16,
			KeyLength:   32,
		},
	}
}

func (h *Hash) HashPassword(password string) (hash []byte, encodedHash string, err error) {
	return generateFromPassword(password, h.Config)
}

func generateFromPassword(password string, config HashConfig) (hash []byte, encodedHash string, err error) {
	salt, err := generateRandomBytes(config.SaltLength)
	if err != nil {
		return nil, "", err
	}
	hash = argon2.IDKey([]byte(password), salt, config.Iterations, config.Memory, config.Parallelism, config.KeyLength)
	b64salt := base64.RawStdEncoding.EncodeToString(salt)
	b64hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, config.Memory, config.Iterations, config.Parallelism, b64salt, b64hash)
	return hash, encodedHash, nil
}

func generateRandomBytes(length uint32) ([]byte, error) {
	b := make([]byte, length)
	_, err := cryptorand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (h *Hash) VerifyPassword(encodedHash, password string) (match bool, err error) {
	return compareHashAndPassword(encodedHash, password)
}

func compareHashAndPassword(encodedHash, password string) (match bool, err error) {
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

func decodeHash(encodedHash string) (params HashConfig, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return params, salt, hash, errors.New("auth: invalid hash")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return HashConfig{}, nil, nil, err
	}

	if version != argon2.Version {
		return params, salt, hash, errors.New("auth: incompatible version of argon2")
	}

	params = HashConfig{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return HashConfig{}, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return HashConfig{}, nil, nil, err
	}
	params.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return HashConfig{}, nil, nil, err
	}
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}
