# Plan: graftcxt — Convert dg.sh to Go CLI

## Context

`scripts/dg.sh` manages a docs context array in `.project.json` and syncs GitHub repos into `./docs/` via `tiged`/`degit`. Converting it to `graftcxt` (Go CLI) gives type-safe JSON handling, interactive huh forms for missing args, and consistent lipgloss styling — following the patterns established in `../begin/internal/ui/`.

Key behaviour: the go-github Git Tree API returns working-tree content only, so no `.git/` directory ever lands in the output. Synced repos integrate cleanly into the parent git repo — the same guarantee tiged/degit provides via tarball download, but with no external binary dependency.

Standalone Go module at `graftcxt/` in the project root (separate from `src/`).

---

## Target File Structure

```
graftcxt/
├── go.mod                      (module github.com/rigerc/graftcxt, go 1.26)
├── main.go                     (entry: cmd.Execute())
├── cmd/
│   ├── root.go                 (cobra root, persistent --project flag)
│   ├── add.go                  (add subcommand)
│   ├── rem.go                  (rem subcommand)
│   ├── ls.go                   (ls subcommand)
│   └── sync.go                 (sync subcommand)
└── internal/
    ├── context/
    │   └── context.go          (JSON schema, CRUD ops, ParseRepoRef, SyncRepo)
    └── ui/
        └── forms.go            (huh form builders + graftcxtTheme)
```

---

## Dependencies

```
charm.land/huh/v2
charm.land/lipgloss/v2
github.com/spf13/cobra
github.com/google/go-github/v68      (already in mkproject src/go.mod)
golang.org/x/oauth2                  (optional: GITHUB_TOKEN auth to raise rate limit)
```

No bubbletea needed — huh runs standalone for simple one-shot forms.

---

## internal/context/context.go

`.project.json` already contains `project` and `skills` keys. The `context` array is added alongside them:

```json
{
  "project": { ... },
  "skills": {},
  "context": [
    { "repo": "owner/repo", "name": "repo", "last_sync": "2026-05-06T..." }
  ]
}
```

Go types — `json.RawMessage` for `project` and `skills` preserves existing content on round-trip:

```go
type ContextEntry struct {
    Repo     string  `json:"repo"`
    Name     string  `json:"name"`
    LastSync *string `json:"last_sync"`
}

type ProjectFile struct {
    Project json.RawMessage `json:"project"`
    Skills  json.RawMessage `json:"skills"`
    Context []ContextEntry  `json:"context"`
}
```

Functions:
- `Load(path string) (*ProjectFile, error)` — reads `.project.json`; initialises `Context: []` if key absent
- `Save(path string, pf *ProjectFile) error` — atomic write via temp file + `os.Rename`
- `ParseRepoRef(repoID string) (owner, repo, subdir, ref string, err error)` — parses `owner/repo[/subdir][#ref]`
- `ParseRepoName(repoID string) (string, error)` — returns the leaf dir name used under `docs/`
- `HasEntry(pf *ProjectFile, repoID string) bool`
- `AddEntry(pf *ProjectFile, e ContextEntry)`
- `RemoveEntry(pf *ProjectFile, repoID string)`
- `SyncRepo(repoID, destPath string, gh *github.Client) error`

**SyncRepo via go-github:**
1. `ParseRepoRef()` → `owner`, `repo`, `subdir`, `ref` (defaults to `"HEAD"`)
2. `gh.Git.GetTree(ctx, owner, repo, ref, true)` — full recursive tree
3. Filter blob entries to `subdir` prefix if set
4. For each blob: `gh.Git.GetBlob(ctx, owner, repo, sha)` → decode → write to `destPath/relative-path`
5. Delete files in `destPath` not present in the fetched tree (clean sync)

Auth: reads `GITHUB_TOKEN` env var; if set, builds an oauth2 HTTP client (5000 req/hr); otherwise unauthenticated (60 req/hr).

---

## internal/ui/forms.go

Following `begin/internal/ui/picker.go` patterns:

```go
// InputRepoForm shows a huh input for a repo identifier.
// Returns the entered value, or errAborted if user exits.
func InputRepoForm() (string, error)

// SelectRepoForm shows a huh select over tracked entries for removal.
// Returns selected repo identifier, or errAborted if user exits.
func SelectRepoForm(entries []context.ContextEntry) (string, error)

// graftcxtTheme returns a lipgloss-styled huh.Theme (mirrors beginTheme()).
func graftcxtTheme() *huh.Theme
```

Each form: `huh.NewForm(huh.NewGroup(...)).WithTheme(graftcxtTheme()).Run()`.

---

## cmd/ — Cobra Commands

### root.go
- Persistent flag `--project` (default `".project.json"`)
- `cobra.Command{Use: "graftcxt", Short: "Graft external GitHub repos into your project context"}`

### add.go
```
graftcxt add [repo]
```
- No arg → `ui.InputRepoForm()` prompt
- `context.ParseRepoName()` to validate
- `context.HasEntry()` → error if already tracked
- `context.SyncRepo()` → `context.AddEntry()` → `context.Save()`

### rem.go
```
graftcxt rem [repo]
```
- No arg → `ui.SelectRepoForm(pf.Context)`
- `context.HasEntry()` → error if not tracked
- `os.RemoveAll(docsDir/name)` + `context.RemoveEntry()` → `context.Save()`

### ls.go
```
graftcxt ls
```
- Tabular lipgloss output, columns: `REPO`, `NAME`, `LAST_SYNC`

### sync.go
```
graftcxt sync
```
- Iterate `pf.Context`, call `context.SyncRepo()` for each, update `last_sync`
- Print count on completion

---

## Key Design Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| huh mode | standalone `form.Run()` | One-shot inputs; no full Elm loop needed |
| Aborted form | sentinel `errAborted` | Clean exit without error message |
| JSON write | temp file + `os.Rename` | Atomic, mirrors shell's `mktemp` + `mv` |
| Repo syncing | go-github Git Trees API + GetBlob | No external binary; no `.git/` in output |
| Auth | `GITHUB_TOKEN` → oauth2 client | 60 req/hr unauth → 5000 req/hr with token |
| docsDir | derived from project file dir + `docs/` | Mirrors shell's `DOCS_DIR` derivation |

---

## Critical Files to Create

1. `graftcxt/go.mod`
2. `graftcxt/main.go`
3. `graftcxt/cmd/root.go`
4. `graftcxt/cmd/add.go`
5. `graftcxt/cmd/rem.go`
6. `graftcxt/cmd/ls.go`
7. `graftcxt/cmd/sync.go`
8. `graftcxt/internal/context/context.go`
9. `graftcxt/internal/ui/forms.go`

---

## Verification

```bash
cd graftcxt && go mod tidy && go build -o graftcxt .

# List on empty context
./graftcxt ls

# Add interactively (no arg → huh prompt)
./graftcxt add

# Add with arg
./graftcxt add charmbracelet/bubbletea/tree/main/examples

# Confirm entry listed
./graftcxt ls

# Sync all
./graftcxt sync

# Remove interactively (no arg → huh select)
./graftcxt rem

# Verify docs/ cleaned and .project.json updated
cat ../.project.json
```
