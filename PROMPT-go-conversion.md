# Agent prompt: Python `authsec-bridge` to Go `harness-bridge`

Use this document as the **single source of truth** for implementing a Go rewrite of [authsec-ai/authsec-bridge](https://github.com/authsec-ai/authsec-bridge). Work in this repository (`harness-bridge`). Do not port line-by-line; **preserve behavior and extend** where noted below.

**Upstream repo:** [github.com/mlaccetti/harness-bridge](https://github.com/mlaccetti/harness-bridge) (public).  
**Homebrew:** [github.com/mlaccetti/homebrew-tap](https://github.com/mlaccetti/homebrew-tap).  
**License:** MIT (same as inspiration project).  
**Inspiration:** [authsec-ai/authsec-bridge](https://github.com/authsec-ai/authsec-bridge) (Python); this repo is a Go rewrite with extended CLIs and tooling.

---

## Mission

Build a **standalone, installable CLI** that bridges coding-agent sessions across tools so users can `resume` in another CLI without re-explaining context.

**Primary agents (must support parse + write + discover + list):**

| ID (`--from` / `--to`) | User-facing name | Binary / notes |
|------------------------|------------------|----------------|
| `claude` | Claude Code | `claude` |
| `codex` | Codex CLI | `codex` |
| `cursor` | Cursor Agent | `cursor-agent` |
| `antigravity` | Google Antigravity CLI | `agy` (alias flag value `agy` maps to `antigravity`) |
| `gemini` | Gemini CLI | `gemini` (parity with Python upstream) |

**Paths (high level):**

| CLI | Storage (discover empirically; centralize in `internal/paths`) |
|-----|------------------------------------------------------------------|
| `claude` | `~/.claude/projects/<encoded-cwd>/*.jsonl` |
| `codex` | `~/.codex/sessions/**/rollout-*.jsonl` |
| `cursor` | `~/.cursor/projects/<encoded-workspace>/agent-transcripts/<id>/<id>.jsonl` |
| `antigravity` | Inspect Antigravity (`agy`) data dirs on macOS/Linux; document in README |
| `gemini` | `~/.gemini/...` |

**Universal target:** `markdown` writes `HANDOFF.md` in target cwd (works for any LLM tool).

**Binary names:** primary `harness-bridge`, short alias `hb`. Optional compatibility aliases `session-bridge` / `sb` if cheap (document in README).

---

## Non-goals (unchanged from upstream)

- Do **not** preserve structured tool calls across CLIs (causes load errors). Flatten to **prose summaries** on assistant turns.
- Do **not** preserve source session UUID in target; always mint a new ID.
- Do **not** require Python at runtime.

---

## CLI framework (required choice)

Use **Cobra** as the command tree (subcommands, POSIX flags, help groups, aliases). Layer:

1. **[Charm Fang](https://github.com/charmbracelet/fang)** (`charm.land/fang/v2`) for execution UX:
   - Styled help and errors (TTY-aware)
   - Built-in `--version` from Go build info
   - `completion` subcommand (bash/zsh/fish)
   - Hidden `man` generator
   - Does not dump full usage after a user error (cleaner failures)

2. **[Carapace](https://github.com/carapace-sh/carapace)** for **rich, context-aware completions**:
   - Complete `--from` / `--to` with fixed CLI enum values
   - Complete `--session` with recent session IDs for the selected `--from` (read from discover, cap list size)
   - Complete `--cwd` with sensible defaults (PWD, recent session cwds)
   - Document install: `harness-bridge completion bash|zsh|fish` (Fang) plus Carapace wiring per their Cobra integration guide

Do **not** use a second CLI framework (Kong, urfave/cli) alongside Cobra. Keep command definitions in `internal/cli/` or `cmd/harness-bridge/cmd/` packages; `main` only calls `fang.Execute`.

**UX requirements:**

- Every subcommand: `Short`, `Long`, and `Example` blocks with copy-paste commands
- `transfer` examples for each supported direction
- Suggest typo fixes via Cobra (`Did you mean ...?`)
- Exit codes: `0` success, `1` user/runtime error, `2` usage (Cobra/Fang default)

---

## Reference implementation (read before coding)

```bash
git clone https://github.com/authsec-ai/authsec-bridge.git /tmp/authsec-bridge
```

Python layout to mirror conceptually:

```
model          -> internal/session
paths          -> internal/paths
discover       -> internal/discover
parsers/*      -> internal/parser/*
writers/*      -> internal/writer/*
_render        -> internal/render
cli            -> cmd/harness-bridge + internal/cli (cobra commands)
```

---

## Go architecture (required)

### Module & layout

```
harness-bridge/
├── go.mod
├── LICENSE                   # MIT (see License section)
├── mise.toml
├── justfile                  # developer tasks (no Makefile)
├── .goreleaser.yaml          # release binaries + homebrew tap PR
├── .github/
│   ├── CODEOWNERS
│   ├── workflows/ci.yml
│   └── workflows/release.yml # tag push -> goreleaser
├── README.md                 # includes Attribution / credits
├── cmd/
│   └── harness-bridge/
│       └── main.go
├── internal/
│   ├── cli/                  # cobra command tree
│   ├── session/
│   ├── paths/
│   ├── discover/
│   ├── render/
│   ├── parser/
│   │   ├── registry.go
│   │   ├── claude.go
│   │   ├── codex.go
│   │   ├── gemini.go
│   │   ├── cursor.go
│   │   └── antigravity.go
│   └── writer/
│       ├── registry.go
│       ├── claude.go
│       ├── codex.go
│       ├── gemini.go
│       ├── cursor.go
│       ├── antigravity.go
│       └── markdown.go
└── testdata/
```

### Interfaces (adapters)

```go
type Parser interface {
    CLI() string
    Parse(path string) (session.Session, error)
}

type Writer interface {
    CLI() string
    Write(sess session.Session, cwd string) (outPath, newSessionID string, err error)
}
```

- Register parsers/writers in `registry.go`.
- **DRY:** `internal/parser/toolsummarize.go` for tool prose across CLIs.
- **Antigravity:** if on-disk format matches Claude JSONL, reuse Claude parse/write via wrappers that only change roots and `SourceCLI`.

### Subcommands

| Command    | Flags | Behavior |
|------------|-------|----------|
| `list`     | `--from`, `--cwd`, `--limit` (default 10) | List recent sessions |
| `show`     | `--from` (required), `--session`, `--cwd` | Parsed summary + turn preview |
| `transfer` | `--from`, `--to`, `--session`, `--cwd` | Bridge session; print path, new id, resume hints |

`--from` / sources: `claude`, `codex`, `gemini`, `cursor`, `antigravity` (accept `agy` as synonym in flag validation).

`--to` targets: same set + `markdown`.

**Discover rules** (port from Python `discover.py`):

- If `--session` given: search all sessions for that CLI.
- Else: prefer most recent for `--cwd`, then global most recent.

### Enhanced `list` output

Per session row: mtime age, size, path, **cwd** (if known), **turn count**, **task hint** (first user line ~80 chars, strip `<user_query>` / timestamps; fallback `last:` + last user line).

---

## Tooling: mise + just (required)

**mise** pins runtimes and heavy tools:

- `go` (1.22+)
- `golangci-lint`
- `gofumpt`
- `goreleaser`
- `just`

**justfile** is the only task runner (no Makefile). Example:

```just
default:
    just check

fmt:
    gofumpt -w .

lint:
    golangci-lint run ./...

test:
    go test ./... -race -count=1

build:
    go build -o dist/harness-bridge ./cmd/harness-bridge

install:
    go install ./cmd/harness-bridge

check: fmt lint test

release-dry:
    goreleaser release --snapshot --clean
```

README documents: `mise trust && mise install` then `just check`.

---

## License and attribution (required)

### MIT license

- Add a root **`LICENSE`** file: standard [MIT License](https://opensource.org/license/mit) text.
- Copyright holder: `Michael Laccetti` (or `mlaccetti` if you prefer GitHub username only; pick one and use consistently in LICENSE and README).
- Set `license = "MIT"` in `go.mod` if using the `go mod` license directive (Go 1.24+), and keep Homebrew `license: "MIT"` in GoReleaser.
- Release archives must include `LICENSE` (already in `.goreleaser.yaml` `archives.files`).

### README credits

Include a short **"Credits"** or **"Attribution"** section near the top or bottom of `README.md` (not buried in a footnote). Required content:

1. State that **harness-bridge** is a Go reimplementation inspired by **[authsec-bridge](https://github.com/authsec-ai/authsec-bridge)** by AuthSec AI.
2. Link to the upstream repo and note it was the original Python `session-bridge` / `sb` tool for bridging Claude, Codex, and Gemini sessions.
3. Clarify this project extends that idea (additional CLIs, list UX, Homebrew, etc.) and is maintained separately under `mlaccetti/harness-bridge`.
4. Retain MIT compatibility: upstream is MIT; do not imply endorsement by AuthSec AI unless they grant it.

**Example wording (adapt, do not remove the link):**

```markdown
## Credits

This project is inspired by [authsec-bridge](https://github.com/authsec-ai/authsec-bridge),
the Python CLI for bridging coding-agent sessions between Claude Code, Codex, and Gemini.
**harness-bridge** is an independent Go rewrite with additional agents and packaging;
it is not affiliated with AuthSec AI.
```

---

## Quality bar

- `gofumpt` + `golangci-lint` (errcheck, govet, staticcheck)
- Golden parser tests under `testdata/`
- `just check` must pass in CI

---

## GitHub repository and access control

### Create and publish

1. Initialize git in `harness-bridge`; remote `git@github.com:mlaccetti/harness-bridge.git`.
2. Public repo with root **`LICENSE` (MIT)** and README (install paths + **Attribution** to authsec-bridge; see License section).
3. **No collaborators with write access** on the origin repo (default for a solo public repo: only the owner can push branches to `github.com/mlaccetti/harness-bridge`).
4. Forking is **allowed and expected** for contributions. Forks cannot push to origin; they only open PRs from the fork.

### Write access model (what we want)

| Actor | Push to origin branches | Open PR into origin | Merge to `main` |
|-------|-------------------------|---------------------|-----------------|
| `mlaccetti` | Yes (via feature branches; avoid direct `main` once ruleset is on) | Yes | Yes (after self-approval if required) |
| Everyone else | **No** | Yes (from their fork) | **No** |

Enforcement:

1. **Repository Settings -> Collaborators:** do not grant write; keep the repo owner-only for pushes.
2. **Ruleset** targeting branch `main` (Settings -> Rules -> Rulesets):

| Rule | Setting |
|------|---------|
| Require pull request before merging | Yes |
| Required approvals | 1 |
| Require review from Code Owners | Yes |
| Restrict who can dismiss reviews | Repository admins |
| Block force pushes | Yes |
| Require status checks to pass | Yes (job name `check`, see CI below) |
| Restrict who can push to matching branches | **Only `mlaccetti`** (or "Repository admin" role if solo) |

3. Optional ruleset on **all branches** (or `*`): **Restrict branch creation** to repository admins only, so random users cannot create `feature/*` on origin even if GitHub ever granted broader permissions.

4. `.github/CODEOWNERS`:

```
* @mlaccetti
```

Day-to-day: `git checkout -b feat/...` on origin -> PR -> CI `check` green -> you approve and merge. Fork PRs follow the same path; you are never required to merge them.

Optional: Dependabot for Go modules (auto-merge off), signed commits if you use them.

---

## CI workflow (implement `.github/workflows/ci.yml`)

**Triggers:**

- `pull_request` (all branches)
- `push` to `main` (post-merge sanity)

**Job name:** `check` (must match the required status check in the ruleset).

**Runner:** `ubuntu-latest` (sufficient; no macOS-only code in v1).

**Steps (use mise for reproducibility, mirror local dev):**

1. Checkout
2. Install mise (official installer action or `jdx/mise-action`)
3. `mise trust` and `mise install` (reads repo `mise.toml`)
4. `just check` (fmt + lint + test with race)

**Permissions:** `contents: read` only.

**Concurrency:** cancel in-progress runs for the same PR branch:

```yaml
concurrency:
  group: ci-${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

**Reference workflow (commit this file verbatim, adjust action pin versions if needed):**

```yaml
name: CI

on:
  pull_request:
  push:
    branches: [main]

permissions:
  contents: read

concurrency:
  group: ci-${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  check:
    name: check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: jdx/mise-action@v2
        with:
          install: true

      - name: Verify toolchain
        run: mise exec -- go version

      - name: just check
        run: just check
```

**CI acceptance:** PRs cannot merge until `check` is green. Local and CI must use the same `mise.toml` + `justfile`.

---

## Release workflow (implement `.github/workflows/release.yml`)

**Triggers:** push tags matching `v*` (e.g. `v0.1.0`). Tags must be created on commits already on `main` (document: `git tag v0.1.0 && git push origin v0.1.0`).

**Job:** single `release` job running GoReleaser.

**Secrets (repository settings):**

| Secret | Purpose |
|--------|---------|
| `GITHUB_TOKEN` | Provided by Actions; used for GitHub Release assets (default permissions) |
| `HOMEBREW_TAP_GITHUB_TOKEN` | PAT (classic) or fine-grained token with **contents: write** on `mlaccetti/homebrew-tap` and **contents: write** on `mlaccetti/harness-bridge` for releases |

Fine-grained PAT scope: only those two repos, minimum permissions.

**Permissions for the workflow:**

```yaml
permissions:
  contents: write   # upload release assets, create release
```

**Reference workflow:**

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: jdx/mise-action@v2
        with:
          install: true

      - name: GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

**Manual preflight (document in README):** `just release-dry` (`goreleaser release --snapshot --clean`) before tagging.

**Release artifacts:** GitHub Release per tag with tarballs/zip per GOOS/GOARCH, `checksums.txt`, optional changelog from tag message.

---

## GoReleaser and Homebrew tap

### `.goreleaser.yaml` (implement in repo root)

Minimum expectations:

```yaml
version: 2

project_name: harness-bridge

before:
  hooks:
    - go mod tidy
    - go test ./... -race -count=1

builds:
  - id: harness-bridge
    main: ./cmd/harness-bridge
    binary: harness-bridge
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - formats: [tar.gz]
    files:
      - LICENSE
      - README.md

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^chore:"

release:
  github:
    owner: mlaccetti
    name: harness-bridge
  draft: false
  prerelease: auto

brews:
  - name: harness-bridge
    repository:
      owner: mlaccetti
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    directory: Formula
    homepage: "https://github.com/mlaccetti/harness-bridge"
    description: "Bridge coding-agent sessions between CLIs"
    license: "MIT"
    install: |
      bin.install "harness-bridge"
    test: |
      system "#{bin}/harness-bridge", "--version"
```

Wire `main.version` / build info in `cmd/harness-bridge` so `--version` matches the tag.

**Release flow:**

1. Merge feature work to `main`; CI green.
2. Tag `v0.1.0` on `main`, push tag.
3. `release.yml` runs GoReleaser: GitHub Release + commits updated `Formula/harness-bridge.rb` to `mlaccetti/homebrew-tap`.
4. Users: `brew update && brew upgrade harness-bridge` (after tap pull).

Match existing tap layout (`Formula/opencode-usage.rb`) but ship **prebuilt Go binaries**, not Python virtualenv.

### User install docs (README)

```bash
brew tap mlaccetti/tap
brew install harness-bridge

# or
go install github.com/mlaccetti/harness-bridge/cmd/harness-bridge@latest
```

---

## Resume hints (after `transfer`)

| Target | Hint |
|--------|------|
| `claude` | `/resume` in Claude Code |
| `codex` | `codex resume <id>` |
| `gemini` | `gemini --resume <id>` |
| `cursor` | `cursor-agent --resume <id>` |
| `antigravity` | `agy --conversation <id>` or `agy --continue` (verify against installed Antigravity) |
| `markdown` | Read `HANDOFF.md` and continue |

---

## Implementation sequence

1. Scaffold `go.mod` (`github.com/mlaccetti/harness-bridge`), mise, justfile, Fang+Cobra root.
2. Add `.github/workflows/ci.yml` and `.github/workflows/release.yml` per sections above; add `.goreleaser.yaml`, `CODEOWNERS`.
3. Port session, render, paths (claude, codex, gemini) + tests.
4. Parsers/writers + `transfer` for parity CLIs.
5. `cursor` + `antigravity` adapters.
6. Enhanced `list` + Carapace completions.
7. Document ruleset + secrets (`HOMEBREW_TAP_GITHUB_TOKEN`) in README; `just release-dry` once locally.
8. `just check` green; smoke test with real session files.

---

## Acceptance criteria

- [ ] Python parity for claude/codex/gemini subcommands and flags
- [ ] `cursor` + `antigravity` (`agy`) supported
- [ ] Fang help/errors + shell completion documented
- [ ] Carapace completes CLI enums and session IDs where feasible
- [ ] `just check` passes locally; CI job `check` matches `ci.yml`
- [ ] Tag push runs `release.yml`; GoReleaser publishes GitHub Release + `homebrew-tap` formula
- [ ] Origin repo: only owner can push branches; fork PRs allowed; `main` ruleset + CODEOWNERS
- [ ] Enhanced `list` with task hints
- [ ] Root `LICENSE` (MIT); README credits [authsec-bridge](https://github.com/authsec-ai/authsec-bridge)

---

## One-shot kickoff message

```
Implement harness-bridge per PROMPT-go-conversion.md.

Module: github.com/mlaccetti/harness-bridge. CLI: Cobra + Fang + Carapace.
Tasks: mise + justfile only (no Makefile). Antigravity = Google CLI binary `agy`, id `antigravity`.

Clone authsec-bridge to /tmp for behavior reference.

Implement ci.yml (job name `check`), release.yml (tags v*), and .goreleaser.yaml (homebrew tap mlaccetti/homebrew-tap).
Document HOMEBREW_TAP_GITHUB_TOKEN and owner-only push / fork-PR workflow in README.
Add MIT LICENSE and README Attribution pointing to https://github.com/authsec-ai/authsec-bridge.

Run `just check` before stopping.
```
