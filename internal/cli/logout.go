package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"cpcli/internal/session"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Encerra a sessão no servidor e apaga a sessão local",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := clientFromSession()
			if err != nil {
				if errors.Is(err, session.ErrNotLoggedIn) {
					return nil
				}
				return err
			}
			// Always clear the local session even if the remote logout call
			// fails (server unreachable, sid already expired, ...) — but
			// warn instead of silently pretending the server-side session
			// was actually closed.
			res, callErr := client.ApiCallSimple("logout", map[string]interface{}{})
			if callErr != nil {
				fmt.Fprintf(os.Stderr, "aviso: falha ao encerrar a sessão no servidor: %v\n", callErr)
			} else if !res.Success {
				fmt.Fprintf(os.Stderr, "aviso: o servidor recusou o logout: %s\n", res.ErrorMsg)
			}
			return session.Clear(activeProfile())
		},
	}
}
