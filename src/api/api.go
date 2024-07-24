package api

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/secnex/secnex-api-gateway/auth"
	apitypes "github.com/secnex/secnex-api-gateway/types"
)

// Forward forwards the request to the target URL
func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	route, remainingPath, err := s.CheckProxyRequest(w, r)
	if err != nil {
		return
	}
	targetURL, err := s.constructTargetURL(route.URL, remainingPath, r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.proxyRequest(w, r, targetURL)
}

// Handler to refresh the routes
func (s *Server) Refresh(w http.ResponseWriter, r *http.Request) {
	if !auth.CheckAuthentication(w, r) {
		return
	}
	routes, err := s.RefreshRoutes(s.Database)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.SetRoutes(routes)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	result := apitypes.Result{
		Code:    http.StatusOK,
		Message: "Routes refreshed",
	}
	w.Write([]byte(result.String()))
}

// Handler to get health status
func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	result := apitypes.ResultHealth{
		Code:    http.StatusOK,
		Message: "OK",
		Status:  "Healthy",
	}
	w.Write([]byte(result.String()))
}

// Check of request
func (s *Server) CheckProxyRequest(w http.ResponseWriter, r *http.Request) (Route, string, error) {
	clientIP := r.RemoteAddr
	w.Header().Set("Content-Type", "application/json")
	_, routePath, remainingPath, err := s.extractPaths(r.URL.Path)
	if err != nil {
		result := apitypes.ResultError{
			Code:    http.StatusBadRequest,
			Message: "Bad request",
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(result.String()))
		return Route{}, "", err
	}

	route, err := s.GetRoute(routePath)
	if err != nil {
		result := apitypes.ResultError{
			Code:    http.StatusNotFound,
			Message: "Route not found",
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(result.String()))
		return Route{}, "", err
	}
	// Check if the client IP is allowed
	if !s.CheckAllowedIP(route, clientIP) {
		result := apitypes.ResultError{
			Code:    http.StatusForbidden,
			Message: "Forbidden",
			Error:   "IP not allowed",
		}
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(result.String()))
		return Route{}, "", fmt.Errorf("ip not allowed")
	}

	// Check if the client IP is blocked
	if s.CheckBlockedIPs(route, clientIP) {
		result := apitypes.ResultError{
			Code:    http.StatusForbidden,
			Message: "Forbidden",
			Error:   "IP blocked",
		}
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(result.String()))
		return Route{}, "", fmt.Errorf("ip blocked")
	}

	// Check if the user agent is allowed
	if !s.CheckUserAgent(route, r.UserAgent()) {
		result := apitypes.ResultError{
			Code:    http.StatusForbidden,
			Message: "Forbidden",
			Error:   "User agent not allowed",
		}
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(result.String()))
		return Route{}, "", fmt.Errorf("user agent not allowed")
	}

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
			result := apitypes.ResultError{
				Code:    http.StatusMethodNotAllowed,
				Message: "Method not allowed",
				Error:   "Method not allowed",
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(result.String()))
			return Route{}, "", fmt.Errorf("method not allowed")
		}
	}

	// Check if the Authorization header is required
	if route.RequiredAuth {
		if err := s.CheckAuthorizationHeader(r); err != nil {
			result := apitypes.ResultError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
				Error:   err.Error(),
			}
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(result.String()))
			return Route{}, "", err
		}
	}

	return route, remainingPath, nil
}

// checkAuthorizationHeader checks if the Authorization header is present
func (s *Server) CheckAuthorizationHeader(r *http.Request) error {
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
	// Check if IPv4 or IPv6
	if strings.Contains(clientIP, "[") || strings.Contains(clientIP, "]") {
		// Get all between []
		clientIP = strings.Split(clientIP, "[")[1]
		clientIP = strings.Split(clientIP, "]")[0]
	} else {
		// Get all before :
		clientIP = strings.Split(clientIP, ":")[0]
	}

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

// Check blocked IPs
func (s *Server) CheckBlockedIPs(r Route, clientIP string) bool {
	// Check if IPv4 or IPv6
	if strings.Contains(clientIP, "[") || strings.Contains(clientIP, "]") {
		// Get all between []
		clientIP = strings.Split(clientIP, "[")[1]
		clientIP = strings.Split(clientIP, "]")[0]
	} else {
		// Get all before :
		clientIP = strings.Split(clientIP, ":")[0]
	}

	clientAddr := net.ParseIP(clientIP)
	for _, ip := range r.BlockedIPs {
		if net.ParseIP(string(ip)).Equal(clientAddr) {
			return true
		}
	}

	return false
}

// Check user agent is allowed or rejected
func (s *Server) CheckUserAgent(r Route, userAgent string) bool {
	status := true
	if len(r.AllowedUserAgents) > 0 {
		status = false
		for _, allowedUserAgent := range r.AllowedUserAgents {
			if UserAgent(userAgent) == allowedUserAgent {
				status = true
				break
			}
		}
	}

	if !status {
		return false
	}

	if len(r.RejectedUserAgents) > 0 {
		for _, rejectedUserAgent := range r.RejectedUserAgents {
			if UserAgent(userAgent) == rejectedUserAgent {
				return false
			}
		}
	}

	return true
}

// extractPaths extracts the route path and the remaining path
func (s *Server) extractPaths(urlPath string) (string, string, string, error) {
	path := strings.TrimPrefix(urlPath, s.BasePath)
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

// constructTargetURL constructs the target URL
func (s *Server) constructTargetURL(baseURL, remainingPath, rawQuery string) (*url.URL, error) {
	targetURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	targetURL.Path = fmt.Sprintf("%s/%s", strings.TrimRight(targetURL.Path, "/"), remainingPath)
	targetURL.RawQuery = rawQuery

	return targetURL, nil
}

func (s *Server) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL *url.URL) {
	s.Proxy.Director = func(req *http.Request) {
		req.URL = targetURL
		req.Host = targetURL.Host
		req.Method = r.Method
	}

	rec := httptest.NewRecorder()
	s.Proxy.ServeHTTP(rec, r)

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
