package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// objectType describes one Check Point object family. See entitySpec in
// helpers.go (shared with vpn.go's VPN community commands).
type objectType = entitySpec

var (
	objTypeHost = objectType{
		name: "host", short: "Objetos host (endereço IP único)",
		addCmd: "add-host", showCmd: "show-host", setCmd: "set-host", deleteCmd: "delete-host", listCmd: "show-hosts",
		fieldHint: "ip-address, comments, color, groups",
	}
	objTypeNetwork = objectType{
		name: "network", short: "Objetos de rede (subnet)",
		addCmd: "add-network", showCmd: "show-network", setCmd: "set-network", deleteCmd: "delete-network", listCmd: "show-networks",
		fieldHint: "subnet4, mask-length4 (ou subnet-mask), comments, color",
	}
	objTypeGroup = objectType{
		name: "group", short: "Grupos de objetos",
		addCmd: "add-group", showCmd: "show-group", setCmd: "set-group", deleteCmd: "delete-group", listCmd: "show-groups",
		fieldHint: `members (ex: --field members='["host1","host2"]')`,
	}
	objTypeServiceTCP = objectType{
		name: "service-tcp", short: "Serviços TCP",
		addCmd: "add-service-tcp", showCmd: "show-service-tcp", setCmd: "set-service-tcp", deleteCmd: "delete-service-tcp", listCmd: "show-services-tcp",
		fieldHint: "port, comments",
	}
	objTypeServiceUDP = objectType{
		name: "service-udp", short: "Serviços UDP",
		addCmd: "add-service-udp", showCmd: "show-service-udp", setCmd: "set-service-udp", deleteCmd: "delete-service-udp", listCmd: "show-services-udp",
		fieldHint: "port, comments",
	}
)

func newObjectCmd(ot objectType) *cobra.Command {
	root := &cobra.Command{
		Use:   ot.name,
		Short: ot.short,
	}
	root.AddCommand(
		newObjectAddCmd(ot),
		newObjectShowCmd(ot),
		newObjectSetCmd(ot),
		newObjectDeleteCmd(ot),
		newObjectListCmd(ot),
	)
	return root
}

func newObjectAddCmd(ot objectType) *cobra.Command {
	var name string
	var fields []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: fmt.Sprintf("Cria um objeto %s (campos comuns: %s)", ot.name, ot.fieldHint),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name é obrigatório")
			}
			payload, err := parseFields(fields)
			if err != nil {
				return err
			}
			payload["name"] = name
			return callAndPrint(ot.addCmd, payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome do objeto (obrigatório)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "Campo chave=valor (repetível). Valores em JSON válido são interpretados como tal.")
	return cmd
}

func newObjectShowCmd(ot objectType) *cobra.Command {
	var name, uid string
	cmd := &cobra.Command{
		Use:   "show",
		Short: fmt.Sprintf("Mostra um objeto %s", ot.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["details-level"] = "full"
			return callAndPrint(ot.showCmd, payload, false, false)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome do objeto")
	cmd.Flags().StringVar(&uid, "uid", "", "UID do objeto")
	return cmd
}

func newObjectSetCmd(ot objectType) *cobra.Command {
	var name, uid string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: fmt.Sprintf("Altera um objeto %s existente", ot.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			extra, err := parseFields(fields)
			if err != nil {
				return err
			}
			for k, v := range extra {
				payload[k] = v
			}
			return callAndPrint(ot.setCmd, payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome do objeto")
	cmd.Flags().StringVar(&uid, "uid", "", "UID do objeto")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "Campo chave=valor a alterar (repetível)")
	return cmd
}

func newObjectDeleteCmd(ot objectType) *cobra.Command {
	var name, uid string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: fmt.Sprintf("Apaga um objeto %s", ot.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			return callAndPrint(ot.deleteCmd, payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome do objeto")
	cmd.Flags().StringVar(&uid, "uid", "", "UID do objeto")
	return cmd
}

func newObjectListCmd(ot objectType) *cobra.Command {
	var filter, detailsLevel string
	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("Lista objetos %s", ot.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{}
			if filter != "" {
				payload["filter"] = filter
			}
			return listAndPrint(ot.listCmd, detailsLevel, "objects", payload)
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Texto de busca (filtro do Check Point)")
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Nível de detalhe: uid | standard | full")
	return cmd
}
