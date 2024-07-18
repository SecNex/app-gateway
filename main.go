package main

import "log"

func main() {
	log.Println("Starting application gateway...")
	webhook := NewRoute("webhook", "http://localhost:8081", []string{"GET"}, []string{"127.0.0.1", "::1"}, false, true, true)
	gateway := NewServer(8080, []Route{webhook}, "/api/v1")
	log.Println("Server started on port 8080")
	gateway.RunServer()
}
