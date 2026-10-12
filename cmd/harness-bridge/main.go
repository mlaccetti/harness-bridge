package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"

	"github.com/mlaccetti/harness-bridge/internal/cli"
	"github.com/mlaccetti/harness-bridge/internal/upgrade"
	"github.com/mlaccetti/harness-bridge/internal/version"
)

func main() {
	info := version.Get()

	root := cli.NewRoot()
	if name := invokedName(); name != "" {
		root.Use = name
	}
	root.Version = info.Version
	root.SetVersionTemplate(info.Line())

	ctx := context.Background()
	upgrade.MaybeSuggestUpgrade(ctx, info.Version, os.Args, os.Stderr)

	// Fang would overwrite root.Version; Cobra --version uses Version + VersionTemplate.
	if err := fang.Execute(ctx, root, fang.WithoutVersion()); err != nil {
		if root.SilenceErrors {
			cobra.CheckErr(err)
		}
		os.Exit(1)
	}
}

func invokedName() string {
	base := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	switch base {
	case "harness-bridge", "":
		return ""
	case "hb", "sb", "session-bridge":
		return base
	default:
		return ""
	}
}
