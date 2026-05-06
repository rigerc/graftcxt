package cmd

import (
	"errors"
	"fmt"

	ctx "github.com/rigerc/graftcxt/internal/context"
	"github.com/rigerc/graftcxt/internal/ui"
	"github.com/spf13/cobra"
)

var (
	addDir string

	addCmd = &cobra.Command{
		Use:   "add [repo]",
		Short: "Add and sync a GitHub repo (default to docs/context/)",
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
			dest := contextEntryPath(projectFile, ctx.ContextEntry{Name: name, Dir: addDir})
			if err := ctx.SyncRepo(repoID, dest, ctx.NewGitHubClient()); err != nil {
				return err
			}
			now := ctx.NowString()
			ctx.AddEntry(pf, ctx.ContextEntry{Repo: repoID, Name: name, LastSync: &now, Dir: addDir})
			if err := ctx.Save(projectFile, pf); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added %s -> %s\n", repoID, dest)
			return nil
		},
	}
)

func init() {
	addCmd.Flags().StringVar(&addDir, "dir", "", "custom directory for this repo (relative to project dir)")
	rootCmd.AddCommand(addCmd)
}
