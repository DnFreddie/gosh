package sessionizer

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var ErrNotInTmux = errors.New("not in a tmux session")

type Tmux struct {
	Current      *TmuxSession
	CurrentIndex int
	Sessions     []TmuxSession
}

type TmuxSession struct {
	Name     string
	Created  time.Time
	Group    string
	Attached bool
	AbsPath  string
	Windows  []TmuxWindow
	// "#{session_name}: #{session_windows} windows "	\
	// "(created #{t:session_created})"		\
	// "#{?session_grouped, (group ,}"			\
	// "#{session_group}#{?session_grouped,),}"	\
	// "#{?session_attached, (attached),}"
}

type TmuxWindow struct {
	Index       int
	Name        string
	SessionName string
}

// Get windows for specific session(s)
// If no sessions provided or "*" in list, returns windows for all sessions
func (t *Tmux) GetWindows(sessionNames ...string) ([]TmuxWindow, error) {
	format := "#{session_name}|#{window_index}|#{window_name}"
	var args []string

	getAllSessions := len(sessionNames) == 0
	for _, name := range sessionNames {
		if name == "*" {
			getAllSessions = true
			break
		}
	}

	if getAllSessions {
		args = []string{"list-windows", "-a", "-F", format}
	} else if len(sessionNames) == 1 {
		args = []string{"list-windows", "-t", sessionNames[0], "-F", format}
	} else {
		var allWindows []TmuxWindow
		for _, sessionName := range sessionNames {
			windows, err := t.GetWindows(sessionName)
			if err != nil {
				return nil, err
			}
			allWindows = append(allWindows, windows...)
		}
		return allWindows, nil
	}

	stdout, err := t.Run(args...)
	if err != nil {
		return nil, err
	}

	var windows []TmuxWindow
	lines := strings.Split(strings.TrimSpace(stdout), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		index, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		windows = append(windows, TmuxWindow{
			SessionName: parts[0],
			Index:       index,
			Name:        parts[2],
		})
	}

	return windows, nil
}
func (t *Tmux) UpdatePosition(sessionName string) {
	for i := range t.Sessions {
		if t.Sessions[i].Name == sessionName {
			t.Current = &t.Sessions[i]
			t.CurrentIndex = i
			break
		}
	}
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

func (t *Tmux) HasSession(session_name string) (bool, error) {
	_, err := t.Run("has-session", "-t", session_name)
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func parseSessions(line string) (TmuxSession, error) {
	fields := strings.Split(line, "|")
	if len(fields) < 4 {
		return TmuxSession{}, fmt.Errorf("invalid session line: %q", line)
	}

	createdStr := fields[0]
	name := fields[1]
	group := fields[2]
	attachedStr := fields[3]

	createdUnix, err := strconv.ParseInt(createdStr, 10, 64)
	if err != nil {
		return TmuxSession{}, fmt.Errorf("invalid creation timestamp %q: %w", createdStr, err)
	}

	attached := attachedStr == "1"

	return TmuxSession{
		Name:     name,
		Created:  time.Unix(createdUnix, 0),
		Group:    group,
		Attached: attached,
		Windows:  nil,
	}, nil
}

func NewTmux() (*Tmux, error) {
	t := &Tmux{}

	if err := t.LoadSessions(); err != nil {
		return nil, err
	}

	for i := range t.Sessions {
		if t.Sessions[i].Attached {
			t.Current = &t.Sessions[i]
			break
		}
	}

	if t.Current == nil && len(t.Sessions) > 0 {
		t.Current = &t.Sessions[0]
	}

	return t, nil
}

func (t *Tmux) LoadSessions() error {
	sessions, err := t.GetSessions()
	if err != nil {
		return err
	}
	t.Sessions = sessions
	return nil
}

func (t *Tmux) GetSessions() ([]TmuxSession, error) {
	stdout, err := t.Run("list-sessions", "-F", "#{session_created}|#{session_name}|#{session_group}|#{?session_attached,1,0}")
	if err != nil {
		return nil, err
	}

	var tmuxSessions []TmuxSession
	trimmed := strings.TrimSpace(stdout)
	sessions := strings.Split(trimmed, "\n")

	for _, s := range sessions {
		if s == "" {
			continue
		}
		parsed_session, err := parseSessions(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tmux session: %w", err)
		}
		tmuxSessions = append(tmuxSessions, parsed_session)
	}

	return tmuxSessions, nil
}

func (t *Tmux) SwitchSession(sessionName string) error {
	t.UpdatePosition(sessionName)

	var err error
	if os.Getenv("TMUX") == "" {
		_, err = t.Run("attach-session", "-t", sessionName)
	} else {
		_, err = t.Run("switch-client", "-t", sessionName)
	}

	if err != nil {
		return fmt.Errorf("failed to switch to the session %v: %w", sessionName, err)
	}
	return nil
}

// Switch to a specific window in a session
func (t *Tmux) SwitchWindow(sessionName string, windowIndex int) error {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)

	var err error
	if os.Getenv("TMUX") == "" {
		_, err = t.Run("attach-session", "-t", target)
	} else {
		_, err = t.Run("switch-client", "-t", target)
	}

	if err != nil {
		return fmt.Errorf("failed to switch to window %d in session %v: %w", windowIndex, sessionName, err)
	}

	t.UpdatePosition(sessionName)
	return nil
}

func (t *Tmux) CreateSession(name string, absPath string) error {
	sessionName := name
	if strings.HasPrefix(name, ".") {
		sessionName = name[1:]
	}

	exists, err := t.HasSession(sessionName)
	if err != nil {
		return errors.New("failed to check if session exists")
	}

	if !exists {
		_, err = t.Run("new-session", "-d", "-s", sessionName, "-c", absPath)
		if err != nil {
			return fmt.Errorf("failed to create the session name:%v, path:%v: %w", sessionName, absPath, err)
		}

		if err := t.LoadSessions(); err != nil {
			return fmt.Errorf("failed to reload sessions: %w", err)
		}

		t.UpdatePosition(sessionName)
	}

	return t.SwitchSession(sessionName)
}
