package main

import (
	"fmt"

	"github.com/Obedience-Corp/camp-buzz/internal/buzzcli"
	"github.com/Obedience-Corp/camp-buzz/internal/config"
	"github.com/spf13/cobra"
)

func newBindCmd() *cobra.Command {
	var relay, channel, festivalID, festivalPath string

	cmd := &cobra.Command{
		Use:   "bind",
		Short: "Write non-secret Buzz bind config for this campaign",
		Long:  "Upserts .campaign/integrations/buzz.yaml. Never writes private keys.",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := campaignRoot()
			if root == "" {
				return fmt.Errorf("not inside a campaign (no .campaign / CAMP_ROOT)")
			}
			// Merge with existing file only (do not bake env into the file).
			cfg, _, err := config.LoadFile(root)
			if err != nil {
				return err
			}
			if relay != "" {
				cfg.RelayURL = relay
			}
			if channel != "" {
				cfg.ChannelID = channel
			}
			if festivalID != "" {
				cfg.FestivalID = festivalID
			}
			if festivalPath != "" {
				cfg.FestivalPath = festivalPath
			}
			if cfg.ChannelID == "" {
				return fmt.Errorf("--channel is required when no channel_id is already bound")
			}
			cfg.RelayURL = config.NormalizeRelayURL(cfg.RelayURL)
			if err := validateBindConfig(cfg); err != nil {
				return err
			}
			path, err := config.Write(root, cfg)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", safeDisplay(path, "(unknown)"))
			return nil
		},
	}

	addBindFlags(cmd, &relay, &channel, &festivalID, &festivalPath)
	return cmd
}

func addBindFlags(cmd *cobra.Command, relay, channel, festivalID, festivalPath *string) {
	cmd.Flags().StringVar(relay, "relay", "", "Buzz relay HTTP base URL (e.g. http://localhost:3000)")
	cmd.Flags().StringVar(channel, "channel", "", "Buzz channel UUID")
	cmd.Flags().StringVar(festivalID, "festival", "", "Default festival id for footers")
	cmd.Flags().StringVar(festivalPath, "festival-path", "", "Optional festival path relative to campaign")
}

func validateBindConfig(cfg config.Config) error {
	if err := buzzcli.ValidateChannelID(cfg.ChannelID); err != nil {
		return err
	}
	if err := buzzcli.ValidateRelayURL(cfg.RelayURL); err != nil {
		return err
	}
	return validateFooter(valueOr(cfg.FestivalID, "-"), "-", valueOr(cfg.FestivalPath, "-"), "n/a")
}
