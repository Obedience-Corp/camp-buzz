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
