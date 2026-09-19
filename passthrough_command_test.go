package kong_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestPassthroughPartialCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		flag    string
		cmdArgs []string
		err     string
	}{
		{"InheritedFlag", []string{"command", "--flag", "value", "program"}, "value", []string{"program"}, ""},
		{"InheritedShortFlag", []string{"command", "-f", "value", "program"}, "value", []string{"program"}, ""},
		{"InheritedEqualsFlag", []string{"command", "--flag=value", "program"}, "value", []string{"program"}, ""},
		{"FlagBeforeCommand", []string{"--flag", "value", "command", "program"}, "value", []string{"program"}, ""},
		{"FlagsAfterPositional", []string{"command", "program", "--flag", "value", "--unknown", "-x"}, "", []string{"program", "--flag", "value", "--unknown", "-x"}, ""},
		{"FlagBeforeAndAfterPositional", []string{"command", "--flag", "parsed", "program", "--flag", "forwarded"}, "parsed", []string{"program", "--flag", "forwarded"}, ""},
		{"UnknownLongFlag", []string{"command", "--unknown", "program"}, "", nil, "unknown flag --unknown"},
		{"UnknownShortFlag", []string{"command", "-x", "program"}, "", nil, "unknown flag -x"},
		{"DashDash", []string{"command", "--", "--unknown", "program"}, "", []string{"--", "--unknown", "program"}, ""},
		{"DashDashAfterPositional", []string{"command", "program", "--", "--flag"}, "", []string{"program", "--", "--flag"}, ""},
		{"Dash", []string{"command", "-", "--flag", "value"}, "", []string{"-", "--flag", "value"}, ""},
		{"Empty", []string{"command"}, "", nil, ""},
		{"OnlyFlag", []string{"command", "--flag", "value"}, "value", nil, ""},
		{"Alias", []string{"exec", "--flag", "value", "program"}, "value", []string{"program"}, ""},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var cli struct {
				Flag    string `short:"f"`
				Command struct {
					Args []string `arg:"" optional:""`
				} `cmd:"" aliases:"exec" passthrough:"partial"`
			}
			parser := mustNew(t, &cli)
			_, err := parser.Parse(test.args)
			if test.err != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.flag, cli.Flag)
			assert.Equal(t, test.cmdArgs, cli.Command.Args)
		})
	}
}

func TestPassthroughAllCommand(t *testing.T) {
	var cli struct {
		Flag    string
		Command struct {
			Args []string `arg:"" optional:""`
		} `cmd:"" passthrough:"all"`
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"--flag", "parent", "command", "--flag", "child", "--unknown"})
	assert.NoError(t, err)
	assert.Equal(t, "parent", cli.Flag)
	assert.Equal(t, []string{"--flag", "child", "--unknown"}, cli.Command.Args)
}
