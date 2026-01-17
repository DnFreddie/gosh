package busybox

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// ItemFormatter allows customizing display with optional grouping
type ItemFormatter[T any] struct {
	ToString  func(T) string
	GetGroup  func(T) string // Optional: return group name for grouping
	Separator string         // Optional: custom separator between groups (default: blank line)
}

// RunTerm is the backward-compatible simple version
func RunTerm[T any](items []T, toString func(T) string) (T, error) {
	return RunTermGrouped(items, ItemFormatter[T]{
		ToString: toString,
	})
}

func RunTermGrouped[T any](items []T, formatter ItemFormatter[T]) (T, error) {
	var zero T
	if len(items) == 0 {
		return zero, fmt.Errorf("no items available")
	}

	lookup := make(map[string]T)
	displayItems := make([]string, 0, len(items))
	lastGroup := ""

	for _, item := range items {
		// Add group separator/header
		if formatter.GetGroup != nil {
			group := formatter.GetGroup(item)
			if group != lastGroup {
				if lastGroup != "" {
					displayItems = append(displayItems, "")
				}
				// Add custom separator with group name replacement
				if formatter.Separator != "" {
					sep := strings.ReplaceAll(formatter.Separator, "{{GROUP}}", group)
					displayItems = append(displayItems, sep) // Separator - NOT in lookup
				}
				lastGroup = group
			}
		}

		s := formatter.ToString(item)
		lookup[s] = item // Only actual items go in lookup
		displayItems = append(displayItems, s)
	}

	term := NewTerm()
	defer term.Close()

	if err := term.GetSize(); err != nil {
		return zero, fmt.Errorf("getting terminal size: %w", err)
	}

	termHeight := int(float64(term.Height()) * 0.8)
	termHeight = max(5, termHeight)

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

	sel := &selector[T]{
		items:      displayItems,
		termHeight: termHeight,
	}

	for {
		filtered := sel.filter(lookup)
		sel.clampSelection(len(filtered))
		sel.adjustScroll(filtered, lookup)

		fmt.Print(sel.render(filtered))

		key, r := term.Read()
		if sel.handleInput(key, r, filtered, lookup) {
			if sel.cancelled {
				return zero, fmt.Errorf("cancelled")
			}
			return sel.selected, nil
		}
	}
}

type selector[T any] struct {
	items          []string
	input          string
	selectionIndex int
	scrollOffset   int
	termHeight     int
	selected       T
	cancelled      bool
}

func (s *selector[T]) filter(lookup map[string]T) []string {
	if s.input == "" {
		return s.items
	}

	inputLower := strings.ToLower(s.input)

	// Track matching items and their preceding separators
	filtered := make([]string, 0)
	pendingSeparators := make([]string, 0) // Store separators until we find a match
	groupMatches := false                  // Track if the current group header matches

	for _, item := range s.items {
		// Check if this is a separator (not in lookup)
		_, isActualItem := lookup[item]

		if !isActualItem {

			// Check if this separator (group name) matches the filter
			if item != "" && strings.Contains(strings.ToLower(item), inputLower) {
				groupMatches = true
			} else {
				groupMatches = false
			}

			pendingSeparators = append(pendingSeparators, item)
			continue
		}

		itemMatches := strings.Contains(strings.ToLower(item), inputLower)

		if itemMatches || groupMatches {

			filtered = append(filtered, pendingSeparators...)
			pendingSeparators = nil

			filtered = append(filtered, item)
		} else {
			pendingSeparators = nil
		}
	}

	return filtered
}

func (s *selector[T]) clampSelection(max int) {
	if max == 0 {
		s.selectionIndex = 0
		return
	}
	if s.selectionIndex >= max {
		s.selectionIndex = max - 1
	}
}

func (s *selector[T]) adjustScroll(filtered []string, lookup map[string]T) {
	if len(filtered) == 0 {
		s.selectionIndex = 0
		s.scrollOffset = 0
		return
	}

	// Clamp selection to valid range first
	if s.selectionIndex < 0 {
		s.selectionIndex = 0
	}
	if s.selectionIndex >= len(filtered) {
		s.selectionIndex = len(filtered) - 1
	}

	// Skip forward over separators (items not in lookup)
	for s.selectionIndex < len(filtered) {
		if _, exists := lookup[filtered[s.selectionIndex]]; exists {
			break // Found a valid item
		}
		s.selectionIndex++
	}

	// If we went past the end, go backwards to find a valid item
	if s.selectionIndex >= len(filtered) {
		s.selectionIndex = len(filtered) - 1
		for s.selectionIndex >= 0 {
			if _, exists := lookup[filtered[s.selectionIndex]]; exists {
				break // Found a valid item
			}
			s.selectionIndex--
		}
	}

	// If still no valid selection found, reset to 0
	if s.selectionIndex < 0 {
		s.selectionIndex = 0
	}

	// Adjust scroll to keep selection visible
	if s.selectionIndex < s.scrollOffset {
		s.scrollOffset = s.selectionIndex
	} else if s.selectionIndex >= s.scrollOffset+s.termHeight {
		s.scrollOffset = s.selectionIndex - s.termHeight + 1
	}

	// Ensure scroll offset is valid
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
}

func (s *selector[T]) render(filtered []string) string {
	var buf bytes.Buffer

	buf.WriteString(MoveCursorHome + ClearToEOS)
	buf.WriteString(InColors(Cyan, "> ") + s.input + CRLF + CRLF)

	if len(filtered) == 0 {
		buf.WriteString(InColors(Red, "No results found.") + CRLF)
		return buf.String()
	}

	end := min(s.scrollOffset+s.termHeight, len(filtered))
	for i := s.scrollOffset; i < end; i++ {
		// Empty separator line
		if filtered[i] == "" {
			buf.WriteString(CRLF)
			continue
		}

		if i == s.selectionIndex {
			buf.WriteString(InColors(Blue, "> "+filtered[i]))
		} else {
			buf.WriteString("  " + filtered[i])
		}
		buf.WriteString(CRLF)
	}

	if len(filtered) > s.termHeight {
		buf.WriteString(CRLF + InColors(Cyan, fmt.Sprintf("[%d/%d]", end, len(filtered))))
	}

	return buf.String()
}

func (s *selector[T]) handleInput(key Key, r rune, filtered []string, lookup map[string]T) bool {
	switch key {
	case CtrlC, Escape:
		s.cancelled = true
		return true

	case Backspace:
		if len(s.input) > 0 {
			s.input = s.input[:len(s.input)-1]
			s.selectionIndex, s.scrollOffset = 0, 0
		}

	case Enter:
		if s.selectionIndex < len(filtered) {
			selectedStr := filtered[s.selectionIndex]
			// Only return if it's actually in the lookup (not a separator)
			if item, exists := lookup[selectedStr]; exists {
				s.selected = item
				return true
			}
		}

	case UpArrow:
		// Move up, skipping separators (items not in lookup)
		for i := s.selectionIndex - 1; i >= 0; i-- {
			if _, exists := lookup[filtered[i]]; exists {
				s.selectionIndex = i
				break
			}
		}

	case DownArrow:
		// Move down, skipping separators (items not in lookup)
		for i := s.selectionIndex + 1; i < len(filtered); i++ {
			if _, exists := lookup[filtered[i]]; exists {
				s.selectionIndex = i
				break
			}
		}

	case Other:
		if r != 0 && !isControlRune(r) {
			s.input += string(r)
			s.selectionIndex, s.scrollOffset = 0, 0
		}
	}
	return false
}

func isControlRune(r rune) bool {
	return r < 32 || r == 127
}
