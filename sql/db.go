package sql

import (
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ApiKey struct {
}

func NewDBCOnfig(host string, port int, user, password, dbName, sslMode string) DBConfig {
	return DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	}
}

func NewDB(cfg DBConfig) (*DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func generateApiKey(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (db *DB) NewAuthentication(expiresIn int) (uuid.UUID, ApiKey, error) {
	__id := uuid.New()
	__type := "Bearer"
	__key := generateApiKey(32)
	__expiresIn := expiresIn
	if expiresIn == 0 {
		_, err := db.Exec("INSERT INTO authentications (id, prefix, key) VALUES ($1, $2, $3)", __id, __type, __key)
		if err != nil {
			return uuid.Nil, ApiKey{}, err
		}
	} else {
		_, err := db.Exec("INSERT INTO authentications (id, prefix, key, expires_in) VALUES ($1, $2, $3, $4)", __id, __type, __key, __expiresIn)
		if err != nil {
			return uuid.Nil, ApiKey{}, err
		}
	}

	return __id, ApiKey{}, nil
}

func (db *DB) GetAuthentication(id uuid.UUID) (Authentication, error) {
	var auth Authentication
	err := db.QueryRow("SELECT * FROM authentications WHERE id = $1", id).Scan(&auth.ID, &auth.Prefix, &auth.Key, &auth.ExpiresIn)
	if err != nil {
		return Authentication{}, err
	}

	return auth, nil
}
