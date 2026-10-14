# harness-bridge

Bridge coding-agent sessions between CLIs so you can **resume** in another tool without re-explaining context.

Short names **`hb`**, **`sb`**, and **`session-bridge`** are installed as symlinks to the same binary (Homebrew). With `go install`, only the `harness-bridge` command is on your `PATH` unless you add your own symlinks.

## Standalone scope

This repository is intentionally **standalone**: it does not depend on rckt-skills, rckt-factory, muster, or agent marketplaces. You can install and run it with Go or Homebrew alone. Other harness tools may optionally shell out to `harness-bridge`; that is composition, not a required dependency. See [docs/STANDALONE.md](docs/STANDALONE.md) for the full boundary and optional integration notes.

## Credits

This project is inspired by [authsec-bridge](https://github.com/authsec-ai/authsec-bridge),
the Python CLI for bridging coding-agent sessions between Claude Code, Codex, and Gemini.
**harness-bridge** is an independent Go rewrite with additional agents and packaging;
it is not affiliated with AuthSec AI.

## Install

```bash
brew tap mlaccetti/tap
brew install harness-bridge

# or (Go module; published on each release tag)
go install github.com/mlaccetti/harness-bridge/cmd/harness-bridge@latest
# pin a version: go install github.com/mlaccetti/harness-bridge/cmd/harness-bridge@v0.1.2
```

Module docs: [pkg.go.dev/github.com/mlaccetti/harness-bridge](https://pkg.go.dev/github.com/mlaccetti/harness-bridge).

**Windows:** no installer yet. Download `harness-bridge_*_windows_amd64.zip` (or `arm64`) from [GitHub Releases](https://github.com/mlaccetti/harness-bridge/releases) and add `harness-bridge.exe` to your `PATH`. Release builds are cross-compiled on Linux CI (no Windows runner).

Homebrew builds may print a one-line `# ...` upgrade hint on startup when a newer release exists (`HARNESS_BRIDGE_NO_UPGRADE_NOTICE=1` to disable).

## Usage

```bash
harness-bridge list --from claude
harness-bridge show --from codex --session abc123
harness-bridge transfer --from claude --to cursor --cwd .
harness-bridge transfer --from cursor --to markdown
```

Use `--cwd` when the session is tied to a project directory. `transfer` supports planning flags such as `--dry-run` / `--plan`, `--stats`, `--last-turns`, and `--since-compaction` (see `harness-bridge transfer --help`).

### Supported agents

Use these values for `--from` and `--to` (except `markdown`, which is a transfer target only).

| ID | Tool | Notes |
|----|------|--------|
| `claude` | [Claude Code](https://docs.anthropic.com/en/docs/claude-code) | Sessions under `~/.claude/projects/<encoded-cwd>/*.jsonl`. Resume in Claude with `/resume`. |
| `codex` | [Codex CLI](https://github.com/openai/codex) | Rollouts under `~/.codex/sessions/.../rollout-*.jsonl`. Resume with `codex resume <session-id>`. |
| `gemini` | [Gemini CLI](https://github.com/google-gemini/gemini-cli) | Project data under `~/.gemini/`. Resume with `gemini --resume <session-id>`. |
| `cursor` | Cursor Agent (`cursor-agent`) | Transcripts under `~/.cursor/projects/<workspace>/agent-transcripts/`. Resume with `cursor-agent --resume <session-id>`. |
| `antigravity` | Google Antigravity (`agy`) | Alias: `agy` for `--from` / `--to`. Claude-style JSONL under the Antigravity config dir (e.g. `~/.config/antigravity/projects/...`, or `~/Library/Application Support/Antigravity` on macOS). Run `agy` once so dirs exist. Resume with `agy --conversation <id>` or `agy --continue`. |
| `markdown` | Handoff file | Writes `HANDOFF.md` in the target cwd (not a source for `list` / `show`). |

After `transfer`, the CLI prints a short resume hint for the target you chose.

### Shell completion

```bash
harness-bridge completion bash
harness-bridge completion zsh
harness-bridge completion fish
```

Carapace completes `--from`, `--to`, `--session`, and `--cwd`.

## Development

```bash
mise trust && mise install
just check
just run --version
just run list --from claude
```

Run `just` for all recipes. Use `just run` when the binary is not installed yet. Tasks live in the `justfile` (no Makefile).

## Contributing

Contributions are welcome.

1. **Fork** [mlaccetti/harness-bridge](https://github.com/mlaccetti/harness-bridge) on GitHub.
2. Create a **feature branch** on your fork (`feat/...`, `fix/...`, etc.).
3. Install tooling (`mise trust && mise install`) and run **`just check`** (format, lint, tests).
4. Open a **pull request** against `main` with a clear description and test notes.
5. Ensure the CI job **`check`** passes on the PR. A maintainer will review before merge.

Please do not commit secrets or local session JSONL from your machine. Use `testdata/` for fixtures.

**Branch protection:** `main` requires a PR, passing **`check`**, and code-owner review (see [`.github/rulesets/`](.github/rulesets/)). External contributors are expected to work from forks; maintainers merge after review.

## Maintainers

### CI and release

- **CI** (`.github/workflows/ci.yml`): `pull_request` and pushes to `main` run `just check` via mise.
- **Release** (`.github/workflows/release.yml`): pushing a tag `v*` runs GoReleaser (GitHub release + Homebrew tap), then verifies the tag on **proxy.golang.org** and requests **pkg.go.dev** indexing (`go install …@vX.Y.Z`).

| Secret | Purpose |
|--------|---------|
| `GITHUB_TOKEN` | Release assets (provided by Actions) |
| `HOMEBREW_TAP_GITHUB_TOKEN` | PAT with **contents: write** on `mlaccetti/homebrew-tap` and this repo |

**Release checklist:** merge to `main` with green **`check`** → `just release-dry` → signed tag `vX.Y.Z` → `git push origin vX.Y.Z` → release workflow publishes binaries, Homebrew formula, and the Go module → users `brew update && brew upgrade harness-bridge` or `go install github.com/mlaccetti/harness-bridge/cmd/harness-bridge@vX.Y.Z`.

Repository rulesets are exported under [`.github/rulesets/`](.github/rulesets/). Solo-owner merges on `main` use administrator bypass once CI is green; fork PRs still need approval.

## License

MIT. See [LICENSE](LICENSE).
