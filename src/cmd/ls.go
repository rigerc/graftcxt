package cmd

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	ctx "github.com/rigerc/graftcxt/internal/context"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tracked context repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectFile, err := resolvedProjectPath()
		if err != nil {
			return err
		}
		pf, err := ctx.Load(projectFile)
		if err != nil {
			return err
		}
		header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
		rows := []string{fmt.Sprintf("%-42s  %-20s  %s  %s", header.Render("REPO"), header.Render("NAME"), header.Render("LAST_SYNC"), header.Render("DIR"))}
		for _, e := range pf.Context {
			last := ""
			if e.LastSync != nil {
				last = *e.LastSync
			}
			dir := "(default)"
			if e.Dir != "" {
				dir = e.Dir
			}
			rows = append(rows, fmt.Sprintf("%-42s  %-20s  %s  %s", e.Repo, e.Name, last, dir))
		}
		fmt.Fprintln(cmd.OutOrStdout(), strings.Join(rows, "\n"))
		return nil
	},
}

func init() { rootCmd.AddCommand(lsCmd) }
