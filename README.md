# harness-bridge

Bridge coding-agent sessions between CLIs so you can **resume** in another tool without re-explaining context.

Supported sources and targets: **Claude Code**, **Codex**, **Cursor Agent** (`cursor-agent`), **Gemini CLI**, **Google Antigravity** (`agy`), plus universal **`markdown`** (`HANDOFF.md`).

Aliases: `hb`, `session-bridge`, `sb`.

## Credits

This project is inspired by [authsec-bridge](https://github.com/authsec-ai/authsec-bridge),
the Python CLI for bridging coding-agent sessions between Claude Code, Codex, and Gemini.
**harness-bridge** is an independent Go rewrite with additional agents and packaging;
it is not affiliated with AuthSec AI.

## Install

```bash
brew tap mlaccetti/tap
brew install harness-bridge

# or
go install github.com/mlaccetti/harness-bridge/cmd/harness-bridge@latest
```

## Development

```bash
mise trust && mise install
just check
just run --version
just run list --from claude
```

Run `just` to see all recipes. Use `just run` (via `go run`) when you have not installed the binary yet.

Tasks are defined in the `justfile` (no Makefile).

## Usage

```bash
harness-bridge list --from claude
harness-bridge show --from codex --session abc123
harness-bridge transfer --from claude --to cursor --cwd .
harness-bridge transfer --from cursor --to markdown
```

Shell completion (Fang):

```bash
harness-bridge completion bash
harness-bridge completion zsh
harness-bridge completion fish
```

Carapace provides context-aware completions for `--from`, `--to`, `--session`, and `--cwd`.

### Antigravity (`agy`)

Session files use Claude-style JSONL under the Antigravity config directory (typically `~/.config/antigravity/projects/...` on Linux, `~/Library/Application Support/Antigravity` on macOS when present). Run `agy` once so directories exist, then use `--from antigravity` (or `agy`).

## Repository access

- **Origin** (`github.com/mlaccetti/harness-bridge`): only the owner (`mlaccetti`) can push branches.
- **Contributions**: fork the repo, open a PR from your fork; you cannot push to origin directly.
- **`main`**: protected by a ruleset requiring PR, review, CODEOWNERS approval, and CI job **`check`**.

## CI and release

- **CI** (`.github/workflows/ci.yml`): `pull_request` and pushes to `main` run `just check` via mise.
- **Release** (`.github/workflows/release.yml`): pushing a tag `v*` runs GoReleaser.

### Secrets

| Secret | Purpose |
|--------|---------|
| `GITHUB_TOKEN` | GitHub Release assets (provided by Actions) |
| `HOMEBREW_TAP_GITHUB_TOKEN` | PAT with **contents: write** on `mlaccetti/homebrew-tap` (and releases on this repo) so GoReleaser can commit `Formula/harness-bridge.rb` |

Create a fine-grained PAT scoped to those two repositories with minimum permissions.

### Release checklist

1. Merge to `main`; ensure CI **`check`** is green.
2. Dry run locally: `just release-dry`
3. Tag: `git tag v0.1.0 && git push origin v0.1.0`
4. Users: `brew update && brew upgrade harness-bridge`

When you run a **Homebrew-installed** release build, the CLI pings GitHub once per invocation (interactive stderr only) and may print a standalone `# ...` comment with `brew update && brew upgrade harness-bridge` if a newer tag exists. Disable with `HARNESS_BRIDGE_NO_UPGRADE_NOTICE=1`.

## License

MIT. See [LICENSE](LICENSE).
