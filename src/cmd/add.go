package cmd

import (
	"errors"
	"fmt"

	ctx "github.com/rigerc/graftcxt/internal/context"
	"github.com/rigerc/graftcxt/internal/ui"
	"github.com/spf13/cobra"
)

var (
	addDir   string
	addDryRun bool

	addCmd = &cobra.Command{
		Use:   "add [repo]",
		Short: "Add a GitHub repo to project tracking (use 'sync' to download)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoID := ""
			if len(args) == 1 {
				repoID = args[0]
			} else {
				var err error
				repoID, err = ui.InputRepoForm()
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

			// Calculate destination path
			entry := ctx.ContextEntry{Repo: repoID, Name: name, Dir: addDir}
			dest := contextEntryPath(projectFile, entry)

			if addDryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "[DRY-RUN] Would add to project file:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  Repo: %s\n", repoID)
				fmt.Fprintf(cmd.OutOrStdout(), "  Name: %s\n", name)
				fmt.Fprintf(cmd.OutOrStdout(), "  Dir:  %s\n", dest)
				fmt.Fprintf(cmd.OutOrStdout(), "\nRun 'graftcxt sync' to download the repository files.\n")
				return nil
			}

			// Just add to project file (no syncing)
			ctx.AddEntry(pf, entry)
			if err := ctx.Save(projectFile, pf); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added %s -> %s\n", repoID, dest)
			fmt.Fprintf(cmd.OutOrStdout(), "Run 'graftcxt sync' to download the repository files.\n")
			return nil
		},
	}
)

func init() {
	addCmd.Flags().StringVar(&addDir, "dir", "", "custom directory for this repo (relative to project dir)")
	addCmd.Flags().BoolVar(&addDryRun, "dry-run", false, "preview addition without modifying project file")
	rootCmd.AddCommand(addCmd)
}
