package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Obedience-Corp/camp-buzz/internal/assets"
	"github.com/Obedience-Corp/camp-buzz/internal/buzzcli"
	"github.com/Obedience-Corp/camp-buzz/internal/config"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check install, buzz CLI, env, and bind config",
		Long:  "Exit 0 if ready to post; non-zero with actionable fixes. Never prints secrets.",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ok := true

			exe, err := os.Executable()
			if err != nil {
				fmt.Fprintf(out, "binary: unknown (%v)\n", err)
				ok = false
			} else {
				fmt.Fprintf(out, "binary: %s\n", filepath.Clean(exe))
			}

			root := campaignRoot()
			if root == "" {
				fmt.Fprintln(out, "campaign root: not found (set CAMP_ROOT or run inside a campaign)")
			} else {
				fmt.Fprintf(out, "campaign root: %s\n", root)
			}

			assetRoot, reason := assets.Root()
			fmt.Fprintf(out, "asset root: %s (%s)\n", assetRoot, reason)

			if p, err := buzzcli.LookPath(); err != nil {
				fmt.Fprintf(out, "buzz CLI: MISSING — %v\n", err)
				ok = false
			} else {
				fmt.Fprintf(out, "buzz CLI: %s\n", p)
			}

			if buzzcli.HasPrivateKey() {
				fmt.Fprintln(out, "BUZZ_PRIVATE_KEY: set")
			} else {
				fmt.Fprintln(out, "BUZZ_PRIVATE_KEY: NOT set (required to post)")
				ok = false
			}

			cfg, bindPath, err := config.Resolve(root)
			if err != nil {
				fmt.Fprintf(out, "bind config: error — %v\n", err)
				ok = false
			} else {
				if bindPath != "" {
					if _, err := os.Stat(bindPath); err == nil {
						fmt.Fprintf(out, "bind file: %s\n", bindPath)
					} else {
						fmt.Fprintf(out, "bind file: %s (absent — use camp buzz bind)\n", bindPath)
					}
				} else {
					fmt.Fprintln(out, "bind file: n/a (no campaign root)")
				}
				if cfg.RelayURL == "" {
					fmt.Fprintln(out, "relay_url: NOT set (BUZZ_RELAY_URL or bind file)")
					ok = false
				} else {
					fmt.Fprintf(out, "relay_url: %s\n", cfg.RelayURL)
				}
				if cfg.ChannelID == "" {
					fmt.Fprintln(out, "channel_id: NOT set (BUZZ_CHANNEL or bind file)")
					ok = false
				} else {
					fmt.Fprintf(out, "channel_id: %s\n", cfg.ChannelID)
				}
				if cfg.FestivalID != "" {
					fmt.Fprintf(out, "festival_id: %s\n", cfg.FestivalID)
				}
			}

			// camp discovery hint
			if _, err := exec.LookPath("camp"); err != nil {
				fmt.Fprintln(out, "camp: not on PATH (optional for direct camp-buzz use)")
			} else {
				fmt.Fprintln(out, "camp: found — after install, invoke as: camp buzz …")
			}

			if ok {
				fmt.Fprintln(out, "status: ready")
				return nil
			}
			fmt.Fprintln(out, "status: not ready")
			return fmt.Errorf("doctor found issues (see above)")
		},
	}
}
