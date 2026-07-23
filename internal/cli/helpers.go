package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	api "github.com/CheckPointSW/cp-mgmt-api-go-sdk/APIFiles"

	"cpcli/internal/session"
)

// entitySpec names the add/show/set/delete/list commands for one Check
// Point object family. Command names don't follow a single naming
// convention (e.g. "show-hosts" vs "show-services-tcp" vs
// "show-vpn-communities-star"), so each family is spelled out explicitly
// instead of derived from a pluralization rule. Shared by both plain
// objects (object.go) and VPN communities (vpn.go).
type entitySpec struct {
	name      string
	short     string
	addCmd    string
	showCmd   string
	setCmd    string
	deleteCmd string
	listCmd   string
	fieldHint string
}

// asyncCallTimeout bounds how long cpcli waits for a command that polls
// "show-task" internally (add/set/delete/publish/install-policy/...).
//
// The vendored Check Point SDK's internal wait-for-task loop has a bug: if
// "show-task" itself keeps failing (e.g. the session is revoked mid-call),
// after 5 attempts it falls into a branch with no sleep and no exit
// condition — a 100%-CPU busy loop that never returns. Running the call in
// a goroutine with a hard deadline turns that into a clear, bounded error
// instead of a silent, unbounded hang; main.go's os.Exit on error also
// reclaims whatever goroutine is still spinning.
const asyncCallTimeout = 5 * time.Minute

// queryPageLimit mirrors the fixed page size the SDK's own pagination
// helper uses.
const queryPageLimit = 50

// clientFromSession loads the persisted session for the active profile and
// builds an API client ready to make authenticated calls.
//
// AcceptServerCertificate is intentionally left false: the SDK's auto-accept
// path persists an empty fingerprint on first use instead of the real one
// (it saves c.fingerprint before that field is populated), which silently
// defeats trust-on-first-use pinning. Leaving it false routes through the
// SDK's interactive compare-and-prompt path instead, which pins the real
// fingerprint correctly. IgnoreServerCertificate is carried over from
// whatever was chosen at login time (Session.Insecure), so an insecure
// login doesn't turn into a fingerprint prompt (or failure, in
// non-interactive use) on every later command.
func clientFromSession() (*api.ApiClient, *session.Session, error) {
	sess, err := session.Load(activeProfile())
	if err != nil {
		return nil, nil, err
	}
	args := api.APIClientArgs(sess.Port, "", sess.Sid, sess.Server, "", -1, sess.ApiVersion, sess.Insecure, false, "", api.WebContext, api.TimeOut, api.SleepTime, "cpcli", "", -1)
	client := api.APIClient(args)
	return client, sess, nil
}

// apiCallWithTimeout runs client.ApiCall on its own goroutine and bounds it
// with asyncCallTimeout. See the doc comment on asyncCallTimeout for why.
func apiCallWithTimeout(client *api.ApiClient, command string, payload map[string]interface{}, waitForTask bool) (api.APIResponse, error) {
	type outcome struct {
		res api.APIResponse
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		res, err := client.ApiCall(command, payload, client.GetSessionID(), waitForTask, false)
		done <- outcome{res, err}
	}()
	select {
	case o := <-done:
		return o.res, o.err
	case <-time.After(asyncCallTimeout):
		return api.APIResponse{}, fmt.Errorf("o comando %q não respondeu em %s — o servidor pode estar com problemas ao processar a task; verifique manualmente com `cpcli task <task-id>`", command, asyncCallTimeout)
	}
}

// callAndPrint executes a single API command and prints the result. When
// mutates is true and the call succeeds, it reminds the user that a publish
// is required for the change to take effect on the gateways.
func callAndPrint(command string, payload map[string]interface{}, waitForTask bool, mutates bool) error {
	client, _, err := clientFromSession()
	if err != nil {
		return err
	}
	res, err := apiCallWithTimeout(client, command, payload, waitForTask)
	if err != nil {
		return err
	}
	if err := printResult(res); err != nil {
		return err
	}
	if mutates {
		fmt.Fprintln(os.Stderr, "Lembre-se de rodar `cpcli session publish` para efetivar a mudança nos gateways.")
	}
	return nil
}

func printResult(res api.APIResponse) error {
	if !res.Success {
		return errors.New(res.ErrorMsg)
	}
	data := res.GetData()
	if len(data) == 0 {
		fmt.Println("OK")
		return nil
	}
	return printJSON(data)
}

func printJSON(v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// queryAll lists every item matching payload for a command that returns a
// paginated container (e.g. show-hosts, show-access-rulebase), aggregating
// every page itself via plain client.ApiCall calls.
//
// It deliberately does not use the SDK's ApiQuery/genApiQuery helpers:
// those call os.Exit(1) directly when any page after the first one fails
// (e.g. the session expires partway through a large listing), which would
// kill the whole cpcli process instead of returning a normal error.
func queryAll(client *api.ApiClient, command, detailsLevel, containerKey string, payload map[string]interface{}) ([]interface{}, error) {
	if containerKey == "" {
		containerKey = "objects"
	}

	// Non-nil so an empty result set prints as "[]" instead of "null".
	all := []interface{}{}
	offset := 0
	for {
		page := make(map[string]interface{}, len(payload)+3)
		for k, v := range payload {
			page[k] = v
		}
		page["limit"] = queryPageLimit
		page["offset"] = offset
		page["details-level"] = detailsLevel

		res, err := apiCallWithTimeout(client, command, page, false)
		if err != nil {
			return nil, err
		}
		if !res.Success {
			return nil, errors.New(res.ErrorMsg)
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

// listAndPrint is the shared body of every "list" subcommand: paginate
// through queryAll and print the aggregated results as JSON.
func listAndPrint(command, detailsLevel, containerKey string, payload map[string]interface{}) error {
	client, _, err := clientFromSession()
	if err != nil {
		return err
	}
	items, err := queryAll(client, command, detailsLevel, containerKey, payload)
	if err != nil {
		return err
	}
	return printJSON(items)
}

// parseFields turns repeated --field key=value flags into a JSON payload.
// A value that parses as valid JSON (numbers, booleans, arrays, objects) is
// kept as that type; anything else is kept as a plain string.
func parseFields(fields []string) (map[string]interface{}, error) {
	payload := map[string]interface{}{}
	for _, f := range fields {
		key, value, ok := splitKV(f)
		if !ok {
			return nil, fmt.Errorf("--field inválido (esperado chave=valor): %q", f)
		}
		var decoded interface{}
		if err := json.Unmarshal([]byte(value), &decoded); err == nil {
			payload[key] = decoded
		} else {
			payload[key] = value
		}
	}
	return payload, nil
}

func splitKV(s string) (key, value string, ok bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:], true
		}
	}
	return "", "", false
}

// nameOrUIDPayload builds the identifying payload for show/set/delete calls
// that accept either a "name" or a "uid".
func nameOrUIDPayload(name, uid string) (map[string]interface{}, error) {
	if uid == "" && name == "" {
		return nil, errors.New("informe --name ou --uid")
	}
	p := map[string]interface{}{}
	if uid != "" {
		p["uid"] = uid
	} else {
		p["name"] = name
	}
	return p, nil
}
