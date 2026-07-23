package mgmt

import (
	"errors"
	"strings"
	"testing"
	"time"

	api "github.com/CheckPointSW/cp-mgmt-api-go-sdk/APIFiles"
)

// recordedCall captures the arguments the fake caller received, so tests can
// assert how mgmt builds each request (the SDK's APIResponse.data is
// unexported and can't be populated from here, so the request side is what we
// can meaningfully verify).
type recordedCall struct {
	command     string
	payload     map[string]interface{}
	sid         string
	waitForTask bool
}

type fakeCaller struct {
	sid       string
	responses []api.APIResponse
	errs      []error
	block     chan struct{} // when non-nil, ApiCall blocks until it's closed
	calls     []recordedCall
}

func (f *fakeCaller) GetSessionID() string { return f.sid }

func (f *fakeCaller) ApiCall(command string, payload map[string]interface{}, sid string, waitForTask bool, useProxy bool, args ...string) (api.APIResponse, error) {
	idx := len(f.calls)
	f.calls = append(f.calls, recordedCall{command, payload, sid, waitForTask})
	if f.block != nil {
		<-f.block
	}
	var err error
	if idx < len(f.errs) {
		err = f.errs[idx]
	}
	if idx < len(f.responses) {
		return f.responses[idx], err
	}
	return api.APIResponse{Success: true}, err
}

func TestCallPassesCommandAndWait(t *testing.T) {
	f := &fakeCaller{sid: "SID-123", responses: []api.APIResponse{{Success: true}}}
	c := newClient(f)

	if _, err := c.Call("add-host", map[string]interface{}{"name": "web-01"}, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.calls) != 1 {
		t.Fatalf("expected exactly 1 ApiCall, got %d", len(f.calls))
	}
	got := f.calls[0]
	if got.command != "add-host" {
		t.Errorf("command = %q, want %q", got.command, "add-host")
	}
	if !got.waitForTask {
		t.Error("waitForTask = false, want true")
	}
	if got.sid != "SID-123" {
		t.Errorf("sid = %q, want the session's sid", got.sid)
	}
}

func TestCallMapsFailureToAPIError(t *testing.T) {
	f := &fakeCaller{sid: "S", responses: []api.APIResponse{{Success: false, ErrorMsg: "Unrecognized parameter [details-level]"}}}
	c := newClient(f)

	_, err := c.Call("show-session", map[string]interface{}{}, false)
	if err == nil {
		t.Fatal("expected an error for a failed API call")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *APIError", err)
	}
	if apiErr.Command != "show-session" {
		t.Errorf("APIError.Command = %q, want %q", apiErr.Command, "show-session")
	}
	if apiErr.Message != "Unrecognized parameter [details-level]" {
		t.Errorf("APIError.Message = %q, want the server message", apiErr.Message)
	}
}

func TestListEmptyResultIsNonNil(t *testing.T) {
	f := &fakeCaller{sid: "S", responses: []api.APIResponse{{Success: true}}}
	c := newClient(f)

	items, err := c.List("show-hosts", "standard", "objects", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items == nil {
		t.Fatal("List returned nil; empty results must be a non-nil slice so they serialize as [] not null")
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestListBuildsPaginationParams(t *testing.T) {
	f := &fakeCaller{sid: "S", responses: []api.APIResponse{{Success: true}}}
	c := newClient(f)

	if _, err := c.List("show-hosts", "full", "objects", map[string]interface{}{"filter": "web"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("expected 1 page request, got %d", len(f.calls))
	}
	p := f.calls[0].payload
	if p["limit"] != queryPageLimit {
		t.Errorf("limit = %v, want %d", p["limit"], queryPageLimit)
	}
	if p["offset"] != 0 {
		t.Errorf("offset = %v, want 0", p["offset"])
	}
	if p["details-level"] != "full" {
		t.Errorf("details-level = %v, want %q", p["details-level"], "full")
	}
	if p["filter"] != "web" {
		t.Errorf("caller payload not preserved: filter = %v, want %q", p["filter"], "web")
	}
}

func TestListPropagatesAPIError(t *testing.T) {
	f := &fakeCaller{sid: "S", responses: []api.APIResponse{{Success: false, ErrorMsg: "nope"}}}
	c := newClient(f)

	items, err := c.List("show-hosts", "standard", "objects", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected the failed page to surface an error")
	}
	if items != nil {
		t.Errorf("expected nil items on error, got %v", items)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *APIError", err)
	}
}

func TestCallTimesOut(t *testing.T) {
	f := &fakeCaller{sid: "S", block: make(chan struct{})}
	c := newClient(f)
	c.callTimeout = 20 * time.Millisecond

	_, err := c.Call("publish", map[string]interface{}{}, true)
	close(f.block) // release the blocked goroutine so it doesn't leak

	if err == nil {
		t.Fatal("expected a timeout error when the call never returns")
	}
	if !strings.Contains(err.Error(), "não respondeu") {
		t.Errorf("error = %q, want a timeout message", err.Error())
	}
}

func TestSessionIDPassesThrough(t *testing.T) {
	c := newClient(&fakeCaller{sid: "abc-123"})
	if c.SessionID() != "abc-123" {
		t.Errorf("SessionID() = %q, want %q", c.SessionID(), "abc-123")
	}
}
