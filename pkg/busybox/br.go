package busybox

import (
	"fmt"
	"strings"
)

func ProcessLine(bracket, line string) {
	leadingSpaces := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " ")))
	trailingSpaces := strings.Repeat(" ", len(line)-len(strings.TrimRight(line, " ")))
	trimmedLine := strings.TrimSpace(line)

	var formatted string
	switch bracket {
	case "(", ")", "b":
		formatted = fmt.Sprintf("(%s)", trimmedLine)
	case "{", "}", "sb":
		formatted = fmt.Sprintf("{%s}", trimmedLine)
	case "[", "]":
		formatted = fmt.Sprintf("[%s]", trimmedLine)
	case "'", "q":
		formatted = fmt.Sprintf("'%s'", trimmedLine)
	case "#":
		formatted = fmt.Sprintf("#%s", trimmedLine)
	case "s":
		formatted = fmt.Sprintf("\"%s\"", trimmedLine)
	case "c":
		formatted = fmt.Sprintf("```bash\n%s\n```", line)
	case "a":
		formatted = fmt.Sprintf("\"{{ %s }}\"", line)
	default:
		formatted = fmt.Sprintf("%s%s%s", bracket, trimmedLine, bracket)
	}

	fmt.Printf("%s%s%s", leadingSpaces, formatted, trailingSpaces)
}
