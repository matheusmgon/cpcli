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
		Short: "Autentica no Management Server e guarda a sessão localmente",
		RunE: func(cmd *cobra.Command, args []string) error {
			if server == "" {
				return fmt.Errorf("--server é obrigatório")
			}
			if apiKey == "" && user == "" {
				return fmt.Errorf("informe --user (login com usuário/senha) ou --api-key")
			}

			password := ""
			if apiKey == "" {
				password = os.Getenv("CPCLI_PASSWORD")
				if password == "" {
					var perr error
					password, perr = promptPassword()
					if perr != nil {
						return fmt.Errorf("falha ao ler senha: %w", perr)
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
			fmt.Printf("Login efetuado em %s (perfil %q). Sessão salva.\n", server, activeProfile())
			return nil
		},
	}

	cmd.Flags().StringVar(&server, "server", "", "Endereço IP ou hostname do Management Server (obrigatório)")
	cmd.Flags().IntVar(&port, "port", mgmt.DefaultPort, "Porta da API")
	cmd.Flags().StringVar(&user, "user", "", "Usuário administrador (senha via CPCLI_PASSWORD ou prompt interativo)")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "Autentica via API key ao invés de usuário/senha")
	cmd.Flags().StringVar(&domain, "domain", "", "Nome/UID/IP do domínio (Multi-Domain Server)")
	cmd.Flags().BoolVar(&readOnly, "read-only", false, "Login em modo somente leitura")
	cmd.Flags().BoolVar(&continueSession, "continue-last-session", false, "Continua a última sessão ao invés de criar uma nova")
	cmd.Flags().BoolVar(&insecure, "insecure", false, "NÃO verifica o fingerprint TLS do servidor (inseguro — use só em laboratório)")
	return cmd
}

func promptPassword() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("stdin não é um terminal interativo (comum ao rodar via wrapper/CI/pipe) — defina a senha com a variável de ambiente CPCLI_PASSWORD em vez de deixar o prompt pedir")
	}
	fmt.Fprint(os.Stderr, "Senha: ")
	b, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
