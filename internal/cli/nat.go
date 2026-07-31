package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newNatCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "nat",
		Short: "NAT rules (nat-rule)",
	}
	root.AddCommand(
		newNatAddCmd(),
		newNatShowCmd(),
		newNatSetCmd(),
		newNatDeleteCmd(),
		newNatListCmd(),
	)
	return root
}

func newNatAddCmd() *cobra.Command {
	var pkg, position, comments string
	var fields []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a manual NAT rule to a policy package",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg == "" {
				return fmt.Errorf("--package is required (policy package name/uid)")
			}
			payload, err := parseFields(fields)
			if err != nil {
				return err
			}
			payload["package"] = pkg
			payload["position"] = position
			if comments != "" {
				payload["comments"] = comments
			}
			return callAndPrint("add-nat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Policy package name/UID (required)")
	cmd.Flags().StringVar(&position, "position", "top", `Position: "top", "bottom", or a number`)
	cmd.Flags().StringVar(&comments, "comments", "", "Rule comment")
	cmd.Flags().StringArrayVar(&fields, "field", nil, `key=value field (e.g. --field original-source='"host1"' --field translated-source='"nat-host1"' --field method='"static"')`)
	return cmd
}

func newNatShowCmd() *cobra.Command {
	var uid, pkg, ruleNumber string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show a NAT rule (by --uid, or --package + --rule-number)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{}
			if uid != "" {
				payload["uid"] = uid
			} else {
				if pkg == "" || ruleNumber == "" {
					return fmt.Errorf("provide --uid, or --package + --rule-number")
				}
				payload["package"] = pkg
				payload["rule-number"] = ruleNumber
			}
			return callAndPrint("show-nat-rule", payload, false, false)
		},
	}
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&pkg, "package", "", "Policy package name/UID")
	cmd.Flags().StringVar(&ruleNumber, "rule-number", "", "Rule number within the package")
	return cmd
}

func newNatSetCmd() *cobra.Command {
	var uid, pkg, ruleNumber string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update an existing NAT rule (by --uid, or --package + --rule-number)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{}
			if uid != "" {
				payload["uid"] = uid
			} else {
				if pkg == "" || ruleNumber == "" {
					return fmt.Errorf("provide --uid, or --package + --rule-number")
				}
				payload["package"] = pkg
				payload["rule-number"] = ruleNumber
			}
			extra, err := parseFields(fields)
			if err != nil {
				return err
			}
			for k, v := range extra {
				payload[k] = v
			}
			return callAndPrint("set-nat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&pkg, "package", "", "Policy package name/UID")
	cmd.Flags().StringVar(&ruleNumber, "rule-number", "", "Rule number within the package")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "key=value field to modify (repeatable)")
	return cmd
}

func newNatDeleteCmd() *cobra.Command {
	var uid, pkg, ruleNumber string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a NAT rule (--package required; --uid or --rule-number)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg == "" {
				return fmt.Errorf("--package is required (delete-nat-rule needs the package even with --uid)")
			}
			payload := map[string]interface{}{"package": pkg}
			if uid != "" {
				payload["uid"] = uid
			} else {
				if ruleNumber == "" {
					return fmt.Errorf("provide --uid, or --rule-number")
				}
				payload["rule-number"] = ruleNumber
			}
			return callAndPrint("delete-nat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&uid, "uid", "", "Rule UID")
	cmd.Flags().StringVar(&pkg, "package", "", "Policy package name/UID (required)")
	cmd.Flags().StringVar(&ruleNumber, "rule-number", "", "Rule number within the package")
	return cmd
}

func newNatListCmd() *cobra.Command {
	var pkg, detailsLevel string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the NAT rules (rulebase) of a policy package",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg == "" {
				return fmt.Errorf("--package is required")
			}
			payload := map[string]interface{}{"package": pkg}
			return listRulebaseAndPrint("show-nat-rulebase", detailsLevel, "rulebase", payload)
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Policy package name/UID (required)")
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Detail level: uid | standard | full")
	return cmd
}
