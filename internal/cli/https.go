package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newHTTPSCmd groups HTTPS Inspection commands: the rulebase (https-rule on
// an https-layer) and listing the https layers. Mirrors the shape of
// newThreatCmd/newRuleCmd — same rulebase-under-a-layer pattern the
// Management API uses for every blade with its own inspection layer.
func newHTTPSCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "https",
		Short: "HTTPS Inspection (rules and layers)",
	}
	root.AddCommand(
		newHTTPSRuleCmd(),
		newHTTPSLayersCmd(),
	)
	return root
}

func newHTTPSRuleCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "rule",
		Short: "HTTPS Inspection rules (https-rule) in a layer",
	}
	root.AddCommand(
		newHTTPSRuleAddCmd(),
		newHTTPSRuleShowCmd(),
		newHTTPSRuleSetCmd(),
		newHTTPSRuleDeleteCmd(),
		newHTTPSRuleListCmd(),
	)
	return root
}

func newHTTPSRuleAddCmd() *cobra.Command {
	var layer, position, name string
	var fields []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a rule in an HTTPS Inspection layer",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf("--layer is required")
			}
			payload, err := parseFields(fields)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			payload["position"] = position
			if name != "" {
				payload["name"] = name
			}
			return callAndPrint("add-https-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "HTTPS Inspection layer name/UID (required)")
	cmd.Flags().StringVar(&position, "position", "top", `Position: "top", "bottom", a number, or "above:<uid>"/"below:<uid>"`)
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringArrayVar(&fields, "field", nil, `key=value field (e.g. --field source='["any"]' --field action='"Inspect"')`)
	return cmd
}

// HTTPS Inspection's show/set/delete all require "layer" even when the rule
// is identified by "uid" — confirmed against a live Management Server
// (generic_err_missing_required_parameters: [layer] otherwise). Same
// requirement threat-rule has; unlike access-rule, where uid alone is
// enough for show/set.

func newHTTPSRuleShowCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show an HTTPS Inspection rule (--layer required)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer is required (show-https-rule needs the layer even with --uid)`)
			}
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			return callAndPrint("show-https-rule", payload, false, false)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&layer, "layer", "", "Rule layer (required)")
	return cmd
}

func newHTTPSRuleSetCmd() *cobra.Command {
	var name, uid, layer string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update an HTTPS Inspection rule (--layer required)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer is required (set-https-rule needs the layer even with --uid)`)
			}
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			extra, err := parseFields(fields)
			if err != nil {
				return err
			}
			for k, v := range extra {
				payload[k] = v
			}
			return callAndPrint("set-https-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&layer, "layer", "", "Rule layer (required)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "key=value field to modify (repeatable)")
	return cmd
}

func newHTTPSRuleDeleteCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an HTTPS Inspection rule (--layer required)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer is required (delete-https-rule needs the layer even with --uid)`)
			}
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			return callAndPrint("delete-https-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&layer, "layer", "", "Rule layer (required)")
	return cmd
}

func newHTTPSRuleListCmd() *cobra.Command {
	var layer, detailsLevel string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the rules (rulebase) of an HTTPS Inspection layer",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf("--layer is required")
			}
			payload := map[string]interface{}{"name": layer}
			return listRulebaseAndPrint("show-https-rulebase", detailsLevel, "rulebase", payload)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "Layer name/UID (required)")
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Detail level: uid | standard | full")
	return cmd
}

func newHTTPSLayersCmd() *cobra.Command {
	var detailsLevel string
	cmd := &cobra.Command{
		Use:   "layers",
		Short: "List the available HTTPS Inspection layers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAndPrint("show-https-layers", detailsLevel, "https-layers", map[string]interface{}{})
		},
	}
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Detail level: uid | standard | full")
	return cmd
}
