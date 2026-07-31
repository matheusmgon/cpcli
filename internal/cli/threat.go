package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newThreatCmd groups Threat Prevention commands: the rulebase (threat-rule
// on a threat layer), the threat profiles, and listing the threat layers.
func newThreatCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "threat",
		Short: "Threat Prevention (rules, profiles and layers)",
	}
	root.AddCommand(
		newThreatRuleCmd(),
		newThreatProfileCmd(),
		newThreatLayersCmd(),
	)
	return root
}

// --- Threat rules ---------------------------------------------------------

func newThreatRuleCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "rule",
		Short: "Threat Prevention rules (threat-rule) in a layer",
	}
	root.AddCommand(
		newThreatRuleAddCmd(),
		newThreatRuleShowCmd(),
		newThreatRuleSetCmd(),
		newThreatRuleDeleteCmd(),
		newThreatRuleListCmd(),
	)
	return root
}

func newThreatRuleAddCmd() *cobra.Command {
	var layer, position, name string
	var fields []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a rule in a Threat Prevention layer",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf("--layer is required")
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			payload, err := parseFields(fields)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			payload["position"] = position
			payload["name"] = name
			return callAndPrint("add-threat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "Threat Prevention layer name/UID (required)")
	cmd.Flags().StringVar(&position, "position", "top", `Position: "top", "bottom", a number, or "above:<uid>"/"below:<uid>"`)
	cmd.Flags().StringVar(&name, "name", "", "Rule name (required)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, `key=value field (e.g. --field protected-scope='["any"]' --field track='"Log"')`)
	return cmd
}

func newThreatRuleShowCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show a Threat Prevention rule (--layer required)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer is required (show-threat-rule needs the layer even with --uid)`)
			}
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			return callAndPrint("show-threat-rule", payload, false, false)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&layer, "layer", "", "Rule layer (required)")
	return cmd
}

func newThreatRuleSetCmd() *cobra.Command {
	var name, uid, layer string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update a Threat Prevention rule (--layer required)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer is required (set-threat-rule needs the layer even with --uid)`)
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
			return callAndPrint("set-threat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&layer, "layer", "", "Rule layer (required)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "key=value field to modify (repeatable)")
	return cmd
}

func newThreatRuleDeleteCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Threat Prevention rule (--layer required)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer is required (delete-threat-rule needs the layer even with --uid)`)
			}
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			return callAndPrint("delete-threat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&layer, "layer", "", "Rule layer (required)")
	return cmd
}

func newThreatRuleListCmd() *cobra.Command {
	var layer, detailsLevel string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the rules (rulebase) of a Threat Prevention layer",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf("--layer is required")
			}
			payload := map[string]interface{}{"name": layer}
			return listRulebaseAndPrint("show-threat-rulebase", detailsLevel, "rulebase", payload)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "Layer name/UID (required)")
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Detail level: uid | standard | full")
	return cmd
}

// --- Threat profiles ------------------------------------------------------

func newThreatProfileCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "profile",
		Short: "Threat Prevention profiles (threat-profile)",
	}

	var addName string
	var addFields []string
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Create a Threat Prevention profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if addName == "" {
				return fmt.Errorf("--name is required")
			}
			payload, err := parseFields(addFields)
			if err != nil {
				return err
			}
			payload["name"] = addName
			return callAndPrint("add-threat-profile", payload, true, true)
		},
	}
	addCmd.Flags().StringVar(&addName, "name", "", "Profile name (required)")
	addCmd.Flags().StringArrayVar(&addFields, "field", nil, "key=value field (e.g. --field active-protections-performance-impact='\"medium\"')")

	var showName, showUID string
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show a Threat Prevention profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(showName, showUID)
			if err != nil {
				return err
			}
			return callAndPrint("show-threat-profile", payload, false, false)
		},
	}
	showCmd.Flags().StringVar(&showName, "name", "", "Profile name")
	showCmd.Flags().StringVar(&showUID, "uid", "", "Profile UID")

	var setName, setUID string
	var setFields []string
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update an existing Threat Prevention profile",
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
			return callAndPrint("set-threat-profile", payload, true, true)
		},
	}
	setCmd.Flags().StringVar(&setName, "name", "", "Profile name")
	setCmd.Flags().StringVar(&setUID, "uid", "", "Profile UID")
	setCmd.Flags().StringArrayVar(&setFields, "field", nil, "key=value field to modify (repeatable)")

	var delName, delUID string
	delCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Threat Prevention profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(delName, delUID)
			if err != nil {
				return err
			}
			return callAndPrint("delete-threat-profile", payload, true, true)
		},
	}
	delCmd.Flags().StringVar(&delName, "name", "", "Profile name")
	delCmd.Flags().StringVar(&delUID, "uid", "", "Profile UID")

	var listDetails string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the Threat Prevention profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAndPrint("show-threat-profiles", listDetails, "objects", map[string]interface{}{})
		},
	}
	listCmd.Flags().StringVar(&listDetails, "details-level", "standard", "Detail level: uid | standard | full")

	root.AddCommand(addCmd, showCmd, setCmd, delCmd, listCmd)
	return root
}

// --- Threat layers --------------------------------------------------------

func newThreatLayersCmd() *cobra.Command {
	var detailsLevel string
	cmd := &cobra.Command{
		Use:   "layers",
		Short: "List the available Threat Prevention layers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAndPrint("show-threat-layers", detailsLevel, "threat-layers", map[string]interface{}{})
		},
	}
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Detail level: uid | standard | full")
	return cmd
}
