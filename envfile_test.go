package kong_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/alecthomas/kong"
)

func TestParseENVFile(t *testing.T) {
	tests := []struct {
		name                 string
		input                string
		expected             map[string]string
		expectedErrorMessage string
	}{
		{
			name:                 "Empty input",
			input:                "",
			expected:             map[string]string{},
			expectedErrorMessage: "",
		},
		{
			name:  "Single simple string",
			input: "OPTION_A=Text 1",
			expected: map[string]string{
				"OPTION_A": "Text 1",
			},
			expectedErrorMessage: "",
		},
		{
			name:  "Single string with export clause",
			input: "export OPTION_A=Text 1",
			expected: map[string]string{
				"OPTION_A": "Text 1",
			},
			expectedErrorMessage: "",
		},
		{
			name: "Multiple strings",
			input: "OPTION_A=Text 1\n" +
				"OPTION_B=Some other text",
			expected: map[string]string{
				"OPTION_A": "Text 1",
				"OPTION_B": "Some other text",
			},
			expectedErrorMessage: "",
		},
		{
			name: "Multiple strings with commented lines",
			input: "OPTION_A=Text 1\n" +
				"# Comment line\n" +
				"OPTION_B=Some other text",
			expected: map[string]string{
				"OPTION_A": "Text 1",
				"OPTION_B": "Some other text",
			},
			expectedErrorMessage: "",
		},
		{
			name: "Multiple strings with a comment in the end of a line",
			input: "OPTION_A=Text 1\n" +
				"OPTION_B=Some other text # commented segment",
			expected: map[string]string{
				"OPTION_A": "Text 1",
				"OPTION_B": "Some other text",
			},
			expectedErrorMessage: "",
		},
		{
			name: "Multiple strings with double quotes",
			input: "OPTION_A=Text 1\n" +
				"OPTION_B=\"Some other text # not comment anymore\"",
			expected: map[string]string{
				"OPTION_A": "Text 1",
				"OPTION_B": "Some other text # not comment anymore",
			},
			expectedErrorMessage: "",
		},
		{
			name: "Multiple strings with single quotes",
			input: "OPTION_A=Text 1\n" +
				"OPTION_B='Some other text # not comment anymore'",
			expected: map[string]string{
				"OPTION_A": "Text 1",
				"OPTION_B": "Some other text # not comment anymore",
			},
			expectedErrorMessage: "",
		},
		{
			name: "Multiple strings with substitution",
			input: "OPTION_A=Text 1\n" +
				"OPTION_B=\"$OPTION_A -> Some other text # not comment anymore\"",
			expected: map[string]string{
				"OPTION_A": "Text 1",
				"OPTION_B": "Text 1 -> Some other text # not comment anymore",
			},
			expectedErrorMessage: "",
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			actual, err := kong.ParseENVFile(reader)
			switch tt.expectedErrorMessage {
			case "":
				require.NoError(t, err)
				require.EqualValues(t, tt.expected, actual)
			default:
				require.Error(t, err, tt.expectedErrorMessage)
				require.Empty(t, actual)
			}
		})
	}
}
