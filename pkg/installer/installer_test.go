package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// Define the regex patterns as in the main function
var checksumRegex = regexp.MustCompile(`(?i)(checksum|checksums)(\.txt)?$|\.sha256$|\.sha512$`)
var linuxRegex = regexp.MustCompile(`(?i)linux.*(amd64|x86_64|arm|aarch64|musl|gnu)?.*\.tar\.gz\s*$`)

func TestLinuxRegex(t *testing.T) {
	// Test cases for Linux regex
	testCases := []struct {
		name        string
		shouldMatch bool
	}{
		{"fzf-linux_amd64.tar.gz", true},
		{"fzf-linux_arm64.tar.gz", true},
		{"fzf-linux-x86_64.tar.gz", true},
		{"fzf-linux-gnu.tar.gz", true},
		{"fzf-x86_64-linux-musl.tar.gz", true},
		{"fzf-windows.tar.gz", false},
		{"fzf-linux.tar.bz2", false},                           // Different file extension
		{"bat-v0.24.0-x86_64-unknown-linux-gnu.tar.gz ", true}, // Different file extension
	}

	for _, tc := range testCases {
		if linuxRegex.MatchString(tc.name) != tc.shouldMatch {
			t.Errorf("Expected match for '%s': %v, but got %v", tc.name, tc.shouldMatch, !tc.shouldMatch)
		}
	}
}

func TestChecksumRegex(t *testing.T) {
	// Test cases for checksum regex
	testCases := []struct {
		name        string
		shouldMatch bool
	}{
		{"fzf_checksums.txt", true},
		{"fzf.sha256", true},
		{"fzf.sha512", true},
		{"fzf-checksum.txt", true},
		{"fzf-checksums.txt", true},
		{"fzf-windows.sha256", true},
		{"fzf.tar.gz", false},
	}

	for _, tc := range testCases {
		if checksumRegex.MatchString(tc.name) != tc.shouldMatch {
			t.Errorf("Expected match for '%s': %v, but got %v", tc.name, tc.shouldMatch, !tc.shouldMatch)
		}
	}

}

// [NewInstaller] demonstrates creating a new installer instance with configuration
func Example_newInstaller() {
	// Create test configuration
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	config := Config{
		TargetDir: filepath.Join(homeDir, ".local", "bin"),
		TempDir:   "/tmp",
	}

	// Initialize installer with test repository
	repos := []string{"cli/cli"}
	installer, err := NewInstaller(config, repos)
	if err != nil {
		fmt.Printf("Error creating installer: %v\n", err)
		return
	}

	fmt.Printf("Target directory: %s\n", filepath.Base(installer.config.TargetDir))
	fmt.Printf("Number of repos: %d\n", len(installer.repos))

	// Output:
	// Target directory: bin
	// Number of repos: 1
}

// [NewRepo] shows how to create and validate a new repository instance
func Example_newRepo() {
	// Test valid repository URL
	validRepo := "cli/cli"
	repo, err := NewRepo(validRepo)
	if err != nil {
		fmt.Printf("Error with valid repo: %v\n", err)
	} else {
		fmt.Printf("Valid repo - Owner: %s, Name: %s\n", repo.Owner, repo.Name)
	}

	invalidRepo := "invalid://repo"
	_, err = NewRepo(invalidRepo)
	if err != nil {
		fmt.Printf("Invalid repo error: %v\n", err)
	}

	// Output:
	// Valid repo - Owner: cli, Name: cli
	// Invalid repo error: invalid repository URL format. Must be 'owner/repo'
}

// [ValidateRepoUrl] demonstrates repository URL validation logic
func Example_validateRepoUrl() {
	// Test cases for URL validation
	urls := []string{
		"owner/repo",
		"owner/repo/extra",
		"",
		"owner",
	}

	for _, url := range urls {
		parts, err := validateRepoUrl(url)
		if err != nil {
			fmt.Printf("URL '%s' invalid: %v\n", url, err)
		} else {
			fmt.Printf("URL '%s' valid: owner='%s' repo='%s'\n", url, parts[0], parts[1])
		}
	}

	// Output:
	// URL 'owner/repo' valid: owner='owner' repo='repo'
	// URL 'owner/repo/extra' invalid: invalid repository URL format. Must be 'owner/repo'
	// URL '' invalid: invalid repository URL format. Must be 'owner/repo'
	// URL 'owner' invalid: invalid repository URL format. Must be 'owner/repo'
}

// [CreateTempDir] shows temporary directory creation and handling
func Example_createTempDir() {
	installer := &Installer{
		config: Config{
			TempDir: "/tmp",
		},
	}

	// Demonstrate directory pattern
	pattern := filepath.Join(installer.config.TempDir, "install_*")
	fmt.Printf("Directory pattern: %s\n", pattern)
	fmt.Printf("Base temp dir: %s\n", filepath.Base(installer.config.TempDir))

	// Output:
	// Directory pattern: /tmp/install_*
	// Base temp dir: tmp
}
