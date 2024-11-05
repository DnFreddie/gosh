package installer

import (
	"fmt"
	"strings"
)

// Test the [getName] function, which returns the name of the snippet
func Example_getName() {
	line := "/*##NAME:example.go##*/"
	fileName := getName(line)
	fmt.Printf("%s", fileName)
	// Output:
	// example.go
}

// Test the [getCodeFormat] function, which returns the format of the snippet
func Example_getCodeFormat() {
	line := "```bash"
	codeFormat := getCodeFormat(line)
	fmt.Printf("%s", codeFormat)
	// Output:
	// bash
}

// Tests the [GetSnippet] which returns a [Snippet]
func Example_GetSnippet() {

	test := " ```bash\n##NAME:TEST.t##\nthis\nis\ntest\n```\n"

	reader := strings.NewReader(test)
	snippet, err := GetSnippet(reader)
	if err != nil {
		fmt.Printf("Error getting snippet: %v\n", err)
		return
	}

	fmt.Printf("Snippet Name: %s\n", snippet.Name)
	fmt.Printf("Snippet Language: %s\n", snippet.Lang)
	fmt.Printf("Snippet Content: %s\n", snippet.Content.String())

	// Output:
	// Snippet Name: TEST.t
	// Snippet Language: bash
	// Snippet Content: this
	// is
	// test
}
