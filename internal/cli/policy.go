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
	)
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
