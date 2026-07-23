package service

import (
	"errors"
	"strings"
	"testing"
)

type fakeClient struct {
	listItems []interface{}
	listErr   error
	callData  map[string]interface{}
	callErr   error
	logoutErr error
	loggedOut bool

	lastCallCommand string
	lastCallPayload map[string]interface{}
	lastCallWait    bool

	lastListCommand   string
	lastListDetails   string
	lastListContainer string
	lastListPayload   map[string]interface{}
}

func (f *fakeClient) Call(command string, payload map[string]interface{}, waitForTask bool) (map[string]interface{}, error) {
	f.lastCallCommand = command
	f.lastCallPayload = payload
	f.lastCallWait = waitForTask
	return f.callData, f.callErr
}

func (f *fakeClient) List(command, detailsLevel, containerKey string, payload map[string]interface{}) ([]interface{}, error) {
	f.lastListCommand = command
	f.lastListDetails = detailsLevel
	f.lastListContainer = containerKey
	f.lastListPayload = payload
	return f.listItems, f.listErr
}

func (f *fakeClient) Logout() error {
	f.loggedOut = true
	return f.logoutErr
}

// connected returns a Service wired to the given fake, as if Login had
// succeeded.
func connected(f *fakeClient) *Service {
	return &Service{client: f, info: SessionInfo{Connected: true, Server: "srv", User: "admin"}}
}

func TestOperationsRequireLogin(t *testing.T) {
	s := New()
	if _, err := s.ListObjects("host", ""); !errors.Is(err, ErrNotConnected) {
		t.Errorf("ListObjects on disconnected = %v, want ErrNotConnected", err)
	}
	if _, err := s.GetObject("host", "x"); !errors.Is(err, ErrNotConnected) {
		t.Errorf("GetObject on disconnected = %v, want ErrNotConnected", err)
	}
	if err := s.DeleteObject("host", "x"); !errors.Is(err, ErrNotConnected) {
		t.Errorf("DeleteObject on disconnected = %v, want ErrNotConnected", err)
	}
	if _, err := s.ListAccessRulebase("Network"); !errors.Is(err, ErrNotConnected) {
		t.Errorf("ListAccessRulebase on disconnected = %v, want ErrNotConnected", err)
	}
	if _, err := s.Publish(); !errors.Is(err, ErrNotConnected) {
		t.Errorf("Publish on disconnected = %v, want ErrNotConnected", err)
	}
}

func TestUnknownKindIsRejected(t *testing.T) {
	f := &fakeClient{}
	s := connected(f)
	_, err := s.ListObjects("dragon", "")
	if err == nil || !strings.Contains(err.Error(), "desconhecido") {
		t.Fatalf("ListObjects on unknown kind = %v, want an 'unknown kind' error", err)
	}
	if f.lastListCommand != "" {
		t.Error("an unknown kind must not reach the API")
	}
}

func TestListObjectsPassesQueryAndMaps(t *testing.T) {
	f := &fakeClient{listItems: []interface{}{
		map[string]interface{}{"name": "web-01"},
		"not-a-map", // skipped by toMaps
		map[string]interface{}{"name": "web-02"},
	}}
	s := connected(f)

	rows, err := s.ListObjects("host", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 mapped rows, got %d", len(rows))
	}
	if f.lastListCommand != "show-hosts" || f.lastListContainer != "objects" {
		t.Errorf("List called (%q,%q), want (show-hosts,objects)", f.lastListCommand, f.lastListContainer)
	}
	if f.lastListPayload["filter"] != "web" {
		t.Errorf("filter = %v, want %q", f.lastListPayload["filter"], "web")
	}
}

func TestAddObjectBuildsPayloadAndWaits(t *testing.T) {
	f := &fakeClient{callData: map[string]interface{}{"name": "web-01"}}
	s := connected(f)

	fields := map[string]interface{}{"name": "web-01", "ip-address": "10.0.0.10"}
	if _, err := s.AddObject("host", fields); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.lastCallCommand != "add-host" {
		t.Errorf("command = %q, want add-host", f.lastCallCommand)
	}
	if !f.lastCallWait {
		t.Error("AddObject should wait for the task")
	}
	if f.lastCallPayload["ip-address"] != "10.0.0.10" {
		t.Errorf("payload not forwarded: %v", f.lastCallPayload)
	}
}

func TestAddObjectRequiresName(t *testing.T) {
	f := &fakeClient{}
	s := connected(f)
	if _, err := s.AddObject("host", map[string]interface{}{"ip-address": "1.2.3.4"}); err == nil {
		t.Fatal("AddObject without a name should error")
	}
	if f.lastCallCommand != "" {
		t.Error("a nameless add must not reach the API")
	}
}

func TestDeleteObjectUsesKindCommand(t *testing.T) {
	f := &fakeClient{}
	s := connected(f)
	if err := s.DeleteObject("network", "lan"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.lastCallCommand != "delete-network" {
		t.Errorf("command = %q, want delete-network", f.lastCallCommand)
	}
	if f.lastCallPayload["name"] != "lan" {
		t.Errorf("payload name = %v, want lan", f.lastCallPayload["name"])
	}
}

func TestListAccessRulebaseTargetsLayer(t *testing.T) {
	f := &fakeClient{}
	s := connected(f)
	if _, err := s.ListAccessRulebase("Network"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.lastListCommand != "show-access-rulebase" || f.lastListContainer != "rulebase" {
		t.Errorf("List called (%q,%q), want (show-access-rulebase,rulebase)", f.lastListCommand, f.lastListContainer)
	}
	if f.lastListPayload["name"] != "Network" {
		t.Errorf("layer = %v, want Network", f.lastListPayload["name"])
	}
}

func TestInstallPolicyBuildsPayload(t *testing.T) {
	f := &fakeClient{}
	s := connected(f)
	if _, err := s.InstallPolicy("Standard", []string{"gw-lab"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.lastCallCommand != "install-policy" || !f.lastCallWait {
		t.Errorf("install-policy call = (%q, wait=%v), want (install-policy, true)", f.lastCallCommand, f.lastCallWait)
	}
	if f.lastCallPayload["policy-package"] != "Standard" {
		t.Errorf("policy-package = %v, want Standard", f.lastCallPayload["policy-package"])
	}
}

func TestObjectKindsOrder(t *testing.T) {
	got := New().ObjectKinds()
	want := []string{"host", "network", "group", "service-tcp", "service-udp"}
	if len(got) != len(want) {
		t.Fatalf("ObjectKinds len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ObjectKinds[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLogoutClearsState(t *testing.T) {
	f := &fakeClient{}
	s := connected(f)

	if err := s.Logout(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.loggedOut {
		t.Error("Logout should call the server-side logout")
	}
	if s.Status().Connected {
		t.Error("Status should report disconnected after Logout")
	}
	if _, err := s.ListObjects("host", ""); !errors.Is(err, ErrNotConnected) {
		t.Error("operations after Logout should require login again")
	}
}

func TestLogoutWhenDisconnectedIsNoOp(t *testing.T) {
	if err := New().Logout(); err != nil {
		t.Errorf("Logout on a disconnected service = %v, want nil", err)
	}
}
