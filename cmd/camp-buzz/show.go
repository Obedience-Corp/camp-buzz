package main

import (
	"fmt"
	"strconv"

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
			fmt.Fprintf(out, "campaign_root: %s\n", safeDisplay(root, "(none)"))
			fmt.Fprintf(out, "bind_file: %s\n", safeDisplay(path, "(none)"))
			fmt.Fprintf(out, "relay_url: %s\n", displayRelay(cfg.RelayURL))
			fmt.Fprintf(out, "channel_id: %s\n", displayChannel(cfg.ChannelID))
			fmt.Fprintf(out, "festival_id: %s\n", safeDisplay(cfg.FestivalID, "(unset)"))
			fmt.Fprintf(out, "festival_path: %s\n", safeDisplay(cfg.FestivalPath, "(unset)"))
			if buzzcli.HasPrivateKey() {
				fmt.Fprintln(out, "BUZZ_PRIVATE_KEY: [set]")
			} else {
				fmt.Fprintln(out, "BUZZ_PRIVATE_KEY: [unset]")
			}
			ar, reason := assets.Root()
			fmt.Fprintf(out, "asset_root: %s (%s)\n", safeDisplay(ar, "(none)"), safeDisplay(reason, "unknown"))
			return nil
		},
	}
}

func safeDisplay(value, fallback string) string {
	if value == "" {
		return fallback
	}
	quoted := strconv.Quote(value)
	return quoted[1 : len(quoted)-1]
}

func displayRelay(value string) string {
	if value == "" {
		return "(unset)"
	}
	if err := buzzcli.ValidateRelayURL(value); err != nil {
		return "(invalid: " + err.Error() + ")"
	}
	return safeDisplay(value, "(unset)")
}

func displayChannel(value string) string {
	if value == "" {
		return "(unset)"
	}
	if err := buzzcli.ValidateChannelID(value); err != nil {
		return "(invalid: " + err.Error() + ")"
	}
	return value
}
