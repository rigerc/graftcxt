# Implementation Plan: Switch to koanf for JSON Handling

## Overview
Replace the standard `encoding/json` package with `github.com/knadh/koanf/v2` for JSON configuration handling in `src/internal/context/context.go`. Koanf provides a cleaner API, better support for nested structures, and additional features like merging and watching for changes.

## Why Koanf?

1. **Cleaner API**: Koanf provides a unified interface for reading/writing configuration
2. **Better structure**: Supports nested keys and complex structures more elegantly
3. **Extensible**: Can easily add support for other formats (YAML, TOML) in the future
4. **Merge support**: Can merge multiple config sources

## Changes Required

### 1. Add koanf dependency

```bash
cd src/
go get github.com/knadh/koanf/v2
go get github.com/knadh/koanf/v2/parsers/json
go get github.com/knadh/koanf/v2/providers/rawbytes
```

### 2. Update `src/internal/context/context.go`

Replace the current JSON handling:

**Current implementation:**
```go
import "encoding/json"

type ProjectFile struct {
    Project json.RawMessage `json:"project,omitempty"`
    Skills  json.RawMessage `json:"skills,omitempty"`
    Context []ContextEntry  `json:"context"`
}

func Load(path string) (*ProjectFile, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var pf ProjectFile
    if err := json.Unmarshal(b, &pf); err != nil {
        return nil, err
    }
    if pf.Context == nil {
        pf.Context = []ContextEntry{}
    }
    return &pf, nil
}

func Save(path string, pf *ProjectFile) error {
    if pf.Context == nil {
        pf.Context = []ContextEntry{}
    }
    b, err := json.MarshalIndent(pf, "", "  ")
    if err != nil {
        return err
    }
    b = append(b, '\n')
    // ... write to temp file and rename
}
```

**New implementation with koanf:**
```go
import (
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/v2/parsers/json"
    "github.com/knadh/koanf/v2/providers/rawbytes"
)

type ProjectFile struct {
    Project json.RawMessage `json:"project,omitempty"`
    Skills  json.RawMessage `json:"skills,omitempty"`
    Context []ContextEntry  `json:"context"`
}

func Load(path string) (*ProjectFile, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    k := koanf.New(".")
    if err := k.Load(rawbytes.Provider(b), json.Parser()); err != nil {
        return nil, err
    }
    
    var pf ProjectFile
    if err := k.Unmarshal("", &pf); err != nil {
        return nil, err
    }
    
    if pf.Context == nil {
        pf.Context = []ContextEntry{}
    }
    return &pf, nil
}

func Save(path string, pf *ProjectFile) error {
    if pf.Context == nil {
        pf.Context = []ContextEntry{}
    }
    
    // Marshal using koanf
    k := koanf.New(".")
    if err := k.Unmarshal("", &pf); err != nil {
        return err
    }
    
    b, err := k.Marshal(json.Parser())
    if err != nil {
        return err
    }
    
    b = append(b, '\n')
    
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return err
    }
    
    tmp, err := os.CreateTemp(dir, ".project-*.json")
    if err != nil {
        return err
    }
    tmpName := tmp.Name()
    defer os.Remove(tmpName)
    
    if _, err := tmp.Write(b); err != nil {
        _ = tmp.Close()
        return err
    }
    if err := tmp.Close(); err != nil {
        return err
    }
    
    return os.Rename(tmpName, path)
}
```

### 3. Update `go.mod` and `go.sum`

Run `go mod tidy` after adding the dependencies.

## Acceptance Criteria

1. `graftcxt add owner/repo` - still works correctly, project file is saved in valid JSON
2. `graftcxt sync` - still reads project file correctly
3. `graftcxt ls` - still displays tracked repos correctly
4. Project file format remains unchanged (valid JSON with same structure)
5. All existing tests pass (if any)
6. Koanf dependency is properly added to `go.mod`

## Files to Modify

- `src/go.mod` - Add koanf dependencies
- `src/go.sum` - Updated automatically by go mod tidy
- `src/internal/context/context.go` - Replace json with koanf

## Notes

- The `json.RawMessage` type for `Project` and `Skills` fields can still be used with koanf
- Koanf's `rawbytes` provider is used to load from byte slices (read from file)
- The JSON parser ensures we maintain compatibility with the existing JSON format
- Consider adding a wrapper or helper if more complex operations are needed in the future
