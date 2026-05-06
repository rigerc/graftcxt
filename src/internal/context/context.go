package context

import (
	stdctx "context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

type ContextEntry struct {
	Repo     string  `json:"repo"`
	Name     string  `json:"name"`
	LastSync *string `json:"last_sync"`
	Dir      string  `json:"dir,omitempty"`
}

type ProjectFile struct {
	Project json.RawMessage `json:"project,omitempty"`
	Skills  json.RawMessage `json:"skills,omitempty"`
	Context []ContextEntry  `json:"context"`
}

func Load(path string) (*ProjectFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pf ProjectFile
	if err := json.Unmarshal(b, &pf); err != nil {
		return nil, err
	}
	if pf.Context == nil {
		pf.Context = []ContextEntry{}
	}
	return &pf, nil
}

func Save(path string, pf *ProjectFile) error {
	if pf.Context == nil {
		pf.Context = []ContextEntry{}
	}
	b, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".project-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func ParseRepoRef(repoID string) (owner, repo, subdir, ref string, err error) {
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		err = errors.New("repo is required")
		return
	}
	ref = "HEAD"
	if before, after, ok := strings.Cut(repoID, "#"); ok {
		repoID = before
		if strings.TrimSpace(after) != "" {
			ref = strings.TrimSpace(after)
		}
	}
	parts := strings.Split(strings.Trim(repoID, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		err = fmt.Errorf("repo must look like owner/repo[/subdir][#ref]")
		return
	}
	owner, repo = parts[0], parts[1]
	if len(parts) > 2 {
		if (parts[2] == "tree" || parts[2] == "blob") && len(parts) >= 4 {
			ref = parts[3]
			if len(parts) > 4 {
				subdir = strings.Join(parts[4:], "/")
			}
		} else {
			subdir = strings.Join(parts[2:], "/")
		}
	}
	return
}

func ParseRepoName(repoID string) (string, error) {
	_, repo, subdir, _, err := ParseRepoRef(repoID)
	if err != nil {
		return "", err
	}
	name := repo
	if subdir != "" {
		name = filepath.Base(subdir)
	}
	if name == "." || name == string(filepath.Separator) || name == "" {
		return "", fmt.Errorf("could not derive repo name from %q", repoID)
	}
	return name, nil
}

func HasEntry(pf *ProjectFile, repoID string) bool {
	for _, e := range pf.Context {
		if e.Repo == repoID {
			return true
		}
	}
	return false
}

func ContextEntryPath(projectFile string, entry ContextEntry) string {
	baseDir := filepath.Dir(projectFile)
	if entry.Dir != "" {
		return filepath.Join(baseDir, entry.Dir)
	}
	return filepath.Join(baseDir, "docs", "context", entry.Name)
}

func AddEntry(pf *ProjectFile, e ContextEntry) { pf.Context = append(pf.Context, e) }

func RemoveEntry(pf *ProjectFile, repoID string) {
	out := pf.Context[:0]
	for _, e := range pf.Context {
		if e.Repo != repoID {
			out = append(out, e)
		}
	}
	pf.Context = out
}

func NewGitHubClient() *github.Client {
	var hc *http.Client
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		hc = oauth2.NewClient(stdctx.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
	}
	return github.NewClient(hc)
}

func SyncRepo(repoID, destPath string, gh *github.Client) error {
	owner, repo, subdir, ref, err := ParseRepoRef(repoID)
	if err != nil {
		return err
	}
	if gh == nil {
		gh = NewGitHubClient()
	}
	tree, _, err := gh.Git.GetTree(stdctx.Background(), owner, repo, ref, true)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	prefix := strings.Trim(subdir, "/")
	if prefix != "" {
		prefix += "/"
	}
	for _, entry := range tree.Entries {
		if entry.GetType() != "blob" {
			continue
		}
		path := entry.GetPath()
		if prefix != "" {
			if !strings.HasPrefix(path, prefix) {
				continue
			}
			path = strings.TrimPrefix(path, prefix)
		}
		if path == "" {
			continue
		}
		clean := filepath.Clean(path)
		if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
			return fmt.Errorf("unsafe path from GitHub tree: %s", path)
		}
		blob, _, err := gh.Git.GetBlob(stdctx.Background(), owner, repo, entry.GetSHA())
		if err != nil {
			return err
		}
		content, err := decodeBlob(blob)
		if err != nil {
			return fmt.Errorf("decode %s: %w", path, err)
		}
		out := filepath.Join(destPath, clean)
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(out, content, 0o644); err != nil {
			return err
		}
		seen[clean] = true
	}
	return cleanDest(destPath, seen)
}

func decodeBlob(blob *github.Blob) ([]byte, error) {
	if blob == nil {
		return nil, errors.New("nil blob")
	}
	content := blob.GetContent()
	if blob.GetEncoding() == "base64" {
		return base64.StdEncoding.DecodeString(strings.ReplaceAll(content, "\n", ""))
	}
	return []byte(content), nil
}

func cleanDest(destPath string, seen map[string]bool) error {
	if err := os.MkdirAll(destPath, 0o755); err != nil {
		return err
	}
	var dirs []string
	err := filepath.WalkDir(destPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == destPath {
			return nil
		}
		rel, err := filepath.Rel(destPath, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, path)
			return nil
		}
		if !seen[filepath.Clean(rel)] {
			return os.Remove(path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	for _, dir := range dirs {
		_ = os.Remove(dir)
	}
	return nil
}

func NowString() string { return time.Now().UTC().Format(time.RFC3339) }
