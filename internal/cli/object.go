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
	objTypeAddressRange = objectType{
		name: "address-range", short: "Faixas de endereços (address-range)",
		addCmd: "add-address-range", showCmd: "show-address-range", setCmd: "set-address-range", deleteCmd: "delete-address-range", listCmd: "show-address-ranges",
		fieldHint: "ip-address-first, ip-address-last, comments, color",
	}
	objTypeServiceGroup = objectType{
		name: "service-group", short: "Grupos de serviços",
		addCmd: "add-service-group", showCmd: "show-service-group", setCmd: "set-service-group", deleteCmd: "delete-service-group", listCmd: "show-service-groups",
		fieldHint: `members (ex: --field members='["https","ssh"]')`,
	}
	objTypeServiceICMP = objectType{
		name: "service-icmp", short: "Serviços ICMP",
		addCmd: "add-service-icmp", showCmd: "show-service-icmp", setCmd: "set-service-icmp", deleteCmd: "delete-service-icmp", listCmd: "show-services-icmp",
		fieldHint: "icmp-type, icmp-code, comments",
	}
	objTypeServiceOther = objectType{
		name: "service-other", short: "Serviços de outros protocolos IP",
		addCmd: "add-service-other", showCmd: "show-service-other", setCmd: "set-service-other", deleteCmd: "delete-service-other", listCmd: "show-services-other",
		fieldHint: "ip-protocol, match, comments",
	}
	objTypeSecurityZone = objectType{
		name: "security-zone", short: "Zonas de segurança (security-zone)",
		addCmd: "add-security-zone", showCmd: "show-security-zone", setCmd: "set-security-zone", deleteCmd: "delete-security-zone", listCmd: "show-security-zones",
		fieldHint: "comments, color",
	}
	objTypeDNSDomain = objectType{
		name: "dns-domain", short: "Domínios DNS (dns-domain)",
		addCmd: "add-dns-domain", showCmd: "show-dns-domain", setCmd: "set-dns-domain", deleteCmd: "delete-dns-domain", listCmd: "show-dns-domains",
		fieldHint: `name (ex: ".example.com"), is-sub-domain (bool)`,
	}
	objTypeWildcard = objectType{
		name: "wildcard", short: "Objetos wildcard (IP + máscara curinga)",
		addCmd: "add-wildcard", showCmd: "show-wildcard", setCmd: "set-wildcard", deleteCmd: "delete-wildcard", listCmd: "show-wildcards",
		fieldHint: "ipv4-address, ipv4-mask-wildcard, comments",
	}
	objTypeTag = objectType{
		name: "tag", short: "Tags (rótulos de objetos)",
		addCmd: "add-tag", showCmd: "show-tag", setCmd: "set-tag", deleteCmd: "delete-tag", listCmd: "show-tags",
		fieldHint: "comments, color",
	}
	objTypeTime = objectType{
		name: "time", short: "Objetos de tempo (time)",
		addCmd: "add-time", showCmd: "show-time", setCmd: "set-time", deleteCmd: "delete-time", listCmd: "show-times",
		fieldHint: "start, end, recurrence (ver docs da API)",
	}
	objTypeDynamicObject = objectType{
		name: "dynamic-object", short: "Objetos dinâmicos (dynamic-object)",
		addCmd: "add-dynamic-object", showCmd: "show-dynamic-object", setCmd: "set-dynamic-object", deleteCmd: "delete-dynamic-object", listCmd: "show-dynamic-objects",
		fieldHint: "comments, color",
	}
	objTypeAccessRole = objectType{
		name: "access-role", short: "Access Roles (identidade: usuários/máquinas/redes)",
		addCmd: "add-access-role", showCmd: "show-access-role", setCmd: "set-access-role", deleteCmd: "delete-access-role", listCmd: "show-access-roles",
		fieldHint: `networks, users, machines (ex: --field networks='["any"]')`,
	}
	objTypeApplicationSite = objectType{
		name: "application-site", short: "Aplicações/sites customizados (application-site)",
		addCmd: "add-application-site", showCmd: "show-application-site", setCmd: "set-application-site", deleteCmd: "delete-application-site", listCmd: "show-application-sites",
		fieldHint: "url-list, application-signature, primary-category, comments",
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
