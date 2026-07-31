package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// objectType describes one Check Point object family. See entitySpec in
// helpers.go (shared with vpn.go's VPN community commands).
type objectType = entitySpec

var (
	objTypeHost = objectType{
		name: "host", short: "Host objects (single IP address)",
		addCmd: "add-host", showCmd: "show-host", setCmd: "set-host", deleteCmd: "delete-host", listCmd: "show-hosts",
		fieldHint: "ip-address, comments, color, groups",
	}
	objTypeNetwork = objectType{
		name: "network", short: "Network objects (subnet)",
		addCmd: "add-network", showCmd: "show-network", setCmd: "set-network", deleteCmd: "delete-network", listCmd: "show-networks",
		fieldHint: "subnet4, mask-length4 (or subnet-mask), comments, color",
	}
	objTypeGroup = objectType{
		name: "group", short: "Object groups",
		addCmd: "add-group", showCmd: "show-group", setCmd: "set-group", deleteCmd: "delete-group", listCmd: "show-groups",
		fieldHint: `members (e.g. --field members='["host1","host2"]')`,
	}
	objTypeServiceTCP = objectType{
		name: "service-tcp", short: "TCP services",
		addCmd: "add-service-tcp", showCmd: "show-service-tcp", setCmd: "set-service-tcp", deleteCmd: "delete-service-tcp", listCmd: "show-services-tcp",
		fieldHint: "port, comments",
	}
	objTypeServiceUDP = objectType{
		name: "service-udp", short: "UDP services",
		addCmd: "add-service-udp", showCmd: "show-service-udp", setCmd: "set-service-udp", deleteCmd: "delete-service-udp", listCmd: "show-services-udp",
		fieldHint: "port, comments",
	}
	objTypeAddressRange = objectType{
		name: "address-range", short: "Address ranges (address-range)",
		addCmd: "add-address-range", showCmd: "show-address-range", setCmd: "set-address-range", deleteCmd: "delete-address-range", listCmd: "show-address-ranges",
		fieldHint: "ip-address-first, ip-address-last, comments, color",
	}
	objTypeServiceGroup = objectType{
		name: "service-group", short: "Service groups",
		addCmd: "add-service-group", showCmd: "show-service-group", setCmd: "set-service-group", deleteCmd: "delete-service-group", listCmd: "show-service-groups",
		fieldHint: `members (e.g. --field members='["https","ssh"]')`,
	}
	objTypeServiceICMP = objectType{
		name: "service-icmp", short: "ICMP services",
		addCmd: "add-service-icmp", showCmd: "show-service-icmp", setCmd: "set-service-icmp", deleteCmd: "delete-service-icmp", listCmd: "show-services-icmp",
		fieldHint: "icmp-type, icmp-code, comments",
	}
	objTypeServiceOther = objectType{
		name: "service-other", short: "Services on other IP protocols",
		addCmd: "add-service-other", showCmd: "show-service-other", setCmd: "set-service-other", deleteCmd: "delete-service-other", listCmd: "show-services-other",
		fieldHint: "ip-protocol, match, comments",
	}
	objTypeSecurityZone = objectType{
		name: "security-zone", short: "Security zones (security-zone)",
		addCmd: "add-security-zone", showCmd: "show-security-zone", setCmd: "set-security-zone", deleteCmd: "delete-security-zone", listCmd: "show-security-zones",
		fieldHint: "comments, color",
	}
	objTypeDNSDomain = objectType{
		name: "dns-domain", short: "DNS domains (dns-domain)",
		addCmd: "add-dns-domain", showCmd: "show-dns-domain", setCmd: "set-dns-domain", deleteCmd: "delete-dns-domain", listCmd: "show-dns-domains",
		fieldHint: `name (e.g. ".example.com"), is-sub-domain (bool)`,
	}
	objTypeWildcard = objectType{
		name: "wildcard", short: "Wildcard objects (IP + wildcard mask)",
		addCmd: "add-wildcard", showCmd: "show-wildcard", setCmd: "set-wildcard", deleteCmd: "delete-wildcard", listCmd: "show-wildcards",
		fieldHint: "ipv4-address, ipv4-mask-wildcard, comments",
	}
	objTypeTag = objectType{
		name: "tag", short: "Tags (object labels)",
		addCmd: "add-tag", showCmd: "show-tag", setCmd: "set-tag", deleteCmd: "delete-tag", listCmd: "show-tags",
		fieldHint: "comments, color",
	}
	objTypeTime = objectType{
		name: "time", short: "Time objects (time)",
		addCmd: "add-time", showCmd: "show-time", setCmd: "set-time", deleteCmd: "delete-time", listCmd: "show-times",
		fieldHint: "start, end, recurrence (see API docs)",
	}
	objTypeDynamicObject = objectType{
		name: "dynamic-object", short: "Dynamic objects (dynamic-object)",
		addCmd: "add-dynamic-object", showCmd: "show-dynamic-object", setCmd: "set-dynamic-object", deleteCmd: "delete-dynamic-object", listCmd: "show-dynamic-objects",
		fieldHint: "comments, color",
	}
	objTypeAccessRole = objectType{
		name: "access-role", short: "Access Roles (identity: users/machines/networks)",
		addCmd: "add-access-role", showCmd: "show-access-role", setCmd: "set-access-role", deleteCmd: "delete-access-role", listCmd: "show-access-roles",
		fieldHint: `networks, users, machines (e.g. --field networks='["any"]')`,
	}
	objTypeApplicationSite = objectType{
		name: "application-site", short: "Custom applications/sites (application-site)",
		addCmd: "add-application-site", showCmd: "show-application-site", setCmd: "set-application-site", deleteCmd: "delete-application-site", listCmd: "show-application-sites",
		fieldHint: "url-list, application-signature, primary-category, comments",
	}
)

func newObjectCmd(ot objectType) *cobra.Command {
	root := &cobra.Command{
		Use:   ot.name,
		Short: ot.short,
	}
	root.AddCommand(
		newObjectAddCmd(ot),
		newObjectShowCmd(ot),
		newObjectSetCmd(ot),
		newObjectDeleteCmd(ot),
		newObjectListCmd(ot),
	)
	return root
}

func newObjectAddCmd(ot objectType) *cobra.Command {
	var name string
	var fields []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: fmt.Sprintf("Create a %s object (common fields: %s)", ot.name, ot.fieldHint),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			payload, err := parseFields(fields)
			if err != nil {
				return err
			}
			payload["name"] = name
			return callAndPrint(ot.addCmd, payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Object name (required)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "key=value field (repeatable). Values that parse as valid JSON are interpreted as such.")
	return cmd
}

func newObjectShowCmd(ot objectType) *cobra.Command {
	var name, uid string
	cmd := &cobra.Command{
		Use:   "show",
		Short: fmt.Sprintf("Show a %s object", ot.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["details-level"] = "full"
			return callAndPrint(ot.showCmd, payload, false, false)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Object name")
	cmd.Flags().StringVar(&uid, "uid", "", "Object UID")
	return cmd
}

func newObjectSetCmd(ot objectType) *cobra.Command {
	var name, uid string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: fmt.Sprintf("Update an existing %s object", ot.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			extra, err := parseFields(fields)
			if err != nil {
				return err
			}
			for k, v := range extra {
				payload[k] = v
			}
			return callAndPrint(ot.setCmd, payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Object name")
	cmd.Flags().StringVar(&uid, "uid", "", "Object UID")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "key=value field to modify (repeatable)")
	return cmd
}

func newObjectDeleteCmd(ot objectType) *cobra.Command {
	var name, uid string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: fmt.Sprintf("Delete a %s object", ot.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			return callAndPrint(ot.deleteCmd, payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Object name")
	cmd.Flags().StringVar(&uid, "uid", "", "Object UID")
	return cmd
}

func newObjectListCmd(ot objectType) *cobra.Command {
	var filter, detailsLevel string
	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %s objects", ot.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{}
			if filter != "" {
				payload["filter"] = filter
			}
			return listAndPrint(ot.listCmd, detailsLevel, "objects", payload)
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Search text (Check Point filter)")
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Detail level: uid | standard | full")
	return cmd
}
