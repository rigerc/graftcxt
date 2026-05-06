package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	ctx "github.com/rigerc/graftcxt/internal/context"
	"github.com/spf13/cobra"
)

var projectPath string

var rootCmd = &cobra.Command{
	Use:          "graftcxt",
	Short:        "Graft external GitHub repos into your project context",
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&projectPath, "project", ".project.json", "path to project JSON file")
}

func resolvedProjectPath() (string, error) {
	if projectPath != ".project.json" {
		if _, err := os.Stat(projectPath); err != nil {
			return "", fmt.Errorf("project file %q not found: %w", projectPath, err)
		}
		return projectPath, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, ".project.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find .project.json in current directory or any parent; pass --project /path/to/.project.json")
		}
		dir = parent
	}
}

func docsDir(projectFile string) string { return filepath.Join(filepath.Dir(projectFile), "docs") }

func contextEntryPath(projectFile string, entry ctx.ContextEntry) string {
	return ctx.ContextEntryPath(projectFile, entry)
}
