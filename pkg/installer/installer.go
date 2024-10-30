package installer

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var TOOLBOX = []string{
	"cli/cli",
	"mikefarah/yq",
	"junegunn/fzf",
}

// Config holds configuration for the installer
type Config struct {
	TargetDir string // Directory where executables will be installed (default: ~/.local/bin)
	TempDir   string // Directory for temporary files (default: system temp)
}

// Installer manages installation of repositories
type Installer struct {
	config Config
	client *http.Client
	repos  []*Repo
}

type Repo struct {
	Owner   string
	Name    string
	Version string
	Links   DownloadLinks
}

// DownloadLinks holds URLs for downloading assets and checksums
type DownloadLinks struct {
	ArchiveUrl string
}

// GitHubRelease holds release information fetched from the GitHub API
type GitHubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []GitHubAsset `json:"assets"`
}

// GitHubAsset represents an individual asset in a release
type GitHubAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

func NewInstaller(config Config, repoUrls []string) (*Installer, error) {
	if config.TargetDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		config.TargetDir = filepath.Join(homeDir, ".local", "bin")
	}

	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}

	var repos []*Repo
	for _, repoUrl := range repoUrls {
		repo, err := NewRepo(repoUrl)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}

	return &Installer{
		config: config,
		client: &http.Client{CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		}},
		repos: repos,
	}, nil
}

func NewRepo(repoUrl string) (*Repo, error) {
	parts := strings.Split(repoUrl, ":")
	repoPath := parts[0]

	repoParts, err := validateRepoUrl(repoPath)
	if err != nil {
		return nil, err
	}

	return &Repo{
		Owner: repoParts[0],
		Name:  repoParts[1],
	}, nil
}

func (i *Installer) fetchReleases() error {
	for _, repo := range i.repos {
		if err := repo.fetchRelease(i.client); err != nil {
			return err
		}

	}
	return nil
}

func (r *Repo) fetchRelease(client *http.Client) error {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", r.Owner, r.Name)
	slog.Info("Trying to fetch", "url", apiURL)
	resp, err := client.Get(apiURL)
	if err != nil {
		return fmt.Errorf("error fetching release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: received non-200 response code %d\nfor %s", resp.StatusCode, apiURL)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	linuxRegex := regexp.MustCompile(`(?i)linux.*(amd64|x86_64).*\.tar\.gz\s*$`)

	for _, asset := range release.Assets {
		if linuxRegex.MatchString(asset.Name) {
			r.Links.ArchiveUrl = asset.DownloadURL
		}
	}

	if r.Links.ArchiveUrl == "" {
		return fmt.Errorf("could not find a matching Linux AMD64 asset in release %s", release.TagName)
	}

	r.Version = release.TagName
	return nil
}

func validateRepoUrl(repoUrl string) ([]string, error) {
	parts := strings.Split(repoUrl, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, errors.New("invalid repository URL format. Must be 'owner/repo'")
	}
	return parts, nil
}

func (i *Installer) Install() error {
	if err := i.fetchReleases(); err != nil {
		return err
	}

	tempDir, err := i.createTempDir()
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	for _, repo := range i.repos {
		if err := i.installRepo(repo, tempDir); err != nil {
			return fmt.Errorf("failed to install %s/%s: %w", repo.Owner, repo.Name, err)
		}
	}
	return nil
}
