default:
    @just --list

git_version := `git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo dev`
git_commit := `git rev-parse HEAD 2>/dev/null || echo none`
git_date := `git log -1 --format=%cI 2>/dev/null || echo unknown`
version_ldflags := "-s -w -X github.com/mlaccetti/harness-bridge/internal/version.version=" + git_version + " -X github.com/mlaccetti/harness-bridge/internal/version.commit=" + git_commit + " -X github.com/mlaccetti/harness-bridge/internal/version.date=" + git_date

fmt:
    gofumpt -w .

lint:
    golangci-lint run ./...

test:
    go test ./... -race -count=1

build:
    go build -ldflags '{{version_ldflags}}' -o dist/harness-bridge ./cmd/harness-bridge

# Run the CLI without installing (pass flags after the recipe name).
run *args:
    go run -ldflags '{{version_ldflags}}' ./cmd/harness-bridge {{args}}

install:
    go install -ldflags '{{version_ldflags}}' ./cmd/harness-bridge

check: fmt lint test

release-dry:
    goreleaser release --snapshot --clean
