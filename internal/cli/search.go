package cli

import (
	"github.com/spf13/cobra"
)

// newSearchCmd groups cross-type lookups: a generic object search
// (show-objects) and a reference finder (where-used).
func newSearchCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "find",
		Short: "Busca objetos de qualquer tipo (show-objects) e referências (where-used)",
		RunE:  runFindObjects,
	}
	root.Flags().String("filter", "", "Texto de busca (filtro do Check Point)")
	root.Flags().String("type", "", "Restringe a um tipo de objeto (ex: host, network, service-tcp)")
	root.Flags().String("details-level", "standard", "Nível de detalhe: uid | standard | full")

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
		Short: "Mostra onde um objeto é referenciado (regras, grupos, etc.)",
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
	cmd.Flags().StringVar(&name, "name", "", "Nome do objeto")
	cmd.Flags().StringVar(&uid, "uid", "", "UID do objeto")
	cmd.Flags().BoolVar(&indirect, "indirect", false, "Inclui referências indiretas (via grupos)")
	return cmd
}
