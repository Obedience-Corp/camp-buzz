// Package buzzcli locates and invokes the external buzz CLI.
package buzzcli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DefaultCommandTimeout bounds calls to the external Buzz CLI when the caller
// has not supplied an earlier deadline.
const DefaultCommandTimeout = 30 * time.Second

type commandFactory func(context.Context, string, ...string) *exec.Cmd

// LookPath finds the buzz binary on PATH.
func LookPath() (string, error) {
	p, err := exec.LookPath("buzz")
	if err != nil {
		return "", fmt.Errorf("buzz CLI not found on PATH (install buzz, then re-run camp buzz doctor): %w", err)
	}
	return p, nil
}

// HasPrivateKey reports whether BUZZ_PRIVATE_KEY is set (does not print it).
func HasPrivateKey() bool {
	return strings.TrimSpace(os.Getenv("BUZZ_PRIVATE_KEY")) != ""
}

// SendMessage runs: buzz messages send --channel <id> --content -.
// body is written to stdin. relayURL is set via BUZZ_RELAY_URL if non-empty.
// The command is limited to DefaultCommandTimeout; an earlier caller deadline
// or cancellation always wins.
func SendMessage(ctx context.Context, channelID, body, relayURL string) error {
	return sendMessage(ctx, channelID, body, relayURL, DefaultCommandTimeout, LookPath, exec.CommandContext)
}

func sendMessage(
	ctx context.Context,
	channelID, body, relayURL string,
	timeout time.Duration,
	lookPath func() (string, error),
	newCommand commandFactory,
) error {
	if channelID == "" {
		return fmt.Errorf("channel id required")
	}
	if !HasPrivateKey() {
		return fmt.Errorf("BUZZ_PRIVATE_KEY is not set")
	}
	bin, err := lookPath()
	if err != nil {
		return err
	}
	commandCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := newCommand(commandCtx, bin, "messages", "send", "--channel", channelID, "--content", "-")
	cmd.Stdin = strings.NewReader(body)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	env := os.Environ()
	if relayURL != "" {
		env = append(env, "BUZZ_RELAY_URL="+relayURL)
	}
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		if ctxErr := commandCtx.Err(); ctxErr != nil {
			return fmt.Errorf("buzz messages send for channel %q: %w (process: %v)", channelID, ctxErr, err)
		}
		return fmt.Errorf("buzz messages send for channel %q: %w", channelID, err)
	}
	return nil
}
