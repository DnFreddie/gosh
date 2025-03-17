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
	tmux, err := NewTmux("", "")

	home := os.Getenv("HOME")
	if len(home) == 0 {
		return errors.New("Failed to find ssh config")
	}
	sshConfig := path.Join(home, ".ssh/config")
	f, err := os.Open(sshConfig)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()

	reader := io.Reader(f)
	hosts, err := GetHosts(reader)

	if err != nil {
		return err
	}
	choice, err := busybox.RunTerm(hosts)

	if err != nil {
		return err

	}
	tmux.session_name = fmt.Sprintf("ssh@%s", string(choice))
	tmux.abs_path = home
	if err := tmux.CreateSession(); err != nil {
		return err
	}
	command := fmt.Sprintf("ssh %s", string(choice))
	if _, err := tmux.Run("send-keys", "-t", tmux.session_name, command, "C-m"); err != nil {
		return err
	}
	return nil

}

type Host string

func (h Host) String() string {
	return string(h)
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
		hosts = append(hosts, Host(host))
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
	tmux, err := NewTmux("", "")
	home := os.Getenv("HOME")
	if len(home) == 0 {
		return errors.New("Failed to find HOMEDIR")
	}
	// so this add now the full path
	dirs, err := Find(home)
	var display []Path

	for _, v := range dirs {
		shorten := strings.Replace(v.String(), home, "", 1)
		fmt.Println(shorten)
		display = append(display, Path(shorten))
	}

	choice, err := busybox.RunTerm(display)
	if err != nil {
		return fmt.Errorf("Directory not found or not selected: %v\n", err)
	}

	abs_choice := path.Join(home, choice.String())
	tmux.abs_path = abs_choice
	tmux.SetName(path.Base(choice.String()))
	err = tmux.CreateSession()
	if err != nil {
		return err
	}
	return nil

}

func Vf() error {

	tmux, err := NewTmux("", "")

	home := os.Getenv("HOME")
	if len(home) == 0 {
		return errors.New("Failed to find HOMEDIR")
	}
	fmt.Println(home)
	dirs, err := Find(home)

	choice, err := busybox.RunTerm(dirs)
	if err != nil {
		return fmt.Errorf("Directory not found or not selected: %v\n", err)
	}

	tmux.SetName(string(choice))
	tmux.abs_path = string(choice)

	if _, err := tmux.Run("new-window", "-c", tmux.abs_path, "$EDITOR ."); err != nil {
		return err
	}
	return nil

}

// Find searches for directories within the specified directory up to a depth of 3.
// It returns a slice of maps where each map contains the basename and absolute path of a directory.
// The function limits concurrent processing to 5 goroutines.
type Path string

func (p Path) String() string {
	return string(p)
}

func Find(dir string) ([]Path, error) {
	toSkip := []string{".cache", ".local", "node_modules", ".git"}
	var absolutePaths []Path
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
			absolutePaths = append(absolutePaths, Path(absPath))
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

	repo, err := busybox.RunTerm(rm.Repos)
	if err != nil {
		return fmt.Errorf("failed to run terminal command: %w", err)
	}

	if err = repo.Clone(gitDir); !errors.Is(err, github.RepoExistErr{}) {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	t, err := NewTmux(repo.Name, repo.Path)
	if err != nil {
		return fmt.Errorf("failed to create new Tmux session: %w", err)
	}

	if err := t.CreateSession(); err != nil {
		return fmt.Errorf("failed to create Tmux session: %w", err)
	}

	return nil
}
