package main

import (
	"context"
	"os"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"

	"github.com/mlaccetti/harness-bridge/internal/cli"
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

	if err := fang.Execute(context.Background(), root, fang.WithVersion(version)); err != nil {
		if root.SilenceErrors {
			cobra.CheckErr(err)
		}
		os.Exit(1)
	}
}
