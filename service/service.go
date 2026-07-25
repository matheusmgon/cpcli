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
	ListRulebase(command, detailsLevel, containerKey string, payload map[string]interface{}) ([]interface{}, error)
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
	"host":          {"show-hosts", "show-host", "add-host", "set-host", "delete-host"},
	"network":       {"show-networks", "show-network", "add-network", "set-network", "delete-network"},
	"group":         {"show-groups", "show-group", "add-group", "set-group", "delete-group"},
	"service-tcp":   {"show-services-tcp", "show-service-tcp", "add-service-tcp", "set-service-tcp", "delete-service-tcp"},
	"service-udp":   {"show-services-udp", "show-service-udp", "add-service-udp", "set-service-udp", "delete-service-udp"},
	"address-range": {"show-address-ranges", "show-address-range", "add-address-range", "set-address-range", "delete-address-range"},
	"service-group": {"show-service-groups", "show-service-group", "add-service-group", "set-service-group", "delete-service-group"},
	"access-role":   {"show-access-roles", "show-access-role", "add-access-role", "set-access-role", "delete-access-role"},
}

// ObjectKinds returns the supported object type keys, in display order.
func (s *Service) ObjectKinds() []string {
	return []string{"host", "network", "group", "service-tcp", "service-udp", "address-range", "service-group", "access-role"}
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

// --- Search --------------------------------------------------------------------

// SearchObjects searches for objects across every type (or narrowed to
// objType) matching filter — powers the desktop UI's object picker used to
// find source/destination/service/etc. values by name instead of typing
// them blind. Deliberately bounded to a single page (unlike listSimple,
// which aggregates every page): a picker only needs "good enough matches"
// for the current search text, not an exhaustive list — if what the caller
// wants isn't in the first batch, they refine the filter, the same way
// SmartConsole's own object-picker dropdown behaves.
func (s *Service) SearchObjects(filter, objType string) ([]map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	payload := map[string]interface{}{"limit": 50}
	if filter != "" {
		payload["filter"] = filter
	}
	if objType != "" {
		payload["type"] = objType
	}
	data, err := c.Call("show-objects", payload, false)
	if err != nil {
		return nil, err
	}
	items, _ := data["objects"].([]interface{})
	return toMaps(items), nil
}

// --- Access control / NAT (read) --------------------------------------------

// ListAccessLayers returns the Access Control layers.
func (s *Service) ListAccessLayers() ([]map[string]interface{}, error) {
	return s.listSimple("show-access-layers", "access-layers", map[string]interface{}{})
}

// ListAccessRulebase returns the rules of one Access Control layer.
func (s *Service) ListAccessRulebase(layer string) ([]map[string]interface{}, error) {
	return s.listRulebase("show-access-rulebase", "rulebase", map[string]interface{}{"name": layer})
}

// ListNatRulebase returns the NAT rules of one policy package.
func (s *Service) ListNatRulebase(pkg string) ([]map[string]interface{}, error) {
	return s.listRulebase("show-nat-rulebase", "rulebase", map[string]interface{}{"package": pkg})
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

// --- Access rules (CRUD) -----------------------------------------------------

// AddAccessRule creates an access rule in the given layer. fields carries the
// API body (name, action, source, destination, service, position, …).
func (s *Service) AddAccessRule(layer string, fields map[string]interface{}) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	p := cloneFields(fields)
	p["layer"] = layer
	return c.Call("add-access-rule", p, true)
}

// SetAccessRule updates the rule identified by uid in the given layer.
func (s *Service) SetAccessRule(layer, uid string, fields map[string]interface{}) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	p := cloneFields(fields)
	p["layer"] = layer
	p["uid"] = uid
	return c.Call("set-access-rule", p, true)
}

// DeleteAccessRule removes the rule identified by uid in the given layer.
func (s *Service) DeleteAccessRule(layer, uid string) error {
	c, err := s.conn()
	if err != nil {
		return err
	}
	_, err = c.Call("delete-access-rule", map[string]interface{}{"layer": layer, "uid": uid}, true)
	return err
}

// --- NAT rules (CRUD) --------------------------------------------------------

// AddNatRule creates a manual NAT rule in the given policy package.
func (s *Service) AddNatRule(pkg string, fields map[string]interface{}) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	p := cloneFields(fields)
	p["package"] = pkg
	return c.Call("add-nat-rule", p, true)
}

// SetNatRule updates the NAT rule identified by uid in the given package.
func (s *Service) SetNatRule(pkg, uid string, fields map[string]interface{}) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	p := cloneFields(fields)
	p["package"] = pkg
	p["uid"] = uid
	return c.Call("set-nat-rule", p, true)
}

// DeleteNatRule removes the NAT rule identified by uid in the given package.
func (s *Service) DeleteNatRule(pkg, uid string) error {
	c, err := s.conn()
	if err != nil {
		return err
	}
	_, err = c.Call("delete-nat-rule", map[string]interface{}{"package": pkg, "uid": uid}, true)
	return err
}

// --- Threat Prevention (rules + profiles) -------------------------------------

// ListThreatLayers returns the Threat Prevention layers.
func (s *Service) ListThreatLayers() ([]map[string]interface{}, error) {
	return s.listSimple("show-threat-layers", "threat-layers", map[string]interface{}{})
}

// ListThreatRulebase returns the rules of one Threat Prevention layer.
func (s *Service) ListThreatRulebase(layer string) ([]map[string]interface{}, error) {
	return s.listRulebase("show-threat-rulebase", "rulebase", map[string]interface{}{"name": layer})
}

// AddThreatRule creates a rule in the given Threat Prevention layer. fields
// carries the API body (name is required by the API, plus action,
// protected-scope, position, …).
func (s *Service) AddThreatRule(layer string, fields map[string]interface{}) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	p := cloneFields(fields)
	p["layer"] = layer
	return c.Call("add-threat-rule", p, true)
}

// SetThreatRule updates the rule identified by uid in the given layer.
// show/set/delete-threat-rule all require "layer" even when the rule is
// identified by uid — confirmed against a live Management Server.
func (s *Service) SetThreatRule(layer, uid string, fields map[string]interface{}) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	p := cloneFields(fields)
	p["layer"] = layer
	p["uid"] = uid
	return c.Call("set-threat-rule", p, true)
}

// DeleteThreatRule removes the rule identified by uid in the given layer.
func (s *Service) DeleteThreatRule(layer, uid string) error {
	c, err := s.conn()
	if err != nil {
		return err
	}
	_, err = c.Call("delete-threat-rule", map[string]interface{}{"layer": layer, "uid": uid}, true)
	return err
}

// ListThreatProfiles returns the Threat Prevention profiles.
func (s *Service) ListThreatProfiles() ([]map[string]interface{}, error) {
	return s.listSimple("show-threat-profiles", "objects", map[string]interface{}{})
}

// AddThreatProfile creates a Threat Prevention profile.
func (s *Service) AddThreatProfile(fields map[string]interface{}) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call("add-threat-profile", cloneFields(fields), true)
}

// SetThreatProfile updates an existing Threat Prevention profile. fields
// must identify the profile (by "name" or "uid") plus the fields to change.
func (s *Service) SetThreatProfile(fields map[string]interface{}) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call("set-threat-profile", cloneFields(fields), true)
}

// DeleteThreatProfile removes a Threat Prevention profile by name.
func (s *Service) DeleteThreatProfile(name string) error {
	c, err := s.conn()
	if err != nil {
		return err
	}
	_, err = c.Call("delete-threat-profile", map[string]interface{}{"name": name}, true)
	return err
}

// --- HTTPS Inspection ----------------------------------------------------------

// ListHttpsLayers returns the HTTPS Inspection layers.
func (s *Service) ListHttpsLayers() ([]map[string]interface{}, error) {
	return s.listSimple("show-https-layers", "https-layers", map[string]interface{}{})
}

// ListHttpsRulebase returns the rules of one HTTPS Inspection layer.
func (s *Service) ListHttpsRulebase(layer string) ([]map[string]interface{}, error) {
	return s.listRulebase("show-https-rulebase", "rulebase", map[string]interface{}{"name": layer})
}

// AddHttpsRule creates a rule in the given HTTPS Inspection layer.
func (s *Service) AddHttpsRule(layer string, fields map[string]interface{}) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	p := cloneFields(fields)
	p["layer"] = layer
	return c.Call("add-https-rule", p, true)
}

// SetHttpsRule updates the rule identified by uid in the given layer.
// show/set/delete-https-rule all require "layer" even when the rule is
// identified by uid — confirmed against a live Management Server.
func (s *Service) SetHttpsRule(layer, uid string, fields map[string]interface{}) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	p := cloneFields(fields)
	p["layer"] = layer
	p["uid"] = uid
	return c.Call("set-https-rule", p, true)
}

// DeleteHttpsRule removes the rule identified by uid in the given layer.
func (s *Service) DeleteHttpsRule(layer, uid string) error {
	c, err := s.conn()
	if err != nil {
		return err
	}
	_, err = c.Call("delete-https-rule", map[string]interface{}{"layer": layer, "uid": uid}, true)
	return err
}

// --- Gateway interfaces (topology / anti-spoofing) ----------------------------

// ListGatewayInterfaces returns the interfaces of a simple-gateway (IP,
// anti-spoofing, topology settings).
func (s *Service) ListGatewayInterfaces(gateway string) ([]map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	data, err := c.Call("show-simple-gateway", map[string]interface{}{
		"name":          gateway,
		"details-level": "full",
	}, false)
	if err != nil {
		return nil, err
	}
	ifaces, _ := data["interfaces"].([]interface{})
	return toMaps(ifaces), nil
}

// SetGatewayInterface changes fields (e.g. "anti-spoofing") on one named
// interface of a gateway. set-simple-gateway replaces the WHOLE
// "interfaces" array rather than patching a single entry — sending back
// only the changed interface silently deletes every other one on the
// gateway (confirmed against a live Management Server). So this reads the
// current interfaces first and sends the complete, merged array back via
// mgmt.MergeGatewayInterface (shared with the CLI's "gateway interface set").
func (s *Service) SetGatewayInterface(gateway, ifaceName string, fields map[string]interface{}) (map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	data, err := c.Call("show-simple-gateway", map[string]interface{}{
		"name":          gateway,
		"details-level": "full",
	}, false)
	if err != nil {
		return nil, err
	}
	ifaces, _ := data["interfaces"].([]interface{})
	updated, found := mgmt.MergeGatewayInterface(ifaces, ifaceName, fields)
	if !found {
		return nil, fmt.Errorf("interface %q não encontrada no gateway %q", ifaceName, gateway)
	}
	return c.Call("set-simple-gateway", map[string]interface{}{
		"name":       gateway,
		"interfaces": updated,
	}, true)
}

// --- VPN communities ---------------------------------------------------------

var vpnRegistry = map[string]objectCommands{
	"meshed": {"show-vpn-communities-meshed", "show-vpn-community-meshed", "add-vpn-community-meshed", "set-vpn-community-meshed", "delete-vpn-community-meshed"},
	"star":   {"show-vpn-communities-star", "show-vpn-community-star", "add-vpn-community-star", "set-vpn-community-star", "delete-vpn-community-star"},
}

func resolveVPN(kind string) (objectCommands, error) {
	oc, ok := vpnRegistry[kind]
	if !ok {
		return objectCommands{}, fmt.Errorf("tipo de comunidade VPN desconhecido: %q", kind)
	}
	return oc, nil
}

// VpnKinds returns the supported VPN community topologies, in display order.
func (s *Service) VpnKinds() []string { return []string{"star", "meshed"} }

// ListVpnCommunities returns the VPN communities of the given topology.
func (s *Service) ListVpnCommunities(kind string) ([]map[string]interface{}, error) {
	oc, err := resolveVPN(kind)
	if err != nil {
		return nil, err
	}
	return s.listSimple(oc.list, "objects", map[string]interface{}{})
}

// GetVpnCommunity returns the full detail of one VPN community by name.
func (s *Service) GetVpnCommunity(kind, name string) (map[string]interface{}, error) {
	oc, err := resolveVPN(kind)
	if err != nil {
		return nil, err
	}
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call(oc.show, map[string]interface{}{"name": name}, false)
}

// AddVpnCommunity creates a VPN community of the given topology from API fields
// (e.g. {"name","gateways"} for meshed, {"name","center-gateways",
// "satellite-gateways"} for star).
func (s *Service) AddVpnCommunity(kind string, fields map[string]interface{}) (map[string]interface{}, error) {
	oc, err := resolveVPN(kind)
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

// SetVpnCommunity updates an existing VPN community. fields must include the
// identifying "name" plus the fields to change.
func (s *Service) SetVpnCommunity(kind string, fields map[string]interface{}) (map[string]interface{}, error) {
	oc, err := resolveVPN(kind)
	if err != nil {
		return nil, err
	}
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	return c.Call(oc.set, fields, true)
}

// DeleteVpnCommunity removes a VPN community by name.
func (s *Service) DeleteVpnCommunity(kind, name string) error {
	oc, err := resolveVPN(kind)
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

// listRulebase is listSimple's counterpart for rulebase commands: it
// resolves action/source/destination/service/etc. UIDs to display names
// (via mgmt.Client.ListRulebase's "objects-dictionary" lookup) instead of
// returning raw UIDs for those fields — confirmed against a live server
// that rulebase rows never include them pre-resolved, even at
// details-level=full.
func (s *Service) listRulebase(command, containerKey string, payload map[string]interface{}) ([]map[string]interface{}, error) {
	c, err := s.conn()
	if err != nil {
		return nil, err
	}
	items, err := c.ListRulebase(command, "standard", containerKey, payload)
	if err != nil {
		return nil, err
	}
	return toMaps(items), nil
}

// cloneFields returns a shallow copy of f so a method can add identifying keys
// (layer, package, uid, …) without mutating the caller's map.
func cloneFields(f map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(f)+2)
	for k, v := range f {
		out[k] = v
	}
	return out
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
