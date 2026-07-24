package mgmt

// referenceFieldNames are the rulebase row fields that hold object
// references (as bare UID strings or arrays of them) rather than literal
// values. Only these are substituted — fields like "uid" (the rule's own
// identity) or "layer"/"domain" must never be touched by name resolution.
var referenceFieldNames = map[string]bool{
	"action":                 true,
	"track":                  true,
	"source":                 true,
	"destination":            true,
	"service":                true,
	"vpn":                    true,
	"content":                true,
	"protected-scope":        true,
	"install-on":             true,
	"original-source":        true,
	"original-destination":   true,
	"original-service":       true,
	"translated-source":      true,
	"translated-destination": true,
	"translated-service":     true,
}

// ListRulebase aggregates a rulebase-style command ("show-access-rulebase",
// "show-nat-rulebase", "show-threat-rulebase", "show-https-rulebase") into a
// flat slice, the same way List does, but also asks the API for its
// "objects-dictionary" (via use-object-dictionary) and substitutes the
// dictionary's names into each row's reference fields.
//
// Without this, rulebase rows come back with action/source/destination/
// service/etc. as bare UID strings (confirmed against a live server — even
// at details-level=full) instead of names, which is why the raw client and
// UI showed UIDs instead of "Drop"/"Any"/etc. SmartConsole itself resolves
// these the same way, via the same dictionary mechanism.
func (c *Client) ListRulebase(command, detailsLevel, containerKey string, payload map[string]interface{}) ([]interface{}, error) {
	if containerKey == "" {
		containerKey = "rulebase"
	}

	all := []interface{}{}
	var dictionary []interface{}
	offset := 0
	for {
		page := make(map[string]interface{}, len(payload)+4)
		for k, v := range payload {
			page[k] = v
		}
		page["limit"] = queryPageLimit
		page["offset"] = offset
		page["use-object-dictionary"] = true
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
		if dict, ok := data["objects-dictionary"].([]interface{}); ok {
			dictionary = append(dictionary, dict...)
		}

		to, _ := data["to"].(float64)
		total, _ := data["total"].(float64)
		if len(items) == 0 || to >= total {
			break
		}
		offset = int(to)
	}
	return resolveRulebaseRefs(all, dictionary), nil
}

// resolveRulebaseRefs returns a copy of rows with every referenceFieldNames
// value that's a known uid (per dictionary) replaced by its display name —
// recursing into section rows' own nested "rulebase" array too (see
// resolveRulebaseRefsWithIndex). Rows that aren't objects (shouldn't
// happen, but List's containerKey extraction is untyped) pass through
// unchanged.
func resolveRulebaseRefs(rows []interface{}, dictionary []interface{}) []interface{} {
	idx := buildUIDIndex(dictionary)
	if len(idx) == 0 {
		return rows
	}
	return resolveRulebaseRefsWithIndex(rows, idx)
}

// resolveRulebaseRefsWithIndex does the actual substitution, given an
// already-built uid->name index — factored out so it can recurse into
// section rows' nested "rulebase" array (e.g. an "Automatic Generated
// Rules" NAT section hides its real rules there — confirmed live, a plain
// top-level walk would never see or resolve them) without rebuilding the
// index per section.
func resolveRulebaseRefsWithIndex(rows []interface{}, idx map[string]string) []interface{} {
	out := make([]interface{}, len(rows))
	for i, raw := range rows {
		row, ok := raw.(map[string]interface{})
		if !ok {
			out[i] = raw
			continue
		}
		resolved := make(map[string]interface{}, len(row))
		for k, v := range row {
			switch {
			case k == "rulebase":
				if nested, ok := v.([]interface{}); ok {
					resolved[k] = resolveRulebaseRefsWithIndex(nested, idx)
				} else {
					resolved[k] = v
				}
			case referenceFieldNames[k]:
				resolved[k] = substituteUIDs(v, idx)
			default:
				resolved[k] = v
			}
		}
		out[i] = resolved
	}
	return out
}

func buildUIDIndex(dictionary []interface{}) map[string]string {
	idx := make(map[string]string, len(dictionary))
	for _, raw := range dictionary {
		obj, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		uid, _ := obj["uid"].(string)
		name, _ := obj["name"].(string)
		if uid != "" && name != "" {
			idx[uid] = name
		}
	}
	return idx
}

func substituteUIDs(v interface{}, idx map[string]string) interface{} {
	switch val := v.(type) {
	case string:
		if name, ok := idx[val]; ok {
			return name
		}
		return val
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, item := range val {
			out[i] = substituteUIDs(item, idx)
		}
		return out
	case map[string]interface{}:
		// Recurse into nested objects too — e.g. "track" comes back as
		// {"type": "<uid>", "alert": ..., ...}, not a bare uid/array like
		// action/source/destination. Safe because we're already scoped to
		// one allowlisted reference field's contents, not the rule row
		// itself (so this never touches the row's own "uid"/"layer"/etc).
		out := make(map[string]interface{}, len(val))
		for k, item := range val {
			out[k] = substituteUIDs(item, idx)
		}
		return out
	default:
		return v
	}
}
