package api

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/secnex/secnex-api-gateway/db"
	"github.com/secnex/secnex-api-gateway/middleware"
)

// Server struct
type Server struct {
	ID       string
	Name     string
	Port     string
	BasePath string
	Proxy    *httputil.ReverseProxy
	Routes   []Route
	Database *db.Connection
	MU       sync.Mutex
}

// NewServer creates a new server
func NewServer(server db.Server, database *db.Connection) *Server {
	return &Server{
		ID:       server.ID,
		Name:     server.Name,
		Port:     fmt.Sprintf(":%d", server.Port),
		Proxy:    &httputil.ReverseProxy{},
		Routes:   []Route{},
		BasePath: server.BasePath,
		Database: database,
	}
}

// RunServer runs the server
func (s *Server) RunServer() {
	r := http.NewServeMux()

	s.MU.Lock()
	for _, route := range s.Routes {
		if route.ForwardSubPath {
			r.HandleFunc(fmt.Sprintf("%s/%s/", s.BasePath, route.Path), s.Handler)
		}
		r.HandleFunc(fmt.Sprintf("%s/%s", s.BasePath, route.Path), s.Handler)
		log.Printf("New route registered: %s/%s -> %s\n", s.BasePath, route.Path, route.URL)
	}

	r.HandleFunc("/api/health", s.Health)

	r.HandleFunc("/api/gateway/refresh", s.Refresh)

	s.MU.Unlock()

	go s.StartRouteRefresher(5)

	loggedRouter := middleware.LoggingMiddleware(r)
	log.Printf("Starting %s (%s) on port %s\n", s.Name, s.ID, s.Port)
	log.Fatal(http.ListenAndServe(s.Port, loggedRouter))
}

func (s *Server) StartRouteRefresher(minutes int) {
	ticker := time.NewTicker(time.Duration(minutes) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Refreshing routes...")
		s.RefreshRoutesPeriodically()
		log.Println("Routes refreshed! Next refresh in", minutes, "minutes.")
	}
}

func (s *Server) RefreshRoutesPeriodically() {
	routes, err := s.RefreshRoutes(s.Database)
	if err != nil {
		log.Printf("Error refreshing routes: %s\n", err)
		return
	}

	s.SetRoutes(routes)
}
