package cmd

import (
	"fmt"

	ctx "github.com/rigerc/graftcxt/internal/context"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all tracked context repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectFile, err := resolvedProjectPath()
		if err != nil {
			return err
		}
		pf, err := ctx.Load(projectFile)
		if err != nil {
			return err
		}
		gh := ctx.NewGitHubClient()
		for i, e := range pf.Context {
			fmt.Fprintf(cmd.OutOrStdout(), "Syncing %s...\n", e.Repo)
			dest := contextEntryPath(projectFile, e)
			if err := ctx.SyncRepo(e.Repo, dest, gh); err != nil {
				return err
			}
			now := ctx.NowString()
			pf.Context[i].LastSync = &now
		}
		if err := ctx.Save(projectFile, pf); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Synced %d repos\n", len(pf.Context))
		return nil
	},
}

func init() { rootCmd.AddCommand(syncCmd) }
