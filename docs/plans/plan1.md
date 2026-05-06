# Implementation Plan: Per-Repo Configurable Save Directory

## Goal
Add ability to specify a save directory **per configured repository** (not globally), with a default of `docs/context/` for each repo.

## Current Architecture
- CLI tool written in Go using cobra
- Commands: `add`, `sync`, `ls`, `rem`
- ProjectFile contains `[]ContextEntry` where each entry has: `Repo`, `Name`, `LastSync`
- All commands use `docsDir()` which returns `filepath.Dir(projectFile) + "/docs"` as base
- Destination path: `filepath.Join(docsDir(projectFile), name)`

## Implementation Steps

### 1. Update ContextEntry Struct (src/internal/context/context.go)
- Add `Dir` or `Path` field to `ContextEntry` struct
- JSON field name: `dir` (relative path from project file directory)
- Type: `string`
- This stores the per-repo directory override

```go
type ContextEntry struct {
    Repo     string  `json:"repo"`
    Name     string  `json:"name"`
    LastSync *string `json:"last_sync"`
    Dir      string  `json:"dir,omitempty"`  // NEW: per-repo directory (relative to project file dir)
}
```

### 2. Create Helper Function for Path Resolution (src/internal/context/context.go)
- Add function to compute full path for a context entry
- Respects per-repo Dir if set, falls back to default
- Default: `filepath.Join(filepath.Dir(projectFile), "docs", "context", entry.Name)`
- Or simpler default: `filepath.Join(filepath.Dir(projectFile), "docs", "context", entry.Name)`
- Wait - need to clarify: should default be `docs/context/` as stated, or should we keep per-repo base as configurable too?
- Since requirement says "per configured repo", the Dir field stores the RELATIVE path from project file directory

### 3. Create Context Path Resolver Function (src/cmd/root.go)
- New function: `contextEntryPath(projectFile string, entry ContextEntry) string`
- Computes full path based on entry.Dir or default
- If entry.Dir is set: `filepath.Join(filepath.Dir(projectFile), entry.Dir)`
- If entry.Dir is empty: `filepath.Join(filepath.Dir(projectFile), "docs", "context", entry.Name)`

### 4. Update Add Command (src/cmd/add.go)
- Add flag: `--dir` to specify custom directory for this repo
- Flag accepts relative path (relative to project file directory)
- When adding, set entry.Dir = flag value if provided
- Update dest calculation to use entry.Dir
- Default behavior (no --dir): saves to docs/context/{name}

### 5. Update Sync Command (src/cmd/sync.go)
- Iterate through all entries, compute path using `contextEntryPath()`
- Each repo syncs to its own configured directory

### 6. Update Remove Command (src/cmd/rem.go)
- Compute removal path using `contextEntryPath()`
- Removes from correct custom directory

### 7. Update List Command (src/cmd/ls.go)
- Add column showing the custom directory path (if not default)
- Shows relative path or "default" indicator

### 8. Backward Compatibility
- When loading old project files (entries without Dir field), Dir is empty string
- Empty Dir triggers default behavior: docs/context/{name}
- Existing repos in old location need migration path
- Option: Add `migrate` command or provide manual migration instructions

### 9. Add Migration Command (Optional but Recommended)
- `graftcxt migrate` - moves repos from old location to new default
- Or: detect repos in old docs/{name} location and offer to move

### 10. Update Tests
- Test adding repo with custom --dir flag
- Test adding repo without --dir (default to docs/context/)
- Test sync with mixed custom and default directories
- Test backward compatibility (loading old entries)
- Test list with custom directories

### 11. Documentation Updates
- Update command help strings to mention --dir flag
- Update README with examples

## Technical Details

### ContextEntry Enhancement
```go
type ContextEntry struct {
    Repo     string  `json:"repo"`
    Name     string  `json:"name"`
    LastSync *string `json:"last_sync"`
    Dir      string  `json:"dir,omitempty"`  // relative to project file directory
}
```

### Path Resolution Logic
```go
func ContextEntryPath(projectFile string, entry ContextEntry) string {
    baseDir := filepath.Dir(projectFile)
    if entry.Dir != "" {
        return filepath.Join(baseDir, entry.Dir)
    }
    // Default: docs/context/{name}
    return filepath.Join(baseDir, "docs", "context", entry.Name)
}
```

### Add Command Flag
```go
var dirOverride string
addCmd.Flags().StringVar(&dirOverride, "dir", "", "custom directory for this repo (relative to project dir)")
```

When creating entry:
```go
entry := ContextEntry{
    Repo: repoID,
    Name: name,
    Dir: dirOverride,  // empty = use default
}
```

## File-by-File Changes

### 1. src/internal/context/context.go
- Add `Dir string` field to `ContextEntry`
- Consider adding: `func (e ContextEntry) FullPath(projectFile string) string`

### 2. src/cmd/root.go  
- Add `contextEntryPath(projectFile string, entry ContextEntry) string`
- Deprecate/keep `docsDir()` for other uses if needed

### 3. src/cmd/add.go
- Add `--dir` flag
- Set entry.Dir = flag value
- Update dest to use `contextEntryPath()`
- Update help text: "Add and sync a GitHub repo (default to docs/context/)"

### 4. src/cmd/sync.go
- Replace `filepath.Join(docsDir(projectFile), e.Name)` with `contextEntryPath(projectFile, e)`

### 5. src/cmd/rem.go
- Replace with `contextEntryPath(projectFile, e)` for removal

### 6. src/cmd/ls.go
- Add column showing directory or "(default)"
- Format nicely

### 7. src/cmd/root_test.go
- Add test for contextEntryPath
- Test with and without Dir set
- Test backward compatibility

## Backward Compatibility Strategy

**Old format project file:**
```json
{
  "context": [
    {"repo": "owner/repo", "name": "repo", "last_sync": "..."}
  ]
}
```

**New format project file (after adding repo with --dir):**
```json
{
  "context": [
    {"repo": "owner/repo1", "name": "repo1", "last_sync": "..."},  // default
    {"repo": "owner/repo2", "name": "repo2", "last_sync": "...", "dir": "custom/path"}  // custom
  ]
}
```

**Migration:**
- Old repos remain in `docs/repo_name` unless user manually moves them or uses new --dir flag
- New repos go to `docs/context/repo_name` by default
- Users can specify `--dir docs/repo_name` to match old behavior

## Example Usage

### Default Behavior (new repos)
```bash
./graftcxt add owner/repo1
# Saves to: docs/context/repo1

./graftcxt add owner/repo2 --dir my-context
# Saves to: my-context/repo2 (relative to project file)

./graftcxt add owner/repo3 --dir docs/legacy
# Saves to: docs/legacy/repo3
```

### Listing
```bash
./graftcxt ls

REPO                                      NAME          LAST_SYNC           DIR
owner/repo1                               repo1         2026-05-06T16:14:40Z  (default)
owner/repo2                               repo2         2026-05-06T16:15:00Z  my-context
```

### Syncing
```bash
./graftcxt sync
# Syncs repo1 to docs/context/repo1
# Syncs repo2 to my-context/repo2
# etc.
```

## Edge Cases to Handle

1. **Empty Dir field** → use default (docs/context/{name})
2. **Relative paths** → always relative to project file directory
3. **Absolute paths** → should we allow? Probably not, keep relative
4. **Parent directory traversal (`../`)** → validate/sanitize path
5. **Changing Dir after repo is synced** → old files remain in old location
6. **Removing repo with custom Dir** → remove from correct location
7. **Multiple repos with same Dir** → allowed, they share directory
8. **Dir set but doesn't exist** → create on sync

## Open Questions

1. Should we add validation to prevent Dir from being empty string (use omitempty)?
2. Should we provide a `--dir` flag on `sync` to override temporarily? (probably not needed)
3. Should we add a `graftcxt move` command to migrate repos between directories?
4. Should the default be `docs/context/` or should it be configurable globally too?

## Success Criteria

- [ ] Can add repo with custom --dir flag
- [ ] Can add repo without --dir (defaults to docs/context/)
- [ ] Sync works with mixed custom and default directories
- [ ] Remove works with custom directories
- [ ] List shows directory info
- [ ] Backward compatible with old project files
- [ ] Tests pass
- [ ] Documentation updated
