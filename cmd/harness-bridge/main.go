package main

import (
	"context"
	"os"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"

	"github.com/mlaccetti/harness-bridge/internal/cli"
	"github.com/mlaccetti/harness-bridge/internal/upgrade"
	"github.com/mlaccetti/harness-bridge/internal/version"
)

func main() {
	info := version.Get()

	root := cli.NewRoot()
	root.Version = info.Version
	root.SetVersionTemplate(info.Line())

	ctx := context.Background()
	upgrade.MaybeSuggestUpgrade(ctx, info.Version, os.Args, os.Stderr)

	if err := fang.Execute(ctx, root, fang.WithVersion(info.Version)); err != nil {
		if root.SilenceErrors {
			cobra.CheckErr(err)
		}
		os.Exit(1)
	}
}
