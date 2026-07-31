package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// vpnCommunityType names the two VPN community families the Management API
// exposes: meshed and star topologies. See entitySpec in helpers.go (shared
// with object.go's plain-object commands).
type vpnCommunityType = entitySpec

var (
	vpnMeshed = vpnCommunityType{
		name: "meshed", short: "Meshed VPN communities (site-to-site mesh)",
		addCmd: "add-vpn-community-meshed", showCmd: "show-vpn-community-meshed", setCmd: "set-vpn-community-meshed", deleteCmd: "delete-vpn-community-meshed", listCmd: "show-vpn-communities-meshed",
		fieldHint: `gateways (e.g. --field gateways='["gw1","gw2"]'), encryption-method`,
	}
	vpnStar = vpnCommunityType{
		name: "star", short: "Star VPN communities (hub-and-spoke)",
		addCmd: "add-vpn-community-star", showCmd: "show-vpn-community-star", setCmd: "set-vpn-community-star", deleteCmd: "delete-vpn-community-star", listCmd: "show-vpn-communities-star",
		fieldHint: `center-gateways, satellite-gateways (lists of names/uids), encryption-method`,
	}
)

func newVPNCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "vpn",
		Short: "Site-to-site VPN communities (meshed and star)",
	}
	root.AddCommand(
		newVPNCommunityCmd(vpnMeshed),
		newVPNCommunityCmd(vpnStar),
	)
	return root
}

func newVPNCommunityCmd(ct vpnCommunityType) *cobra.Command {
	root := &cobra.Command{
		Use:   ct.name,
		Short: ct.short,
	}

	var addName string
	var addFields []string
	addCmd := &cobra.Command{
		Use:   "add",
		Short: fmt.Sprintf("Create a %s VPN community (common fields: %s)", ct.name, ct.fieldHint),
		RunE: func(cmd *cobra.Command, args []string) error {
			if addName == "" {
				return fmt.Errorf("--name is required")
			}
			payload, err := parseFields(addFields)
			if err != nil {
				return err
			}
			payload["name"] = addName
			return callAndPrint(ct.addCmd, payload, true, true)
		},
	}
	addCmd.Flags().StringVar(&addName, "name", "", "Community name (required)")
	addCmd.Flags().StringArrayVar(&addFields, "field", nil, "key=value field (repeatable)")

	var showName, showUID string
	showCmd := &cobra.Command{
		Use:   "show",
		Short: fmt.Sprintf("Show a %s VPN community", ct.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(showName, showUID)
			if err != nil {
				return err
			}
			return callAndPrint(ct.showCmd, payload, false, false)
		},
	}
	showCmd.Flags().StringVar(&showName, "name", "", "Community name")
	showCmd.Flags().StringVar(&showUID, "uid", "", "Community UID")

	var setName, setUID string
	var setFields []string
	setCmd := &cobra.Command{
		Use:   "set",
		Short: fmt.Sprintf("Update an existing %s VPN community", ct.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(setName, setUID)
			if err != nil {
				return err
			}
			extra, err := parseFields(setFields)
			if err != nil {
				return err
			}
			for k, v := range extra {
				payload[k] = v
			}
			return callAndPrint(ct.setCmd, payload, true, true)
		},
	}
	setCmd.Flags().StringVar(&setName, "name", "", "Community name")
	setCmd.Flags().StringVar(&setUID, "uid", "", "Community UID")
	setCmd.Flags().StringArrayVar(&setFields, "field", nil, "key=value field to modify (repeatable)")

	var delName, delUID string
	delCmd := &cobra.Command{
		Use:   "delete",
		Short: fmt.Sprintf("Delete a %s VPN community", ct.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(delName, delUID)
			if err != nil {
				return err
			}
			return callAndPrint(ct.deleteCmd, payload, true, true)
		},
	}
	delCmd.Flags().StringVar(&delName, "name", "", "Community name")
	delCmd.Flags().StringVar(&delUID, "uid", "", "Community UID")

	var listFilter, listDetails string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %s VPN communities", ct.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{}
			if listFilter != "" {
				payload["filter"] = listFilter
			}
			return listAndPrint(ct.listCmd, listDetails, "objects", payload)
		},
	}
	listCmd.Flags().StringVar(&listFilter, "filter", "", "Search text (Check Point filter)")
	listCmd.Flags().StringVar(&listDetails, "details-level", "standard", "Detail level: uid | standard | full")

	root.AddCommand(addCmd, showCmd, setCmd, delCmd, listCmd)
	return root
}
