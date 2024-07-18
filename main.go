package main

import "log"

func main() {
	log.Println("Starting application gateway...")
	webhook := Route{
		Path: "webhook",
		URL:  "http://localhost:3000",
	}
	gateway := NewServer([]Route{webhook}, "/api/v1")
	log.Println("Server started on port 8080")
	gateway.RunServer()
}
