package api

import (
	"fmt"
	"log"

	"github.com/secnex/secnex-api-gateway/db"
)

// Route struct
type Route struct {
	Path               string
	URL                string
	AllowedMethods     []Method
	AllowedIPs         []IPAddress
	BlockedIPs         []IPAddress
	AllowedUserAgents  []UserAgent
	RejectedUserAgents []UserAgent
	DefaultAllowed     bool
	RequiredAuth       bool
	ForwardSubPath     bool
}

type Method string
type IPAddress string
type UserAgent string

func NewRoute(path string, url string, methods []Method, allowedIPs []IPAddress, blockedIPs []IPAddress, allowedUserAgents []UserAgent, rejectedUserAgents []UserAgent, defaultAllowed bool, requiredAuth bool, forwardSubPath bool) Route {
	return Route{
		Path:               path,
		URL:                url,
		AllowedMethods:     methods,
		AllowedIPs:         allowedIPs,
		BlockedIPs:         blockedIPs,
		AllowedUserAgents:  allowedUserAgents,
		RejectedUserAgents: rejectedUserAgents,
		DefaultAllowed:     defaultAllowed,
		RequiredAuth:       requiredAuth,
		ForwardSubPath:     forwardSubPath,
	}
}

func (s *Server) SetRoutes(routes []Route) {
	s.Routes = routes
}

// getRoute returns the route for the given path
func (s *Server) GetRoute(path string) (Route, error) {
	for _, route := range s.Routes {
		if route.Path == path {
			return route, nil
		}
	}
	return Route{}, fmt.Errorf("Route not found")
}

func GetRoutes(cnx *db.Connection, serverId string) ([]Route, error) {
	routes, err := cnx.GetRoutes(serverId)
	if err != nil {
		return nil, err
	}

	var __routes []Route
	for _, __route := range routes {
		firewall, err := cnx.GetFirewall(__route.FirewallID)
		if err != nil {
			return nil, err
		}
		var __methods []Method
		methods, err := cnx.GetMethods(firewall.ID, __route.ID)
		if err != nil {
			return nil, err
		}
		for _, method := range methods {
			__methods = append(__methods, Method(method.Method))
		}
		var __allowedIPs []IPAddress
		ips, err := cnx.GetIPs(firewall.ID, __route.ID, db.ACTION_ALLOW)
		if err != nil {
			return nil, err
		}
		for _, ip := range ips {
			__allowedIPs = append(__allowedIPs, IPAddress(ip.IP))
		}
		var __blockedIPs []IPAddress
		blockedIPs, err := cnx.GetIPs(firewall.ID, __route.ID, db.ACTION_REJECT)
		if err != nil {
			return nil, err
		}
		for _, blockedIP := range blockedIPs {
			__blockedIPs = append(__blockedIPs, IPAddress(blockedIP.IP))
		}
		var __useragents []UserAgent
		useragents, err := cnx.GetUserAgents(firewall.ID, __route.ID, db.ACTION_ALLOW)
		if err != nil {
			return nil, err
		}
		for _, useragent := range useragents {
			__useragents = append(__useragents, UserAgent(useragent.UserAgent))
		}
		var __rejectedUserAgents []UserAgent
		rejectedUserAgents, err := cnx.GetUserAgents(firewall.ID, __route.ID, db.ACTION_REJECT)
		if err != nil {
			return nil, err
		}
		for _, rejectedUserAgent := range rejectedUserAgents {
			__rejectedUserAgents = append(__rejectedUserAgents, UserAgent(rejectedUserAgent.UserAgent))
		}
		__route := NewRoute(__route.Path, __route.URL, __methods, __allowedIPs, __blockedIPs, __useragents, __rejectedUserAgents, firewall.AllowAll, firewall.RequireAuth, __route.ForwardSubPath)
		__routes = append(__routes, __route)
	}

	log.Printf("Loaded %d routes.\n", len(__routes))

	return __routes, nil
}

func (s *Server) RefreshRoutes(cnx *db.Connection) ([]Route, error) {
	routes, err := GetRoutes(cnx, s.ID)
	if err != nil {
		return nil, err
	}

	return routes, nil
}
