package installer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rogpeppe/go-internal/lockedfile"
)

type Completer struct {
	command     string
	shellType   string
	outputDir   string
	completions [][]string
}

func NewCompleter(command string) *Completer {
	return &Completer{
		command:   command,
		shellType: "bash",
		outputDir: os.ExpandEnv("${HOME}/.local/share/completions"),
		completions: [][]string{
			{"completion"},
			{"completion", "-s"},
		},
	}
}

func (c *Completer) SetShell(shell string) *Completer {
	if len(shell) > 0 {
		c.shellType = shell
	}
	return c
}

func (c *Completer) SetOutputDir(dir string) *Completer {
	if len(dir) > 0 {
		c.outputDir = os.ExpandEnv(dir)
	}
	return c
}

func (c *Completer) Generate() (string, error) {
	var lastErr error

	for _, baseArgs := range c.completions {
		args := append(baseArgs, c.shellType)
		cmd := exec.Command(c.command, args...)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			lastErr = fmt.Errorf("command failed: %w\nstderr: %s", err, stderr.String())
			continue
		}

		return stdout.String(), nil
	}

	return "", fmt.Errorf("all completion attempts failed: %w", lastErr)
}

func (c *Completer) Save() error {

	if err := os.MkdirAll(c.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	content, err := c.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate completion: %w", err)
	}

	filename := filepath.Join(c.outputDir, fmt.Sprintf("%s.%s", c.command, c.shellType))
	reader := bytes.NewReader([]byte(content))

	return lockedfile.Write(filename, reader, 0644)
}
