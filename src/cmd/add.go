package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	ctx "github.com/rigerc/graftcxt/internal/context"
	"github.com/rigerc/graftcxt/internal/output"
	"github.com/rigerc/graftcxt/internal/ui"
	"github.com/spf13/cobra"
)

var (
	addDir    string
	addDryRun bool

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
				repoID, dir, err = ui.InputRepoForm()
				if errors.Is(err, ui.ErrAborted) {
					return nil
				}
				if err != nil {
					return err
				}
			}
			if err := validateAddDir(dir); err != nil {
				return err
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

			// Calculate destination path
			entry := ctx.ContextEntry{Repo: repoID, Name: name, Dir: dir}
			dest := contextEntryPath(projectFile, entry)

			if addDryRun {
				output.Fprintf(cmd.OutOrStdout(), "[DRY-RUN] Would add to project file:\n")
				output.Fprintf(cmd.OutOrStdout(), "  Repo: %s\n", repoID)
				output.Fprintf(cmd.OutOrStdout(), "  Name: %s\n", name)
				output.Fprintf(cmd.OutOrStdout(), "  Dir:  %s\n", dest)
				output.Fprintf(cmd.OutOrStdout(), "\nRun 'graftcxt sync' to download the repository files.\n")
				return nil
			}

			// Just add to project file (no syncing)
			ctx.AddEntry(pf, entry)
			if err := ctx.Save(projectFile, pf); err != nil {
				return err
			}
			output.Fprintf(cmd.OutOrStdout(), "Added %s -> %s\n", repoID, dest)
			output.Fprintf(cmd.OutOrStdout(), "Run 'graftcxt sync' to download the repository files.\n")
			return nil
		},
	}
)

func validateAddDir(dir string) error {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil
	}
	if filepath.IsAbs(dir) || strings.HasPrefix(dir, "~") {
		return fmt.Errorf("directory must be relative to the project")
	}
	if strings.Contains("/"+filepath.ToSlash(dir)+"/", "/../") {
		return fmt.Errorf("directory must not contain ..")
	}
	return nil
}

func init() {
	addCmd.Flags().StringVar(&addDir, "dir", "", "custom directory for this repo (relative to project dir)")
	addCmd.Flags().BoolVar(&addDryRun, "dry-run", false, "preview addition without modifying project file")
	rootCmd.AddCommand(addCmd)
}
