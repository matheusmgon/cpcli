package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"cpcli/internal/mgmt"
	"cpcli/internal/session"
)

func newLoginCmd() *cobra.Command {
	var (
		server          string
		port            int
		user            string
		apiKey          string
		domain          string
		readOnly        bool
		continueSession bool
		insecure        bool
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to the Management Server and store the session locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			if server == "" {
				return fmt.Errorf("--server is required")
			}
			if apiKey == "" && user == "" {
				return fmt.Errorf("provide --user (username/password login) or --api-key")
			}

			password := ""
			if apiKey == "" {
				password = os.Getenv("CPCLI_PASSWORD")
				if password == "" {
					var perr error
					password, perr = promptPassword()
					if perr != nil {
						return fmt.Errorf("failed to read password: %w", perr)
					}
				}
			}

			_, res, err := mgmt.Login(mgmt.LoginOptions{
				Server:          server,
				Port:            port,
				User:            user,
				Password:        password,
				APIKey:          apiKey,
				Domain:          domain,
				ReadOnly:        readOnly,
				ContinueSession: continueSession,
				Insecure:        insecure,
			})
			if err != nil {
				return err
			}

			sess := &session.Session{
				Server:     server,
				Port:       port,
				Sid:        res.Sid,
				ApiVersion: res.APIVersion,
				Domain:     domain,
				User:       user,
				ReadOnly:   readOnly,
				Insecure:   insecure,
			}
			if err := session.Save(activeProfile(), sess); err != nil {
				return err
			}
			fmt.Printf("Logged in to %s (profile %q). Session saved.\n", server, activeProfile())
			return nil
		},
	}

	cmd.Flags().StringVar(&server, "server", "", "IP address or hostname of the Management Server (required)")
	cmd.Flags().IntVar(&port, "port", mgmt.DefaultPort, "API port")
	cmd.Flags().StringVar(&user, "user", "", "Administrator user (password via CPCLI_PASSWORD env or interactive prompt)")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "Authenticate via API key instead of username/password")
	cmd.Flags().StringVar(&domain, "domain", "", "Domain name/UID/IP (Multi-Domain Server)")
	cmd.Flags().BoolVar(&readOnly, "read-only", false, "Log in in read-only mode")
	cmd.Flags().BoolVar(&continueSession, "continue-last-session", false, "Continue the last session instead of creating a new one")
	cmd.Flags().BoolVar(&insecure, "insecure", false, "do NOT verify the server's TLS fingerprint (insecure — lab use only)")
	return cmd
}

func promptPassword() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("stdin is not an interactive terminal (common when running via wrapper/CI/pipe) — set the password with the CPCLI_PASSWORD environment variable instead of relying on the prompt")
	}
	fmt.Fprint(os.Stderr, "Password: ")
	b, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
