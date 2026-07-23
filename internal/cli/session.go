package cli

import (
	"github.com/spf13/cobra"
)

func newSessionCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "session",
		Short: "Publica, descarta ou mostra a sessão de mudanças atual",
	}
	root.AddCommand(
		&cobra.Command{
			Use:   "publish",
			Short: "Publica as mudanças pendentes (necessário após add/set/delete)",
			RunE: func(cmd *cobra.Command, args []string) error {
				return callAndPrint("publish", map[string]interface{}{}, true, false)
			},
		},
		&cobra.Command{
			Use:   "discard",
			Short: "Descarta as mudanças pendentes não publicadas",
			RunE: func(cmd *cobra.Command, args []string) error {
				return callAndPrint("discard", map[string]interface{}{}, false, false)
			},
		},
		&cobra.Command{
			Use:   "show",
			Short: "Mostra detalhes da sessão atual",
			RunE: func(cmd *cobra.Command, args []string) error {
				// show-session não aceita "details-level" (retorna
				// generic_err_invalid_parameter_name); ele já devolve a
				// sessão atual por completo sem parâmetro nenhum.
				return callAndPrint("show-session", map[string]interface{}{}, false, false)
			},
		},
	)
	return root
}
