package busybox

import (
	"io"
	"os"
	"path/filepath"

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
)

type language string

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

type languageInfo struct {
	Name      string
	Extension string
}

func findLanguage(fileName string) string {
	ext := filepath.Ext(fileName)
	if lang, exists := languages[ext]; exists {
		return lang
	}
	return "unknown"
}

type Highlighter struct {
	language   string
	fileName   string
	fileReader io.Reader
}

func NewHighlighter(fileName string, r io.Reader) *Highlighter {
	return &Highlighter{
		fileName:   fileName,
		fileReader: r,
	}

}
func (h *Highlighter) Highlight() error {
	lang := findLanguage(h.fileName)
	content, err := io.ReadAll(h.fileReader)
	if err != nil {
		return err
	}
	return highlightCode(lang, string(content))
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
