package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Obedience-Corp/camp-buzz/internal/buzzcli"
	"github.com/Obedience-Corp/camp-buzz/internal/config"
	"github.com/spf13/cobra"
)

func newPostCmd() *cobra.Command {
	var (
		message      string
		festival     string
		task         string
		pathFlag     string
		gate         string
		channel      string
		relay        string
		fromHook     bool
		noFooter     bool
	)

	cmd := &cobra.Command{
		Use:   "post",
		Short: "Post a Festival status message to Buzz",
		Long: `Builds an optional status footer and runs:
  buzz messages send --channel <id> --content -

Reads body from --message or stdin. Never logs private keys.
Used by humans, fest hooks (--from-hook), and agents.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := campaignRoot()
			cfg, _, err := config.Resolve(root)
			if err != nil {
				return err
			}
			if channel != "" {
				cfg.ChannelID = channel
			}
			if relay != "" {
				cfg.RelayURL = relay
			}
			if festival != "" {
				cfg.FestivalID = festival
			}
			if pathFlag != "" {
				cfg.FestivalPath = pathFlag
			}
			if cfg.ChannelID == "" {
				return fmt.Errorf("channel required (--channel, BUZZ_CHANNEL, or camp buzz bind)")
			}

			body := strings.TrimSpace(message)
			if body == "" {
				// stdin
				if fromHook {
					// hooks may pass empty message with only metadata
					body = "Festival status update"
				} else {
					b, err := io.ReadAll(os.Stdin)
					if err != nil {
						return fmt.Errorf("read stdin: %w", err)
					}
					body = strings.TrimSpace(string(b))
				}
			}
			if body == "" {
				return fmt.Errorf("message body required (--message or stdin)")
			}

			if !noFooter {
				fest := cfg.FestivalID
				if fest == "" {
					fest = "-"
				}
				t := task
				if t == "" {
					t = "-"
				}
				p := cfg.FestivalPath
				if p == "" {
					p = "-"
				}
				g := gate
				if g == "" {
					g = "n/a"
				}
				footer := fmt.Sprintf("\n\n---\nfestival: %s\ntask: %s\npath: %s\ngate: %s\n---\n", fest, t, p, g)
				body = body + footer
			}

			if err := buzzcli.SendMessage(cfg.ChannelID, body, cfg.RelayURL); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "posted")
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Message body (else stdin)")
	cmd.Flags().StringVar(&festival, "festival", "", "Festival id for footer")
	cmd.Flags().StringVar(&task, "task", "", "Task ref for footer")
	cmd.Flags().StringVar(&pathFlag, "path", "", "Festival path for footer")
	cmd.Flags().StringVar(&gate, "gate", "", "Gate status: pending|pass|fail|n/a")
	cmd.Flags().StringVar(&channel, "channel", "", "Override channel UUID")
	cmd.Flags().StringVar(&relay, "relay", "", "Override relay URL")
	cmd.Flags().BoolVar(&fromHook, "from-hook", false, "Invoked from a fest lifecycle hook")
	cmd.Flags().BoolVar(&noFooter, "no-footer", false, "Do not append status footer")
	return cmd
}
