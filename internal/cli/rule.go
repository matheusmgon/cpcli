package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRuleCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "rule",
		Short: "Access Control rules (access-rule)",
	}
	root.AddCommand(
		newRuleAddCmd(),
		newRuleShowCmd(),
		newRuleSetCmd(),
		newRuleDeleteCmd(),
		newRuleListCmd(),
		newLayerListCmd(),
	)
	return root
}

func newRuleAddCmd() *cobra.Command {
	var layer, position, name, action, comments string
	var fields []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a rule in an Access Control layer",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer is required (e.g. "Network")`)
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
			if action != "" {
				payload["action"] = action
			}
			if comments != "" {
				payload["comments"] = comments
			}
			return callAndPrint("add-access-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "Policy layer name/UID (required)")
	cmd.Flags().StringVar(&position, "position", "top", `Position: "top", "bottom", a number, or "above:<uid>"/"below:<uid>"`)
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringVar(&action, "action", "", "Action: accept, drop, reject, ...")
	cmd.Flags().StringVar(&comments, "comments", "", "Rule comment")
	cmd.Flags().StringArrayVar(&fields, "field", nil, `key=value field (e.g. --field source='["any"]' --field service='["https"]' --field destination='["any"]')`)
	return cmd
}

func newRuleShowCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show a rule (by --uid, or --name + --layer)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			if uid == "" && layer != "" {
				payload["layer"] = layer
			}
			return callAndPrint("show-access-rule", payload, false, false)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&layer, "layer", "", "Rule layer (required when identifying by --name)")
	return cmd
}

func newRuleSetCmd() *cobra.Command {
	var name, uid, layer string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update an existing rule (by --uid, or --name + --layer)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			if uid == "" && layer != "" {
				payload["layer"] = layer
			}
			extra, err := parseFields(fields)
			if err != nil {
				return err
			}
			for k, v := range extra {
				payload[k] = v
			}
			return callAndPrint("set-access-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&layer, "layer", "", "Rule layer (required when identifying by --name)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "key=value field to modify (repeatable)")
	return cmd
}

func newRuleDeleteCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a rule (--uid + --layer, or --name + --layer)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer is required (delete-access-rule needs the layer even with --uid)`)
			}
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			return callAndPrint("delete-access-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&layer, "layer", "", "Rule layer (required)")
	return cmd
}

func newRuleListCmd() *cobra.Command {
	var layer, detailsLevel string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the rules (rulebase) of a layer",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf("--layer is required")
			}
			payload := map[string]interface{}{"name": layer}
			return listRulebaseAndPrint("show-access-rulebase", detailsLevel, "rulebase", payload)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "Layer name/UID (required)")
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Detail level: uid | standard | full")
	return cmd
}

func newLayerListCmd() *cobra.Command {
	var detailsLevel string
	cmd := &cobra.Command{
		Use:   "layers",
		Short: "List the available Access Control (access) layers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAndPrint("show-access-layers", detailsLevel, "access-layers", map[string]interface{}{})
		},
	}
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Detail level: uid | standard | full")
	return cmd
}
