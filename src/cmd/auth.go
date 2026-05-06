package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var authStatusCmd = &cobra.Command{
	Use:   "auth-status",
	Short: "Check GitHub authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check GITHUB_TOKEN env var
		if os.Getenv("GITHUB_TOKEN") != "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Authenticated via GITHUB_TOKEN environment variable")
			return nil
		}
		// Check gh CLI
		token, err := exec.Command("gh", "auth", "token").Output()
		if err == nil && strings.TrimSpace(string(token)) != "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Authenticated via gh CLI")
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Not authenticated. Set GITHUB_TOKEN or run 'gh auth login'.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authStatusCmd)
}
