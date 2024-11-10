package busybox

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	_ "github.com/alecthomas/chroma/lexers/a" // for Bash
	_ "github.com/alecthomas/chroma/lexers/g" // for Go
	_ "github.com/alecthomas/chroma/lexers/i" // for INI
	_ "github.com/alecthomas/chroma/lexers/j" // for JSON, Java, JavaScript
	_ "github.com/alecthomas/chroma/lexers/m" // for Markdown
	_ "github.com/alecthomas/chroma/lexers/p" // for Python, Perl
	_ "github.com/alecthomas/chroma/lexers/r" // for Rust
	_ "github.com/alecthomas/chroma/lexers/t" // for TOML
	_ "github.com/alecthomas/chroma/lexers/y" // for YAML
	"github.com/alecthomas/chroma/styles"
	"golang.org/x/term"
)

// Language mappings
var languages = map[string]string{
	".sh":   "bash",
	".rs":   "rust",
	".go":   "go",
	".yaml": "yaml",
	".yml":  "yaml",
	".json": "json",
	".pl":   "perl",
	".py":   "python",
	".js":   "javascript",
	".java": "java",
	".ini":  "ini",
	".toml": "toml",
	".md":   "markdown",
}

func findLanguage(fileName string) string {
	ext := filepath.Ext(fileName)
	if lang, exists := languages[ext]; exists {
		return lang
	}
	return "unknown"
}

type HighlightedPager struct {
	term     *Terminal
	filename string
	content  io.Reader
}

func NewHighlightedPager(filename string, content io.Reader) *HighlightedPager {
	return &HighlightedPager{
		filename: filename,
		content:  content,
	}
}

func (hp *HighlightedPager) Run() error {
	// If stdout is not a terminal, just highlight and print
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		content, err := io.ReadAll(hp.content)
		if err != nil {
			return fmt.Errorf("reading content: %w", err)
		}
		return highlightCode(findLanguage(hp.filename), string(content))
	}

	lines, err := hp.getHighlightedLines()
	if err != nil {
		return fmt.Errorf("highlighting content: %w", err)
	}

	hp.term = NewTerm().(*Terminal)
	defer hp.term.Close()

	if err := hp.term.GetSize(); err != nil {
		return fmt.Errorf("getting terminal size: %w", err)
	}

	hp.displayLoop(lines)
	return nil
}

func (hp *HighlightedPager) getHighlightedLines() ([]string, error) {
	var buf bytes.Buffer

	lang := findLanguage(hp.filename)
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	content, err := io.ReadAll(hp.content)
	if err != nil {
		return nil, fmt.Errorf("reading content: %w", err)
	}

	formatter := formatters.Get("terminal")
	style := styles.Vim

	iterator, err := lexer.Tokenise(nil, string(content))
	if err != nil {
		return nil, fmt.Errorf("tokenizing content: %w", err)
	}

	if err := formatter.Format(&buf, style, iterator); err != nil {
		return nil, fmt.Errorf("formatting content: %w", err)
	}

	lines := strings.Split(buf.String(), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines, nil
}

func (hp *HighlightedPager) displayLoop(lines []string) {
	offset := 0
	pageSize := hp.term.Height - 1 // Save space for status line

	for {
		hp.term.Clear()
		hp.displayPage(lines, offset, pageSize)
		hp.displayStatusLine(offset, len(lines))

		if !hp.handleNavigation(&offset, lines, pageSize) {
			return
		}
	}
}

func (hp *HighlightedPager) displayPage(lines []string, offset, pageSize int) {
	for i := 0; i < pageSize && offset+i < len(lines); i++ {
		// Reset cursor to start of line before printing
		fmt.Print(ResetCursor)
		fmt.Println(lines[offset+i])
	}
}

func (hp *HighlightedPager) displayStatusLine(offset, totalLines int) {
	//currentPage := (offset / (hp.term.Height - 1)) + 1
	totalPages := (totalLines / (hp.term.Height - 1)) + 1
	if totalLines%(hp.term.Height-1) == 0 {
		totalPages--
	}

	fmt.Printf("\033[%d;0H", hp.term.Height)
	fmt.Print(ResetCursor)

	fmt.Print("\033[7m")
	statusText := fmt.Sprintf(" File: %s | Line: %d/%d | Press q/Esc to quit ",
		hp.filename, offset+1, totalLines)

	// Truncate status if  too long
	if len(statusText) > hp.term.Width {
		statusText = statusText[:hp.term.Width-3] + "..."
	}

	fmt.Print(statusText)

	padding := hp.term.Width - len(statusText)
	if padding > 0 {
		fmt.Print(strings.Repeat(" ", padding))
	}

	fmt.Print("\033[0m")
}

func highlightCode(lang string, content string) error {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	formatter := formatters.Get("terminal")
	style := styles.Monokai

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return err
	}

	return formatter.Format(os.Stdout, style, iterator)
}

func (hp *HighlightedPager) handleNavigation(offset *int, lines []string, pageSize int) bool {
	key, ch := read()

	switch key {
	case CtrlC, Escape:
		return false
	case Other:
		switch ch {
		case 'q', 'Q':
			return false
		case 'j':
			if *offset < len(lines)-pageSize {
				*offset++
			}
		case 'k':
			if *offset > 0 {
				*offset--
			}
		case 'g':
			*offset = 0
		case 'G':
			*offset = len(lines) - pageSize
			if *offset < 0 {
				*offset = 0
			}
		case 'u':
			*offset -= pageSize / 2
			if *offset < 0 {
				*offset = 0
			}
		case 'd':
			*offset += pageSize / 2
			maxOffset := len(lines) - pageSize
			if maxOffset < 0 {
				maxOffset = 0
			}
			if *offset > maxOffset {
				*offset = maxOffset
			}
		}
	case UpArrow:
		if *offset > 0 {
			*offset--
		}
	case DownArrow:
		maxOffset := len(lines) - pageSize
		if maxOffset < 0 {
			maxOffset = 0
		}
		if *offset < maxOffset {
			*offset++
		}
	}

	return true
}

