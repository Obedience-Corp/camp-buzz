package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printHelp(os.Stdout)
		return nil
	}

	switch args[0] {
	case "help", "--help", "-h":
		printHelp(os.Stdout)
		return nil
	case "current":
		return runCurrent()
	case "ready":
		fs := flag.NewFlagSet("ready", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		tag := fs.String("tag", "", "release tag to verify against origin/main")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return runReady(*tag)
	case "artifacts":
		fs := flag.NewFlagSet("artifacts", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		dir := fs.String("dir", "dist", "GoReleaser distribution directory")
		tag := fs.String("tag", "", "stable release tag")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return verifyArtifacts(*dir, *tag)
	case "stable":
		fs := flag.NewFlagSet("stable", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		level := fs.String("level", "patch", "release increment: patch, minor, or major")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return runStable(*level)
	case "tag":
		fs := flag.NewFlagSet("tag", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		version := fs.String("version", "", "explicit release tag to create and push")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *version == "" {
			return fmt.Errorf("missing required --version")
		}
		return runTag(*version)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printHelp(out io.Writer) {
	fmt.Fprintf(out, "%s release commands\n\n", releaseToolName)
	fmt.Fprintln(out, "Primary:")
	fmt.Fprintln(out, "  just release stable [patch|minor|major]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Support:")
	fmt.Fprintln(out, "  just release current")
	fmt.Fprintln(out, "  just release check")
	fmt.Fprintln(out, "  just release snapshot")
	fmt.Fprintln(out, "  just release package-check v0.1.0")
	fmt.Fprintln(out, "  just release tag v0.1.0")
	if !releasesEnabled {
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Status: disabled - %s\n", releaseDisabledReason)
	}
}
