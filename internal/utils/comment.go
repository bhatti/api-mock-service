package utils

import (
	"bufio"
	"bytes"
	"strings"
)

// ToYAMLComment converts a multiline string to YAML comment style
// by prefixing each line with "# ".
// Empty lines will have just "#" without the space.
func ToYAMLComment(input string) string {
	// If input is empty, return empty string
	if len(strings.TrimSpace(input)) == 0 {
		return ""
	}

	var buf bytes.Buffer
	scanner := bufio.NewScanner(strings.NewReader(input))
	first := true

	for scanner.Scan() {
		line := scanner.Text()

		// Add newline before all lines except the first
		if !first {
			buf.WriteString("\n")
		}
		first = false

		// Handle empty lines
		if len(strings.TrimSpace(line)) == 0 {
			buf.WriteString("#")
			continue
		}

		// Add "# " prefix to non-empty lines
		buf.WriteString("# ")
		buf.WriteString(line)
	}

	return buf.String()
}

// FromYAMLComment converts YAML comments back to regular multiline text
// by removing the "# " prefix from each line.
func FromYAMLComment(input string) string {
	if len(strings.TrimSpace(input)) == 0 {
		return ""
	}

	var buf bytes.Buffer
	scanner := bufio.NewScanner(strings.NewReader(input))
	first := true

	for scanner.Scan() {
		line := scanner.Text()

		// Add newline before all lines except the first
		if !first {
			buf.WriteString("\n")
		}
		first = false

		// Remove "# " or "#" prefix
		line = strings.TrimPrefix(line, "# ")
		line = strings.TrimPrefix(line, "#")

		buf.WriteString(line)
	}

	return buf.String()
}
