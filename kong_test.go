package kong

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func mustNew(t *testing.T, cli interface{}) *Kong {
	t.Helper()
	parser, err := New("", "", cli)
	require.NoError(t, err)
	return parser
}

func TestArgumentSequence(t *testing.T) {
	var cli struct {
		User struct {
			Create struct {
				ID    int    `arg:"" help:""`
				First string `arg:"" help:""`
				Last  string `arg:"" help:""`
			} `help:""`
		} `help:""`
	}
	p := mustNew(t, &cli)
	cmd, err := p.Parse([]string{"user", "create", "10", "Alec", "Thomas"})
	require.NoError(t, err)
	require.Equal(t, "user create <id> <first> <last>", cmd)
}

func TestBranchingArgument(t *testing.T) {
	/*
		app user create <id> <first> <last>
		app	user <id> delete
		app	user <id> rename <to>

	*/
	var cli struct {
		User struct {
			Create struct {
				ID    string `arg:"" help:""`
				First string `arg:"" help:""`
				Last  string `arg:"" help:""`
			} `help:""`

			// Branching argument.
			ID struct {
				ID     int      `arg:"" help:""`
				Flag   int      `help:""`
				Delete struct{} `help:""`
				Rename struct {
					To string `help:""`
				} `help:""`
			} `arg:"" help:""`
		} `help:"Manage users."`
	}
	p := mustNew(t, &cli)
	cmd, err := p.Parse([]string{"user", "10", "delete"})
	require.NoError(t, err)
	require.Equal(t, 10, cli.User.ID.ID)
	require.Equal(t, "user <id> delete", cmd)
}

func TestResetWithDefaults(t *testing.T) {
	var cli struct {
		Flag            string `help:""`
		FlagWithDefault string `default:"default" help:""`
	}
	cli.Flag = "BLAH"
	cli.FlagWithDefault = "BLAH"
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "", cli.Flag)
	require.Equal(t, "default", cli.FlagWithDefault)
}

func TestFlagSlice(t *testing.T) {
	var cli struct {
		Slice []int `help:""`
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"--slice=1,2,3"})
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, cli.Slice)
}

func TestArgSlice(t *testing.T) {
	var cli struct {
		Slice []int `help:"" arg:""`
		Flag  bool  `help:""`
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"1", "2", "3", "--flag"})
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, cli.Slice)
	require.Equal(t, true, cli.Flag)
}

func TestUnsupportedfieldErrors(t *testing.T) {
	var cli struct {
		Keys map[string]string `help:""`
	}
	require.Panics(t, func() { mustNew(t, &cli) })
}
