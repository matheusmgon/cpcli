package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newNatCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "nat",
		Short: "Regras de NAT (nat-rule)",
	}
	root.AddCommand(
		newNatAddCmd(),
		newNatShowCmd(),
		newNatSetCmd(),
		newNatDeleteCmd(),
		newNatListCmd(),
	)
	return root
}

func newNatAddCmd() *cobra.Command {
	var pkg, position, comments string
	var fields []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Adiciona uma regra de NAT manual a um pacote de política",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg == "" {
				return fmt.Errorf("--package é obrigatório (nome/uid do pacote de política)")
			}
			payload, err := parseFields(fields)
			if err != nil {
				return err
			}
			payload["package"] = pkg
			payload["position"] = position
			if comments != "" {
				payload["comments"] = comments
			}
			return callAndPrint("add-nat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Nome/UID do pacote de política (obrigatório)")
	cmd.Flags().StringVar(&position, "position", "top", `Posição: "top", "bottom" ou número`)
	cmd.Flags().StringVar(&comments, "comments", "", "Comentário da regra")
	cmd.Flags().StringArrayVar(&fields, "field", nil, `Campo chave=valor (ex: --field original-source='"host1"' --field translated-source='"nat-host1"' --field method='"static"')`)
	return cmd
}

func newNatShowCmd() *cobra.Command {
	var uid, pkg, ruleNumber string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Mostra uma regra de NAT (por --uid, ou --package + --rule-number)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{}
			if uid != "" {
				payload["uid"] = uid
			} else {
				if pkg == "" || ruleNumber == "" {
					return fmt.Errorf("informe --uid, ou --package + --rule-number")
				}
				payload["package"] = pkg
				payload["rule-number"] = ruleNumber
			}
			return callAndPrint("show-nat-rule", payload, false, false)
		},
	}
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&pkg, "package", "", "Nome/UID do pacote de política")
	cmd.Flags().StringVar(&ruleNumber, "rule-number", "", "Número da regra dentro do pacote")
	return cmd
}

func newNatSetCmd() *cobra.Command {
	var uid, pkg, ruleNumber string
	var fields []string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Altera uma regra de NAT existente (por --uid, ou --package + --rule-number)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{}
			if uid != "" {
				payload["uid"] = uid
			} else {
				if pkg == "" || ruleNumber == "" {
					return fmt.Errorf("informe --uid, ou --package + --rule-number")
				}
				payload["package"] = pkg
				payload["rule-number"] = ruleNumber
			}
			extra, err := parseFields(fields)
			if err != nil {
				return err
			}
			for k, v := range extra {
				payload[k] = v
			}
			return callAndPrint("set-nat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&pkg, "package", "", "Nome/UID do pacote de política")
	cmd.Flags().StringVar(&ruleNumber, "rule-number", "", "Número da regra dentro do pacote")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "Campo chave=valor a alterar (repetível)")
	return cmd
}

func newNatDeleteCmd() *cobra.Command {
	var uid, pkg, ruleNumber string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Apaga uma regra de NAT (--package obrigatório; --uid ou --rule-number)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg == "" {
				return fmt.Errorf("--package é obrigatório (delete-nat-rule exige o pacote mesmo com --uid)")
			}
			payload := map[string]interface{}{"package": pkg}
			if uid != "" {
				payload["uid"] = uid
			} else {
				if ruleNumber == "" {
					return fmt.Errorf("informe --uid, ou --rule-number")
				}
				payload["rule-number"] = ruleNumber
			}
			return callAndPrint("delete-nat-rule", payload, true, true)
		},
	}
	cmd.Flags().StringVar(&uid, "uid", "", "UID da regra")
	cmd.Flags().StringVar(&pkg, "package", "", "Nome/UID do pacote de política (obrigatório)")
	cmd.Flags().StringVar(&ruleNumber, "rule-number", "", "Número da regra dentro do pacote")
	return cmd
}

func newNatListCmd() *cobra.Command {
	var pkg, detailsLevel string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista as regras de NAT (rulebase) de um pacote de política",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg == "" {
				return fmt.Errorf("--package é obrigatório")
			}
			payload := map[string]interface{}{"package": pkg}
			return listRulebaseAndPrint("show-nat-rulebase", detailsLevel, "rulebase", payload)
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Nome/UID do pacote de política (obrigatório)")
	cmd.Flags().StringVar(&detailsLevel, "details-level", "standard", "Nível de detalhe: uid | standard | full")
	return cmd
}
