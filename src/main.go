package main

import (
	"log"

	"github.com/secnex/secnex-api-gateway/api"
	"github.com/secnex/secnex-api-gateway/db"
)

const SERVER = "SGW01"

func main() {
	log.Println("Starting application gateway...")
	db := db.NewDB("localhost", 5432, "postgres", "postgres", "secnex_core")
	cnx, err := db.DB.Connect()
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}
	defer cnx.Connection.Close()
	serverConfig, err := cnx.GetServerConfiguration(SERVER)
	if err != nil {
		log.Fatalf("Error getting server configuration: %s", err)
	}

	routes, err := api.GetRoutes(cnx, serverConfig.ID)
	if err != nil {
		log.Fatalf("Error getting routes: %s", err)
	}

	server := api.NewServer(serverConfig, cnx)
	server.SetRoutes(routes)
	server.RunServer()
}
