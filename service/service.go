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
	"fmt"
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

// --- Generic objects ---------------------------------------------------------

// objectCommands is the set of Management API command names for one object
// family. Command names don't follow a single pluralization rule, so each is
// spelled out (mirrors the CLI's entitySpec).
type objectCommands struct {
	list, show, add, set, del string
}

var objectRegistry = map[string]objectCommands{
	"host":        {"show-hosts", "show-host", "add-host", "set-host", "delete-host"},
	"network":     {"show-networks", "show-network", "add-network", "set-network", "delete-network"},
	"group":       {"show-groups", "show-group", "add-group", "set-group", "delete-group"},
	"service-tcp": {"show-services-tcp", "show-service-tcp", "add-service-tcp", "set-service-tcp", "delete-service-tcp"},
	"service-udp": {"show-services-udp", "show-service-udp", "add-service-udp", "set-service-udp", "delete-service-udp"},
}

// ObjectKinds returns the supported object type keys, in display order.
func (s *Service) ObjectKinds() []string {
	return []string{"host", "network", "group", "service-tcp", "service-udp"}
}

func resolveKind(kind string) (objectCommands, error) {
	oc, ok := objectRegistry[kind]
	if !ok {
		return objectCommands{}, fmt.Errorf("tipo de objeto desconhecido: %q", kind)
	}
	return oc, nil
}

// ListObjects returns every object of the given kind, optionally narrowed by
// the Check Point text filter.
func (s *Service) ListObjects(kind, filter string) ([]map[string]interface{}, error) {
	oc, err := resolveKind(kind)
	if err != nil {
		return nil, err
	}
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	payload := map[string]interface{}{}
	if filter != "" {
		payload["filter"] = filter
	}
	items, err := c.List(oc.list, "standard", "objects", payload)
	if err != nil {
		return nil, err
	}
	return toMaps(items), nil
}

// GetObject returns the full detail of one object by name.
func (s *Service) GetObject(kind, name string) (map[string]interface{}, error) {
	oc, err := resolveKind(kind)
	if err != nil {
		return nil, err
	}
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call(oc.show, map[string]interface{}{"name": name, "details-level": "full"}, false)
}

// AddObject creates an object of the given kind from arbitrary API fields
// (e.g. {"name","ip-address"} for a host, {"name","subnet4","mask-length4"}
// for a network). The change is pending until Publish.
func (s *Service) AddObject(kind string, fields map[string]interface{}) (map[string]interface{}, error) {
	oc, err := resolveKind(kind)
	if err != nil {
		return nil, err
	}
	if name, _ := fields["name"].(string); name == "" {
		return nil, errors.New("o campo 'name' é obrigatório")
	}
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call(oc.add, fields, true)
}

// SetObject updates an existing object of the given kind. fields must identify
// the object (by "name" or "uid") plus the fields to change.
func (s *Service) SetObject(kind string, fields map[string]interface{}) (map[string]interface{}, error) {
	oc, err := resolveKind(kind)
	if err != nil {
		return nil, err
	}
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call(oc.set, fields, true)
}

// DeleteObject removes an object of the given kind by name. Pending until Publish.
func (s *Service) DeleteObject(kind, name string) error {
	oc, err := resolveKind(kind)
	if err != nil {
		return err
	}
	c, err := s.conn()
	if err != nil {
		return err
	}
	_, err = c.Call(oc.del, map[string]interface{}{"name": name}, true)
	return err
}

// --- Access control / NAT (read) --------------------------------------------

// ListAccessLayers returns the Access Control layers.
func (s *Service) ListAccessLayers() ([]map[string]interface{}, error) {
	return s.listSimple("show-access-layers", "objects", map[string]interface{}{})
}

// ListAccessRulebase returns the rules of one Access Control layer.
func (s *Service) ListAccessRulebase(layer string) ([]map[string]interface{}, error) {
	return s.listSimple("show-access-rulebase", "rulebase", map[string]interface{}{"name": layer})
}

// ListNatRulebase returns the NAT rules of one policy package.
func (s *Service) ListNatRulebase(pkg string) ([]map[string]interface{}, error) {
	return s.listSimple("show-nat-rulebase", "rulebase", map[string]interface{}{"package": pkg})
}

// --- Policy / gateways -------------------------------------------------------

// ListPackages returns the policy packages.
func (s *Service) ListPackages() ([]map[string]interface{}, error) {
	return s.listSimple("show-packages", "packages", map[string]interface{}{})
}

// ListGateways returns gateways and servers known to the management.
func (s *Service) ListGateways() ([]map[string]interface{}, error) {
	return s.listSimple("show-gateways-and-servers", "objects", map[string]interface{}{})
}

// InstallPolicy installs a policy package on the given gateway targets and
// waits for the task to finish.
func (s *Service) InstallPolicy(pkg string, targets []string) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call("install-policy", map[string]interface{}{"policy-package": pkg, "targets": targets}, true)
}

// VerifyPolicy checks a policy package for errors before installing.
func (s *Service) VerifyPolicy(pkg string) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call("verify-policy", map[string]interface{}{"policy-package": pkg}, true)
}

// --- Session changes ---------------------------------------------------------

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

// listSimple is the shared body of the read-only list helpers.
func (s *Service) listSimple(command, containerKey string, payload map[string]interface{}) ([]map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	items, err := c.List(command, "standard", containerKey, payload)
	if err != nil {
		return nil, err
	}
	return toMaps(items), nil
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
