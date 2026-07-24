package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"cpcli/internal/mgmt"
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

// clientFromSession loads the persisted session for the active profile and
// builds a mgmt.Client ready to make authenticated calls. All the SDK/
// transport plumbing (timeouts, pagination, fingerprint handling) lives in
// package mgmt so the same core can back a UI.
func clientFromSession() (*mgmt.Client, *session.Session, error) {
	sess, err := session.Load(activeProfile())
	if err != nil {
		return nil, nil, err
	}
	client := mgmt.Connect(mgmt.Conn{
		Server:     sess.Server,
		Port:       sess.Port,
		Sid:        sess.Sid,
		APIVersion: sess.ApiVersion,
		Insecure:   sess.Insecure,
	})
	return client, sess, nil
}

// callAndPrint executes a single API command and prints the result. When
// mutates is true and the call succeeds, it reminds the user that a publish
// is required for the change to take effect on the gateways.
func callAndPrint(command string, payload map[string]interface{}, waitForTask bool, mutates bool) error {
	client, _, err := clientFromSession()
	if err != nil {
		return err
	}
	data, err := client.Call(command, payload, waitForTask)
	if err != nil {
		return err
	}
	if err := printData(data); err != nil {
		return err
	}
	if mutates {
		fmt.Fprintln(os.Stderr, "Lembre-se de rodar `cpcli session publish` para efetivar a mudança nos gateways.")
	}
	return nil
}

// printData renders a command's data payload, or "OK" when the command
// returned no data (some mutating calls do).
func printData(data map[string]interface{}) error {
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

// listAndPrint is the shared body of every "list" subcommand: paginate
// through mgmt.Client.List and print the aggregated results as JSON.
func listAndPrint(command, detailsLevel, containerKey string, payload map[string]interface{}) error {
	client, _, err := clientFromSession()
	if err != nil {
		return err
	}
	items, err := client.List(command, detailsLevel, containerKey, payload)
	if err != nil {
		return err
	}
	return printJSON(items)
}

// listRulebaseAndPrint is listAndPrint's counterpart for rulebase commands
// (show-access-rulebase, show-nat-rulebase, show-threat-rulebase,
// show-https-rulebase): it resolves action/source/destination/service/etc.
// UIDs to display names via mgmt.Client.ListRulebase, instead of printing
// raw UIDs like a plain List would.
func listRulebaseAndPrint(command, detailsLevel, containerKey string, payload map[string]interface{}) error {
	client, _, err := clientFromSession()
	if err != nil {
		return err
	}
	items, err := client.ListRulebase(command, detailsLevel, containerKey, payload)
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
