// Package buzzcli locates and invokes the external buzz CLI.
package buzzcli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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

// SendMessage runs: buzz messages send --channel <id> --content -
// body is written to stdin. relayURL is set via BUZZ_RELAY_URL if non-empty.
func SendMessage(channelID, body, relayURL string) error {
	if channelID == "" {
		return fmt.Errorf("channel id required")
	}
	if !HasPrivateKey() {
		return fmt.Errorf("BUZZ_PRIVATE_KEY is not set")
	}
	bin, err := LookPath()
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, "messages", "send", "--channel", channelID, "--content", "-")
	cmd.Stdin = strings.NewReader(body)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	env := os.Environ()
	if relayURL != "" {
		env = append(env, "BUZZ_RELAY_URL="+relayURL)
	}
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("buzz messages send: %w", err)
	}
	return nil
}
