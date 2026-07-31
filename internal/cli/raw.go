package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"cpcli/internal/config"
)

// newRawCmd is the escape hatch for any Management API command not modeled
// by a dedicated subcommand — the API surface has hundreds of commands and
// this project only wraps the common ones explicitly.
func newRawCmd() *cobra.Command {
	var jsonPayload, filePayload, detailsLevel, containerKey string
	var list bool
	var noWait bool

	cmd := &cobra.Command{
		Use:   "raw <command>",
		Short: "Run any Management API command directly (e.g. show-gateways-and-servers)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := args[0]

			payload, err := rawPayload(jsonPayload, filePayload)
			if err != nil {
				return err
			}

			if list {
				return listAndPrint(command, detailsLevel, containerKey, payload)
			}

			client, _, err := clientFromSession()
			if err != nil {
				return err
			}
			data, err := client.Call(command, payload, !noWait)
			if err != nil {
				return err
			}
			return printData(data)
		},
	}

	cmd.Flags().StringVar(&jsonPayload, "json", "", `JSON body of the command, e.g. --json '{"name":"x"}'`)
	cmd.Flags().StringVar(&filePayload, "file", "", "File containing the JSON body of the command")
	cmd.Flags().BoolVar(&list, "list", false, `Treat the command as a paginated query (e.g. show-*s), aggregating all pages`)
	cmd.Flags().StringVar(&containerKey, "container-key", "objects", `JSON key holding the list, when --list (e.g. "rulebase")`)
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Detail level (--list): uid | standard | full")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "Do not wait for async task completion (returns the task-id)")
	return cmd
}

func rawPayload(jsonPayload, filePayload string) (map[string]interface{}, error) {
	var raw []byte
	switch {
	case jsonPayload != "" && filePayload != "":
		return nil, fmt.Errorf("use --json or --file, not both")
	case jsonPayload != "":
		raw = []byte(jsonPayload)
	case filePayload != "":
		data, err := os.ReadFile(config.ResolvePath(filePayload))
		if err != nil {
			return nil, err
		}
		raw = data
	default:
		return map[string]interface{}{}, nil
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}
	return payload, nil
}
