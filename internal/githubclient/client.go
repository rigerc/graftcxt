package githubclient

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
)

type Service struct {
	token  string
	client *github.Client
}

func NewService(token string) *Service {
	return &Service{
		token:  token,
		client: github.NewClient(nil).WithAuthToken(token),
	}
}

func (s *Service) Token() string { return s.token }

func (s *Service) Client() *github.Client { return s.client }

// CreateRepo creates a repository under the authenticated user and returns the clone URL.
func (s *Service) CreateRepo(ctx context.Context, name, description string, private bool) (string, error) {
	if s.token == "" {
		return "", fmt.Errorf("GITHUB_TOKEN is not set")
	}
	repo, _, err := s.client.Repositories.Create(ctx, "", &github.Repository{
		Name:        github.Ptr(name),
		Description: github.Ptr(description),
		Private:     github.Ptr(private),
	})
	if err != nil {
		return "", fmt.Errorf("create github repo: %w", err)
	}
	return repo.GetCloneURL(), nil
}
