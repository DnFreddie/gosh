package sessionizer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/DnFreddie/gosh/pkg/busybox"
)

func Fd() error {
	tmux, err := NewTmux("", "")
	home := os.Getenv("HOME")
	if len(home) == 0 {
		return errors.New("Failed to find HOMEDIR")
	}
	dirs, err := Find(home)
	choice, err := busybox.RunTerm(dirs)

	if err != nil {
		return fmt.Errorf("Directory not found or not selected: %v\n", err)
	}
	for k, v := range choice {
		tmux.SetName(k)
		tmux.abs_path = v
	}
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

	for k, v := range choice {
		tmux.SetName(k)
		tmux.abs_path = v
	}

	if _, err := tmux.Run("new-window", "-c", tmux.abs_path, "$EDITOR ."); err != nil {
		return err
	}
	return nil

}

// Find searches for directories within the specified directory up to a depth of 3.
// It returns a slice of maps where each map contains the basename and absolute path of a directory.
// The function limits concurrent processing to 5 goroutines.
func Find(dir string) ([]map[string]string, error) {

	var dirMaps []map[string]string
	var dirMapsMutex sync.Mutex

	var errorsArr []error
	var errorsMutex sync.Mutex

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

		if depth > 2 {
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

			basename := filepath.Base(p)
			dirMap := map[string]string{basename: absPath}

			dirMapsMutex.Lock()
			dirMaps = append(dirMaps, dirMap)
			dirMapsMutex.Unlock()
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

	return dirMaps, nil
}
