package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Obedience-Corp/camp-buzz/internal/assets"
	"github.com/Obedience-Corp/camp-buzz/internal/buzzcli"
	"github.com/Obedience-Corp/camp-buzz/internal/config"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check install, buzz CLI, env, and bind config",
		Long:  "Exit 0 if ready to post; non-zero with actionable fixes. Never prints secrets.",
		RunE:  runDoctor,
	}
}

func runDoctor(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	root, installationOK := doctorInstallation(out)
	ready := installationOK
	if !doctorBuzz(out) {
		ready = false
	}
	if !doctorConfig(out, root) {
		ready = false
	}
	reportCamp(out)
	if ready {
		fmt.Fprintln(out, "status: ready")
		return nil
	}
	fmt.Fprintln(out, "status: not ready")
	return fmt.Errorf("doctor found issues (see above)")
}

func doctorInstallation(out io.Writer) (string, bool) {
	ready := true
	executable, err := os.Executable()
	if err != nil {
		fmt.Fprintf(out, "binary: unknown (%v)\n", err)
		ready = false
	} else {
		fmt.Fprintf(out, "binary: %s\n", safeDisplay(filepath.Clean(executable), "unknown"))
	}
	root := campaignRoot()
	if root == "" {
		fmt.Fprintln(out, "campaign root: not found (set CAMP_ROOT or run inside a campaign)")
	} else {
		fmt.Fprintf(out, "campaign root: %s\n", safeDisplay(root, "unknown"))
	}
	assetRoot, reason := assets.Root()
	fmt.Fprintf(out, "asset root: %s (%s)\n", safeDisplay(assetRoot, "unknown"), safeDisplay(reason, "unknown"))
	return root, ready
}

func doctorBuzz(out io.Writer) bool {
	ready := true
	if path, err := buzzcli.LookPath(); err != nil {
		fmt.Fprintf(out, "buzz CLI: MISSING — %v\n", err)
		ready = false
	} else {
		fmt.Fprintf(out, "buzz CLI: %s\n", safeDisplay(path, "unknown"))
	}
	if buzzcli.HasPrivateKey() {
		fmt.Fprintln(out, "BUZZ_PRIVATE_KEY: set")
	} else {
		fmt.Fprintln(out, "BUZZ_PRIVATE_KEY: NOT set (required to post)")
		ready = false
	}
	return ready
}

func doctorConfig(out io.Writer, root string) bool {
	cfg, bindPath, err := config.Resolve(root)
	if err != nil {
		fmt.Fprintf(out, "bind config: error — %v\n", err)
		return false
	}
	reportBindFile(out, bindPath)
	return validateDoctorConfig(out, cfg)
}

func reportBindFile(out io.Writer, path string) {
	if path == "" {
		fmt.Fprintln(out, "bind file: n/a (no campaign root)")
		return
	}
	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(out, "bind file: %s\n", safeDisplay(path, "unknown"))
	} else {
		fmt.Fprintf(out, "bind file: %s (absent — use camp buzz bind)\n", safeDisplay(path, "unknown"))
	}
}

func validateDoctorConfig(out io.Writer, cfg config.Config) bool {
	ready := reportRelay(out, cfg.RelayURL)
	if !reportChannel(out, cfg.ChannelID) {
		ready = false
	}
	if err := validateFooter(valueOr(cfg.FestivalID, "-"), "-", valueOr(cfg.FestivalPath, "-"), "n/a"); err != nil {
		fmt.Fprintf(out, "footer config: INVALID — %v\n", err)
		ready = false
	} else if cfg.FestivalID != "" {
		fmt.Fprintf(out, "festival_id: %s\n", cfg.FestivalID)
	}
	return ready
}

func reportRelay(out io.Writer, relayURL string) bool {
	if relayURL == "" {
		fmt.Fprintln(out, "relay_url: NOT set (BUZZ_RELAY_URL or bind file; use http://…)")
		return false
	}
	if err := buzzcli.ValidateRelayURL(relayURL); err != nil {
		fmt.Fprintf(out, "relay_url: INVALID — %v\n", err)
		return false
	}
	fmt.Fprintf(out, "relay_url: %s\n", relayURL)
	return true
}

func reportChannel(out io.Writer, channelID string) bool {
	if channelID == "" {
		fmt.Fprintln(out, "channel_id: NOT set (BUZZ_CHANNEL or bind file)")
		return false
	}
	if err := buzzcli.ValidateChannelID(channelID); err != nil {
		fmt.Fprintf(out, "channel_id: INVALID — %v\n", err)
		return false
	}
	fmt.Fprintf(out, "channel_id: %s\n", channelID)
	return true
}

func reportCamp(out io.Writer) {
	if _, err := exec.LookPath("camp"); err != nil {
		fmt.Fprintln(out, "camp: not on PATH (optional for direct camp-buzz use)")
		return
	}
	fmt.Fprintln(out, "camp: found — after install, invoke as: camp buzz …")
}
