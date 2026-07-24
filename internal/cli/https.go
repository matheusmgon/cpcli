package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newHTTPSCmd groups HTTPS Inspection commands: the rulebase (https-rule on
// an https-layer) and listing the https layers. Mirrors the shape of
// newThreatCmd/newRuleCmd — same rulebase-under-a-layer pattern the
// Management API uses for every blade with its own inspection layer.
func newHTTPSCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "https",
		Short: "HTTPS Inspection (regras e camadas)",
	}
	root.AddCommand(
		newHTTPSRuleCmd(),
		newHTTPSLayersCmd(),
	)
	return root
}

func newHTTPSRuleCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "rule",
		Short: "Regras de HTTPS Inspection (https-rule) em uma camada",
	}
	root.AddCommand(
		newHTTPSRuleAddCmd(),
		newHTTPSRuleShowCmd(),
		newHTTPSRuleSetCmd(),
		newHTTPSRuleDeleteCmd(),
		newHTTPSRuleListCmd(),
	)
	return root
}

func newHTTPSRuleAddCmd() *cobra.Command {
	var layer, position, name string
	var fields []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Adiciona uma regra em uma camada de HTTPS Inspection",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf("--layer é obrigatório")
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
			return callAndPrint("add-https-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "Nome/UID da camada de HTTPS Inspection (obrigatório)")
	cmd.Flags().StringVar(&position, "position", "top", `Posição: "top", "bottom", número ou "above:<uid>"/"below:<uid>"`)
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra")
	cmd.Flags().StringArrayVar(&fields, "field", nil, `Campo chave=valor (ex: --field source='["any"]' --field action='"Inspect"')`)
	return cmd
}

// HTTPS Inspection's show/set/delete all require "layer" even when the rule
// is identified by "uid" — confirmed against a live Management Server
// (generic_err_missing_required_parameters: [layer] otherwise). Same
// requirement threat-rule has; unlike access-rule, where uid alone is
// enough for show/set.

func newHTTPSRuleShowCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Mostra uma regra de HTTPS Inspection (--layer obrigatório)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer é obrigatório (show-https-rule exige a camada mesmo com --uid)`)
			}
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			return callAndPrint("show-https-rule", payload, false, false)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra")
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&layer, "layer", "", "Camada da regra (obrigatório)")
	return cmd
}

func newHTTPSRuleSetCmd() *cobra.Command {
	var name, uid, layer string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Altera uma regra de HTTPS Inspection (--layer obrigatório)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer é obrigatório (set-https-rule exige a camada mesmo com --uid)`)
			}
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			extra, err := parseFields(fields)
			if err != nil {
				return err
			}
			for k, v := range extra {
				payload[k] = v
			}
			return callAndPrint("set-https-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra")
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&layer, "layer", "", "Camada da regra (obrigatório)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "Campo chave=valor a alterar (repetível)")
	return cmd
}

func newHTTPSRuleDeleteCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Apaga uma regra de HTTPS Inspection (--layer obrigatório)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer é obrigatório (delete-https-rule exige a camada mesmo com --uid)`)
			}
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			return callAndPrint("delete-https-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra")
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&layer, "layer", "", "Camada da regra (obrigatório)")
	return cmd
}

func newHTTPSRuleListCmd() *cobra.Command {
	var layer, detailsLevel string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista as regras (rulebase) de uma camada de HTTPS Inspection",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf("--layer é obrigatório")
			}
			payload := map[string]interface{}{"name": layer}
			return listRulebaseAndPrint("show-https-rulebase", detailsLevel, "rulebase", payload)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "Nome/UID da camada (obrigatório)")
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Nível de detalhe: uid | standard | full")
	return cmd
}

func newHTTPSLayersCmd() *cobra.Command {
	var detailsLevel string
	cmd := &cobra.Command{
		Use:   "layers",
		Short: "Lista as camadas de HTTPS Inspection disponíveis",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAndPrint("show-https-layers", detailsLevel, "https-layers", map[string]interface{}{})
		},
	}
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Nível de detalhe: uid | standard | full")
	return cmd
}
