package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPolicyCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "policy",
		Short: "Install or verify a policy package on the gateways",
	}
	root.AddCommand(
		newPolicyInstallCmd(),
		newPolicyVerifyCmd(),
		newPolicyPackageCmd(),
	)
	return root
}

// newPolicyPackageCmd manages policy packages (the containers that hold the
// Access Control and Threat Prevention rulebases installed on gateways).
func newPolicyPackageCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "package",
		Short: "Policy packages (policy-package)",
	}

	var addName string
	var addFields []string
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Create a policy package (blades: --field access=true --field threat-prevention=true)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if addName == "" {
				return fmt.Errorf("--name is required")
			}
			payload, err := parseFields(addFields)
			if err != nil {
				return err
			}
			payload["name"] = addName
			return callAndPrint("add-package", payload, true, true)
		},
	}
	addCmd.Flags().StringVar(&addName, "name", "", "Package name (required)")
	addCmd.Flags().StringArrayVar(&addFields, "field", nil, "key=value field (repeatable). e.g. --field access=true --field nat=true")

	var showName, showUID string
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show a policy package",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(showName, showUID)
			if err != nil {
				return err
			}
			return callAndPrint("show-package", payload, false, false)
		},
	}
	showCmd.Flags().StringVar(&showName, "name", "", "Package name")
	showCmd.Flags().StringVar(&showUID, "uid", "", "Package UID")

	var setName, setUID string
	var setFields []string
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update an existing policy package",
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
			return callAndPrint("set-package", payload, true, true)
		},
	}
	setCmd.Flags().StringVar(&setName, "name", "", "Package name")
	setCmd.Flags().StringVar(&setUID, "uid", "", "Package UID")
	setCmd.Flags().StringArrayVar(&setFields, "field", nil, "key=value field to modify (repeatable)")

	var delName, delUID string
	delCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a policy package",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(delName, delUID)
			if err != nil {
				return err
			}
			return callAndPrint("delete-package", payload, true, true)
		},
	}
	delCmd.Flags().StringVar(&delName, "name", "", "Package name")
	delCmd.Flags().StringVar(&delUID, "uid", "", "Package UID")

	var listDetails string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the policy packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAndPrint("show-packages", listDetails, "packages", map[string]interface{}{})
		},
	}
	listCmd.Flags().StringVar(&listDetails, "details-level", "standard", "Detail level: uid | standard | full")

	root.AddCommand(addCmd, showCmd, setCmd, delCmd, listCmd)
	return root
}

func newPolicyInstallCmd() *cobra.Command {
	var pkg string
	var targets []string
	var accessOnly bool
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a policy package on the target gateways (waits for completion)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg == "" {
				return fmt.Errorf("--package is required")
			}
			if len(targets) == 0 {
				return fmt.Errorf("provide at least one --target (gateway name/uid)")
			}
			payload := map[string]interface{}{
				"policy-package": pkg,
				"targets":        targets,
			}
			if accessOnly {
				payload["access"] = true
			}
			return callAndPrint("install-policy", payload, true, false)
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Policy package name (required)")
	cmd.Flags().StringArrayVar(&targets, "target", nil, "Name/UID of a target gateway (repeatable, required)")
	cmd.Flags().BoolVar(&accessOnly, "access-only", false, "Install only the Access Control Policy (not the Threat Prevention)")
	return cmd
}

func newPolicyVerifyCmd() *cobra.Command {
	var pkg string
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify a policy package for errors before installing",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg == "" {
				return fmt.Errorf("--package is required")
			}
			return callAndPrint("verify-policy", map[string]interface{}{"policy-package": pkg}, true, false)
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Policy package name (required)")
	return cmd
}
