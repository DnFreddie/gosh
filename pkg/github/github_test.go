package github

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// Test [NewRepo] that  creates new repos with two formats
// normal url and  Owner/Name scenarion
func Example_NewRepo() {
	// Example with a URL
	repo1, err := NewRepo("https://github.com/DnFreddie/gosh.git")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("URL:", repo1.URL)
		fmt.Println("Owner:", repo1.Owner)
		fmt.Println("Name:", repo1.Name)
		fmt.Println("FullName:", repo1.FullName)
	}

	// Example with an "owner/repo" format
	repo2, err := NewRepo("DnFreddie/gosh")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("URL:", repo2.URL)
		fmt.Println("Owner:", repo2.Owner)
		fmt.Println("Name:", repo2.Name)
		fmt.Println("FullName:", repo2.FullName)
	}

	// Example with an invalid format
	_, err = NewRepo("invalid-format")
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Output:
	// URL: https://github.com/DnFreddie/gosh.git
	// Owner: DnFreddie
	// Name: gosh
	// FullName: DnFreddie/gosh
	// URL: https://github.com/DnFreddie/gosh.git
	// Owner: DnFreddie
	// Name: gosh
	// FullName: DnFreddie/gosh
	// Error: invalid repository format, expected owner/repo, got: invalid-format
}

// Test [Repo.Clone]  thats  cloning repo to a specyfied dir
func TestRepo_Clone(t *testing.T) {

	tmpDir, err := os.MkdirTemp("", "repo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repo := &Repo{
		Owner:    "DnFreddie",
		Name:     "gosh",
		URL:      "https://github.com/DnFreddie/gosh.git",
		FullName: "DnFreddie/gosh",
	}

	// Test with empty dir (should use current directory)
	t.Run("Empty dir", func(t *testing.T) {
		err := repo.Clone("")
		if err == nil {
			t.Error("Expected error when cloning with empty dir, got none")
		}
	})

	// Test with specified dir
	t.Run("Specified dir", func(t *testing.T) {
		err := repo.Clone(tmpDir)
		if err == nil {
			expectedPath := filepath.Join(tmpDir, repo.FullName)
			if repo.Path != expectedPath {
				t.Errorf("Expected Path %s, got %s", expectedPath, repo.Path)
			}
		}
	})
}
func ExampleRepoManager_FilterByOwner() {
	rm := &RepoManager{
		Repos: []Repo{
			{Name: "repo1", Owner: "owner1"},
			{Name: "repo2", Owner: "owner2"},
			{Name: "repo3", Owner: "owner1"},
		},
	}
	filtered := rm.FilterByOwner("owner1")
	for _, repo := range filtered.Repos {
		fmt.Println(repo.Name)
	}
	// Output:
	// repo1
	// repo3
}
