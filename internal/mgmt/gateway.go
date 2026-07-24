package mgmt

// interfaceReadOnlyFields are computed/display fields show-simple-gateway
// includes on every interface entry but set-simple-gateway rejects
// (generic_err_invalid_parameter_name) if they're echoed back unchanged —
// confirmed against a live Management Server.
var interfaceReadOnlyFields = []string{"icon", "color", "uid", "network-interface-type", "topology-automatic-calculation"}

// MergeGatewayInterface returns a copy of a simple-gateway's "interfaces"
// array (as returned by "show-simple-gateway" with details-level=full) with
// fields merged into the entry named ifaceName.
//
// set-simple-gateway replaces the WHOLE "interfaces" array rather than
// patching one entry — sending back only the interface being changed
// silently deletes every other interface on the gateway (confirmed live:
// dropped a gateway from 2 interfaces to 1). So every entry is preserved,
// only ifaceName's fields are changed, and every entry has its read-only
// fields stripped since the whole array gets resent.
//
// The bool return reports whether ifaceName was found.
func MergeGatewayInterface(ifaces []interface{}, ifaceName string, fields map[string]interface{}) ([]interface{}, bool) {
	updated := make([]interface{}, len(ifaces))
	found := false
	for i, raw := range ifaces {
		entry, ok := raw.(map[string]interface{})
		if !ok {
			updated[i] = raw
			continue
		}
		clean := make(map[string]interface{}, len(entry))
		for k, v := range entry {
			clean[k] = v
		}
		for _, f := range interfaceReadOnlyFields {
			delete(clean, f)
		}
		if entry["name"] == ifaceName {
			found = true
			for k, v := range fields {
				clean[k] = v
			}
		}
		updated[i] = clean
	}
	return updated, found
}
