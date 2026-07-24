package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPolicyCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "policy",
		Short: "Instala ou verifica um pacote de política nos gateways",
	}
	root.AddCommand(
		newPolicyInstallCmd(),
		newPolicyVerifyCmd(),
		newPolicyPackageCmd(),
	)
	return root
}

// newPolicyPackageCmd manages policy packages (the containers that hold the
// Access Control and Threat Prevention rulebases installed on gateways).
func newPolicyPackageCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "package",
		Short: "Pacotes de política (policy-package)",
	}

	var addName string
	var addFields []string
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Cria um pacote de política (blades: --field access=true --field threat-prevention=true)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if addName == "" {
				return fmt.Errorf("--name é obrigatório")
			}
			payload, err := parseFields(addFields)
			if err != nil {
				return err
			}
			payload["name"] = addName
			return callAndPrint("add-package", payload, true, true)
		},
	}
	addCmd.Flags().StringVar(&addName, "name", "", "Nome do pacote (obrigatório)")
	addCmd.Flags().StringArrayVar(&addFields, "field", nil, "Campo chave=valor (repetível). Ex: --field access=true --field nat=true")

	var showName, showUID string
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Mostra um pacote de política",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(showName, showUID)
			if err != nil {
				return err
			}
			return callAndPrint("show-package", payload, false, false)
		},
	}
	showCmd.Flags().StringVar(&showName, "name", "", "Nome do pacote")
	showCmd.Flags().StringVar(&showUID, "uid", "", "UID do pacote")

	var setName, setUID string
	var setFields []string
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Altera um pacote de política existente",
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
			return callAndPrint("set-package", payload, true, true)
		},
	}
	setCmd.Flags().StringVar(&setName, "name", "", "Nome do pacote")
	setCmd.Flags().StringVar(&setUID, "uid", "", "UID do pacote")
	setCmd.Flags().StringArrayVar(&setFields, "field", nil, "Campo chave=valor a alterar (repetível)")

	var delName, delUID string
	delCmd := &cobra.Command{
		Use:   "delete",
		Short: "Apaga um pacote de política",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := nameOrUIDPayload(delName, delUID)
			if err != nil {
				return err
			}
			return callAndPrint("delete-package", payload, true, true)
		},
	}
	delCmd.Flags().StringVar(&delName, "name", "", "Nome do pacote")
	delCmd.Flags().StringVar(&delUID, "uid", "", "UID do pacote")

	var listDetails string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lista os pacotes de política",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAndPrint("show-packages", listDetails, "packages", map[string]interface{}{})
		},
	}
	listCmd.Flags().StringVar(&listDetails, "details-level", "standard", "Nível de detalhe: uid | standard | full")

	root.AddCommand(addCmd, showCmd, setCmd, delCmd, listCmd)
	return root
}

func newPolicyInstallCmd() *cobra.Command {
	var pkg string
	var targets []string
	var accessOnly bool
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Instala um pacote de política nos gateways de destino (espera a conclusão)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg == "" {
				return fmt.Errorf("--package é obrigatório")
			}
			if len(targets) == 0 {
				return fmt.Errorf("informe pelo menos um --target (nome/uid do gateway)")
			}
			payload := map[string]interface{}{
				"policy-package": pkg,
				"targets":        targets,
			}
			if accessOnly {
				payload["access"] = true
			}
			return callAndPrint("install-policy", payload, true, false)
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Nome do pacote de política (obrigatório)")
	cmd.Flags().StringArrayVar(&targets, "target", nil, "Nome/UID de um gateway de destino (repetível, obrigatório)")
	cmd.Flags().BoolVar(&accessOnly, "access-only", false, "Instala apenas a Access Control Policy (não a Threat Prevention)")
	return cmd
}

func newPolicyVerifyCmd() *cobra.Command {
	var pkg string
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verifica se um pacote de política tem erros antes de instalar",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkg == "" {
				return fmt.Errorf("--package é obrigatório")
			}
			return callAndPrint("verify-policy", map[string]interface{}{"policy-package": pkg}, true, false)
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Nome do pacote de política (obrigatório)")
	return cmd
}
