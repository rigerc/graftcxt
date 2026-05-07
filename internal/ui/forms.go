package ui

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	ctx "github.com/rigerc/graftcxt/internal/context"
)

var ErrAborted = errors.New("aborted")

func InputRepoForm() (repo, dir string, err error) {
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Repository").
			Description("Enter the GitHub repository to add to your context.").
			Placeholder("e.g., owner/repo or owner/repo/subdir#main").
			Value(&repo).
			CharLimit(200).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("repository is required")
				}
				if !isValidRepoFormat(s) {
					return fmt.Errorf("invalid format (expected: owner/repo[/subdir][#ref])")
				}
				return nil
			}),
		huh.NewInput().
			Title("Directory (optional)").
			Description("Custom directory for this repo (relative to project dir). Leave empty for default.").
			Placeholder("e.g., docs/custom-context").
			Value(&dir).
			CharLimit(200).
			Validate(validateRelativeDir),
	)).WithTheme(graftcxtTheme())
	if err := form.Run(); err != nil {
		return "", "", mapAbort(err)
	}
	return repo, strings.TrimSpace(dir), nil
}

func validateRelativeDir(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "~") {
		return fmt.Errorf("directory must be relative to the project")
	}
	if strings.Contains("/"+strings.ReplaceAll(s, "\\", "/")+"/", "/../") {
		return fmt.Errorf("directory must not contain ..")
	}
	return nil
}

func isValidRepoFormat(s string) bool {
	parts := strings.Split(s, "/")
	if len(parts) < 2 {
		return false
	}
	if parts[0] == "" || parts[1] == "" {
		return false
	}
	return true
}

func SelectRepoForm(entries []ctx.ContextEntry) (string, error) {
	if len(entries) == 0 {
		return "", fmt.Errorf("no tracked repos")
	}
	var repo string
	opts := make([]huh.Option[string], 0, len(entries))
	for _, e := range entries {
		dirInfo := ""
		if e.Dir != "" {
			dirInfo = fmt.Sprintf(" → %s", e.Dir)
		} else {
			dirInfo = " → docs/context/"
		}
		opts = append(opts, huh.NewOption(fmt.Sprintf("%s (%s)%s", e.Repo, e.Name, dirInfo), e.Repo))
	}
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Remove repository").
			Description("Select a repository to remove from your context tracking.").
			Options(opts...).
			Value(&repo).
			Height(min(10, len(opts))),
	)).WithTheme(graftcxtTheme())
	if err := form.Run(); err != nil {
		return "", mapAbort(err)
	}
	return repo, nil
}

func graftcxtTheme() huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		t := huh.ThemeCharm(isDark)
		accent := lipgloss.Color("69")
		muted := lipgloss.Color("240")
		t.Focused.Title = t.Focused.Title.Foreground(accent).Bold(true)
		t.Focused.Base = t.Focused.Base.BorderForeground(accent)
		t.Focused.Description = t.Focused.Description.Foreground(muted)
		t.Blurred.Title = t.Blurred.Title.Foreground(muted)
		return t
	})
}

func mapAbort(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		return ErrAborted
	}
	return err
}
