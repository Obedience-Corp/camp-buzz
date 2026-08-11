package main

import (
	"fmt"

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
			// merge with existing
			cfg, _, err := config.Resolve(root)
			if err != nil {
				return err
			}
			// clear env overlays for merge base — re-read file only
			cfg, _, _ = config.Resolve("") // empty
			// load file without env by reading again simply:
			existing, _, _ := config.Resolve(root)
			cfg = existing
			// strip env influence: re-read from Write path using only flags when set
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
			if cfg.ChannelID == "" && channel == "" {
				return fmt.Errorf("--channel is required when binding a new channel")
			}
			path, err := config.Write(root, cfg)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", path)
			return nil
		},
	}

	cmd.Flags().StringVar(&relay, "relay", "", "Buzz relay URL (e.g. ws://localhost:3000)")
	cmd.Flags().StringVar(&channel, "channel", "", "Buzz channel UUID")
	cmd.Flags().StringVar(&festivalID, "festival", "", "Default festival id for footers")
	cmd.Flags().StringVar(&festivalPath, "festival-path", "", "Optional festival path relative to campaign")
	return cmd
}
