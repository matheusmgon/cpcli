package cli

import (
	"github.com/spf13/cobra"
)

func newTaskCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "task <task-id>",
		Short: "Show the status of an async task (e.g. install-policy, publish)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{"task-id": args[0], "details-level": "full"}
			return callAndPrint("show-task", payload, false, false)
		},
	}
}
