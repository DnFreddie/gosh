package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DnFreddie/gosh/pkg/installer"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [owner/repo...]",
	Short: "Install GitHub released binaries",
	Long: `Install downloads and installs the latest released binaries from GitHub repositories.
    
Example usage:
  gosh install mikefarah/yq DnFreddie/gosh
  gosh install --target ~/.local/bin mikefarah/yq
  gosh install cli/cli:gh `,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println("at least one repository must be specified")
			return nil
		}

		targetDir, err := cmd.Flags().GetString("target")
		if err != nil {
			fmt.Println("error getting target directory: %w", err)
			return nil
		}

		tempDir, err := cmd.Flags().GetString("temp")
		if err != nil {
			fmt.Println("error getting temp directory: %w", err)
			return nil
		}

		config := installer.Config{
			TargetDir: targetDir,
			TempDir:   tempDir,
		}

		inst, err := installer.NewInstaller(config, args)
		if err != nil {
			fmt.Println("failed to create installer: %w", err)
			return nil
		}

		if err := inst.Install(); err != nil {
			fmt.Println("installation failed: %w", err)
			return nil
		}

		fmt.Println("Installation completed successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not determine home directory: %v\n", err)
		return
	}

	defaultTargetDir := filepath.Join(homeDir, ".local", "bin")
	defaultTempDir := os.TempDir()

	installCmd.Flags().StringP("target", "t", defaultTargetDir, "Target directory for installed binaries")
	installCmd.Flags().String("temp", defaultTempDir, "Temporary directory for downloads")
}
