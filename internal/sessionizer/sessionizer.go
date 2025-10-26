package sessionizer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/DnFreddie/gosh/pkg/busybox"
	"github.com/DnFreddie/gosh/pkg/github"
)

func Fs() error {
	tmux, err := NewTmux()
	if err != nil {
		return err
	}

	home := os.Getenv("HOME")
	if len(home) == 0 {
		return errors.New("Failed to find ssh config")
	}
	sshConfig := path.Join(home, ".ssh/config")
	f, err := os.Open(sshConfig)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := io.Reader(f)
	hosts, err := GetHosts(reader)
	if err != nil {
		return err
	}

	choice, err := busybox.RunTerm(hosts, func(h Host) string {
		return fmt.Sprintf("%s", h.Name)
	})
	if err != nil {
		return err
	}

	sessionName := fmt.Sprintf("ssh@%s", choice.Name)
	if err := tmux.CreateSession(sessionName, home); err != nil {
		return err
	}

	command := fmt.Sprintf("ssh %s", choice.Name)
	if _, err := tmux.Run("send-keys", "-t", sessionName, command, "C-m"); err != nil {
		return err
	}
	return nil
}

type Host struct {
	Name string
}

func GetHosts(r io.Reader) ([]Host, error) {
	var hosts []Host
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Bytes()
		host := findHost(line)
		if len(host) == 0 {
			continue
		}
		hosts = append(hosts, Host{Name: host})
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if len(hosts) == 0 {
		return hosts, errors.New("No hosts found")
	}
	return hosts, nil
}

func findHost(line []byte) string {
	regx := `(?i)^Host\s+(\w+)`
	re := regexp.MustCompile(regx)
	matches := re.FindSubmatch(line)
	if len(matches) < 2 {
		return ""
	}
	return string(matches[1])
}

func Fd() error {
	tmux, err := NewTmux()
	if err != nil {
		return err
	}

	home := os.Getenv("HOME")
	if len(home) == 0 {
		return errors.New("Failed to find HOMEDIR")
	}

	dirs, err := Find(home)
	if err != nil {
		return err
	}

	var display []string
	for _, v := range dirs {
		shorten_path := strings.Replace(v, home, "", 1)
		display = append(display, shorten_path)
	}

	choice, err := busybox.RunTerm(display, func(s string) string {
		return s
	})
	if err != nil {
		return fmt.Errorf("Directory not found or not selected: %v\n", err)
	}

	absChoice := path.Join(home, choice)
	sessionName := path.Base(choice)

	err = tmux.CreateSession(sessionName, absChoice)
	if err != nil {
		return err
	}
	return nil
}

func Vf() error {
	tmux, err := NewTmux()
	if err != nil {
		return err
	}

	home := os.Getenv("HOME")
	if len(home) == 0 {
		return errors.New("Failed to find HOMEDIR")
	}

	dirs, err := Find(home)
	if err != nil {
		return err
	}

	choice, err := busybox.RunTerm(dirs, func(s string) string { return s })
	if err != nil {
		return fmt.Errorf("Directory not found or not selected: %v\n", err)
	}

	if _, err := tmux.Run("new-window", "-c", choice, "$EDITOR ."); err != nil {
		return err
	}
	return nil
}

func Find(dir string) ([]string, error) {
	toSkip := []string{".cache", ".local", "node_modules", ".git"}
	var absolutePaths []string
	var errorsArr []error
	var errorsMutex sync.Mutex
	var pathsMutex sync.Mutex
	buffered := make(chan struct{}, 5)
	var wg sync.WaitGroup

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			errorsMutex.Lock()
			errorsArr = append(errorsArr, fmt.Errorf("error accessing path %s: %w", path, err))
			errorsMutex.Unlock()
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		if slices.Contains(toSkip, filepath.Base(path)) {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			errorsMutex.Lock()
			errorsArr = append(errorsArr, fmt.Errorf("error getting relative path for %s: %w", path, err))
			errorsMutex.Unlock()
			return nil
		}

		depth := 1
		if relPath != "." {
			depth = strings.Count(relPath, string(os.PathSeparator)) + 1
		}
		if depth > 3 {
			return filepath.SkipDir
		}

		buffered <- struct{}{}
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			defer func() { <-buffered }()

			absPath, err := filepath.Abs(p)
			if err != nil {
				errorsMutex.Lock()
				errorsArr = append(errorsArr, fmt.Errorf("error getting absolute path for %s: %w", p, err))
				errorsMutex.Unlock()
				return
			}

			pathsMutex.Lock()
			absolutePaths = append(absolutePaths, absPath)
			pathsMutex.Unlock()
		}(path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the path %s: %w", dir, err)
	}

	wg.Wait()

	if len(errorsArr) > 0 {
		return nil, errors.Join(errorsArr...)
	}

	return absolutePaths, nil
}

func Fg(gitDir string) error {
	rm, err := github.NewRepoManager(github.USER_REPOS, &http.Client{})
	if err != nil {
		return fmt.Errorf("failed to initialize RepoManager: %w", err)
	}

	repo, err := busybox.RunTerm(rm.Repos, func(r github.Repo) string { return r.Name })
	if err != nil {
		return fmt.Errorf("failed to run terminal command: %w", err)
	}

	if err = repo.Clone(gitDir); err != nil && !errors.Is(err, github.RepoExistErr{}) {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	t, err := NewTmux()
	if err != nil {
		return fmt.Errorf("failed to create new Tmux: %w", err)
	}

	if err := t.CreateSession(repo.Name, repo.Path); err != nil {
		return fmt.Errorf("failed to create Tmux session: %w", err)
	}

	return nil
}

func (t *Tmux) Tn() error {

	choice, err := busybox.RunTerm(t.Sessions, func(s TmuxSession) string { return s.Name })
	if err != nil {
		return err
	}

	err = t.SwitchSession(choice.Name)
	if err != nil {
		return err
	}

	return nil
}
