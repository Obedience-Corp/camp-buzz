// Package buzzcli locates and invokes the external buzz CLI.
package buzzcli

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DefaultCommandTimeout bounds calls to the external Buzz CLI when the caller
// has not supplied an earlier deadline.
const DefaultCommandTimeout = 30 * time.Second

// MaxContentBytes matches Buzz CLI's MAX_CONTENT_BYTES contract.
const MaxContentBytes = 65_536

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
	if err := validateMessageRequest(channelID, body, relayURL); err != nil {
		return err
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

func validateMessageRequest(channelID, body, relayURL string) error {
	if err := ValidateChannelID(channelID); err != nil {
		return err
	}
	if err := ValidateContent(body); err != nil {
		return err
	}
	return ValidateRelayURL(relayURL)
}

// ValidateChannelID enforces the UUID form accepted by the Buzz CLI.
func ValidateChannelID(value string) error {
	if !validUUID(value) {
		return fmt.Errorf("channel id must be a UUID")
	}
	return nil
}

// ValidateContent enforces Buzz's non-empty 65,536-byte message contract.
func ValidateContent(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("message body required")
	}
	if len(value) > MaxContentBytes {
		return fmt.Errorf("message body exceeds Buzz limit (%d > %d bytes)", len(value), MaxContentBytes)
	}
	return nil
}

func validUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i, char := range []byte(value) {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if char != '-' {
				return false
			}
			continue
		}
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}

// ValidateRelayURL accepts an empty value for Buzz's default relay; explicit
// values must be absolute HTTP(S) base URLs without credentials, queries, or fragments.
func ValidateRelayURL(value string) error {
	if value == "" {
		return nil
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("relay URL must be an absolute HTTP or HTTPS URL")
	}
	if parsed.User != nil || parsed.Fragment != "" || parsed.RawQuery != "" || parsed.ForceQuery {
		return fmt.Errorf("relay URL must not contain credentials, a query, or a fragment")
	}
	return nil
}
