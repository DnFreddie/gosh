package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DnFreddie/gosh/pkg/installer"
	"github.com/spf13/cobra"
)

// completionCmd handles shell completion generation
var completionCmd = &cobra.Command{
	Use:   "compl [command-name]",
	Short: "Generate shell completions for a command",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell, err := cmd.Parent().Flags().GetString("shell")
		if err != nil {
			return fmt.Errorf("error getting shell type: %w", err)
		}

		completionDir, err := cmd.Parent().Flags().GetString("completion-dir")
		if err != nil {
			return fmt.Errorf("error getting completion directory: %w", err)
		}

		completer := installer.NewCompleter(args[0]).
			SetShell(shell).
			SetOutputDir(completionDir)

		if err := completer.Save(); err != nil {
			return fmt.Errorf("failed to generate completion: %w", err)
		}

		fmt.Printf("Generated %s completion for %s\n", shell, args[0])
		return nil
	},
}

// installCmd handles GitHub binary installation
var installCmd = &cobra.Command{
	Use:   "install [owner/repo...]",
	Short: "Install GitHub released binaries",
	Long: `Install downloads and installs the latest released binaries from GitHub repositories.
    
Example usage:
  gosh install mikefarah/yq DnFreddie/gosh
  gosh install --target ~/.local/bin mikefarah/yq
  gosh install cli/cli:gh
  gosh install --toolbox`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, err := cmd.Flags().GetString("target")
		if err != nil {
			return fmt.Errorf("error getting target directory: %w", err)
		}

		tempDir, err := cmd.Flags().GetString("temp")
		if err != nil {
			return fmt.Errorf("error getting temp directory: %w", err)
		}

		toolbox, err := cmd.Flags().GetBool("toolbox")
		if err != nil {
			return fmt.Errorf("error getting toolbox flag: %w", err)
		}

		if !toolbox && len(args) == 0 {
			return fmt.Errorf("at least one repository must be specified")
		}

		config := installer.Config{
			TargetDir: targetDir,
			TempDir:   tempDir,
		}

		repos := args
		if toolbox {
			repos = installer.TOOLBOX
		}

		inst, err := installer.NewInstaller(config, repos)
		if err != nil {
			return fmt.Errorf("failed to create installer: %w", err)
		}

		if err := inst.Install(); err != nil {
			return fmt.Errorf("installation failed: %w", err)
		}

		fmt.Println("Installation completed successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(completionCmd)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not determine home directory: %v\n", err)
		return
	}

	defaultTargetDir := filepath.Join(homeDir, ".local", "bin")
	defaultTempDir := os.TempDir()
	defaultCompletionDir := filepath.Join(homeDir, ".local", "share", "completions")

	installCmd.Flags().StringP("target", "t", defaultTargetDir, "Target directory for installed binaries")
	installCmd.Flags().String("temp", defaultTempDir, "Temporary directory for downloads")
	installCmd.Flags().Bool("toolbox", false, "Whether to download an entire toolbox")
	installCmd.Flags().String("shell", "bash", "Shell type for completion generation (bash, zsh, fish)")
	installCmd.Flags().String("completion-dir", defaultCompletionDir, "Directory for storing completion files")
}
