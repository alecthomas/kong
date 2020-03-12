package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			got := posCompleter.Options(newCompleterArgs("app " + args))

			assert.Equal(t, want, got)
		})
	}
}

func TestPositionalCompleter_completerIndex(t *testing.T) {
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
			got := posCompleter.completerIndex(newCompleterArgs("foo " + args))
			assert.Equal(t, want, got)
		})
	}
}
