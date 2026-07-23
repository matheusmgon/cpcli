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
		Short: "Executa qualquer comando da Management API diretamente (ex: show-gateways-and-servers)",
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
			res, err := apiCallWithTimeout(client, command, payload, !noWait)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	}

	cmd.Flags().StringVar(&jsonPayload, "json", "", `Corpo JSON do comando, ex: --json '{"name":"x"}'`)
	cmd.Flags().StringVar(&filePayload, "file", "", "Arquivo contendo o corpo JSON do comando")
	cmd.Flags().BoolVar(&list, "list", false, `Trata o comando como uma consulta paginada (ex: show-*s), agregando todas as páginas`)
	cmd.Flags().StringVar(&containerKey, "container-key", "objects", `Chave do JSON que contém a lista, quando --list (ex: "rulebase")`)
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Nível de detalhe (--list): uid | standard | full")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "Não espera a conclusão de tasks assíncronas (retorna o task-id)")
	return cmd
}

func rawPayload(jsonPayload, filePayload string) (map[string]interface{}, error) {
	var raw []byte
	switch {
	case jsonPayload != "" && filePayload != "":
		return nil, fmt.Errorf("use --json ou --file, não os dois")
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
		return nil, fmt.Errorf("payload inválido: %w", err)
	}
	return payload, nil
}
