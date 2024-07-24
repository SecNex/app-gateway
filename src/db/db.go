package db

import (
	"database/sql"
	"fmt"
	"log"
)

type Server struct {
	ID        string
	Name      string
	Address   string
	Port      int
	BasePath  string
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
	DeletedAt sql.NullString
}

type Route struct {
	ID              string
	Name            string
	Path            string
	URL             string
	FirewallID      string
	ServerID        string
	GlobalAvailable bool
	ForwardSubPath  bool
	CreatedAt       sql.NullString
	UpdatedAt       sql.NullString
	DeletedAt       sql.NullString
}

type Firewall struct {
	ID          string
	Name        string
	AllowAll    bool
	RequireAuth bool
	CreatedAt   sql.NullString
	UpdatedAt   sql.NullString
	DeletedAt   sql.NullString
}

type Method struct {
	FirewallID string
	RouteID    string
	Method     string
	Action     string
	CreatedAt  sql.NullString
	UpdatedAt  sql.NullString
	DeletedAt  sql.NullString
}

type IP struct {
	FirewallID string
	RouteID    string
	IP         string
	Action     string
	CreatedAt  sql.NullString
	UpdatedAt  sql.NullString
	DeletedAt  sql.NullString
}

type UserAgent struct {
	FirewallID string
	RouteID    string
	UserAgent  string
	Action     string
	CreatedAt  sql.NullString
	UpdatedAt  sql.NullString
	DeletedAt  sql.NullString
}

const ACTION_ALLOW = "ALLOW"
const ACTION_REJECT = "BLOCK"

func (db *DB) ConnectInit() (*Connection, error) {
	log.Printf("Connecting to database %s...\n", "postgres")
	cnx, err := db.ConnectDatabase("postgres")
	if err != nil {
		return nil, err
	}

	log.Printf("Creating database %s...\n", db.Database)
	err = cnx.CreateDatabase(db.Database)
	if err != nil {
		return nil, err
	}

	cnx.Connection.Close()

	log.Printf("Connecting to database %s...\n", db.Database)
	return db.ConnectDatabase(db.Database)
}

func (c *Connection) CreateDatabase(name string) error {
	log.Printf("Dropping and creating database %s...\n", name)
	_, err := c.Connection.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", name))
	if err != nil {
		return err
	}

	_, err = c.Connection.Exec(fmt.Sprintf("CREATE DATABASE %s", name))
	if err != nil {
		return err
	}

	return nil
}

func (c *Connection) GetRoutes(server string) ([]Route, error) {
	rows, err := c.Connection.Query("SELECT * FROM routes WHERE deleted_at IS NULL AND server_id = $1", server)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	routes := []Route{}
	for rows.Next() {
		route := Route{}
		// id, name, route, target, firewall_id, include_subroutes, created_at, updated_at, deleted_at
		err := rows.Scan(&route.ID, &route.Name, &route.Path, &route.URL, &route.FirewallID, &route.ServerID, &route.GlobalAvailable, &route.ForwardSubPath, &route.CreatedAt, &route.UpdatedAt, &route.DeletedAt)
		if err != nil {
			return nil, err
		}

		routes = append(routes, route)
	}

	return routes, nil
}

func (c *Connection) GetFirewall(id string) (Firewall, error) {
	rows, err := c.Connection.Query("SELECT * FROM firewalls WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return Firewall{}, err
	}
	defer rows.Close()

	firewall := Firewall{}
	for rows.Next() {
		err := rows.Scan(&firewall.ID, &firewall.Name, &firewall.AllowAll, &firewall.RequireAuth, &firewall.CreatedAt, &firewall.UpdatedAt, &firewall.DeletedAt)
		if err != nil {
			return Firewall{}, err
		}
	}

	return firewall, nil
}

func (c *Connection) GetMethods(firewall string, route string) ([]Method, error) {
	rows, err := c.Connection.Query("SELECT * FROM methods WHERE firewall_id = $1 AND route_id = $2 AND action = $3 AND deleted_at IS NULL", firewall, route, ACTION_ALLOW)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	methods := []Method{}
	for rows.Next() {
		method := Method{}
		err := rows.Scan(&method.FirewallID, &method.RouteID, &method.Method, &method.Action, &method.CreatedAt, &method.UpdatedAt, &method.DeletedAt)
		if err != nil {
			return nil, err
		}

		methods = append(methods, method)
	}

	return methods, nil
}

func (c *Connection) GetIPs(firewall string, route string, action string) ([]IP, error) {
	rows, err := c.Connection.Query("SELECT * FROM ips WHERE firewall_id = $1 AND route_id = $2 AND action = $3 AND deleted_at IS NULL", firewall, route, action)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ips := []IP{}
	for rows.Next() {
		ip := IP{}
		err := rows.Scan(&ip.FirewallID, &ip.RouteID, &ip.IP, &ip.Action, &ip.CreatedAt, &ip.UpdatedAt, &ip.DeletedAt)
		if err != nil {
			return nil, err
		}

		ips = append(ips, ip)
	}

	return ips, nil
}

func (c *Connection) GetUserAgents(firewall string, route string, action string) ([]UserAgent, error) {
	rows, err := c.Connection.Query("SELECT * FROM useragents WHERE firewall_id = $1 AND route_id = $2 AND action = $3 AND deleted_at IS NULL", firewall, route, action)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userAgents := []UserAgent{}
	for rows.Next() {
		userAgent := UserAgent{}
		err := rows.Scan(&userAgent.FirewallID, &userAgent.RouteID, &userAgent.UserAgent, &userAgent.Action, &userAgent.CreatedAt, &userAgent.UpdatedAt, &userAgent.DeletedAt)
		if err != nil {
			return nil, err
		}

		userAgents = append(userAgents, userAgent)
	}

	return userAgents, nil
}

func (c *Connection) GetServerConfiguration(name string) (Server, error) {
	rows, err := c.Connection.Query("SELECT * FROM servers WHERE name = $1 AND deleted_at IS NULL LIMIT 1", name)
	if err != nil {
		return Server{}, err
	}
	defer rows.Close()

	server := Server{}
	for rows.Next() {
		err := rows.Scan(&server.ID, &server.Name, &server.Address, &server.Port, &server.BasePath, &server.CreatedAt, &server.UpdatedAt, &server.DeletedAt)
		if err != nil {
			return Server{}, err
		}
	}

	return server, nil
}
