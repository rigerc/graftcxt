set dotenv-load := true

# Show available commands
default:
    just --list

# Run the CLI. Pass args after --, e.g. `just run -- ls --project .project.json`
run *args:
    go run . {{args}}

# Build the CLI binary into ./bin/graftcxt
build:
    mkdir -p bin
    go build -o bin/graftcxt .

# Install the CLI into GOBIN/GOPATH/bin from the local checkout
install:
    go install .

# Run Go tests
test:
    go test ./...

# Format Go source
fmt:
    gofmt -w .

# Tidy module dependencies
tidy:
    go mod tidy

# Run formatting, dependency tidy, tests, and build
check: fmt tidy test build
