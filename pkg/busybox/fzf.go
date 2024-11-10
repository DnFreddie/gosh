package busybox

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
)

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

func isControlRune(r rune) bool {
	return r < 32 || r == 127
}
