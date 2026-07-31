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
		Short: "Command-line client for the Check Point Management API",
		// main.go prints every returned error itself ("error: ..."), so
		// Cobra's own error printing is silenced to avoid showing it twice.
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&profileFlag, "profile", "default", "Session/server profile (lets you manage more than one Management Server)")

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
		newObjectCmd(objTypeAddressRange),
		newObjectCmd(objTypeServiceGroup),
		newObjectCmd(objTypeServiceICMP),
		newObjectCmd(objTypeServiceOther),
		newObjectCmd(objTypeSecurityZone),
		newObjectCmd(objTypeDNSDomain),
		newObjectCmd(objTypeWildcard),
		newObjectCmd(objTypeTag),
		newObjectCmd(objTypeTime),
		newObjectCmd(objTypeDynamicObject),
		newObjectCmd(objTypeAccessRole),
		newObjectCmd(objTypeApplicationSite),
		newRuleCmd(),
		newNatCmd(),
		newPolicyCmd(),
		newVPNCmd(),
		newGatewayCmd(),
		newThreatCmd(),
		newHTTPSCmd(),
		newSearchCmd(),
		newRawCmd(),
	)

	return root.Execute()
}
