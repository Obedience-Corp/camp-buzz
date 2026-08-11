package main

import (
	"fmt"

	"github.com/Obedience-Corp/camp-buzz/internal/assets"
	"github.com/Obedience-Corp/camp-buzz/internal/buzzcli"
	"github.com/Obedience-Corp/camp-buzz/internal/config"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show resolved bind config (secrets redacted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			root := campaignRoot()
			cfg, path, err := config.Resolve(root)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "campaign_root: %s\n", empty(root, "(none)"))
			fmt.Fprintf(out, "bind_file: %s\n", empty(path, "(none)"))
			fmt.Fprintf(out, "relay_url: %s\n", empty(cfg.RelayURL, "(unset)"))
			fmt.Fprintf(out, "channel_id: %s\n", empty(cfg.ChannelID, "(unset)"))
			fmt.Fprintf(out, "festival_id: %s\n", empty(cfg.FestivalID, "(unset)"))
			fmt.Fprintf(out, "festival_path: %s\n", empty(cfg.FestivalPath, "(unset)"))
			if buzzcli.HasPrivateKey() {
				fmt.Fprintln(out, "BUZZ_PRIVATE_KEY: [set]")
			} else {
				fmt.Fprintln(out, "BUZZ_PRIVATE_KEY: [unset]")
			}
			ar, reason := assets.Root()
			fmt.Fprintf(out, "asset_root: %s (%s)\n", ar, reason)
			return nil
		},
	}
}

func empty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
