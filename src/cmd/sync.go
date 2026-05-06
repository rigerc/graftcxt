package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	ctx "github.com/rigerc/graftcxt/internal/context"
	"github.com/spf13/cobra"
)

var syncDryRun bool

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
		if len(pf.Context) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No tracked repos to sync. Use 'graftcxt add' to add repositories.")
			return nil
		}

		gh, err := ctx.NewGitHubClient()
		if err != nil {
			return err
		}

		// Header
		header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", header.Render("🔄 Syncing tracked repositories..."))
		fmt.Fprintln(cmd.OutOrStdout())

		totalFiles := 0
		syncedRepos := 0
		startTime := time.Now()

		for i, e := range pf.Context {
			// Show repo info
			lastSync := "never"
			if e.LastSync != nil {
				lastSync = *e.LastSync
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  📦 %s\n", e.Repo)
			fmt.Fprintf(cmd.OutOrStdout(), "     Last sync: %s\n", lastSync)

			dest := contextEntryPath(projectFile, e)

			if syncDryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "     [DRY-RUN] Would sync to: %s\n", dest)
				continue
			}

			// Count files before sync for comparison
			filesBefore := countFiles(dest)

			fmt.Fprintf(cmd.OutOrStdout(), "     Syncing...\n")
			if err := ctx.SyncRepoWithWriter(e.Repo, dest, gh, cmd.OutOrStdout()); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "     ❌ Error: %v\n", err)
				return err
			}

			filesAfter := countFiles(dest)
			newFiles := filesAfter - filesBefore
			totalFiles += filesAfter

			now := ctx.NowString()
			pf.Context[i].LastSync = &now
			syncedRepos++

			fmt.Fprintf(cmd.OutOrStdout(), "     ✅ Synced: %d files (added %d new)\n", filesAfter, newFiles)
			fmt.Fprintf(cmd.OutOrStdout(), "     Last sync: %s\n", now)
			fmt.Fprintln(cmd.OutOrStdout())
		}

		if !syncDryRun {
			if err := ctx.Save(projectFile, pf); err != nil {
				return err
			}
		}

		// Summary
		elapsed := time.Since(startTime).Round(time.Millisecond)
		fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("─", 50))
		if syncDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "[DRY-RUN] Would sync %d repos\n", len(pf.Context))
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "✅ Synced %d repos, %d total files in %s\n", syncedRepos, totalFiles, elapsed)
		}
		return nil
	},
}

func countFiles(dir string) int {
	count := 0
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			count++
		}
		return nil
	})
	return count
}

func init() {
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "preview sync without downloading files")
	rootCmd.AddCommand(syncCmd)
}
