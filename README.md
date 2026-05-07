# graftcxt

> Graft external GitHub repositories into a project's local context so agents can read them without vendoring `.git/` history.

`graftcxt` is a small Go CLI for tracking external GitHub repos in `.project.json` and syncing their working-tree files into your project. It is designed for AI-agent context: pull docs, examples, or reference repositories into `docs/context/`, keep metadata in one JSON file, and refresh only when the upstream tree changes.

## Features

- Track GitHub repositories in `.project.json` under a `context` array.
- Sync repo contents via the GitHub API—no `git clone`, no nested `.git/` directories.
- Support full repos, subdirectories, branches, tags, and GitHub `tree/...` URLs.
- Skip unchanged repositories using saved tree SHAs; force refresh when needed.
- Choose default `docs/context/<name>` output or a custom project-relative directory.
- Interactive forms for missing `add`/`rem` arguments.
- Dry-run and silent modes for scripts.

## Install

### Go install

```bash
go install github.com/rigerc/graftcxt@latest
```

Requires Go 1.25+.

### From source

```bash
git clone https://github.com/rigerc/graftcxt.git
cd graftcxt
go install .
```

Or build a local binary:

```bash
just build
./bin/graftcxt --help
```

## Authentication

`graftcxt` uses the GitHub API and requires authentication. Use either:

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

or authenticate with the GitHub CLI:

```bash
gh auth login
```

Check what `graftcxt` can see:

```bash
graftcxt auth-status
```

## Quick start

From any directory inside a project that contains `.project.json`:

```bash
# Track a repo. This updates .project.json only.
graftcxt add charmbracelet/bubbletea

# Download tracked repos into docs/context/.
graftcxt sync

# Inspect tracked context repos.
graftcxt ls
```

`.project.json` will contain entries like:

```json
{
  "context": [
    {
      "repo": "charmbracelet/bubbletea",
      "name": "bubbletea",
      "last_sync": "2026-05-06T19:07:13Z",
      "last_tree_sha": "..."
    }
  ]
}
```

## Repo references

Use any of these forms:

```bash
# Whole repo, default branch
graftcxt add owner/repo

# Subdirectory from default branch
graftcxt add owner/repo/path/to/subdir

# Explicit branch, tag, or SHA
graftcxt add owner/repo#main
graftcxt add owner/repo/path/to/subdir#v1.2.3

# GitHub tree-style path
graftcxt add owner/repo/tree/main/docs
```

By default, files sync to `docs/context/<derived-name>`, where `<derived-name>` is the repo name or subdirectory leaf.

## Commands

### `add [repo]`

Track a repo in `.project.json`. `add` does not download files; run `sync` afterward.

```bash
graftcxt add rigerc/repo
graftcxt add rigerc/repo/docs --dir docs/vendor/repo-docs
graftcxt add rigerc/repo --dry-run
```

Flags:

- `--dir <path>`: custom project-relative output directory.
- `--dry-run`: preview without modifying `.project.json`.

### `sync`

Download all tracked repos.

```bash
graftcxt sync
graftcxt sync --dry-run
graftcxt sync --force
```

Flags:

- `--dry-run`: show what would sync.
- `--force`: sync even when the saved tree SHA matches upstream.

### `ls`

List tracked repos, output directories, file counts, and sync metadata.

```bash
graftcxt ls
```

### `rem [repo]`

Remove a tracked repo and delete its synced directory.

```bash
graftcxt rem rigerc/repo
graftcxt rem rigerc/repo --dry-run
```

### Global flags

```bash
--project <path>  Path to .project.json. Defaults to searching current dir and parents.
--silent          Suppress non-essential output.
```

## Development

```bash
just test     # go test ./...
just fmt      # gofmt -w .
just tidy     # go mod tidy
just build    # build ./bin/graftcxt
just install  # go install .
just check    # fmt, tidy, test, build
```

## Why?

AI coding agents often need reference material from other repositories, but cloning those repos into a project creates nested Git metadata and noisy state. `graftcxt` stores only working-tree files in a predictable docs directory and keeps the source-of-truth list in `.project.json`.
