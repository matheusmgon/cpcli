// Package cli wires the cpcli Cobra commands on top of the Check Point
// Management API Go SDK (github.com/CheckPointSW/cp-mgmt-api-go-sdk).
package cli

import (
	"github.com/spf13/cobra"
)

var profileFlag string

func activeProfile() string {
	if profileFlag == "" {
		return "default"
	}
	return profileFlag
}

// Execute builds and runs the root cpcli command.
func Execute() error {
	root := &cobra.Command{
		Use:   "cpcli",
		Short: "Cliente de linha de comando para o Check Point Management API",
		// main.go prints every returned error itself ("erro: ..."), so
		// Cobra's own error printing is silenced to avoid showing it twice.
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&profileFlag, "profile", "default", "Perfil de sessão/servidor (permite gerenciar mais de um Management Server)")

	root.AddCommand(
		newLoginCmd(),
		newLogoutCmd(),
		newSessionCmd(),
		newTaskCmd(),
		newObjectCmd(objTypeHost),
		newObjectCmd(objTypeNetwork),
		newObjectCmd(objTypeGroup),
		newObjectCmd(objTypeServiceTCP),
		newObjectCmd(objTypeServiceUDP),
		newRuleCmd(),
		newNatCmd(),
		newPolicyCmd(),
		newVPNCmd(),
		newRawCmd(),
	)

	return root.Execute()
}
