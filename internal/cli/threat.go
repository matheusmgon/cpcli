package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newThreatCmd groups Threat Prevention commands: the rulebase (threat-rule
// on a threat layer), the threat profiles, and listing the threat layers.
func newThreatCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "threat",
		Short: "Threat Prevention (regras, perfis e camadas)",
	}
	root.AddCommand(
		newThreatRuleCmd(),
		newThreatProfileCmd(),
		newThreatLayersCmd(),
	)
	return root
}

// --- Threat rules ---------------------------------------------------------

func newThreatRuleCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "rule",
		Short: "Regras de Threat Prevention (threat-rule) em uma camada",
	}
	root.AddCommand(
		newThreatRuleAddCmd(),
		newThreatRuleShowCmd(),
		newThreatRuleSetCmd(),
		newThreatRuleDeleteCmd(),
		newThreatRuleListCmd(),
	)
	return root
}

func newThreatRuleAddCmd() *cobra.Command {
	var layer, position, name string
	var fields []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Adiciona uma regra em uma camada de Threat Prevention",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf("--layer é obrigatório")
			}
			if name == "" {
				return fmt.Errorf("--name é obrigatório")
			}
			payload, err := parseFields(fields)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			payload["position"] = position
			payload["name"] = name
			return callAndPrint("add-threat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "Nome/UID da camada de Threat Prevention (obrigatório)")
	cmd.Flags().StringVar(&position, "position", "top", `Posição: "top", "bottom", número ou "above:<uid>"/"below:<uid>"`)
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra (obrigatório)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, `Campo chave=valor (ex: --field protected-scope='["any"]' --field track='"Log"')`)
	return cmd
}

func newThreatRuleShowCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Mostra uma regra de Threat Prevention (--layer obrigatório)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer é obrigatório (show-threat-rule exige a camada mesmo com --uid)`)
			}
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			return callAndPrint("show-threat-rule", payload, false, false)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra")
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&layer, "layer", "", "Camada da regra (obrigatório)")
	return cmd
}

func newThreatRuleSetCmd() *cobra.Command {
	var name, uid, layer string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Altera uma regra de Threat Prevention (--layer obrigatório)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer é obrigatório (set-threat-rule exige a camada mesmo com --uid)`)
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
			return callAndPrint("set-threat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra")
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&layer, "layer", "", "Camada da regra (obrigatório)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "Campo chave=valor a alterar (repetível)")
	return cmd
}

func newThreatRuleDeleteCmd() *cobra.Command {
	var name, uid, layer string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Apaga uma regra de Threat Prevention (--layer obrigatório)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf(`--layer é obrigatório (delete-threat-rule exige a camada mesmo com --uid)`)
			}
			payload, err := nameOrUIDPayload(name, uid)
			if err != nil {
				return err
			}
			payload["layer"] = layer
			return callAndPrint("delete-threat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome da regra")
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&layer, "layer", "", "Camada da regra (obrigatório)")
	return cmd
}

func newThreatRuleListCmd() *cobra.Command {
	var layer, detailsLevel string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista as regras (rulebase) de uma camada de Threat Prevention",
		RunE: func(cmd *cobra.Command, args []string) error {
			if layer == "" {
				return fmt.Errorf("--layer é obrigatório")
			}
			payload := map[string]interface{}{"name": layer}
			return listRulebaseAndPrint("show-threat-rulebase", detailsLevel, "rulebase", payload)
		},
	}
	cmd.Flags().StringVar(&layer, "layer", "", "Nome/UID da camada (obrigatório)")
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Nível de detalhe: uid | standard | full")
	return cmd
}

// --- Threat profiles ------------------------------------------------------

func newThreatProfileCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "profile",
		Short: "Perfis de Threat Prevention (threat-profile)",
	}

	var addName string
	var addFields []string
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Cria um perfil de Threat Prevention",
		RunE: func(cmd *cobra.Command, args []string) error {
			if addName == "" {
				return fmt.Errorf("--name é obrigatório")
			}
			payload, err := parseFields(addFields)
			if err != nil {
				return err
			}
			payload["name"] = addName
			return callAndPrint("add-threat-profile", payload, true, true)
		},
	}
	addCmd.Flags().StringVar(&addName, "name", "", "Nome do perfil (obrigatório)")
	addCmd.Flags().StringArrayVar(&addFields, "field", nil, "Campo chave=valor (ex: --field active-protections-performance-impact='\"medium\"')")

	var showName, showUID string
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Mostra um perfil de Threat Prevention",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(showName, showUID)
			if err != nil {
				return err
			}
			return callAndPrint("show-threat-profile", payload, false, false)
		},
	}
	showCmd.Flags().StringVar(&showName, "name", "", "Nome do perfil")
	showCmd.Flags().StringVar(&showUID, "uid", "", "UID do perfil")

	var setName, setUID string
	var setFields []string
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Altera um perfil de Threat Prevention existente",
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
			return callAndPrint("set-threat-profile", payload, true, true)
		},
	}
	setCmd.Flags().StringVar(&setName, "name", "", "Nome do perfil")
	setCmd.Flags().StringVar(&setUID, "uid", "", "UID do perfil")
	setCmd.Flags().StringArrayVar(&setFields, "field", nil, "Campo chave=valor a alterar (repetível)")

	var delName, delUID string
	delCmd := &cobra.Command{
		Use:   "delete",
		Short: "Apaga um perfil de Threat Prevention",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(delName, delUID)
			if err != nil {
				return err
			}
			return callAndPrint("delete-threat-profile", payload, true, true)
		},
	}
	delCmd.Flags().StringVar(&delName, "name", "", "Nome do perfil")
	delCmd.Flags().StringVar(&delUID, "uid", "", "UID do perfil")

	var listDetails string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lista os perfis de Threat Prevention",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAndPrint("show-threat-profiles", listDetails, "objects", map[string]interface{}{})
		},
	}
	listCmd.Flags().StringVar(&listDetails, "details-level", "standard", "Nível de detalhe: uid | standard | full")

	root.AddCommand(addCmd, showCmd, setCmd, delCmd, listCmd)
	return root
}

// --- Threat layers --------------------------------------------------------

func newThreatLayersCmd() *cobra.Command {
	var detailsLevel string
	cmd := &cobra.Command{
		Use:   "layers",
		Short: "Lista as camadas de Threat Prevention disponíveis",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAndPrint("show-threat-layers", detailsLevel, "threat-layers", map[string]interface{}{})
		},
	}
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Nível de detalhe: uid | standard | full")
	return cmd
}
