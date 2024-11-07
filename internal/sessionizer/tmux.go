package sessionizer

import (
	"bytes"
	"errors"
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
func (t *Tmux) SetName(string) {
	if strings.HasPrefix(t.session_name, ".") {
		t.session_name = t.session_name[1:]
	}

}

func (t *Tmux) Run(args ...string) (string, error) {
	var stderr bytes.Buffer
	cmd := exec.Command("tmux", args...)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}
	return "", nil
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
	hasSession, err := t.HasSession()
	if err != nil {
		return err
	}
	if !hasSession {
		_, err = t.Run("new-session", "-d", "-s", t.session_name, "-c", t.abs_path)
		if err != nil {
			return err
		}
		_, err = t.Run("attach-session", "-t", t.session_name)
		if err != nil {
			return err
		}
	}
	return nil

}
