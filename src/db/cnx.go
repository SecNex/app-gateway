package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

type DB struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Secure   bool
}

type Connection struct {
	DB         *DB
	Connection *sql.DB
}

func NewDBEnv() *Connection {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		dbPort = 5432
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbDatabase := os.Getenv("DB_DATABASE")
	if dbDatabase == "" {
		dbDatabase = "secnex_gateway"
	}

	return NewDB(dbHost, dbPort, dbUser, dbPassword, dbDatabase)
}

func NewDB(host string, port int, user string, password string, database string) *Connection {
	db := &DB{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
		Secure:   false,
	}
	return NewConnection(db)
}

func NewConnection(db *DB) *Connection {
	return &Connection{
		DB: db,
	}
}

func (db *DB) Connect() (*Connection, error) {
	return db.ConnectDatabase(db.Database)
}

func (db *DB) ConnectDatabase(name string) (*Connection, error) {
	log.Printf("Connecting to database %s...\n", name)
	connStr := "host=" + db.Host + " port=" + fmt.Sprintf("%d", db.Port) + " user=" + db.User + " password=" + db.Password + " dbname=" + name
	if db.Secure {
		connStr += " sslmode=require"
	} else {
		connStr += " sslmode=disable"
	}

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	log.Printf("Pinging database %s...\n", name)
	err = conn.Ping()
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to database %s.\n", name)

	return &Connection{
		DB:         db,
		Connection: conn,
	}, nil
}

func (c *Connection) Close() error {
	return c.Connection.Close()
}

func (c *Connection) TestConnection() error {
	return c.Connection.Ping()
}
