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
		Short: "End the server session and remove the local session",
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
			if err := client.Logout(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to end the server session: %v\n", err)
			}
			return session.Clear(activeProfile())
		},
	}
}
