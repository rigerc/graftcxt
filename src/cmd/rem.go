package cmd

import (
	"errors"
	"fmt"
	"os"

	ctx "github.com/rigerc/graftcxt/internal/context"
	"github.com/rigerc/graftcxt/internal/ui"
	"github.com/spf13/cobra"
)

var remCmd = &cobra.Command{
	Use:   "rem [repo]",
	Short: "Remove a tracked repo from docs/ and .project.json",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectFile, err := resolvedProjectPath()
		if err != nil {
			return err
		}
		pf, err := ctx.Load(projectFile)
		if err != nil {
			return err
		}
		repoID := ""
		if len(args) == 1 {
			repoID = args[0]
		} else {
			repoID, err = ui.SelectRepoForm(pf.Context)
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}
		}
		if !ctx.HasEntry(pf, repoID) {
			return fmt.Errorf("%s is not tracked", repoID)
		}
		var entry ctx.ContextEntry
		for _, e := range pf.Context {
			if e.Repo == repoID {
				entry = e
				break
			}
		}
		if err := os.RemoveAll(contextEntryPath(projectFile, entry)); err != nil {
			return err
		}
		ctx.RemoveEntry(pf, repoID)
		if err := ctx.Save(projectFile, pf); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", repoID)
		return nil
	},
}

func init() { rootCmd.AddCommand(remCmd) }
