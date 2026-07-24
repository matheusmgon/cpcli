package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"cpcli/internal/mgmt"
)

// newGatewayCmd groups the gateway/server commands. Listing goes through
// "show-gateways-and-servers" (which returns every gateway, cluster and
// server at once), while add/show/set/delete operate on "simple-gateway"
// objects — the standalone gateway type most labs and edge deployments use.
func newGatewayCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "gateway",
		Short: "Gateways e servidores (listar todos; CRUD de simple-gateway)",
	}
	root.AddCommand(
		newGatewayListCmd(),
		newGatewayShowCmd(),
		newGatewayAddCmd(),
		newGatewaySetCmd(),
		newGatewayDeleteCmd(),
		newGatewayInterfaceCmd(),
	)
	return root
}

// newGatewayInterfaceCmd manages per-interface topology settings
// (anti-spoofing, topology) on a simple-gateway. set-simple-gateway replaces
// the whole "interfaces" array rather than patching one entry — confirmed
// against a live Management Server: sending a single interface silently
// deleted every other one. So "set" always reads the gateway's current
// interfaces first, merges the requested fields into the matching entry,
// and sends the complete array back.
func newGatewayInterfaceCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "interface",
		Short: "Interfaces de um gateway standalone (topologia, anti-spoofing)",
	}
	root.AddCommand(
		newGatewayInterfaceListCmd(),
		newGatewayInterfaceSetCmd(),
	)
	return root
}

func newGatewayInterfaceListCmd() *cobra.Command {
	var gateway string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista as interfaces de um gateway (IP, anti-spoofing, topologia)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if gateway == "" {
				return fmt.Errorf("--gateway é obrigatório")
			}
			ifaces, err := fetchGatewayInterfaces(gateway)
			if err != nil {
				return err
			}
			return printJSON(ifaces)
		},
	}
	cmd.Flags().StringVar(&gateway, "gateway", "", "Nome do gateway (obrigatório)")
	return cmd
}

func newGatewayInterfaceSetCmd() *cobra.Command {
	var gateway, iface string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Altera uma interface do gateway (ex: --field anti-spoofing=true)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if gateway == "" || iface == "" {
				return fmt.Errorf("--gateway e --interface são obrigatórios")
			}
			extra, err := parseFields(fields)
			if err != nil {
				return err
			}
			ifaces, err := fetchGatewayInterfaces(gateway)
			if err != nil {
				return err
			}
			updated, found := mgmt.MergeGatewayInterface(ifaces, iface, extra)
			if !found {
				return fmt.Errorf("interface %q não encontrada no gateway %q", iface, gateway)
			}
			return callAndPrint("set-simple-gateway", map[string]interface{}{
				"name":       gateway,
				"interfaces": updated,
			}, true, true)
		},
	}
	cmd.Flags().StringVar(&gateway, "gateway", "", "Nome do gateway (obrigatório)")
	cmd.Flags().StringVar(&iface, "interface", "", `Nome da interface, ex: "eth0" (obrigatório)`)
	cmd.Flags().StringArrayVar(&fields, "field", nil, `Campo chave=valor (ex: --field anti-spoofing=true --field topology='"external"')`)
	return cmd
}

// fetchGatewayInterfaces returns the current "interfaces" array of a
// simple-gateway, as-is from the API (each entry keeps every field the
// server returned, so a later merge+set round-trips them unchanged).
func fetchGatewayInterfaces(gateway string) ([]interface{}, error) {
	client, _, err := clientFromSession()
	if err != nil {
		return nil, err
	}
	data, err := client.Call("show-simple-gateway", map[string]interface{}{
		"name":          gateway,
		"details-level": "full",
	}, false)
	if err != nil {
		return nil, err
	}
	ifaces, _ := data["interfaces"].([]interface{})
	return ifaces, nil
}

func newGatewayListCmd() *cobra.Command {
	var filter, detailsLevel string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista todos os gateways e servidores gerenciados",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{}
			if filter != "" {
				payload["filter"] = filter
			}
			return listAndPrint("show-gateways-and-servers", detailsLevel, "objects", payload)
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Texto de busca (filtro do Check Point)")
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Nível de detalhe: uid | standard | full")
	return cmd
}

func newGatewayShowCmd() *cobra.Command {
	var name, uid string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Mostra um gateway standalone (simple-gateway)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["details-level"] = "full"
			return callAndPrint("show-simple-gateway", payload, false, false)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome do gateway")
	cmd.Flags().StringVar(&uid, "uid", "", "UID do gateway")
	return cmd
}

func newGatewayAddCmd() *cobra.Command {
	var name, ip string
	var fields []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Cria um gateway standalone (blades comuns: --field firewall=true, --field vpn=true)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name é obrigatório")
			}
			payload, err := parseFields(fields)
			if err != nil {
				return err
			}
			payload["name"] = name
			if ip != "" {
				payload["ip-address"] = ip
			}
			return callAndPrint("add-simple-gateway", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome do gateway (obrigatório)")
	cmd.Flags().StringVar(&ip, "ip", "", "Endereço IP do gateway")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "Campo chave=valor (repetível). Ex: --field firewall=true --field application-control=true")
	return cmd
}

func newGatewaySetCmd() *cobra.Command {
	var name, uid string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Altera um gateway standalone existente",
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
			return callAndPrint("set-simple-gateway", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome do gateway")
	cmd.Flags().StringVar(&uid, "uid", "", "UID do gateway")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "Campo chave=valor a alterar (repetível)")
	return cmd
}

func newGatewayDeleteCmd() *cobra.Command {
	var name, uid string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Apaga um gateway standalone",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			return callAndPrint("delete-simple-gateway", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome do gateway")
	cmd.Flags().StringVar(&uid, "uid", "", "UID do gateway")
	return cmd
}
