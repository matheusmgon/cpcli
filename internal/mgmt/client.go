// Package mgmt is the transport-and-session core for talking to the Check
// Point Management API. It owns the SDK client lifecycle — login, session
// reuse, bounded calls, pagination and logout — and returns plain Go values,
// so that any front end (the cpcli CLI today, a desktop/browser UI next) can
// drive the same management operations without duplicating the SDK plumbing.
package mgmt

import (
	"fmt"
	"strings"
	"time"

	api "github.com/CheckPointSW/cp-mgmt-api-go-sdk/APIFiles"
)

// DefaultPort is the Management API's default TCP port, re-exported so front
// ends don't need to import the SDK just for a flag default.
const DefaultPort = api.DefaultPort

// DefaultCallTimeout bounds how long a single API call may run. See the note
// on the busy-loop guard in callWithTimeout for why this exists.
const DefaultCallTimeout = 5 * time.Minute

// httpTimeout is the per-HTTP-request timeout passed to the Check Point SDK
// via api.APIClientArgs. The SDK's own default (api.TimeOut) is 10s, which
// is too short for slow operations against a laden Management Server —
// install-policy in particular occasionally has a single show-task poll
// stall for more than 10s and fails client-side even though the task
// itself is progressing normally. 5 minutes matches DefaultCallTimeout so
// the two ceilings agree; the SDK's waitForTask loop returns as soon as
// the task reports done, so raising the ceiling doesn't slow anything.
const httpTimeout = 5 * time.Minute

// queryPageLimit mirrors the fixed page size the SDK's own pagination helper
// uses.
const queryPageLimit = 50

// caller is the slice of *api.ApiClient that Client depends on. Expressing it
// as an interface (instead of using *api.ApiClient directly) lets tests drive
// pagination and error handling with a fake, without a live server.
type caller interface {
	ApiCall(command string, payload map[string]interface{}, sid string, waitForTask bool, useProxy bool, args ...string) (api.APIResponse, error)
	GetSessionID() string
}

// Client is an authenticated handle to a Management Server. It is safe to
// reuse across many calls — e.g. for the lifetime of a UI session.
type Client struct {
	api         caller
	callTimeout time.Duration
}

// APIError is returned when the Management API accepts a call but reports a
// failure (res.Success == false). It preserves the server's message verbatim
// so a front end can surface it as-is.
type APIError struct {
	Command string
	Message string
}

func (e *APIError) Error() string { return e.Message }

func newClient(c caller) *Client {
	return &Client{api: c, callTimeout: DefaultCallTimeout}
}

// Conn describes how to reach a Management Server with an already-issued
// session id (sid).
type Conn struct {
	Server     string
	Port       int
	Sid        string
	APIVersion string
	Insecure   bool
}

// Connect builds a Client for an existing, already-authenticated session.
//
// AcceptServerCertificate (the SDK's unsafeAutoAccept argument) is
// intentionally left false: the SDK's auto-accept path persists an empty
// fingerprint on first use instead of the real one (it saves the fingerprint
// field before it is populated), silently defeating trust-on-first-use
// pinning. Leaving it false routes through the SDK's compare-and-prompt path,
// which pins the real fingerprint. IgnoreServerCertificate carries the
// login-time Insecure choice so an insecure login doesn't turn into a
// fingerprint prompt (or a hard failure, in non-interactive use) on every
// later command.
func Connect(c Conn) *Client {
	args := api.APIClientArgs(c.Port, "", c.Sid, c.Server, "", -1, c.APIVersion, c.Insecure, false, "", api.WebContext, httpTimeout, api.SleepTime, "cpcli", "", -1)
	return newClient(api.APIClient(args))
}

// LoginOptions holds the parameters for authenticating to a Management Server.
// Exactly one of Password or APIKey identifies the caller.
type LoginOptions struct {
	Server          string
	Port            int
	User            string
	Password        string
	APIKey          string
	Domain          string
	ReadOnly        bool
	ContinueSession bool
	Insecure        bool
}

// LoginResult carries the session details a caller should persist to reuse the
// session later (cpcli writes these to disk; a UI may keep them in memory).
type LoginResult struct {
	Sid        string
	APIVersion string
}

// Login authenticates and returns a ready-to-use Client together with the
// session details to persist.
func Login(o LoginOptions) (*Client, *LoginResult, error) {
	args := api.APIClientArgs(o.Port, "", "", o.Server, "", -1, "", o.Insecure, false, "", api.WebContext, httpTimeout, api.SleepTime, "cpcli", "", -1)
	sdk := api.APIClient(args)

	var res api.APIResponse
	var err error
	if o.APIKey != "" {
		res, err = sdk.ApiLoginWithApiKey(o.APIKey, o.ContinueSession, o.Domain, o.ReadOnly, nil)
	} else {
		res, err = sdk.ApiLogin(o.User, o.Password, o.ContinueSession, o.Domain, o.ReadOnly, nil)
	}
	if err != nil {
		return nil, nil, err
	}
	if !res.Success {
		return nil, nil, &APIError{Command: "login", Message: res.ErrorMsg}
	}
	return newClient(sdk), &LoginResult{
		Sid:        sdk.GetSessionID(),
		APIVersion: stringField(res.GetData(), "api-server-version"),
	}, nil
}

// SessionID returns the current session id (sid).
func (c *Client) SessionID() string { return c.api.GetSessionID() }

// Call runs one Management API command and returns its data payload. When
// waitForTask is true, an async command (publish, install-policy, ...) is
// polled to completion by the SDK before returning.
func (c *Client) Call(command string, payload map[string]interface{}, waitForTask bool) (map[string]interface{}, error) {
	res, err := c.callWithTimeout(command, payload, waitForTask)
	if err != nil {
		return nil, err
	}
	if !res.Success {
		return nil, &APIError{Command: command, Message: formatFailure(res.ErrorMsg, res.GetData())}
	}
	return res.GetData(), nil
}

// formatFailure builds a useful error message for a failed API response.
// The SDK's ErrorMsg is often empty or minimal for task-based commands
// (e.g. install-policy leaves it as just "Failed to execute API call\n
// Task: X\nMessage: " when task-details don't carry a `fault-message` —
// many task types put the real detail in `statusDescription`,
// `description`, or `message` instead). Fall back to walking the raw
// response data so the user sees the actual reason instead of a blank
// toast.
func formatFailure(sdkErrMsg string, data map[string]interface{}) string {
	extras := extractFailureDetails(data)
	msg := strings.TrimSpace(sdkErrMsg)
	if len(extras) == 0 {
		if msg != "" {
			return msg
		}
		return "operação falhou (o servidor não retornou detalhes)"
	}
	joined := strings.Join(extras, "\n")
	if msg == "" || msg == "Failed to execute API call" {
		return joined
	}
	return msg + "\n" + joined
}

// extractFailureDetails walks a Management API failure response and
// returns every string-valued detail field it recognizes (tasks/task-
// details, top-level errors/warnings/blocking-errors, and common message/
// description fields). Empty strings are dropped.
func extractFailureDetails(data map[string]interface{}) []string {
	var out []string
	push := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	stringField := func(m map[string]interface{}, keys ...string) {
		for _, k := range keys {
			if v, ok := m[k].(string); ok {
				push(v)
			}
		}
	}

	stringField(data, "message", "description", "statusDescription", "error")

	pushMessages := func(key string) {
		items, ok := data[key].([]interface{})
		if !ok {
			return
		}
		for _, it := range items {
			if m, ok := it.(map[string]interface{}); ok {
				stringField(m, "message", "description")
			} else if s, ok := it.(string); ok {
				push(s)
			}
		}
	}
	pushMessages("errors")
	pushMessages("warnings")
	pushMessages("blocking-errors")

	tasks, ok := data["tasks"].([]interface{})
	if !ok {
		return out
	}
	for _, t := range tasks {
		task, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := task["task-name"].(string)
		status, _ := task["status"].(string)
		if name != "" || status != "" {
			push(fmt.Sprintf("Task: %s [%s]", name, status))
		}
		details, _ := task["task-details"].([]interface{})
		for _, d := range details {
			dm, ok := d.(map[string]interface{})
			if !ok {
				continue
			}
			stringField(dm, "fault-message", "statusDescription", "message", "description")
			if stages, ok := dm["stagesInfo"].([]interface{}); ok {
				for _, st := range stages {
					sm, ok := st.(map[string]interface{})
					if !ok {
						continue
					}
					stringField(sm, "statusDescription", "description", "message")
					// stagesInfo entries carry a nested `messages` array
					// where each item is `{message, type}` (type = "err" |
					// "warn"). This is where install-policy hides the real
					// reason it failed (topology unset, license expired,
					// rulebase generation errors, etc.) — the SDK's
					// generic error formatter never looks here.
					if msgs, ok := sm["messages"].([]interface{}); ok {
						for _, m := range msgs {
							mm, ok := m.(map[string]interface{})
							if !ok {
								continue
							}
							kind, _ := mm["type"].(string)
							text, _ := mm["message"].(string)
							if text == "" {
								continue
							}
							if kind != "" {
								push(fmt.Sprintf("[%s] %s", kind, text))
							} else {
								push(text)
							}
						}
					}
				}
			}
		}
	}
	return out
}

// List aggregates every page of a paginated "show-*s" command into one slice.
// An empty result set is returned as a non-nil empty slice, so callers that
// serialize it get "[]" rather than "null".
//
// It deliberately does not use the SDK's ApiQuery/genApiQuery helpers: those
// call os.Exit(1) directly when any page after the first one fails (e.g. the
// session expires partway through a large listing), which would kill the whole
// host process instead of returning a normal error.
func (c *Client) List(command, detailsLevel, containerKey string, payload map[string]interface{}) ([]interface{}, error) {
	if containerKey == "" {
		containerKey = "objects"
	}

	all := []interface{}{}
	offset := 0
	for {
		page := make(map[string]interface{}, len(payload)+3)
		for k, v := range payload {
			page[k] = v
		}
		page["limit"] = queryPageLimit
		page["offset"] = offset
		if detailsLevel != "" {
			page["details-level"] = detailsLevel
		}

		res, err := c.callWithTimeout(command, page, false)
		if err != nil {
			return nil, err
		}
		if !res.Success {
			return nil, &APIError{Command: command, Message: res.ErrorMsg}
		}

		data := res.GetData()
		items, _ := data[containerKey].([]interface{})
		all = append(all, items...)

		to, _ := data["to"].(float64)
		total, _ := data["total"].(float64)
		if len(items) == 0 || to >= total {
			break
		}
		offset = int(to)
	}
	return all, nil
}

// Logout closes the session on the server. It does not touch any locally
// persisted state — the caller decides what to do with that.
func (c *Client) Logout() error {
	res, err := c.callWithTimeout("logout", map[string]interface{}{}, false)
	if err != nil {
		return err
	}
	if !res.Success {
		return &APIError{Command: "logout", Message: res.ErrorMsg}
	}
	return nil
}

// callWithTimeout runs one ApiCall on its own goroutine, bounded by
// c.callTimeout.
//
// The vendored Check Point SDK's internal wait-for-task loop has a bug: if
// "show-task" itself keeps failing (e.g. the session is revoked mid-call),
// after 5 attempts it falls into a branch with no sleep and no exit
// condition — a 100%-CPU busy loop that never returns. Running the call under
// a hard deadline turns that into a clear, bounded error instead of a silent,
// unbounded hang; the host process exiting on error reclaims whatever
// goroutine is still spinning.
func (c *Client) callWithTimeout(command string, payload map[string]interface{}, waitForTask bool) (api.APIResponse, error) {
	type outcome struct {
		res api.APIResponse
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		res, err := c.api.ApiCall(command, payload, c.api.GetSessionID(), waitForTask, false)
		done <- outcome{res, err}
	}()
	select {
	case o := <-done:
		return o.res, o.err
	case <-time.After(c.callTimeout):
		return api.APIResponse{}, fmt.Errorf("o comando %q não respondeu em %s — o servidor pode estar com problemas ao processar a task; verifique o status manualmente (show-task)", command, c.callTimeout)
	}
}

func stringField(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}
