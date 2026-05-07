package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	ctx "github.com/rigerc/graftcxt/internal/context"
	"github.com/rigerc/graftcxt/internal/output"
	"github.com/spf13/cobra"
)

var (
	projectPath string
	silentMode  bool
)

var rootCmd = &cobra.Command{
	Use:          "graftcxt",
	Short:        "Graft external GitHub repos into your project context",
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		output.SetSilent(silentMode)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&projectPath, "project", ".project.json", "path to project JSON file")
	rootCmd.PersistentFlags().BoolVar(&silentMode, "silent", false, "suppress non-essential output")
}

func resolvedProjectPath() (string, error) {
	if projectPath != ".project.json" {
		if _, err := os.Stat(projectPath); err == nil {
			return projectPath, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("project file %q not found: %w", projectPath, err)
		}
		return createProjectFile(projectPath)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		candidate := filepath.Join(dir, ".project.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return createProjectFile(filepath.Join(cwd, ".project.json"))
		}
		dir = parent
	}
}

func createProjectFile(path string) (string, error) {
	if err := ctx.Save(path, &ctx.ProjectFile{}); err != nil {
		return "", err
	}
	return path, nil
}

func contextEntryPath(projectFile string, entry ctx.ContextEntry) string {
	return ctx.ContextEntryPath(projectFile, entry)
}
