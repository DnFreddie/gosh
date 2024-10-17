package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/DnFreddie/gosh/scripts"
)

func RunScript(name scripts.COMMAND, flag ...string) error {

	scriptPath := string(name)
	s, err := scripts.Scripts.Open(scriptPath)
	if err != nil {
		return fmt.Errorf("error opening script file: %w", err)
	}
	defer s.Close()

	script, err := io.ReadAll(s)
	if err != nil {
		return fmt.Errorf("error reading script file: %w", err)
	}

	c := exec.Command("bash")

	c.Stdin = strings.NewReader(string(script))

	if len(flag) > 0 {
		c = exec.Command("bash", append([]string{"-s", "-"}, flag...)...)
		c.Stdin = strings.NewReader(string(script))
	}

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("error executing script")
	}

	return nil
}
