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

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tracked context repos with detailed information",
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
			fmt.Fprintln(cmd.OutOrStdout(), "No tracked repositories.")
			fmt.Fprintln(cmd.OutOrStdout(), "Use 'graftcxt add <repo>' to add a repository.")
			return nil
		}

		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
		subtleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n", headerStyle.Render("📚 Tracked Repositories"))

		totalFiles := 0
		for _, e := range pf.Context {
			// Calculate directory path
			dest := contextEntryPath(projectFile, e)

			// Count files in the directory
			fileCount := 0
			_ = filepath.WalkDir(dest, func(path string, d os.DirEntry, err error) error {
				if err == nil && !d.IsDir() {
					fileCount++
				}
				return nil
			})
			totalFiles += fileCount

			// Format last sync time
			lastSync := "never"
			if e.LastSync != nil {
				// Try to parse and format nicely
				t, err := time.Parse(time.RFC3339, *e.LastSync)
				if err == nil {
					lastSync = t.Local().Format("2006-01-02 15:04:05")
				} else {
					lastSync = *e.LastSync
				}
			}

			// Print repo info
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", headerStyle.Render("● "+e.Repo))
			fmt.Fprintf(cmd.OutOrStdout(), "    Name:      %s\n", e.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "    Directory: %s\n", dest)
			fmt.Fprintf(cmd.OutOrStdout(), "    Files:     %d\n", fileCount)
			fmt.Fprintf(cmd.OutOrStdout(), "    Last sync: %s\n", lastSync)

			// Check if directory exists
			if _, err := os.Stat(dest); os.IsNotExist(err) {
				fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", subtleStyle.Render("⚠️  Not synced yet – run 'graftcxt sync'"))
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}

		// Summary
		fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("─", 50))
		fmt.Fprintf(cmd.OutOrStdout(), "Total: %d repos, %d files\n", len(pf.Context), totalFiles)
		return nil
	},
}

func init() { rootCmd.AddCommand(lsCmd) }
