package fm

import (
	"fmt"
	"io"
	"os"

	"github.com/DnFreddie/gosh/pkg/busybox"
	"github.com/spf13/cobra"
)

// brCmd represents the br command
var Br = &cobra.Command{
	Use:   "br [bracket-type]",
	Short: "Wrap text in brackets",
	Long: `Wraps input text in specified brackets or delimiters.
	
Available bracket types:
  ( or ) or b  - Wraps in parentheses
  { or } or sb - Wraps in curly braces
  [ or ]       - Wraps in square brackets
  ' or q       - Wraps in single quotes
  s            - Wraps in double quotes
  #            - Adds hash prefix
  c            - Wraps in markdown code block
  
Examples:
  echo "hello" | fm br (    -> (hello)
  echo "code" | fm br c     -> bash code
  echo "test" | fm br s     -> "test"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input, err := io.ReadAll(os.Stdin)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		busybox.ProcessLine(args[0], string(input))
	},
}

func init() {
}
