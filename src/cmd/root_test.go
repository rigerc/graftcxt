package cmd

import (
	"os"
	"path/filepath"
	"testing"

	ctx "github.com/rigerc/graftcxt/internal/context"
)

func TestResolvedProjectPathSearchesParents(t *testing.T) {
	oldProjectPath := projectPath
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		projectPath = oldProjectPath
		_ = os.Chdir(oldWd)
	})

	dir := t.TempDir()
	project := filepath.Join(dir, ".project.json")
	if err := os.WriteFile(project, []byte(`{"project":{"name":"x"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(dir, "src", "cmd")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(nested); err != nil {
		t.Fatal(err)
	}

	projectPath = ".project.json"
	got, err := resolvedProjectPath()
	if err != nil {
		t.Fatal(err)
	}
	if got != project {
		t.Fatalf("got %q, want %q", got, project)
	}
}

func TestResolvedProjectPathReportsHelpfulMissingDefault(t *testing.T) {
	oldProjectPath := projectPath
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		projectPath = oldProjectPath
		_ = os.Chdir(oldWd)
	})

	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	projectPath = ".project.json"
	_, err = resolvedProjectPath()
	if err == nil {
		t.Fatal("expected missing project error")
	}
	if got := err.Error(); got == "open .project.json: no such file or directory" {
		t.Fatalf("unhelpful error: %s", got)
	}
}

func TestContextEntryPathDefault(t *testing.T) {
	projectFile := "/path/to/.project.json"
	entry := ctx.ContextEntry{Name: "myrepo"}
	got := ctx.ContextEntryPath(projectFile, entry)
	want := "/path/to/docs/context/myrepo"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestContextEntryPathCustom(t *testing.T) {
	projectFile := "/path/to/.project.json"
	entry := ctx.ContextEntry{Name: "myrepo", Dir: "custom/path"}
	got := ctx.ContextEntryPath(projectFile, entry)
	want := "/path/to/custom/path"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestContextEntryPathCustomNested(t *testing.T) {
	projectFile := "/home/user/project/.project.json"
	entry := ctx.ContextEntry{Name: "repo", Dir: "docs/legacy"}
	got := ctx.ContextEntryPath(projectFile, entry)
	want := "/home/user/project/docs/legacy"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestContextEntryPathEmptyDirUsesDefault(t *testing.T) {
	projectFile := "C:\\projects\\test\\.project.json"
	entry := ctx.ContextEntry{Name: "repo", Dir: ""}
	got := ctx.ContextEntryPath(projectFile, entry)
	want := "C:\\projects\\test\\docs\\context\\repo"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
