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

func TestResolvedProjectPathCreatesMissingDefault(t *testing.T) {
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
	got, err := resolvedProjectPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, ".project.json")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	pf, err := ctx.Load(want)
	if err != nil {
		t.Fatal(err)
	}
	if len(pf.Context) != 0 {
		t.Fatalf("got %d context entries, want 0", len(pf.Context))
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
	projectFile := "/projects/test/.project.json"
	entry := ctx.ContextEntry{Name: "repo", Dir: ""}
	got := ctx.ContextEntryPath(projectFile, entry)
	want := "/projects/test/docs/context/repo"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
