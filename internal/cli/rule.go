package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRuleCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "rule",
		Short: "Regras de Access Control (access-rule)",
	}
	root.AddCommand(
		newRuleAddCmd(),
		newRuleShowCmd(),
		newRuleSetCmd(),
		newRuleDeleteCmd(),
		newRuleListCmd(),
		newLayerListCmd(),
	)
	return root
}

func newRuleAddCmd() *cobra.Command {
	var layer, position, name, action, comments string
	var fields []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Adiciona uma regra em uma camada (layer) de Access Control",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer é obrigatório (ex: "Network")`)
			}
			payload, err := parseFields(fields)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			payload["position"] = position
			if name != "" {
				payload["name"] = name
			}
			if action != "" {
				payload["action"] = action
			}
			if comments != "" {
				payload["comments"] = comments
			}
			return callAndPrint("add-access-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "Nome/UID da camada de política (obrigatório)")
	cmd.Flags().StringVar(&position, "position", "top", `Posição: "top", "bottom", número ou "above:<uid>"/"below:<uid>"`)
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra")
	cmd.Flags().StringVar(&action, "action", "", "Ação: accept, drop, reject, ...")
	cmd.Flags().StringVar(&comments, "comments", "", "Comentário da regra")
	cmd.Flags().StringArrayVar(&fields, "field", nil, `Campo chave=valor (ex: --field source='["any"]' --field service='["https"]' --field destination='["any"]')`)
	return cmd
}

func newRuleShowCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Mostra uma regra (por --uid, ou --name + --layer)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			if uid == "" && layer != "" {
				payload["layer"] = layer
			}
			return callAndPrint("show-access-rule", payload, false, false)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra")
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&layer, "layer", "", "Camada da regra (obrigatório se identificar por --name)")
	return cmd
}

func newRuleSetCmd() *cobra.Command {
	var name, uid, layer string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Altera uma regra existente (por --uid, ou --name + --layer)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			if uid == "" && layer != "" {
				payload["layer"] = layer
			}
			extra, err := parseFields(fields)
			if err != nil {
				return err
			}
			for k, v := range extra {
				payload[k] = v
			}
			return callAndPrint("set-access-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra")
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&layer, "layer", "", "Camada da regra (obrigatório se identificar por --name)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "Campo chave=valor a alterar (repetível)")
	return cmd
}

func newRuleDeleteCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Apaga uma regra (por --uid, ou --name + --layer)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			if uid == "" && layer != "" {
				payload["layer"] = layer
			}
			return callAndPrint("delete-access-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra")
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&layer, "layer", "", "Camada da regra (obrigatório se identificar por --name)")
	return cmd
}

func newRuleListCmd() *cobra.Command {
	var layer, detailsLevel string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista as regras (rulebase) de uma camada",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf("--layer é obrigatório")
			}
			payload := map[string]interface{}{"name": layer}
			return listAndPrint("show-access-rulebase", detailsLevel, "rulebase", payload)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "Nome/UID da camada (obrigatório)")
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Nível de detalhe: uid | standard | full")
	return cmd
}

func newLayerListCmd() *cobra.Command {
	var detailsLevel string
	cmd := &cobra.Command{
		Use:   "layers",
		Short: "Lista as camadas (access layers) de Access Control disponíveis",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAndPrint("show-access-layers", detailsLevel, "objects", map[string]interface{}{})
		},
	}
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Nível de detalhe: uid | standard | full")
	return cmd
}
