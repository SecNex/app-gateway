package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// Route struct
type Route struct {
	Path string
	URL  string
}

// Server struct
type Server struct {
	basePath string
	proxy    *httputil.ReverseProxy
	routes   []Route
}

// NewServer creates a new server
func NewServer(routes []Route, basePath string) *Server {
	return &Server{
		proxy:    &httputil.ReverseProxy{},
		routes:   routes,
		basePath: basePath,
	}
}

// RunServer runs the server
func (s *Server) RunServer() {
	for _, route := range s.routes {
		log.Printf("Registering new route -> %s\n", route.Path)
		log.Printf("Target for %s -> %s\n", route.Path, route.URL)
		http.HandleFunc(fmt.Sprintf("%s/%s/", s.basePath, route.Path), s.handler)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Forward forwards the request to the target URL
func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if err := s.checkAuthorizationHeader(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	_, routePath, remainingPath, err := s.extractPaths(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	route, err := s.getRoute(routePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	targetURL, err := s.constructTargetURL(route.URL, remainingPath, r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Proxying request to %s\n", targetURL.String())

	s.proxyRequest(w, r, targetURL)

	duration := time.Since(start)
	log.Printf("Request duration: %v\n", duration)
}

// Forward forwards the request to the target URL
func (s *Server) Forward(w http.ResponseWriter, r *http.Request) {
	// Get the url query parameter and forward the request to the url
	__url := r.URL.Query().Get("url")
	if __url == "" {
		http.Error(w, "url query parameter missing", http.StatusBadRequest)
		return
	}

	targetURL, err := url.ParseRequestURI(__url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Proxying request to %s\n", targetURL.String())

	s.proxyRequest(w, r, targetURL)
}

// checkAuthorizationHeader checks if the Authorization header is present
func (s *Server) checkAuthorizationHeader(r *http.Request) error {
	log.Println("checkAuthorizationHeader called")
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return fmt.Errorf("authorization header missing")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid Authorization header format")
	}

	if parts[0] != "Bearer" {
		return fmt.Errorf("invalid Authorization header type")
	}

	return nil
}

// extractPaths extracts the route path and the remaining path
func (s *Server) extractPaths(urlPath string) (string, string, string, error) {
	log.Println("extractPaths called")
	path := strings.TrimPrefix(urlPath, s.basePath)
	if path == "" || path == "/" {
		return "", "", "", fmt.Errorf("invalid path")
	}

	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid path")
	}

	routePath := parts[1]
	remainingPath := ""
	if len(parts) == 3 {
		remainingPath = parts[2]
	}

	return path, routePath, remainingPath, nil
}

// getRoute returns the route for the given path
func (s *Server) getRoute(path string) (Route, error) {
	log.Println("getRoute called")
	for _, route := range s.routes {
		if route.Path == path {
			return route, nil
		}
	}
	return Route{}, fmt.Errorf("Route not found")
}

// constructTargetURL constructs the target URL
func (s *Server) constructTargetURL(baseURL, remainingPath, rawQuery string) (*url.URL, error) {
	log.Println("constructTargetURL called")
	targetURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	targetURL.Path = fmt.Sprintf("%s/%s", strings.TrimRight(targetURL.Path, "/"), remainingPath)
	targetURL.RawQuery = rawQuery

	return targetURL, nil
}

// proxyRequest proxies the request to the target URL
func (s *Server) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL *url.URL) {
	log.Println("proxyRequest called")
	s.proxy.Director = func(req *http.Request) {
		req.URL = targetURL
		req.Host = targetURL.Host
		req.Header.Set("X-Forwarded-For", r.RemoteAddr)
	}

	rec := httptest.NewRecorder()
	s.proxy.ServeHTTP(rec, r)

	for k, v := range rec.Header() {
		w.Header()[k] = v
	}
	w.WriteHeader(rec.Code)
	w.Write(rec.Body.Bytes())

	s.logResponseDetails(rec.Body)
}

// logResponseDetails logs the response details
func (s *Server) logResponseDetails(body *bytes.Buffer) {
	log.Println("logResponseDetails called")
	bodyBytes, err := io.ReadAll(body)
	if err == nil {
		bodyString := string(bodyBytes)
		titleStart := strings.Index(bodyString, "<title>")
		titleEnd := strings.Index(bodyString, "</title>")
		if titleStart != -1 && titleEnd != -1 && titleStart < titleEnd {
			title := bodyString[titleStart+len("<title>") : titleEnd]
			log.Printf("Page title: %s\n", title)
		}
	}
}
