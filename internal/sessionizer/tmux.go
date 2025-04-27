package sessionizer

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var ErrNotInTmux = errors.New("not in a tmux session")

type Tmux struct {
	session_name string
	abs_path     string
}

func NewTmux(sessionName, dir string) (*Tmux, error) {
	return &Tmux{
		session_name: sessionName,
		abs_path:     dir,
	}, nil

}

func (t *Tmux) SetName(s string) {
	if strings.HasPrefix(s, ".") {
		t.session_name = s[1:]

	} else {
		t.session_name = s

	}

}
func (t *Tmux) GetName() string {
	return t.session_name
}

func (t *Tmux) Run(args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	var cmd *exec.Cmd

	if os.Getenv("TMUX") == "" {
		cmdArgs := []string{"-c", "tmux " + strings.Join(args, " ")}
		cmd = exec.Command("bash", cmdArgs...)
	} else {
		cmd = exec.Command("tmux", args...)
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		fmt.Println(stderr.String())
		return stderr.String(), err
	}

	return stdout.String(), nil
}

func (t *Tmux) HasSession() (bool, error) {
	_, err := t.Run("has-session", "-t", t.session_name)
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

const (
	ColorIndex       = "\033[38;2;181;126;220m"
	ColorSessionName = "\033[1;38;2;150;150;170m"
	ColorDescription = "\033[38;2;120;120;120m"
	ColorReset       = "\033[0m"
)

func (t *Tmux) ListSessions() ([]string, error) {
	stdout, err := t.Run("list-sessions", "-F#{session_name}")
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(stdout)
	sessions := strings.Split(trimmed, "\n")
	filtered := make([]string, 0, len(sessions))

	for _, s := range sessions {
		if s != "" {
			filtered = append(filtered, s)
		}
	}

	if len(sessions) == 0 {
		return nil, errors.New("No sessions found ")
	}

	for i, session := range filtered {
		fmt.Printf("%s%2d%s: %s%s%s\n",
			ColorIndex, i, ColorReset,
			ColorSessionName, session, ColorReset)
	}

	return filtered, nil
}

func (t *Tmux) SwitchSession() error {
	if os.Getenv("TMUX") == "" {
		cmd := exec.Command("tmux", "attach-session", "-t", t.session_name)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Failed to attach to the session %v: %w", t.session_name, err)
		}
	} else {
		_, err := t.Run("switch-client", "-t", t.session_name)
		if err != nil {
			return fmt.Errorf("Failed to switch to the session %v: %w", t.session_name, err)
		}
	}
	return nil
}

func (t *Tmux) CreateSession() error {
	exists, err := t.HasSession()
	if err != nil {
		return errors.New("Failed to check if session exists")
	}

	if !exists {
		_, err = t.Run("new-session", "-d", "-s", t.session_name, "-c", t.abs_path)
		if err != nil {
			return fmt.Errorf("Failed to create the session name:%v, path:%v", t.session_name, t.abs_path)
		}
	}

	return t.SwitchSession()
}
