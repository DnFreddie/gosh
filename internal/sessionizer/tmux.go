package sessionizer

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"syscall"
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

// GetWindows returns a sorted list of windows for specific session(s).
// If no sessions are provided or "*" is in the list, it returns windows for all sessions.
func (t *Tmux) GetWindows(sessionNames ...string) ([]TmuxWindow, error) {
	var allWindows []TmuxWindow

	getAllSessions := len(sessionNames) == 0 || slices.Contains(sessionNames, "*")

	if getAllSessions {
		windows, err := t.queryWindows("-a")
		if err != nil {
			return nil, err
		}
		allWindows = windows
	} else {
		for _, name := range sessionNames {

			if name == "" {
				continue
			}
			windows, err := t.queryWindows("-t", name)
			if err != nil {
				return nil, err
			}
			allWindows = append(allWindows, windows...)
		}
	}

	slices.SortFunc(allWindows, func(a, b TmuxWindow) int {
		if a.SessionName != b.SessionName {
			return strings.Compare(a.SessionName, b.SessionName)
		}
		return a.Index - b.Index
	})

	return allWindows, nil
}

func (t *Tmux) queryWindows(extraArgs ...string) ([]TmuxWindow, error) {
	format := "#{session_name}|#{window_index}|#{window_name}"
	args := append([]string{"list-windows", "-F", format}, extraArgs...)

	stdout, err := t.Run(args...)
	if err != nil {
		return nil, err
	}

	var windows []TmuxWindow
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		index, _ := strconv.Atoi(parts[1])
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
	if len(args) == 0 {
		return "", fmt.Errorf("no command specified")
	}

	if err := t.requiresTerminal(args[0]); err != nil {
		return "", err
	}

	cmd := t.buildCommand(args)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprint(os.Stderr, stderr.String())
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func (t *Tmux) requiresTerminal(command string) error {
	insideTmux := os.Getenv("TMUX") != ""

	if command == "attach-session" || command == "attach" {
		return fmt.Errorf("'%s' requires direct terminal access", command)
	}
	if command == "switch-client" && !insideTmux {
		return fmt.Errorf("'switch-client' requires terminal when outside tmux")
	}
	return nil
}

func (t *Tmux) buildCommand(args []string) *exec.Cmd {
	if os.Getenv("TMUX") != "" {
		return exec.Command("tmux", args...)
	}

	// Quote args with special characters
	quoted := make([]string, len(args))
	for i, arg := range args {
		if strings.ContainsAny(arg, " |{}()?$`\"\\") {
			quoted[i] = "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
		} else {
			quoted[i] = arg
		}
	}

	return exec.Command("bash", "-c", "tmux "+strings.Join(quoted, " "))
}

func (t *Tmux) HasSession(sessionName string) (bool, error) {
	stdout, err := t.Run("list-sessions", "-F", "#{session_name}")
	if err != nil {
		return false, fmt.Errorf("failed to list sessions: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	return slices.Contains(lines, sessionName), nil
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
		parsedSessions, err := parseSessions(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tmux session: %w", err)
		}
		tmuxSessions = append(tmuxSessions, parsedSessions)
	}

	return tmuxSessions, nil
}

func (t *Tmux) SwitchSession(sessionName string) error {
	t.UpdatePosition(sessionName)

	if os.Getenv("TMUX") == "" {
		tmuxPath, err := exec.LookPath("tmux")
		if err != nil {
			return fmt.Errorf("tmux not found in PATH: %w", err)
		}

		if err := syscall.Exec(tmuxPath, []string{"tmux", "attach-session", "-t", sessionName}, os.Environ()); err != nil {
			return fmt.Errorf("failed to attach to session %s: %w", sessionName, err)
		}
		return nil
	} else {
		_, err := t.Run("switch-client", "-t", sessionName)
		if err != nil {
			return fmt.Errorf("failed to switch to the session %v: %w", sessionName, err)
		}
		return nil
	}
}

// SwitchWindow  to a specific window in a session
func (t *Tmux) SwitchWindow(sessionName string, windowIndex int) error {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)

	if os.Getenv("TMUX") == "" {
		// When not in tmux, use exec to replace the current process with tmux attach-session
		// This ensures tmux gets proper access to the terminal
		tmuxPath, err := exec.LookPath("tmux")
		if err != nil {
			return fmt.Errorf("tmux not found in PATH: %w", err)
		}

		if err := syscall.Exec(tmuxPath, []string{"tmux", "attach-session", "-t", target}, os.Environ()); err != nil {
			return fmt.Errorf("failed to attach to window %s:%d: %w", sessionName, windowIndex, err)
		}
		// This should never be reached
		return fmt.Errorf("terminal error: tmux exec returned without error")
	} else {
		// When already in tmux, use switch-client
		_, err := t.Run("switch-client", "-t", target)
		if err != nil {
			return fmt.Errorf("failed to switch to window %d in session %v: %w", windowIndex, sessionName, err)
		}

		t.UpdatePosition(sessionName)
		return nil
	}
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
