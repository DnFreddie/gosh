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
	if os.Getenv("TMUX") == "" {
		return nil, ErrNotInTmux
	}

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
	cmd := exec.Command("tmux", args...)
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

func (t *Tmux) CreateSession() error {
	exists, err := t.HasSession()
	if err != nil {
		return errors.New("Failed to check if session exists")
	}
	if !exists {
		_, err = t.Run("new-session", "-d", "-s", t.session_name, "-c", t.abs_path)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to create the session name:%v, path:%v\n", t.session_name, t.abs_path))
		}
	}
	if _, err = t.Run("switch-client", "-t", t.session_name); err != nil {
		return errors.New(fmt.Sprintf("Failed to attach to the session %v\n", t.session_name))
	}
	return nil
}
