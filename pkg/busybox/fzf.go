package busybox

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
)

func RunTerm[T any](items []T, toString func(T) string) (T, error) {
	var zero T
	if len(items) == 0 {
		return zero, fmt.Errorf("no items available")
	}

	lookup := make(map[string]T)
	displayItems := make([]string, 0, len(items))
	for _, item := range items {
		s := toString(item)
		lookup[s] = item
		displayItems = append(displayItems, s)
	}

	term := NewTerm()
	defer term.Close()

	if err := term.GetSize(); err != nil {
		return zero, fmt.Errorf("getting terminal size: %w", err)
	}

	terminalHeight := term.Height()
	termHeight := int(float64(terminalHeight) * 0.8)
	if termHeight < 5 {
		termHeight = 5
	}

	term.EnterAltBuffer()
	defer term.ExitAltBuffer()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	go func() {
		<-sigChan
		term.ExitAltBuffer()
		Quit(term)
	}()

	var input string
	var selectionIndex, scrollOffset int

	for {
		// Filter strings using the input
		filteredStrings := filterStrings(displayItems, input)

		if len(filteredStrings) == 0 {
			selectionIndex, scrollOffset = 0, 0
		} else if selectionIndex >= len(filteredStrings) {
			selectionIndex = len(filteredStrings) - 1
		}

		if selectionIndex < scrollOffset {
			scrollOffset = selectionIndex
		} else if selectionIndex >= scrollOffset+termHeight {
			scrollOffset = selectionIndex - termHeight + 1
		}

		screen := buildScreenStrings(input, filteredStrings, selectionIndex, scrollOffset, termHeight)

		fmt.Print(screen)

		key, r := term.Read()

		switch key {
		case CtrlC, Escape:
			return zero, fmt.Errorf("cancelled")
		case Backspace:
			if len(input) > 0 {
				input = input[:len(input)-1]
				selectionIndex, scrollOffset = 0, 0
			}
		case Enter:
			if len(filteredStrings) > 0 {
				selectedStr := filteredStrings[selectionIndex]
				return lookup[selectedStr], nil
			}
		case UpArrow:
			if selectionIndex > 0 {
				selectionIndex--
			}
		case DownArrow:
			if selectionIndex < len(filteredStrings)-1 {
				selectionIndex++
			}
		case Other:
			if r != 0 && !isControlRune(r) {
				input += string(r)
				selectionIndex, scrollOffset = 0, 0
			}
		}
	}
}

func filterStrings(items []string, input string) []string {
	filtered := make([]string, 0, len(items))
	inputLower := strings.ToLower(input)
	for _, s := range items {
		if strings.Contains(strings.ToLower(s), inputLower) {
			filtered = append(filtered, s)
		}
	}
	sort.Strings(filtered)
	return filtered
}

func buildScreenStrings(input string, items []string, selectionIndex, scrollOffset, termHeight int) string {
	var buf bytes.Buffer

	buf.WriteString(MoveCursorHome)
	buf.WriteString(ClearToEOS)

	buf.WriteString(InColors(Cyan, "> "))
	buf.WriteString(input)
	buf.WriteString(CRLF)
	buf.WriteString(CRLF)

	if len(items) == 0 {
		buf.WriteString(InColors(Red, "No results found."))
		buf.WriteString(CRLF)
	} else {
		endIndex := scrollOffset + termHeight
		if endIndex > len(items) {
			endIndex = len(items)
		}

		for i := scrollOffset; i < endIndex; i++ {
			itemStr := items[i]
			if i == selectionIndex {
				buf.WriteString(InColors(Blue, "> "+itemStr))
			} else {
				buf.WriteString("  ")
				buf.WriteString(itemStr)
			}
			buf.WriteString(CRLF)
		}

		if len(items) > termHeight {
			visibleEnd := scrollOffset + termHeight
			if visibleEnd > len(items) {
				visibleEnd = len(items)
			}
			buf.WriteString(CRLF)
			buf.WriteString(InColors(Cyan, fmt.Sprintf("[%d/%d]", visibleEnd, len(items))))
		}
	}

	return buf.String()
}

func isControlRune(r rune) bool {
	return r < 32 || r == 127
}
