package cli

import (
	"github.com/spf13/cobra"
)

// newSearchCmd groups cross-type lookups: a generic object search
// (show-objects) and a reference finder (where-used).
func newSearchCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "find",
		Short: "Search objects of any type (show-objects) and references (where-used)",
		RunE:  runFindObjects,
	}
	root.Flags().String("filter", "", "Search text (Check Point filter)")
	root.Flags().String("type", "", "Restrict to an object type (e.g. host, network, service-tcp)")
	root.Flags().String("details-level", "standard", "Detail level: uid | standard | full")

	root.AddCommand(newWhereUsedCmd())
	return root
}

func runFindObjects(cmd *cobra.Command, args []string) error {
	filter, _ := cmd.Flags().GetString("filter")
	objType, _ := cmd.Flags().GetString("type")
	detailsLevel, _ := cmd.Flags().GetString("details-level")

	payload := map[string]interface{}{}
	if filter != "" {
		payload["filter"] = filter
	}
	if objType != "" {
		payload["type"] = objType
	}
	return listAndPrint("show-objects", detailsLevel, "objects", payload)
}

func newWhereUsedCmd() *cobra.Command {
	var name, uid string
	var indirect bool
	cmd := &cobra.Command{
		Use:   "where-used",
		Short: "Show where an object is referenced (rules, groups, etc.)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			if indirect {
				payload["indirect"] = true
			}
			return callAndPrint("where-used", payload, false, false)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Object name")
	cmd.Flags().StringVar(&uid, "uid", "", "Object UID")
	cmd.Flags().BoolVar(&indirect, "indirect", false, "Include indirect references (via groups)")
	return cmd
}
