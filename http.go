package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// Route struct
type Route struct {
	Path           string
	URL            string
	AllowedMethods []Method
	AllowedIPs     []IPAddress
	DefaultAllowed bool
	RequiredAuth   bool
	ForwardSubPath bool
}

type Method string
type IPAddress string

// Server struct
type Server struct {
	port     string
	basePath string
	proxy    *httputil.ReverseProxy
	routes   []Route
}

func NewRoute(path string, url string, methods []string, ips []string, defaultAllowed bool, requiredAuth bool, forwardSubPath bool) Route {
	allowedMethods := make([]Method, len(methods))
	for i, method := range methods {
		log.Printf("Method: %s\n", method)
		allowedMethods[i] = Method(method)
	}

	allowedIPs := make([]IPAddress, len(ips))
	for i, ip := range ips {
		allowedIPs[i] = IPAddress(ip)
	}

	log.Printf("Route: %s\n", path)
	log.Printf("URL: %s\n", url)
	return Route{
		Path:           path,
		URL:            url,
		AllowedMethods: allowedMethods,
		AllowedIPs:     allowedIPs,
		DefaultAllowed: defaultAllowed,
		RequiredAuth:   requiredAuth,
		ForwardSubPath: forwardSubPath,
	}
}

// NewServer creates a new server
func NewServer(port int, routes []Route, basePath string) *Server {
	return &Server{
		port:     fmt.Sprintf(":%d", port),
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
		log.Printf("Full route path -> %s/%s\n", s.basePath, route.Path)
		if route.ForwardSubPath {
			http.HandleFunc(fmt.Sprintf("%s/%s/", s.basePath, route.Path), s.handler)
		}
		http.HandleFunc(fmt.Sprintf("%s/%s", s.basePath, route.Path), s.handler)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	log.Fatal(http.ListenAndServe(s.port, nil))
}

// Forward forwards the request to the target URL
func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	clientIP := r.RemoteAddr
	method := r.Method

	log.Printf("Received %s request from %s - %s\n", method, clientIP, r.URL.Path)

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

	// Check if the client IP is allowed
	if !s.CheckAllowedIP(route, clientIP) {
		http.Error(w, "IP not allowed", http.StatusForbidden)
		return
	}
	log.Printf("Client IP %s is allowed\n", clientIP)

	log.Printf("Checking if method %s is allowed\n", method)
	// Check if the method is allowed
	if len(route.AllowedMethods) > 0 {
		methodAllowed := false
		for _, allowedMethod := range route.AllowedMethods {
			if Method(r.Method) == allowedMethod {
				methodAllowed = true
				break
			}
		}

		if !methodAllowed {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
	log.Printf("Method %s is allowed\n", r.Method)

	// Check if the Authorization header is required
	if route.RequiredAuth {
		if err := s.checkAuthorizationHeader(r); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
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

// checkAllowedIP checks if the client IP is allowed
func (s *Server) CheckAllowedIP(r Route, clientIP string) bool {
	log.Println("checkAllowedIP called")

	// Check if IPv4 or IPv6
	if strings.Contains(clientIP, "[") || strings.Contains(clientIP, "]") {
		log.Println("Client uses IPv6")
		// Get all between []
		clientIP = strings.Split(clientIP, "[")[1]
		clientIP = strings.Split(clientIP, "]")[0]
	} else {
		log.Println("Client uses IPv4")
		// Get all before :
		clientIP = strings.Split(clientIP, ":")[0]
	}

	log.Printf("Client IP: %s\n", clientIP)

	if len(r.AllowedIPs) == 0 {
		return r.DefaultAllowed
	}

	clientAddr := net.ParseIP(clientIP)
	for _, ip := range r.AllowedIPs {
		if net.ParseIP(string(ip)).Equal(clientAddr) {
			return true
		}
	}

	return false
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

func (s *Server) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL *url.URL) {
	log.Println("proxyRequest called")
	log.Printf("Original Request Method: %s\n", r.Method) // Logging the original method
	s.proxy.Director = func(req *http.Request) {
		req.URL = targetURL
		req.Host = targetURL.Host
		req.Header.Set("X-Forwarded-For", r.RemoteAddr)

		// Ensure the request method is not altered
		req.Method = r.Method

		log.Printf("Forwarded Request Method: %s\n", req.Method) // Logging the forwarded method
	}

	rec := httptest.NewRecorder()
	s.proxy.ServeHTTP(rec, r)

	// Change 502 status to 404 before sending the response
	if rec.Code == http.StatusBadGateway {
		rec.Code = http.StatusNotFound
	}

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
