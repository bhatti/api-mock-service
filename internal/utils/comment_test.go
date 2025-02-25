package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ToYAMLComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "hello world",
			expected: "# hello world",
		},
		{
			name:     "multiple lines",
			input:    "hello\nworld\ntest",
			expected: "# hello\n# world\n# test",
		},
		{
			name:     "empty lines",
			input:    "hello\n\nworld",
			expected: "# hello\n#\n# world",
		},
		{
			name:     "trailing newline",
			input:    "hello\nworld\n",
			expected: "# hello\n# world",
		},
		{
			name:     "spaces and tabs",
			input:    "  hello  \n\tworld  ",
			expected: "#   hello  \n# \tworld  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToYAMLComment(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_FromYAMLComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "# hello world",
			expected: "hello world",
		},
		{
			name:     "multiple lines",
			input:    "# hello\n# world\n# test",
			expected: "hello\nworld\ntest",
		},
		{
			name:     "empty lines",
			input:    "# hello\n#\n# world",
			expected: "hello\n\nworld",
		},
		{
			name:     "spaces and tabs preserved",
			input:    "#   hello  \n# \tworld  ",
			expected: "  hello  \n\tworld  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromYAMLComment(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_RoundTrip(t *testing.T) {
	input := "This is a test\nWith multiple lines\n\nAnd empty lines\n  Indented lines\n\tTabbed lines"
	result := FromYAMLComment(ToYAMLComment(input))
	require.Equal(t, input, result)
}
