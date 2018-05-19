package kong

import (
	"testing"

	"fmt"
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

func TestPropagatedFlags(t *testing.T) {
	var cli struct {
		Flag1    string
		Command1 struct {
			Flag2    bool
			Command2 struct{} `cmd:""`
		} `cmd:""`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"command-1", "command-2", "--flag-2", "--flag-1=moo"})
	require.NoError(t, err)
	require.Equal(t, "moo", cli.Flag1)
	require.Equal(t, true, cli.Command1.Flag2)
}

func TestRequiredFlag(t *testing.T) {
	var cli struct {
		Flag string `required:""`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	fmt.Println(err)
	require.Error(t, err)
}

func TestOptionalArg(t *testing.T) {
	var cli struct {
		Arg string `arg:"" optional:""`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	require.NoError(t, err)
}

func TestRequiredArg(t *testing.T) {
	var cli struct {
		Arg string `arg:""`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	require.Error(t, err)
}

func TestInvalidRequiredAfterOptional(t *testing.T) {
	var cli struct {
		ID   int    `arg:"" optional:""`
		Name string `arg:""`
	}

	_, err := New(&cli)
	require.Error(t, err)
}

func TestOptionalStructArg(t *testing.T) {
	var cli struct {
		Name struct {
			Name    string `arg:"" optional:""`
			Enabled bool
		} `arg:"" optional:""`
	}

	parser := mustNew(t, &cli)

	t.Run("WithFlag", func(t *testing.T) {
		_, err := parser.Parse([]string{"gak", "--enabled"})
		require.NoError(t, err)
		require.Equal(t, "gak", cli.Name.Name)
		require.Equal(t, true, cli.Name.Enabled)
	})

	t.Run("WithoutFlag", func(t *testing.T) {
		_, err := parser.Parse([]string{"gak"})
		require.NoError(t, err)
		require.Equal(t, "gak", cli.Name.Name)
	})

	t.Run("WithNothing", func(t *testing.T) {
		_, err := parser.Parse([]string{})
		require.NoError(t, err)
	})
}

func TestMixedRequiredArgs(t *testing.T) {
	var cli struct {
		Name string `arg:""`
		ID   int    `arg:"" optional:""`
	}

	parser := mustNew(t, &cli)

	t.Run("SingleRequired", func(t *testing.T) {
		_, err := parser.Parse([]string{"gak", "5"})
		require.NoError(t, err)
		require.Equal(t, "gak", cli.Name)
		require.Equal(t, 5, cli.ID)
	})

	t.Run("ExtraOptional", func(t *testing.T) {
		_, err := parser.Parse([]string{"gak"})
		require.NoError(t, err)
		require.Equal(t, "gak", cli.Name)
	})
}
