package busybox

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
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

type EscapeCode string

const (
	Clear       EscapeCode = "\033[H\033[2J\033[H" // Clear screen and reset cursor
	ResetCursor EscapeCode = "\033[0G"             // Move cursor to the beginning of the line
	HideCursor  EscapeCode = "\033[?25l"           // Hide cursor
	ShowCursor  EscapeCode = "\033[?25h"           // Show cursor
)

func clearTerminal() {
	fmt.Print(Clear)
}

type Color string

const (
	Red    Color = "\033[31m"
	Reset  Color = "\033[0m"
	Green  Color = "\033[32m"
	Blue   Color = "\033[34m"
	Cyan   Color = "\033[36m"
	Yellow Color = "\033[33m"
)

func InColors(c Color, s string) {
	fmt.Print(c, s, Reset)
}

type Term interface {
	Start()
	Close()
	Clear()
}

// Quit gracefully exits the program, restoring the terminal state.
func Quit(t Term) {
	t.Close()
	t.Clear()
	defer os.Exit(0)
}

func NewTerm() Term {
	newTerm := &Terminal{}
	newTerm.Start()
	return newTerm
}

// Terminal implements the Term interface.
type Terminal struct {
	oldState *term.State
}

func (t *Terminal) Start() {
	t.startRawMode()
}

func (t *Terminal) startRawMode() {
	var err error
	t.oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
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
		if err := term.Restore(int(os.Stdin.Fd()), t.oldState); err != nil {
			panic(fmt.Sprintf("Failed to restore terminal: %v", err))
		}
		t.oldState = nil
	}
}

func (t *Terminal) Clear() {
	clearTerminal()
}

func read() (Key, rune) {
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
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
	case 13, 10: // Handle both Enter (13) and Line Feed (10)
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

type Stringer interface {
	String() string
}

func RunTerm[T Stringer](items []T) (T, error) {
	var input string
	var selectionIndex int

	if len(items) == 0 {
		var null T
		return null, fmt.Errorf("No items available.")
	}

	term := NewTerm()
	defer term.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		Quit(term)
	}()

	for {
		term.Clear()
		InColors(Cyan, fmt.Sprintf("> %s\n\n", input))

		filteredItems := filterItems(items, input)

		// Adjust selection index to be within bounds.
		if len(filteredItems) == 0 {
			selectionIndex = 0
		} else if selectionIndex >= len(filteredItems) {
			selectionIndex = len(filteredItems) - 1
		}

		displayResults(filteredItems, selectionIndex)

		key, r := read()

		switch key {
		case CtrlC, Escape:
			Quit(term)

		case Backspace:
			if len(input) > 0 {
				input = input[:len(input)-1]
			}
		case Enter:
			if len(filteredItems) > 0 {
				selected := filteredItems[selectionIndex]
				return selected, nil
			}
		case UpArrow:
			if selectionIndex > 0 {
				selectionIndex--
			}
		case DownArrow:
			if selectionIndex < len(filteredItems)-1 {
				selectionIndex++
			}
		case Other:
			if r != 0 && !isControlRune(r) {
				input += string(r)
			}
		}
	}
}

func isControlRune(r rune) bool {
	return r < 32 || r == 127
}

func displayResults[T Stringer](filteredItems []T, selectionIndex int) {
	if len(filteredItems) == 0 {
		InColors(Red, "No results found.\n")
	} else {
		fmt.Print(ResetCursor)
		for i, item := range filteredItems {
			if i == selectionIndex {
				InColors(Blue, fmt.Sprintf("> %v\n", item.String()))
				fmt.Print(ResetCursor)
			} else {
				fmt.Printf("  %v\n", item.String())
				fmt.Print(ResetCursor)
			}
		}
	}
}

func filterItems[T Stringer](items []T, input string) []T {
	filtered := make([]T, 0, len(items))
	inputLower := strings.ToLower(input)

	for _, item := range items {
		if strings.Contains(strings.ToLower(item.String()), inputLower) {
			filtered = append(filtered, item)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].String() < filtered[j].String()
	})

	return filtered
}
