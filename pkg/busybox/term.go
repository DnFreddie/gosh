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
	clearScreen = "\033[H\033[2J\033[H" // Clear screen and reset cursor
	resetCursor = "\033[0G"             // Move cursor to start of line
	hideCursor  = "\033[?25l"           // Hide cursor
	showCursor  = "\033[?25h"           // Show cursor
)

// ANSI color codes
const (
	colorRed    = "\033[31m"
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
)

type Color string

type EscapeCode string

const (
	ResetCursor EscapeCode = "\033[0G"   // Move cursor to the beginning of the line
	HideCursor  EscapeCode = "\033[?25l" // Hide cursor
	Red         Color      = "\033[31m"
	Reset       Color      = "\033[0m"
	Green       Color      = "\033[32m"
	Blue        Color      = "\033[34m"
	Cyan        Color      = "\033[36m"
	Yellow      Color      = "\033[33m"
)

func InColors(c Color, s string) {
	fmt.Print(c, s, Reset)
}

type Term interface {
	Start()
	Close()
	Clear()
}

// Terminal implements the [Term] interface
type Terminal struct {
	oldState *term.State
	Width    int
	Height   int
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

	t.Width = width
	t.Height = height
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
	fmt.Print(showCursor)
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

// Clear clears the terminal screen
func (t *Terminal) Clear() {
	fmt.Print(clearScreen)
}

// read reads a single keystroke from the terminal
func read() (Key, rune) {
	// Read from /dev/tty instead of stdin
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		panic(fmt.Sprintf("Failed to open /dev/tty: %v", err))
	}
	defer tty.Close()

	buf := make([]byte, 3)
	n, err := tty.Read(buf)
	if err != nil {
		panic(fmt.Sprintf("Failed to read input: %v", err))
	}
	if n == 0 {
		return Unknown, 0
	}

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

// parseKey converts raw input bytes into a Key type
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

// PrintColored prints text in the specified color
func PrintColored(color string, text string) {
	fmt.Print(color, text, colorReset)
}
