package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/DnFreddie/gosh/internal/utils"
)

const (
	USER_REPOS = "https://api.github.com/users/DnFreddie/repos"
)

type Repo struct {
	Owner    string `json:"login"`
	URL      string `json:"clone_url"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Version  string `json:"version"`
	Path     string
}

func NewRepo(repoPathOrURL string) (*Repo, error) {
	// Handle HTTP URL case
	if strings.HasPrefix(repoPathOrURL, "http") {
		_url := strings.TrimSuffix(repoPathOrURL, ".git")
		parsedURL, err := url.ParseRequestURI(_url)
		if err != nil {
			return nil, fmt.Errorf("invalid git URL: %w", err)
		}

		urlParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(urlParts) < 2 {
			return nil, errors.New("can't parse it, invalid URL")
		}

		owner := urlParts[len(urlParts)-2]
		name := urlParts[len(urlParts)-1]

		return &Repo{
			URL:      repoPathOrURL,
			Owner:    owner,
			Name:     name,
			FullName: owner + "/" + name,
		}, nil
	}

	// Handle "owner/repo" format (e.g., "DnFreddie/gosh")
	parts := strings.Split(repoPathOrURL, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected owner/repo, got: %s", repoPathOrURL)
	}

	return &Repo{
		URL:      fmt.Sprintf("https://github.com/%s/%s.git", parts[0], parts[1]),
		FullName: repoPathOrURL,
		Owner:    parts[0],
		Name:     parts[1],
	}, nil
}

type RepoManager struct {
	Repos []Repo
}

func NewRepoManager(url string, client *http.Client) (*RepoManager, error) {
	return fetchRepos(url, client)
}

func fetchRepos(url string, client *http.Client) (*RepoManager, error) {
	var repos []Repo
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching %v: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received non-200 response code %d for %s", resp.StatusCode, url)
	}

	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &RepoManager{Repos: repos}, nil
}

type RepoExistErr struct {
	name string
	path string
}

func (e RepoExistErr) Error() string {

	return fmt.Sprintf("The repo %s already exist in %s", e.name, e.path)
}

func (e RepoExistErr) Is(target error) bool {
	_, ok := target.(RepoExistErr)
	return ok
}

func (r *Repo) Clone(dir string) error {

	if dir == "" {
		return errors.New("No directory provided Cloning failed")
	}

	location := filepath.Join(dir, r.FullName)
	r.Path = location

	if _, err := os.Stat(location); err == nil {

		return RepoExistErr{
			path: location,
			name: r.FullName,
		}
	}
	slog.Debug("Clone details",
		"dir", dir,
		"owner", r.Owner,
		"name", r.Name,
		"final_location", r.Path)

	done := make(chan bool)
	defer func() { done <- true }()

	go utils.WaitingScreen(done, "Cloning")

	cmd := exec.Command("git", "clone", r.URL, location)

	slog.Info("Cloning", "path", location, "name", r.FullName)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}
	fmt.Printf("Successfully cloned %s", r.FullName)
	return nil
}

func (rm *RepoManager) PrintNames() {
	for _, repo := range rm.Repos {
		fmt.Println(repo.Name)
	}
}

func (rm *RepoManager) FilterByOwner(owner string) *RepoManager {
	var filtered []Repo
	for _, repo := range rm.Repos {
		if repo.Owner == owner {
			filtered = append(filtered, repo)
		}
	}
	return &RepoManager{Repos: filtered}
}
