package busybox

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

type Key int

const (
	Unknown Key = iota
	CtrlC
	Backspace
	Enter
	Escape
	UpArrow
	DownArrow
	Other
)

// ANSI escape codes for terminal control
const (
	ClearScreen     = "\033[2J\033[H" // Clear screen and move to home
	ClearToEOL      = "\033[K"        // Clear from cursor to end of line
	ClearToEOS      = "\033[J"        // Clear from cursor to end of screen
	MoveCursorHome  = "\033[H"        // Move cursor to home (top-left)
	ResetCursor     = "\033[0G"       // Move cursor to start of line
	HideCursor      = "\033[?25l"     // Hide cursor
	ShowCursor      = "\033[?25h"     // Show cursor
	InverseVideo    = "\033[7m"       // Inverse/reverse video
	ResetFormatting = "\033[0m"       // Reset all formatting
	CRLF            = "\r\n"          // Carriage return + line feed
	EnterAltScreen  = "\033[?1049h"   // Enter alternate screen
	ExitAltScreen   = "\033[?1049l"   // Exit alternate screen
	SaveScreen      = "\033[?47h"     // Save screen
	RestoreScreen   = "\033[?47l"     // Restore screen
)

// ANSI color codes

type Color string

type EscapeCode string

const (
	Red    Color = "\033[31m"
	Reset  Color = "\033[0m"
	Green  Color = "\033[32m"
	Blue   Color = "\033[34m"
	Cyan   Color = "\033[36m"
	Yellow Color = "\033[33m"
)

func InColors(c Color, s string) string {
	return fmt.Sprintf("%s%s%s", c, s, Reset)
}

func MoveCursorTo(row, col int) string {
	return fmt.Sprintf("\033[%d;%dH", row, col)
}

type Term interface {
	Start()
	Close()
	Clear()
	Read() (Key, rune)
	GetSize() error

	// Screen buffer management
	EnterAltBuffer()
	ExitAltBuffer()

	// Access to terminal dimensions
	Width() int
	Height() int
}

// Terminal implements the [Term] interface
type Terminal struct {
	oldState *term.State
	width    int // Changed to lowercase
	height   int // Changed to lowercase
	tty      *os.File
}

func NewTerm() Term {
	t := &Terminal{}
	t.Start()
	return t
}

func (t *Terminal) Start() {
	t.startRawMode()
	t.setupSignalHandler()
}

func (t *Terminal) Width() int {
	return t.width
}

func (t *Terminal) Height() int {
	return t.height
}

func (t *Terminal) EnterAltBuffer() {
	fmt.Print(EnterAltScreen)
}

func (t *Terminal) ExitAltBuffer() {
	fmt.Print(ExitAltScreen)
	fmt.Print(ShowCursor)
}

func (t *Terminal) setupSignalHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		t.Close()
		t.Clear()
		os.Exit(0)
	}()
}

func (t *Terminal) GetSize() error {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return fmt.Errorf("get terminal size: %w", err)
	}

	t.width = width   // Changed to lowercase
	t.height = height // Changed to lowercase
	return nil
}

func Quit(t Term) {
	t.Close()
	t.Clear()
	defer os.Exit(0)
}

func (t *Terminal) startRawMode() {
	// Open /dev/tty for raw input mode
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		panic(fmt.Sprintf("Failed to open /dev/tty: %v", err))
	}
	t.tty = tty

	t.oldState, err = term.MakeRaw(int(tty.Fd()))
	if err != nil {
		panic(fmt.Sprintf("Failed to set raw mode: %v", err))
	}
	fmt.Print(HideCursor)
}

func (t *Terminal) Close() {
	fmt.Print(ShowCursor)
	t.stopRawMode()
}

func (t *Terminal) stopRawMode() {
	if t.oldState != nil {
		if err := term.Restore(int(t.tty.Fd()), t.oldState); err != nil {
			panic(fmt.Sprintf("Failed to restore terminal: %v", err))
		}
		t.oldState = nil
	}
	if t.tty != nil {
		t.tty.Close()
		t.tty = nil
	}
}

func (t *Terminal) Clear() {
	fmt.Print(ClearScreen)
}

// Read reads from the terminal's tty file descriptor
func (t *Terminal) Read() (Key, rune) {
	buf := make([]byte, 3)
	n, err := t.tty.Read(buf)
	if err != nil {
		panic(fmt.Sprintf("Failed to read input: %v", err))
	}
	if n == 0 {
		return Unknown, 0
	}

	return parseKey(buf, n)
}

func parseKey(buf []byte, n int) (Key, rune) {
	switch buf[0] {
	case 3:
		return CtrlC, 0
	case 127:
		return Backspace, 0
	case 13, 10:
		return Enter, 0
	case 27:
		if n > 1 && buf[1] == '[' {
			switch buf[2] {
			case 'A':
				return UpArrow, 0
			case 'B':
				return DownArrow, 0
			}
		}
		return Escape, 0
	default:
		return Other, rune(buf[0])
	}
}
