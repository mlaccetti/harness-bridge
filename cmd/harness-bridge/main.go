package main

import (
	"context"
	"os"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"

	"github.com/mlaccetti/harness-bridge/internal/cli"
	"github.com/mlaccetti/harness-bridge/internal/upgrade"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root := cli.NewRoot()
	root.Version = version
	if commit != "none" {
		root.SetVersionTemplate(version + " (" + commit + ", " + date + ")\n")
	}

	ctx := context.Background()
	upgrade.MaybeSuggestUpgrade(ctx, version, os.Args, os.Stderr)

	if err := fang.Execute(ctx, root, fang.WithVersion(version)); err != nil {
		if root.SilenceErrors {
			cobra.CheckErr(err)
		}
		os.Exit(1)
	}
}
