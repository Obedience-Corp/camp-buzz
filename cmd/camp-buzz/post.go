package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Obedience-Corp/camp-buzz/internal/buzzcli"
	"github.com/Obedience-Corp/camp-buzz/internal/config"
	"github.com/spf13/cobra"
)

type postOptions struct {
	message, festival, task, path, gate string
	channel, relay                      string
	fromHook, noFooter                  bool
}

func newPostCmd() *cobra.Command {
	opts := &postOptions{}
	cmd := &cobra.Command{
		Use:   "post",
		Short: "Post a Festival status message to Buzz",
		Long: `Builds an optional status footer and runs:
  buzz messages send --channel <id> --content -

Reads body from --message or stdin. Never logs private keys.
Used by humans, fest hooks (--from-hook), and agents.`,
		RunE: opts.run,
	}
	cmd.Flags().StringVarP(&opts.message, "message", "m", "", "Message body (else stdin)")
	cmd.Flags().StringVar(&opts.festival, "festival", "", "Festival id for footer")
	cmd.Flags().StringVar(&opts.task, "task", "", "Task ref for footer")
	cmd.Flags().StringVar(&opts.path, "path", "", "Festival path for footer")
	cmd.Flags().StringVar(&opts.gate, "gate", "", "Gate status: pending|pass|fail|n/a")
	cmd.Flags().StringVar(&opts.channel, "channel", "", "Override channel UUID")
	cmd.Flags().StringVar(&opts.relay, "relay", "", "Override relay URL")
	cmd.Flags().BoolVar(&opts.fromHook, "from-hook", false, "Invoked from a fest lifecycle hook")
	cmd.Flags().BoolVar(&opts.noFooter, "no-footer", false, "Do not append status footer")
	return cmd
}

func (opts *postOptions) run(cmd *cobra.Command, _ []string) error {
	cfg, err := opts.resolveConfig()
	if err != nil {
		return err
	}
	body, err := readPostBody(opts.message, cmd.InOrStdin(), opts.fromHook)
	if err != nil {
		return err
	}
	body, err = opts.appendFooter(body, cfg)
	if err != nil {
		return err
	}
	if err := buzzcli.SendMessage(cmd.Context(), cfg.ChannelID, body, cfg.RelayURL); err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), "posted")
	return err
}

func (opts *postOptions) resolveConfig() (config.Config, error) {
	cfg, _, err := config.Resolve(campaignRoot())
	if err != nil {
		return config.Config{}, err
	}
	if opts.channel != "" {
		cfg.ChannelID = opts.channel
	}
	if opts.relay != "" {
		cfg.RelayURL = config.NormalizeRelayURL(opts.relay)
	}
	if opts.festival != "" {
		cfg.FestivalID = opts.festival
	}
	if opts.path != "" {
		cfg.FestivalPath = opts.path
	}
	return cfg, nil
}

func readPostBody(message string, stdin io.Reader, fromHook bool) (string, error) {
	body := strings.TrimSpace(message)
	if body == "" && fromHook {
		return "Festival status update", nil
	}
	if body == "" {
		limited := &io.LimitedReader{R: stdin, N: buzzcli.MaxContentBytes + 1}
		data, err := io.ReadAll(limited)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		if len(data) > buzzcli.MaxContentBytes {
			return "", fmt.Errorf("stdin exceeds Buzz limit of %d bytes", buzzcli.MaxContentBytes)
		}
		body = strings.TrimSpace(string(data))
	}
	if body == "" {
		return "", fmt.Errorf("message body required (--message or stdin)")
	}
	if len(body) > buzzcli.MaxContentBytes {
		return "", fmt.Errorf("message body exceeds Buzz limit of %d bytes", buzzcli.MaxContentBytes)
	}
	return body, nil
}

func (opts *postOptions) appendFooter(body string, cfg config.Config) (string, error) {
	if opts.noFooter {
		return body, nil
	}
	festival := valueOr(cfg.FestivalID, "-")
	task := valueOr(opts.task, "-")
	path := valueOr(cfg.FestivalPath, "-")
	gate := valueOr(opts.gate, "n/a")
	if err := validateFooter(festival, task, path, gate); err != nil {
		return "", err
	}
	footer := fmt.Sprintf("\n\n---\nfestival: %s\ntask: %s\npath: %s\ngate: %s\n---\n", festival, task, path, gate)
	return body + footer, nil
}

func validateFooter(festival, task, path, gate string) error {
	for _, field := range []struct {
		name, value string
		max         int
	}{
		{name: "festival", value: festival, max: 128},
		{name: "task", value: task, max: 256},
		{name: "path", value: path, max: 1024},
	} {
		if len(field.value) > field.max || strings.ContainsAny(field.value, "\r\n\x00") {
			return fmt.Errorf("%s footer field is invalid", field.name)
		}
	}
	if path != "-" && (filepath.IsAbs(path) || filepath.Clean(path) == ".." || strings.HasPrefix(filepath.Clean(path), ".."+string(filepath.Separator))) {
		return fmt.Errorf("path footer field must be campaign-relative")
	}
	if gate != "pending" && gate != "pass" && gate != "fail" && gate != "n/a" {
		return fmt.Errorf("gate must be pending, pass, fail, or n/a")
	}
	return nil
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
