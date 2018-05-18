package kong

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func mustNew(t *testing.T, cli interface{}) *Kong {
	t.Helper()
	parser, err := New(cli)
	require.NoError(t, err)
	return parser
}

func TestPositionalArguments(t *testing.T) {
	var cli struct {
		User struct {
			Create struct {
				ID    int    `arg:""`
				First string `arg:""`
				Last  string `arg:""`
			} `cmd:""`
		} `cmd:""`
	}
	p := mustNew(t, &cli)
	cmd, err := p.Parse([]string{"user", "create", "10", "Alec", "Thomas"})
	require.NoError(t, err)
	require.Equal(t, "user create <id> <first> <last>", cmd)
	t.Run("Missing", func(t *testing.T) {
		_, err := p.Parse([]string{"user", "create", "10"})
		require.Error(t, err)
	})
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
				ID    string `arg:""`
				First string `arg:""`
				Last  string `arg:""`
			} `cmd:""`

			// Branching argument.
			ID struct {
				ID     int `arg:""`
				Flag   int
				Delete struct{} `cmd:""`
				Rename struct {
					To string
				} `cmd:""`
			} `arg:""`
		} `cmd:""  help:"User management."`
	}
	p := mustNew(t, &cli)
	cmd, err := p.Parse([]string{"user", "10", "delete"})
	require.NoError(t, err)
	require.Equal(t, 10, cli.User.ID.ID)
	require.Equal(t, "user <id> delete", cmd)
	t.Run("Missing", func(t *testing.T) {
		_, err = p.Parse([]string{"user"})
		require.Error(t, err)
	})
}

func TestResetWithDefaults(t *testing.T) {
	var cli struct {
		Flag            string
		FlagWithDefault string `default:"default" `
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
		Slice []int
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"--slice=1,2,3"})
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, cli.Slice)
}

func TestArgSlice(t *testing.T) {
	var cli struct {
		Slice []int `arg:""`
		Flag  bool
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"1", "2", "3", "--flag"})
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, cli.Slice)
	require.Equal(t, true, cli.Flag)
}

func TestUnsupportedFieldErrors(t *testing.T) {
	var cli struct {
		Keys map[string]string
	}
	_, err := New(&cli)
	require.Error(t, err)
}

func TestMatchingArgField(t *testing.T) {
	var cli struct {
		ID struct {
			NotID int `arg:""`
		} `arg:""`
	}

	_, err := New(&cli)
	require.Error(t, err)
}

func TestCantMixPositionalAndBranches(t *testing.T) {
	var cli struct {
		Arg     string `arg:""`
		Command struct {
		} `cmd:""`
	}
	_, err := New(&cli)
	require.Error(t, err)
}
