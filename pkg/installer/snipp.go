package installer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/DnFreddie/goseq/pkg/terminal"
)

type Snippet struct {
	Lang    string
	Name    string
	Content bytes.Buffer
}

func (s *Snippet) PrintSnippet() {

	green := "\033[32m"
	blue := "\033[34m"
	reset := "\033[0m"

	fmt.Println(strings.Repeat("-", 30))
	fmt.Println("Chosen Snippet:")
	fmt.Println(strings.Repeat("-", 30))
	fmt.Printf("Language: %s%s%s\n", green, s.Lang, reset)
	fmt.Printf("Name: %s%s%s\n", blue, s.Name, reset)
}

type SnipScanner struct {
	scanner *bufio.Scanner
	curr    Snippet
	err     error
}

func NewSnipScanner(r io.Reader) *SnipScanner {
	return &SnipScanner{
		scanner: bufio.NewScanner(r),
	}
}

func (s *SnipScanner) Snippet() Snippet {
	return s.curr
}

func (s *SnipScanner) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.scanner.Err()
}

func (s *SnipScanner) Scan() bool {
	// Reset current snippet content
	s.curr = Snippet{}

	inCodeBlock := false

	for s.scanner.Scan() {
		line := s.scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				s.curr.Lang = getCodeFormat(trimmedLine)
				continue
			} else {
				return true
			}
		}

		name := getName(line)
		if name != "" {
			s.curr.Name = name
			continue
		}

		if inCodeBlock {
			_, err := s.curr.Content.WriteString(line + "\n")
			if err != nil {
				s.err = fmt.Errorf("error writing to buffer: %v", err)
				return false
			}
		}
	}

	return false
}

func getName(line string) string {
	startTag := "##NAME:"
	endTag := "##"

	startIndex := strings.Index(line, startTag)
	if startIndex == -1 {
		return ""
	}

	startIndex += len(startTag)
	endIndex := strings.Index(line[startIndex:], endTag)
	if endIndex == -1 {
		return ""
	}

	return line[startIndex : startIndex+endIndex]
}

func getCodeFormat(line string) string {
	if strings.HasPrefix(line, "```") {
		var collect string
		for _, v := range line[3:] {
			if v != ' ' {
				collect += string(v)
			} else {
				return collect
			}
		}
		return collect
	}
	return ""
}

func GetSnippet(r io.Reader) (*Snippet, error) {
	s := NewSnipScanner(r)
	for s.Scan() {

		snipp := s.Snippet()
		return &snipp, nil
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("No snippets found\n")

}

func ChoseSnippet() (string, error) {
	snippetsDir := "/home/aura/.dotfiles/snippets/"
	files, err := filepath.Glob(filepath.Join(snippetsDir, "*.md"))
	if err != nil {
		return "", err
	}

	fileNames := make([]map[string]string, len(files))
	for i, file := range files {
		fileNames[i] = map[string]string{
			filepath.Base(file): file,
		}
	}
	choice, err := terminal.RunTerm(fileNames)

	if err != nil {
		return "", err
	}
	for _, v := range choice {

		return v, nil
	}
	return "", fmt.Errorf("This shouldHave happend")

}
