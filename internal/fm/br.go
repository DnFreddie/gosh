package fm

import (
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
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			return cmd.Usage()
		}
		input, err := io.ReadAll(os.Stdin)

		if err != nil {
			return err
		}
		busybox.ProcessLine(args[0], string(input))
		return nil
	},
}

func init() {
}
