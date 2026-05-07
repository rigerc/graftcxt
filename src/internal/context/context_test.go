package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSavePreservesSyncMetadata(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".project.json")
	input := `{
  "project": {"name":"x"},
  "context": [
    {"repo":"a/b","name":"b","last_sync":"2026-01-01T00:00:00Z","last_tree_sha":"abc"}
  ]
}
`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}

	pf, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if pf.Context[0].LastSync == nil || *pf.Context[0].LastSync != "2026-01-01T00:00:00Z" {
		t.Fatalf("LastSync not loaded: %#v", pf.Context[0].LastSync)
	}
	if pf.Context[0].LastTreeSHA == nil || *pf.Context[0].LastTreeSHA != "abc" {
		t.Fatalf("LastTreeSHA not loaded: %#v", pf.Context[0].LastTreeSHA)
	}

	AddEntry(pf, ContextEntry{Repo: "c/d", Name: "d"})
	if err := Save(path, pf); err != nil {
		t.Fatal(err)
	}
	gotBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(gotBytes)
	for _, want := range []string{`"last_sync":"2026-01-01T00:00:00Z"`, `"last_tree_sha":"abc"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("saved project missing %s:\n%s", want, got)
		}
	}
}
