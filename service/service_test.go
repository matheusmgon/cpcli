package service

import (
	"errors"
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
	if _, err := s.ListHosts(""); !errors.Is(err, ErrNotConnected) {
		t.Errorf("ListHosts on disconnected = %v, want ErrNotConnected", err)
	}
	if _, err := s.GetHost("x"); !errors.Is(err, ErrNotConnected) {
		t.Errorf("GetHost on disconnected = %v, want ErrNotConnected", err)
	}
	if _, err := s.AddHost("x", "1.2.3.4"); !errors.Is(err, ErrNotConnected) {
		t.Errorf("AddHost on disconnected = %v, want ErrNotConnected", err)
	}
	if err := s.DeleteHost("x"); !errors.Is(err, ErrNotConnected) {
		t.Errorf("DeleteHost on disconnected = %v, want ErrNotConnected", err)
	}
	if _, err := s.Publish(); !errors.Is(err, ErrNotConnected) {
		t.Errorf("Publish on disconnected = %v, want ErrNotConnected", err)
	}
}

func TestListHostsPassesQueryAndMaps(t *testing.T) {
	f := &fakeClient{listItems: []interface{}{
		map[string]interface{}{"name": "web-01"},
		"not-a-map", // should be skipped by toMaps
		map[string]interface{}{"name": "web-02"},
	}}
	s := connected(f)

	hosts, err := s.ListHosts("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 mapped hosts, got %d", len(hosts))
	}
	if f.lastListCommand != "show-hosts" || f.lastListContainer != "objects" || f.lastListDetails != "standard" {
		t.Errorf("List called with (%q,%q,%q), want (show-hosts,standard,objects)", f.lastListCommand, f.lastListDetails, f.lastListContainer)
	}
	if f.lastListPayload["filter"] != "web" {
		t.Errorf("filter = %v, want %q", f.lastListPayload["filter"], "web")
	}
}

func TestListHostsOmitsEmptyFilter(t *testing.T) {
	f := &fakeClient{}
	s := connected(f)
	if _, err := s.ListHosts(""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := f.lastListPayload["filter"]; ok {
		t.Error("empty filter should not be sent as a query param")
	}
}

func TestAddHostBuildsPayload(t *testing.T) {
	f := &fakeClient{callData: map[string]interface{}{"name": "web-01"}}
	s := connected(f)

	if _, err := s.AddHost("web-01", "10.0.0.10"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.lastCallCommand != "add-host" {
		t.Errorf("command = %q, want add-host", f.lastCallCommand)
	}
	if !f.lastCallWait {
		t.Error("AddHost should wait for the task")
	}
	if f.lastCallPayload["name"] != "web-01" || f.lastCallPayload["ip-address"] != "10.0.0.10" {
		t.Errorf("payload = %v, want name/ip-address set", f.lastCallPayload)
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
	if _, err := s.ListHosts(""); !errors.Is(err, ErrNotConnected) {
		t.Error("operations after Logout should require login again")
	}
}

func TestLogoutWhenDisconnectedIsNoOp(t *testing.T) {
	if err := New().Logout(); err != nil {
		t.Errorf("Logout on a disconnected service = %v, want nil", err)
	}
}
