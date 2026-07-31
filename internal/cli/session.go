package cli

import (
	"github.com/spf13/cobra"
)

func newSessionCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "session",
		Short: "Publish, discard or show the current change session",
	}
	root.AddCommand(
		&cobra.Command{
			Use:   "publish",
			Short: "Publish pending changes (required after add/set/delete)",
			RunE: func(cmd *cobra.Command, args []string) error {
				return callAndPrint("publish", map[string]interface{}{}, true, false)
			},
		},
		&cobra.Command{
			Use:   "discard",
			Short: "Discard pending unpublished changes",
			RunE: func(cmd *cobra.Command, args []string) error {
				return callAndPrint("discard", map[string]interface{}{}, false, false)
			},
		},
		&cobra.Command{
			Use:   "show",
			Short: "Show details of the current session",
			RunE: func(cmd *cobra.Command, args []string) error {
				// show-session does not accept "details-level" (returns
				// generic_err_invalid_parameter_name); it already returns the
				// current session in full with no parameters at all.
				return callAndPrint("show-session", map[string]interface{}{}, false, false)
			},
		},
	)
	return root
}
