default:
    @just --list

fmt:
    gofumpt -w .

lint:
    golangci-lint run ./...

test:
    go test ./... -race -count=1

build:
    go build -o dist/harness-bridge ./cmd/harness-bridge

# Run the CLI without installing (pass flags after the recipe name).
run *args:
    go run ./cmd/harness-bridge {{args}}

install:
    go install ./cmd/harness-bridge

check: fmt lint test

release-dry:
    goreleaser release --snapshot --clean
