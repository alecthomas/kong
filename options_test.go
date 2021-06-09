package kong

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	var cli struct{}
	p, err := New(&cli, Name("name"), Description("description"), Writers(nil, nil), Exit(nil))
	require.NoError(t, err)
	require.Equal(t, "name", p.Model.Name)
	require.Equal(t, "description", p.Model.Help)
	require.Nil(t, p.Stdout)
	require.Nil(t, p.Stderr)
	require.Nil(t, p.Exit)
}

type impl string

func (impl) Method() {}

func TestBindTo(t *testing.T) {
	type iface interface {
		Method()
	}

	saw := ""
	method := func(i iface) error {
		saw = string(i.(impl))
		return nil
	}

	var cli struct{}

	p, err := New(&cli, BindTo(impl("foo"), (*iface)(nil)))
	require.NoError(t, err)
	err = callMethod("method", reflect.ValueOf(impl("??")), reflect.ValueOf(method), p.bindings)
	require.NoError(t, err)
	require.Equal(t, "foo", saw)
}

type bindToProviderCLI struct {
	Called bool
	Cmd    bindToProviderCmd `cmd:""`
}

type boundThing struct {
}

type bindToProviderCmd struct{}

func (*bindToProviderCmd) Run(cli *bindToProviderCLI, b *boundThing) error {
	cli.Called = true
	return nil
}

func TestBindToProvider(t *testing.T) {
	var cli bindToProviderCLI
	app, err := New(&cli, BindToProvider(func() (*boundThing, error) { return &boundThing{}, nil }))
	require.NoError(t, err)
	ctx, err := app.Parse([]string{"cmd"})
	require.NoError(t, err)
	err = ctx.Run()
	require.NoError(t, err)
	require.True(t, cli.Called)
}
