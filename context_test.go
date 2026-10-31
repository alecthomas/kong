package kong_test

import (
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/kong"
)

func TestProvenanceCommandLineFlags(t *testing.T) {
	var cli struct {
		Flag    string `name:"flag" aliases:"falias"`
		Short   string `short:"s"`
		Bool    bool
		NegBool bool   `negatable:""`
		Counter int    `type:"counter" short:"c"`
		Arg     string `arg:""`
	}
	p := mustNew(t, &cli)
	ctx, err := p.Parse([]string{"--flag=foo", "-s", "bar", "--bool", "--no-neg-bool", "-c", "pos"})
	assert.NoError(t, err)

	for _, name := range []string{"flag", "s", "bool", "neg-bool", "counter"} {
		f := ctx.Flag(name)
		assert.True(t, f != nil, "flag %s should be found", name)
		src := ctx.FlagSource(f)
		assert.Equal(t, kong.ValueSourceFlag, src.Type)
		assert.Equal(t, kong.ValueSourceFlag, f.Source)
		assert.Equal(t, kong.ValueSourceFlag, ctx.FlagValueSource(f))
		assert.True(t, src.IsFlag())
		assert.False(t, src.IsEnv())
		assert.False(t, src.IsResolver())
		assert.False(t, src.IsDefault())
		assert.False(t, src.IsUnset())
		assert.Equal(t, "flag", src.String())
		assert.Equal(t, src, f.SourceInfo())
	}
}

func TestProvenanceAliases(t *testing.T) {
	var cli struct {
		Flag string `aliases:"alias-flag,a"`
	}
	p := mustNew(t, &cli)
	ctx, err := p.Parse([]string{"--alias-flag=value"})
	assert.NoError(t, err)
	f := ctx.Flag("flag")
	assert.Equal(t, kong.ValueSourceFlag, ctx.FlagValueSource(f))
	assert.Equal(t, "value", cli.Flag)

	ctx2, err := p.Parse([]string{"-a", "short-alias-value"})
	assert.NoError(t, err)
	f2 := ctx2.Flag("flag")
	assert.Equal(t, kong.ValueSourceFlag, ctx2.FlagValueSource(f2))
	assert.Equal(t, "short-alias-value", cli.Flag)
}

func TestProvenanceEnvironment(t *testing.T) {
	var cli struct {
		Single   string `env:"KONG_PROV_SINGLE"`
		Multiple string `env:"KONG_PROV_UNSET,KONG_PROV_SET"`
	}
	t.Setenv("KONG_PROV_SINGLE", "single_val")
	t.Setenv("KONG_PROV_SET", "multiple_val")

	p := mustNew(t, &cli)
	ctx, err := p.Parse(nil)
	assert.NoError(t, err)

	f1 := ctx.Flag("single")
	src1 := ctx.FlagSource(f1)
	assert.Equal(t, kong.ValueSourceEnv, src1.Type)
	assert.Equal(t, "KONG_PROV_SINGLE", src1.Env)
	assert.Equal(t, "KONG_PROV_SINGLE", f1.SourceEnv)
	assert.True(t, src1.IsEnv())
	assert.Equal(t, "env:KONG_PROV_SINGLE", src1.String())

	f2 := ctx.Flag("multiple")
	src2 := ctx.FlagSource(f2)
	assert.Equal(t, kong.ValueSourceEnv, src2.Type)
	assert.Equal(t, "KONG_PROV_SET", src2.Env)
	assert.Equal(t, "KONG_PROV_SET", f2.SourceEnv)
	assert.Equal(t, "env:KONG_PROV_SET", src2.String())
}

func TestProvenanceResolver(t *testing.T) {
	var cli struct {
		JsonFlag    string
		NamedFlag   string
		FuncFlag    string
		WrappedFlag string
	}
	jsonResolver, err := kong.JSON(strings.NewReader(`{"json_flag": "from_json"}`))
	assert.NoError(t, err)

	namedResolver := kong.NamedResolverFunc("custom-db", func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (any, error) {
		if flag.Name == "named-flag" {
			return "from_db", nil
		}
		return nil, nil
	})

	var unnamed kong.ResolverFunc = func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (any, error) {
		if flag.Name == "func-flag" {
			return "from_func", nil
		}
		return nil, nil
	}

	var innerFunc kong.ResolverFunc = func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (any, error) {
		if flag.Name == "wrapped-flag" {
			return "from_wrapped", nil
		}
		return nil, nil
	}
	wrapped := kong.WithResolverName("wrapped-source", innerFunc)

	p := mustNew(t, &cli, kong.Resolvers(jsonResolver, namedResolver, unnamed, wrapped))
	ctx, err := p.Parse(nil)
	assert.NoError(t, err)

	fJson := ctx.Flag("json-flag")
	srcJson := ctx.FlagSource(fJson)
	assert.Equal(t, kong.ValueSourceResolver, srcJson.Type)
	assert.Equal(t, "json", srcJson.ResolverName)
	assert.Equal(t, "json", fJson.SourceResolverName)
	assert.True(t, srcJson.IsResolver())
	assert.Equal(t, "resolver:json", srcJson.String())
	assert.Equal(t, jsonResolver, srcJson.Resolver)

	fNamed := ctx.Flag("named-flag")
	srcNamed := ctx.FlagSource(fNamed)
	assert.Equal(t, kong.ValueSourceResolver, srcNamed.Type)
	assert.Equal(t, "custom-db", srcNamed.ResolverName)
	assert.Equal(t, "custom-db", fNamed.SourceResolverName)
	assert.Equal(t, "resolver:custom-db", srcNamed.String())

	fFunc := ctx.Flag("func-flag")
	srcFunc := ctx.FlagSource(fFunc)
	assert.Equal(t, kong.ValueSourceResolver, srcFunc.Type)
	assert.Equal(t, "ResolverFunc", srcFunc.ResolverName)

	fWrapped := ctx.Flag("wrapped-flag")
	srcWrapped := ctx.FlagSource(fWrapped)
	assert.Equal(t, kong.ValueSourceResolver, srcWrapped.Type)
	assert.Equal(t, "wrapped-source", srcWrapped.ResolverName)
}

func TestProvenanceDefaultAndUnset(t *testing.T) {
	var cli struct {
		Defaulted string `default:"my-default"`
		Unset     string
	}
	p := mustNew(t, &cli)
	ctx, err := p.Parse(nil)
	assert.NoError(t, err)

	fDef := ctx.Flag("defaulted")
	srcDef := ctx.FlagSource(fDef)
	assert.Equal(t, kong.ValueSourceDefault, srcDef.Type)
	assert.True(t, srcDef.IsDefault())
	assert.Equal(t, "default", srcDef.String())

	fUnset := ctx.Flag("unset")
	srcUnset := ctx.FlagSource(fUnset)
	assert.Equal(t, kong.ValueSourceUnset, srcUnset.Type)
	assert.True(t, srcUnset.IsUnset())
	assert.Equal(t, "unset", srcUnset.String())
}

func TestProvenancePrecedence(t *testing.T) {
	var cli struct {
		Flag string `env:"KONG_TEST_PRECEDENCE" default:"def-val"`
	}
	t.Setenv("KONG_TEST_PRECEDENCE", "env-val")

	jsonRes, err := kong.JSON(strings.NewReader(`{"flag": "resolver-val"}`))
	assert.NoError(t, err)

	// 1. CLI wins over Resolver, Env, Default
	p1 := mustNew(t, &cli, kong.Resolvers(jsonRes))
	ctx1, err := p1.Parse([]string{"--flag=cli-val"})
	assert.NoError(t, err)
	assert.Equal(t, "cli-val", cli.Flag)
	assert.Equal(t, kong.ValueSourceFlag, ctx1.FlagValueSource(ctx1.Flag("flag")))

	// 2. Resolver wins over Env, Default
	p2 := mustNew(t, &cli, kong.Resolvers(jsonRes))
	ctx2, err := p2.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, "resolver-val", cli.Flag)
	src2 := ctx2.FlagSource(ctx2.Flag("flag"))
	assert.Equal(t, kong.ValueSourceResolver, src2.Type)
	assert.Equal(t, "json", src2.ResolverName)

	// 3. Env wins over Default
	p3 := mustNew(t, &cli)
	ctx3, err := p3.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, "env-val", cli.Flag)
	src3 := ctx3.FlagSource(ctx3.Flag("flag"))
	assert.Equal(t, kong.ValueSourceEnv, src3.Type)
	assert.Equal(t, "KONG_TEST_PRECEDENCE", src3.Env)

	// 4. Default wins when no env
	t.Setenv("KONG_TEST_PRECEDENCE", "")
	os.Unsetenv("KONG_TEST_PRECEDENCE")
	p4 := mustNew(t, &cli)
	ctx4, err := p4.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, "def-val", cli.Flag)
	src4 := ctx4.FlagSource(ctx4.Flag("flag"))
	assert.Equal(t, kong.ValueSourceDefault, src4.Type)
}

func TestProvenanceNestedAndDefaultCommands(t *testing.T) {
	type NestedCmd struct {
		NestedFlag string `default:"nested-def" env:"KONG_NESTED_ENV"`
	}
	type DefaultCmd struct {
		DefFlag string `default:"default-cmd-val"`
	}
	var cli struct {
		RootFlag string     `default:"root-def"`
		Nested   NestedCmd  `cmd:"" envprefix:"APP_"`
		Default  DefaultCmd `cmd:"" default:"1"`
	}

	t.Setenv("APP_KONG_NESTED_ENV", "nested-env-val")

	// Test default command selection
	p := mustNew(t, &cli)
	ctxDef, err := p.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, "default", ctxDef.Command())
	defFlag := ctxDef.Flag("def-flag")
	assert.True(t, defFlag != nil)
	assert.Equal(t, kong.ValueSourceDefault, ctxDef.FlagValueSource(defFlag))

	// Test nested command selection with env
	ctxNested, err := p.Parse([]string{"nested"})
	assert.NoError(t, err)
	assert.Equal(t, "nested", ctxNested.Command())
	nestedFlag := ctxNested.Flag("nested-flag")
	assert.True(t, nestedFlag != nil)
	srcNested := ctxNested.FlagSource(nestedFlag)
	assert.Equal(t, kong.ValueSourceEnv, srcNested.Type)
	assert.Equal(t, "APP_KONG_NESTED_ENV", srcNested.Env)

	// Test nested command with CLI flag
	ctxNestedCLI, err := p.Parse([]string{"nested", "--nested-flag=override"})
	assert.NoError(t, err)
	nestedFlagCLI := ctxNestedCLI.Flag("nested-flag")
	assert.Equal(t, kong.ValueSourceFlag, ctxNestedCLI.FlagValueSource(nestedFlagCLI))
}

func TestProvenanceLastWins(t *testing.T) {
	var cli struct {
		One string `lastwins:"group" default:"one-def"`
		Two string `lastwins:"group" default:"two-def"`
	}
	p := mustNew(t, &cli)
	ctx, err := p.Parse([]string{"--one=first", "--two=second"})
	assert.NoError(t, err)

	fOne := ctx.Flag("one")
	fTwo := ctx.Flag("two")

	assert.Equal(t, kong.ValueSourceFlag, ctx.FlagValueSource(fTwo))
	assert.Equal(t, kong.ValueSourceDefault, ctx.FlagValueSource(fOne))
}
