# Implementation Plan: Add --silent Flag for Silent Output

## Overview
Add a global `--silent` flag that suppresses non-essential output across all graftcxt commands. When enabled, only critical errors should be printed to stderr.

## Changes Required

### 1. Add global flag to root command (`src/cmd/root.go` or in `src/cmd/`)

Add a package-level variable and register the flag:

```go
var silentMode bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&silentMode, "silent", false, "suppress non-essential output")
}
```

### 2. Create a output helper package or use a simple approach

Option A: Create `src/internal/output/output.go` with conditional printing:

```go
package output

import "fmt"

var silent bool

func SetSilent(s bool) { silent = s }

func Printf(format string, args ...interface{}) {
    if !silent {
        fmt.Printf(format, args...)
    }
}

func Println(args ...interface{}) {
    if !silent {
        fmt.Println(args...)
    }
}

func Fprintf(w io.Writer, format string, args ...interface{}) {
    if !silent {
        fmt.Fprintf(w, format, args...)
    }
}
```

Option B: Simpler approach - just check the flag before printing in each command.

### 3. Update commands to respect silent mode

In each command that prints output:
- `add.go`: Wrap output with silent check or use output package
- `sync.go`: Wrap sync output messages
- `ls.go`: Only print table if not silent
- `rem.go`: Wrap confirmation messages

### 4. Initialize silent mode in root command

```go
var rootCmd = &cobra.Command{
    Use:   "graftcxt",
    Short: "Track and sync GitHub repos for AI context",
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        output.SetSilent(silentMode)
    },
}
```

## Acceptance Criteria

1. `graftcxt --silent add owner/repo` - adds repo without printing "Added ..." message
2. `graftcxt --silent sync` - syncs without printing progress
3. `graftcxt --silent ls` - still outputs the list (it's the command output, not a message)
4. Errors are still printed to stderr even in silent mode
5. `graftcxt add --silent owner/repo` also works (persistent flag)

## Files to Modify

- `src/cmd/` - Add root command with persistent flag (may need to create root.go if not exists)
- `src/internal/output/` - Create output helper package (recommended)
- `src/cmd/add.go` - Use output package for messages
- `src/cmd/sync.go` - Use output package for messages
- `src/cmd/ls.go` - Use output package for messages
- `src/cmd/rem.go` - Use output package for messages
