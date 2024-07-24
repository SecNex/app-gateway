package auth

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/google/uuid"
)

type AuthenticationHeader struct {
	Type  string
	Value string
	Token string
}

type Authentication struct {
	ID        uuid.UUID
	Header    AuthenticationHeader
	ExpiresIn int
	Hash      *Hash
}

func newRandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func StringToBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func Base64ToIDAndToken(s string) (uuid.UUID, string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return uuid.Nil, "", err
	}
	split := strings.Split(string(data), ":")
	id, err := uuid.Parse(split[0])
	if err != nil {
		return uuid.Nil, "", err
	}
	return id, split[1], nil
}

func Base64ToString(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func NewUUID() uuid.UUID {
	return uuid.New()
}

func NewAuthentication() *Authentication {
	return &Authentication{
		ID:        uuid.New(),
		ExpiresIn: 3600,
		Hash:      NewHash(NewHashConfig(64*1024, 4, 4, 16, 32)),
	}
}

func NewAuthenticationWithID(id uuid.UUID) *Authentication {
	return &Authentication{
		ID:        id,
		ExpiresIn: 3600,
	}
}

func (a *Authentication) GenerateToken() (string, string) {
	log.Printf("Generating token for authentication: %v", a.ID)
	value := newRandomString(32)
	token := fmt.Sprintf("%v:%v", a.ID, value)
	a.Header = AuthenticationHeader{
		Type:  "Bearer",
		Value: value,
		Token: StringToBase64(token),
	}
	_, encodedHash, err := a.Hash.HashPassword(value)
	if err != nil {
		log.Fatalf("Error hashing password: %v", err)
	}
	return a.Header.Token, encodedHash
}

func (a *Authentication) GenerateTokenWithSecret(secret string) (string, string) {
	// log.Printf("Generating token for authentication: %v", a.ID)
	token := fmt.Sprintf("%v:%v", a.ID, secret)
	a.Header = AuthenticationHeader{
		Type:  "Bearer",
		Value: secret,
		Token: StringToBase64(token),
	}
	_, encodedHash, err := a.Hash.HashPassword(secret)
	if err != nil {
		log.Fatalf("Error hashing password: %v", err)
	}
	return a.Header.Token, encodedHash
}

func GenerateRandomString(length int) string {
	return newRandomString(length)
}
