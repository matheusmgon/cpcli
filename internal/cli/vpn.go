package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// vpnCommunityType names the two VPN community families the Management API
// exposes: meshed and star topologies. See entitySpec in helpers.go (shared
// with object.go's plain-object commands).
type vpnCommunityType = entitySpec

var (
	vpnMeshed = vpnCommunityType{
		name: "meshed", short: "Comunidades VPN meshed (site-to-site em malha)",
		addCmd: "add-vpn-community-meshed", showCmd: "show-vpn-community-meshed", setCmd: "set-vpn-community-meshed", deleteCmd: "delete-vpn-community-meshed", listCmd: "show-vpn-communities-meshed",
		fieldHint: `gateways (ex: --field gateways='["gw1","gw2"]'), encryption-method`,
	}
	vpnStar = vpnCommunityType{
		name: "star", short: "Comunidades VPN star (hub-and-spoke)",
		addCmd: "add-vpn-community-star", showCmd: "show-vpn-community-star", setCmd: "set-vpn-community-star", deleteCmd: "delete-vpn-community-star", listCmd: "show-vpn-communities-star",
		fieldHint: `center-gateways, satellite-gateways (listas de nomes/uids), encryption-method`,
	}
)

func newVPNCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "vpn",
		Short: "Comunidades VPN site-to-site (meshed e star)",
	}
	root.AddCommand(
		newVPNCommunityCmd(vpnMeshed),
		newVPNCommunityCmd(vpnStar),
	)
	return root
}

func newVPNCommunityCmd(ct vpnCommunityType) *cobra.Command {
	root := &cobra.Command{
		Use:   ct.name,
		Short: ct.short,
	}

	var addName string
	var addFields []string
	addCmd := &cobra.Command{
		Use:   "add",
		Short: fmt.Sprintf("Cria uma comunidade VPN %s (campos comuns: %s)", ct.name, ct.fieldHint),
		RunE: func(cmd *cobra.Command, args []string) error {
			if addName == "" {
				return fmt.Errorf("--name é obrigatório")
			}
			payload, err := parseFields(addFields)
			if err != nil {
				return err
			}
			payload["name"] = addName
			return callAndPrint(ct.addCmd, payload, true, true)
		},
	}
	addCmd.Flags().StringVar(&addName, "name", "", "Nome da comunidade (obrigatório)")
	addCmd.Flags().StringArrayVar(&addFields, "field", nil, "Campo chave=valor (repetível)")

	var showName, showUID string
	showCmd := &cobra.Command{
		Use:   "show",
		Short: fmt.Sprintf("Mostra uma comunidade VPN %s", ct.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(showName, showUID)
			if err != nil {
				return err
			}
			return callAndPrint(ct.showCmd, payload, false, false)
		},
	}
	showCmd.Flags().StringVar(&showName, "name", "", "Nome da comunidade")
	showCmd.Flags().StringVar(&showUID, "uid", "", "UID da comunidade")

	var setName, setUID string
	var setFields []string
	setCmd := &cobra.Command{
		Use:   "set",
		Short: fmt.Sprintf("Altera uma comunidade VPN %s existente", ct.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(setName, setUID)
			if err != nil {
				return err
			}
			extra, err := parseFields(setFields)
			if err != nil {
				return err
			}
			for k, v := range extra {
				payload[k] = v
			}
			return callAndPrint(ct.setCmd, payload, true, true)
		},
	}
	setCmd.Flags().StringVar(&setName, "name", "", "Nome da comunidade")
	setCmd.Flags().StringVar(&setUID, "uid", "", "UID da comunidade")
	setCmd.Flags().StringArrayVar(&setFields, "field", nil, "Campo chave=valor a alterar (repetível)")

	var delName, delUID string
	delCmd := &cobra.Command{
		Use:   "delete",
		Short: fmt.Sprintf("Apaga uma comunidade VPN %s", ct.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(delName, delUID)
			if err != nil {
				return err
			}
			return callAndPrint(ct.deleteCmd, payload, true, true)
		},
	}
	delCmd.Flags().StringVar(&delName, "name", "", "Nome da comunidade")
	delCmd.Flags().StringVar(&delUID, "uid", "", "UID da comunidade")

	var listFilter, listDetails string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("Lista comunidades VPN %s", ct.name),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{}
			if listFilter != "" {
				payload["filter"] = listFilter
			}
			return listAndPrint(ct.listCmd, listDetails, "objects", payload)
		},
	}
	listCmd.Flags().StringVar(&listFilter, "filter", "", "Texto de busca (filtro do Check Point)")
	listCmd.Flags().StringVar(&listDetails, "details-level", "standard", "Nível de detalhe: uid | standard | full")

	root.AddCommand(addCmd, showCmd, setCmd, delCmd, listCmd)
	return root
}
