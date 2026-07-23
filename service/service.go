// Package service is a UI-facing facade over the mgmt transport core. It
// holds a single long-lived authenticated client — for the lifetime of a GUI
// session — and exposes coarse, bindable operations that a Wails desktop app
// (or an HTTP backend serving a browser UI) can call directly.
//
// It is deliberately free of any UI-framework imports: the same Service can be
// bound into Wails, wrapped by net/http handlers, or driven from tests. All
// methods are safe for concurrent use.
package service

import (
	"errors"
	"sync"

	"cpcli/internal/mgmt"
)

// ErrNotConnected is returned by every operation that needs a session when no
// successful Login has happened yet (or after Logout).
var ErrNotConnected = errors.New("não conectado — faça login primeiro")

// apiClient is the slice of *mgmt.Client the facade depends on. Expressing it
// as an interface lets tests drive Service with a fake, without a live server.
type apiClient interface {
	Call(command string, payload map[string]interface{}, waitForTask bool) (map[string]interface{}, error)
	List(command, detailsLevel, containerKey string, payload map[string]interface{}) ([]interface{}, error)
	Logout() error
}

// Service is the stateful facade a front end binds to.
type Service struct {
	mu     sync.Mutex
	client apiClient // nil until a successful Login
	info   SessionInfo
}

// New returns a disconnected Service. Call Login before any other operation.
func New() *Service { return &Service{} }

// LoginRequest carries the credentials/target for Login. Exactly one of
// Password or APIKey identifies the caller. JSON tags match the shape a
// front end sends.
type LoginRequest struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	APIKey   string `json:"apiKey"`
	Domain   string `json:"domain"`
	ReadOnly bool   `json:"readOnly"`
	Insecure bool   `json:"insecure"`
}

// SessionInfo is the connection state a front end renders (never includes any
// secret).
type SessionInfo struct {
	Connected  bool   `json:"connected"`
	Server     string `json:"server"`
	User       string `json:"user"`
	APIVersion string `json:"apiVersion"`
}

// Login authenticates and stores the resulting client for subsequent calls.
func (s *Service) Login(req LoginRequest) (SessionInfo, error) {
	port := req.Port
	if port == 0 {
		port = mgmt.DefaultPort
	}
	client, res, err := mgmt.Login(mgmt.LoginOptions{
		Server:   req.Server,
		Port:     port,
		User:     req.User,
		Password: req.Password,
		APIKey:   req.APIKey,
		Domain:   req.Domain,
		ReadOnly: req.ReadOnly,
		Insecure: req.Insecure,
	})
	if err != nil {
		return SessionInfo{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = client
	s.info = SessionInfo{Connected: true, Server: req.Server, User: req.User, APIVersion: res.APIVersion}
	return s.info, nil
}

// Status reports the current connection state.
func (s *Service) Status() SessionInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.info
}

// Logout closes the session on the server (if connected) and clears local
// state. It is a no-op when already disconnected.
func (s *Service) Logout() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return nil
	}
	err := s.client.Logout()
	s.client = nil
	s.info = SessionInfo{}
	return err
}

// conn returns the active client or ErrNotConnected.
func (s *Service) conn() (apiClient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return nil, ErrNotConnected
	}
	return s.client, nil
}

// ListHosts returns every host object, optionally narrowed by the Check Point
// text filter.
func (s *Service) ListHosts(filter string) ([]map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	payload := map[string]interface{}{}
	if filter != "" {
		payload["filter"] = filter
	}
	items, err := c.List("show-hosts", "standard", "objects", payload)
	if err != nil {
		return nil, err
	}
	return toMaps(items), nil
}

// GetHost returns the full detail of one host by name.
func (s *Service) GetHost(name string) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call("show-host", map[string]interface{}{"name": name, "details-level": "full"}, false)
}

// AddHost creates a host object. The change is pending until Publish.
func (s *Service) AddHost(name, ipAddress string) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call("add-host", map[string]interface{}{"name": name, "ip-address": ipAddress}, true)
}

// DeleteHost removes a host object by name. The change is pending until Publish.
func (s *Service) DeleteHost(name string) error {
	c, err := s.conn()
	if err != nil {
		return err
	}
	_, err = c.Call("delete-host", map[string]interface{}{"name": name}, true)
	return err
}

// Publish persists all pending changes made in this session.
func (s *Service) Publish() (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call("publish", map[string]interface{}{}, true)
}

// Discard drops all pending, unpublished changes.
func (s *Service) Discard() (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call("discard", map[string]interface{}{}, false)
}

func toMaps(items []interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		if m, ok := it.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out
}
