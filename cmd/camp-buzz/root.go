package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Obedience-Corp/camp-buzz/internal/version"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	build := version.Current()
	cmd := &cobra.Command{
		Use:   "camp-buzz",
		Short: "Optional Buzz integration plugin for camp",
		Long: `camp-buzz is a standalone camp plugin (not camp/fest core).

Install on PATH so camp discovers it as:

  camp buzz …

Posts Festival status into a Buzz channel via the external buzz CLI.
No Buzz logic is compiled into camp or fest.

See: workflow/design/festival-buzz-integration (camp design WI-ca719b).`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "camp-buzz is installed and ready.")
			fmt.Fprintln(cmd.OutOrStdout(), "Invoke through camp: camp buzz <doctor|post|bind|show|hook-install>")
			return cmd.Help()
		},
	}

	cmd.Version = build.Version
	cmd.SetVersionTemplate(fmt.Sprintf("camp-buzz %s (%s) built %s\n", build.Version, build.Commit, build.BuildDate))
	cmd.InitDefaultVersionFlag()
	cmd.InitDefaultCompletionCmd()

	cmd.AddCommand(newVersionCmd(build))
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newBindCmd())
	cmd.AddCommand(newPostCmd())
	cmd.AddCommand(newHookInstallCmd())

	return cmd
}

func newVersionCmd(build version.Build) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "camp-buzz %s (%s) built %s\n", build.Version, build.Commit, build.BuildDate)
		},
	}
}

func campaignRoot() string {
	if v := strings.TrimSpace(os.Getenv("CAMP_ROOT")); v != "" {
		return v
	}
	// walk up from cwd for .campaign
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		if st, err := os.Stat(filepath.Join(dir, ".campaign")); err == nil && st.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
