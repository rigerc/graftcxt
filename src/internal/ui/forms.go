package ui

import (
	"errors"
	"fmt"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	ctx "github.com/rigerc/graftcxt/internal/context"
)

var ErrAborted = errors.New("aborted")

func InputRepoForm() (string, error) {
	var repo string
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Repository").Description("owner/repo[/subdir][#ref]").Value(&repo),
	)).WithTheme(graftcxtTheme())
	if err := form.Run(); err != nil {
		return "", mapAbort(err)
	}
	if repo == "" {
		return "", ErrAborted
	}
	return repo, nil
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
			dirInfo = fmt.Sprintf(" [dir: %s]", e.Dir)
		} else {
			dirInfo = " [default: docs/context/]"
		}
		opts = append(opts, huh.NewOption(fmt.Sprintf("%s (%s)%s", e.Repo, e.Name, dirInfo), e.Repo))
	}
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Remove repository").Options(opts...).Value(&repo),
	)).WithTheme(graftcxtTheme())
	if err := form.Run(); err != nil {
		return "", mapAbort(err)
	}
	if repo == "" {
		return "", ErrAborted
	}
	return repo, nil
}

func graftcxtTheme() huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		t := huh.ThemeCharm(isDark)
		accent := lipgloss.Color("69")
		t.Focused.Title = t.Focused.Title.Foreground(accent).Bold(true)
		t.Focused.Base = t.Focused.Base.BorderForeground(accent)
		return t
	})
}

func mapAbort(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		return ErrAborted
	}
	return err
}
