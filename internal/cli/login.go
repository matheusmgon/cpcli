package cli

import (
	"fmt"
	"os"

	api "github.com/CheckPointSW/cp-mgmt-api-go-sdk/APIFiles"
	"github.com/spf13/cobra"
	"golang.org/x/term"

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

			clientArgs := api.APIClientArgs(port, "", "", server, "", -1, "", insecure, false, "", api.WebContext, api.TimeOut, api.SleepTime, "cpcli", "", -1)
			client := api.APIClient(clientArgs)

			var loginRes api.APIResponse
			var err error
			if apiKey != "" {
				loginRes, err = client.ApiLoginWithApiKey(apiKey, continueSession, domain, readOnly, nil)
			} else {
				password := os.Getenv("CPCLI_PASSWORD")
				if password == "" {
					password, err = promptPassword()
					if err != nil {
						return fmt.Errorf("falha ao ler senha: %w", err)
					}
				}
				loginRes, err = client.ApiLogin(user, password, continueSession, domain, readOnly, nil)
			}
			if err != nil {
				return err
			}
			if !loginRes.Success {
				return fmt.Errorf("login falhou: %s", loginRes.ErrorMsg)
			}

			sess := &session.Session{
				Server:     server,
				Port:       port,
				Sid:        client.GetSessionID(),
				ApiVersion: stringField(loginRes.GetData(), "api-server-version"),
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
	cmd.Flags().IntVar(&port, "port", api.DefaultPort, "Porta da API")
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

func stringField(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}
