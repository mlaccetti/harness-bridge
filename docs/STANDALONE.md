# Standalone project boundary

**harness-bridge** is a self-contained Go CLI. It ships, tests, and releases without any other harness or agent orchestration repo on disk or in `go.mod`.

## What this repo is responsible for

- Discovering coding-agent session stores on the local machine (Claude Code, Codex, Cursor Agent, Gemini CLI, Antigravity, and similar).
- Parsing session transcripts and writing handoff artifacts (including universal `markdown` / `HANDOFF.md`).
- Packaging and distribution (Homebrew tap, GoReleaser, CI).

Runtime dependencies are ordinary Go modules listed in `go.mod` (CLI framework, SQLite, UUID, and related libraries). There is no compile-time or runtime coupling to agent marketplaces, skill packs, or multi-repo harness factories.

## Explicit non-dependencies

The following are **not** required to build, test, install, or use harness-bridge:

| Name | Role elsewhere | Relationship to harness-bridge |
|------|----------------|--------------------------------|
| **rckt-skills** | Agent skills / delivery workflows | Optional; not imported or invoked |
| **rckt-factory** | Harness generation / factory tooling | Optional; not imported or invoked |
| **muster** | Orchestration / workspace tooling | Optional; not imported or invoked |
| **marketplace** (e.g. plugin marketplaces) | Plugin distribution | Optional; not a dependency |

If you grep this repository for those names, you should find no production code references (only this document, if at all).

## Optional composition (your choice)

Teams may **optionally** use harness-bridge alongside other tools in the same developer environment:

- **Delivery harnesses** (skills, worktrees, CI gates) can call `harness-bridge` as an external binary in scripts or agent instructions, same as any other CLI.
- **Session handoff** does not require a particular IDE plugin or monorepo layout; point `--cwd` at a project and choose `--from` / `--to` agents you already use.

That composition is operational convenience, not an architectural dependency. harness-bridge does not assume those tools exist, read their config trees, or embed their paths.

## Boundary checks

When extending this project, keep the standalone contract:

1. **Imports**: Stay within this module and public Go dependencies; do not add modules that exist only to wire in a specific harness.
2. **Paths**: Resolve agent data dirs via `internal/paths` and documented OS conventions, not hardcoded paths into other repos or plugin roots.
3. **Docs & CI**: README and workflows should describe install, `just check`, and release without prerequisite clones of sibling harness repos.

## Verification

From the repository root:

```bash
go test ./...
just check   # if mise/just are installed
```

No sibling checkout or marketplace plugin is needed for those commands to pass.
