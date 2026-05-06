# Implementation Plan: Edit 'add' Prompt/Form to Ask for Directory

## Overview
Currently, the `add` command has a `--dir` flag for command-line usage, but the interactive form (`InputRepoForm()`) doesn't ask for the directory. This plan modifies the form to also prompt for the target directory.

## Changes Required

### 1. Update `InputRepoForm()` in `src/internal/ui/forms.go`

Modify the function to return both repo and directory:

```go
func InputRepoForm() (repo, dir string, err error) {
    var (
        repo string
        dir  string
    )
    
    form := huh.NewForm(huh.NewGroup(
        huh.NewInput().
            Title("Repository").
            Description("Enter the GitHub repository to add to your context.").
            Placeholder("e.g., owner/repo or owner/repo/subdir#main").
            Value(&repo).
            CharLimit(200).
            Validate(func(s string) error {
                if s == "" {
                    return fmt.Errorf("repository is required")
                }
                if !isValidRepoFormat(s) {
                    return fmt.Errorf("invalid format (expected: owner/repo[/subdir][#ref])")
                }
                return nil
            }),
        huh.NewInput().
            Title("Directory (optional)").
            Description("Custom directory for this repo (relative to project dir). Leave empty for default.").
            Placeholder("e.g., docs/custom-context").
            Value(&dir).
            CharLimit(200),
    )).WithTheme(graftcxtTheme())
    
    if err := form.Run(); err != nil {
        return "", "", mapAbort(err)
    }
    
    return repo, dir, nil
}
```

### 2. Update the call site in `src/cmd/add.go`

```go
var (
    addDir   string
    addDryRun bool
)

addCmd = &cobra.Command{
    Use:   "add [repo]",
    Short: "Add a GitHub repo to project tracking (use 'sync' to download)",
    Args:  cobra.MaximumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        repoID := ""
        dir := addDir
        
        if len(args) == 1 {
            repoID = args[0]
        } else {
            var err error
            repoID, dir, err = ui.InputRepoForm()  // Updated to receive dir
            if errors.Is(err, ui.ErrAborted) {
                return nil
            }
            if err != nil {
                return err
            }
        }
        
        name, err := ctx.ParseRepoName(repoID)
        if err != nil {
            return err
        }
        
        projectFile, err := resolvedProjectPath()
        if err != nil {
            return err
        }
        
        pf, err := ctx.Load(projectFile)
        if err != nil {
            return err
        }
        
        if ctx.HasEntry(pf, repoID) {
            return fmt.Errorf("%s is already tracked", repoID)
        }

        entry := ctx.ContextEntry{Repo: repoID, Name: name, Dir: dir}
        dest := contextEntryPath(projectFile, entry)

        if addDryRun {
            fmt.Fprintf(cmd.OutOrStdout(), "[DRY-RUN] Would add to project file:\n")
            fmt.Fprintf(cmd.OutOrStdout(), "  Repo: %s\n", repoID)
            fmt.Fprintf(cmd.OutOrStdout(), "  Name: %s\n", name)
            fmt.Fprintf(cmd.OutOrStdout(), "  Dir:  %s\n", dest)
            fmt.Fprintf(cmd.OutOrStdout(), "\nRun 'graftcxt sync' to download the repository files.\n")
            return nil
        }

        ctx.AddEntry(pf, entry)
        if err := ctx.Save(projectFile, pf); err != nil {
            return err
        }
        
        fmt.Fprintf(cmd.OutOrStdout(), "Added %s -> %s\n", repoID, dest)
        fmt.Fprintf(cmd.OutOrStdout(), "Run 'graftcxt sync' to download the repository files.\n")
        return nil
    },
}
```

## Acceptance Criteria

1. Running `graftcxt add` without arguments shows a two-field form (Repository and Directory)
2. Directory field is optional - pressing enter with empty value uses default
3. The `--dir` flag still works for command-line usage
4. Form validation: directory path should be relative (no `..` or absolute paths)
5. The destination path is correctly calculated based on the provided directory

## Files to Modify

- `src/internal/ui/forms.go` - Update `InputRepoForm()` signature and implementation
- `src/cmd/add.go` - Update call to `InputRepoForm()` to handle the new return value
