set dotenv-load := true

# Show available commands
default:
    just --list

# Run the CLI. Pass args after --, e.g. `just run -- ls --project ../.project.json`
run *args:
    cd src && go run . {{args}}

# Build the CLI binary into ./bin/graftcxt
build:
    mkdir -p bin
    cd src && go build -o ../bin/graftcxt .

# Run Go tests
test:
    cd src && go test ./...

# Format Go source
fmt:
    cd src && gofmt -w .

# Tidy module dependencies
tidy:
    cd src && go mod tidy

# Run formatting, dependency tidy, tests, and build
check: fmt tidy test build
