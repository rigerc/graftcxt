package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	ctx "github.com/rigerc/graftcxt/internal/context"
	"github.com/rigerc/graftcxt/internal/ui"
	"github.com/spf13/cobra"
	"github.com/google/go-github/v68/github"
)

var syncDryRun bool
var syncForce bool

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
		skippedRepos := 0
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
				// Dry-run: check if repo would be synced or skipped
				wouldSkip, reason := wouldSkipSync(gh, e)
				if syncForce {
					wouldSkip = false
					reason = "(force enabled)"
				}
				if wouldSkip {
					fmt.Fprintf(cmd.OutOrStdout(), "     [DRY-RUN] Would skip: %s\n", reason)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "     [DRY-RUN] Would sync to: %s\n", dest)
				}
				continue
			}

			// Check if sync is needed (unless --force is set)
			if !syncForce {
				skip, reason := shouldSkipSync(gh, e)
				if skip {
					fmt.Fprintf(cmd.OutOrStdout(), "     ⏭️  Skipped: %s\n", reason)
					skippedRepos++
					fmt.Fprintln(cmd.OutOrStdout())
					continue
				}
			}

			// Count files before sync for comparison
			filesBefore := countFiles(dest)

			fmt.Fprintf(cmd.OutOrStdout(), "     Syncing...\n")

			// Fetch current tree SHA before sync
			owner, repo, _, ref, err := ctx.ParseRepoRef(e.Repo)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "     ❌ Error: %v\n", err)
				return err
			}
			tree, _, err := gh.Git.GetTree(context.Background(), owner, repo, ref, false)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "     ⚠️  Warning: Could not fetch tree SHA: %v\n", err)
				// Continue with sync - don't let tree SHA check block sync
			}
			currentTreeSHA := ""
			if tree != nil {
				currentTreeSHA = tree.GetSHA()
			}

			// Use a live-updating spinner for visual feedback during sync
			err = ui.RunLiveSpinner(
				fmt.Sprintf("Syncing %s", e.Repo),
				func(updateTitle func(string)) error {
					return ctx.SyncRepo(e.Repo, dest, gh, updateTitle)
				},
			)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "     ❌ Error: %v\n", err)
				return err
			}

			filesAfter := countFiles(dest)
			newFiles := filesAfter - filesBefore
			totalFiles += filesAfter

			now := ctx.NowString()
			pf.Context[i].LastSync = &now
			// Store the tree SHA for future change detection
			if currentTreeSHA != "" {
				pf.Context[i].LastTreeSHA = &currentTreeSHA
			}
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
			fmt.Fprintf(cmd.OutOrStdout(), "✅ Synced %d repos, skipped %d repos, %d total files in %s\n", syncedRepos, skippedRepos, totalFiles, elapsed)
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

// shouldSkipSync checks if a repo has changed since last sync.
// Returns (shouldSkip, reason).
func shouldSkipSync(gh *github.Client, e ctx.ContextEntry) (bool, string) {
	// Never synced before - don't skip
	if e.LastTreeSHA == nil {
		return false, ""
	}

	// Parse repo ref to get owner, repo, ref
	owner, repo, _, ref, err := ctx.ParseRepoRef(e.Repo)
	if err != nil {
		// If we can't parse, don't skip - let sync try
		return false, ""
	}

	// Fetch current tree SHA (not recursive - we just need the root tree SHA)
	tree, _, err := gh.Git.GetTree(context.Background(), owner, repo, ref, false)
	if err != nil {
		// API error - don't skip, let sync handle it
		return false, ""
	}

	currentSHA := tree.GetSHA()
	if currentSHA == *e.LastTreeSHA {
		return true, "no changes since last sync"
	}

	return false, ""
}

// wouldSkipSync checks if a repo would be skipped during sync.
// Unlike shouldSkipSync, this is used for dry-run mode to show accurate status.
func wouldSkipSync(gh *github.Client, e ctx.ContextEntry) (bool, string) {
	// Never synced before - would sync
	if e.LastTreeSHA == nil {
		return false, "never synced"
	}

	// Parse repo ref to get owner, repo, ref
	owner, repo, _, ref, err := ctx.ParseRepoRef(e.Repo)
	if err != nil {
		return false, "parse error, would sync"
	}

	// Fetch current tree SHA
	tree, _, err := gh.Git.GetTree(context.Background(), owner, repo, ref, false)
	if err != nil {
		return false, "API error, would sync"
	}

	currentSHA := tree.GetSHA()
	if currentSHA == *e.LastTreeSHA {
		return true, "no changes since last sync"
	}

	return false, "changes detected"
}

func init() {
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "preview sync without downloading files")
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "force sync even if no changes detected")
	rootCmd.AddCommand(syncCmd)
}
