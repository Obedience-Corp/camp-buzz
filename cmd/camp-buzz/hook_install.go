package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Obedience-Corp/camp-buzz/internal/assets"
	"github.com/spf13/cobra"
)

func newHookInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook-install",
		Short: "Print recommended fest hook YAML for Buzz status posts",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, _ := assets.Root()
			candidates := []string{
				filepath.Join(root, "templates", "fest-hooks.example.yaml"),
				filepath.Join(root, "fest-hooks.example.yaml"),
				"assets/templates/fest-hooks.example.yaml",
			}
			for _, p := range candidates {
				data, err := os.ReadFile(p)
				if err == nil {
					fmt.Fprint(cmd.OutOrStdout(), string(data))
					fmt.Fprintf(cmd.ErrOrStderr(), "# source: %s\n", p)
					return nil
				}
			}
			// embedded fallback
			fmt.Fprint(cmd.OutOrStdout(), `# Example fest hooks — enable when camp-buzz doctor is clean.
# fest takes a single command string and defaults to fail: closed;
# fail: open keeps Buzz outages from ever blocking fest.
hooks:
  definitions:
    buzz_status:
      command: camp buzz post --from-hook
      fail: open
      timeout: 30s
# Bind in a task document: hooks: { post: [buzz_status] }
`)
			fmt.Fprintln(cmd.ErrOrStderr(), "# source: embedded fallback (install plugin assets for full templates)")
			return nil
		},
	}
}
