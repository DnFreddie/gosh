package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/DnFreddie/gosh/scripts"
)

func Edit(toEdit string) {

	var editor = "vi"
	if _, err := exec.LookPath("vim"); err == nil {
		editor = "vim"
	}

	vimrcPath := filepath.Join(os.TempDir(), "embedded_vimrc")
	if _, err := os.Stat(vimrcPath); os.IsNotExist(err) {
		vimrcContent, err := scripts.Embeded.ReadFile("vimrc")
		if err != nil {
			fmt.Printf("Failed to read vimrc file: %v\n", err)
			return
		}
		err = os.WriteFile(vimrcPath, vimrcContent, 0644)
		if err != nil {
			fmt.Printf("Failed to write vimrc file: %v\n", err)
			return
		}
	}

	var cmd *exec.Cmd

	switch editor {
	case "vim":
		cmd = exec.Command(editor, "-N", "-u", vimrcPath, toEdit)
	case "vi":
		cmd = exec.Command(editor, "-u", vimrcPath, toEdit)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to run %s: %v\n", editor, err)
	}
}
