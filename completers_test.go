package kong

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestPositionalCompleter_position(t *testing.T) {
	posCompleter := &positionalCompleter{
		Flags: []*Flag{
			{
				Value: &Value{
					Name:   "mybool",
					Mapper: boolMapper{},
				},
				Short: 'b',
			},
			{
				Value: &Value{
					Name:   "mybool2",
					Mapper: boolMapper{},
				},
				Short: 'c',
			},
			{
				Value: &Value{
					Name: "myarg",
				},
				Short: 'a',
			},
		},
	}

	for args, want := range map[string]int{
		``:                 0,
		`foo`:              0,
		`foo `:             1,
		`-b foo `:          1,
		`-bc foo `:         1,
		`-bd foo `:         1,
		`-a foo `:          0,
		`-a=omg foo `:      1,
		`--myarg omg foo `: 1,
		`--myarg=omg foo `: 1,
		`foo bar`:          1,
		`foo bar `:         2,
	} {
		args := args
		want := want
		t.Run(args, func(t *testing.T) {
			got := posCompleter.completerIndex(newArgs("foo " + args))
			assert.Equal(t, want, got)
		})
	}
}

func TestPositionalCompleter_Predict(t *testing.T) {
	completer1 := CompleteSet("1")
	completer2 := CompleteSet("2")
	posCompleter := &positionalCompleter{
		Completers: []Completer{completer1, completer2},
	}

	for args, want := range map[string][]string{
		``:         {"1"},
		`foo`:      {"1"},
		`foo `:     {"2"},
		`foo bar`:  {"2"},
		`foo bar `: {},
	} {
		args := args
		want := want
		t.Run(args, func(t *testing.T) {
			got := posCompleter.Predict(newArgs("app " + args))

			assert.Equal(t, want, got)
		})
	}
}

// The code below is taken from https://github.com/posener/complete/blob/f6dd29e97e24f8cb51a8d4050781ce2b238776a4/args.go
// to assist in tests.

func newArgs(line string) CompleterArgs {
	var (
		all       []string
		completed []string
	)
	parts := splitFields(line)
	if len(parts) > 0 {
		all = parts[1:]
		completed = removeLast(parts[1:])
	}
	return CompleterArgs{
		All:           all,
		Completed:     completed,
		Last:          last(parts),
		LastCompleted: last(completed),
	}
}

// splitFields returns a list of fields from the given command line.
// If the last character is space, it appends an empty field in the end
// indicating that the field before it was completed.
// If the last field is of the form "a=b", it splits it to two fields: "a", "b",
// So it can be completed.
func splitFields(line string) []string {
	parts := strings.Fields(line)

	// Add empty field if the last field was completed.
	if len(line) > 0 && unicode.IsSpace(rune(line[len(line)-1])) {
		parts = append(parts, "")
	}

	// Treat the last field if it is of the form "a=b"
	parts = splitLastEqual(parts)
	return parts
}

func splitLastEqual(line []string) []string {
	if len(line) == 0 {
		return line
	}
	parts := strings.Split(line[len(line)-1], "=")
	return append(line[:len(line)-1], parts...)
}

func removeLast(a []string) []string {
	if len(a) > 0 {
		return a[:len(a)-1]
	}
	return a
}

func last(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[len(args)-1]
}
