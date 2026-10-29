package kong_test

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/alecthomas/kong"
)

func requirementGrammar(expression string) any {
	fields := []reflect.StructField{
		{Name: "Run", Type: reflect.TypeOf(false), Tag: reflect.StructTag("requires:" + strconv.Quote(expression))},
		{Name: "A", Type: reflect.TypeOf(false)},
		{Name: "B", Type: reflect.TypeOf(false)},
		{Name: "C", Type: reflect.TypeOf(false)},
	}
	return reflect.New(reflect.StructOf(fields)).Interface()
}

func assertRequirementDiagnostic(t *testing.T, err error, owner, expression string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected a requirement error")
	}
	words := func(text string) []string {
		return strings.FieldsFunc(text, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_'
		})
	}
	mentioned := map[string]bool{}
	for _, word := range words(err.Error()) {
		mentioned[strings.TrimLeft(word, "-")] = true
	}
	for _, name := range append([]string{owner}, words(expression)...) {
		if !mentioned[name] {
			t.Fatalf("diagnostic does not identify %q from owner %q and expression %q: %v", name, owner, expression, err)
		}
	}
}

func TestRequirementsBooleanExpressions(t *testing.T) {
	for _, test := range []struct {
		expression string
		present    []string
		valid      bool
	}{
		{"a", nil, false}, {"a", []string{"--a"}, true},
		{"!a", nil, true}, {"!a", []string{"--a"}, false},
		{"a|b&c", []string{"--a"}, true},
		{"a|b&c", []string{"--b"}, false},
		{"a|b&c", []string{"--b", "--c"}, true},
		{"(a|b)&c", []string{"--a"}, false},
		{"(a|b)&c", []string{"--a", "--c"}, true},
		{"!!a & !(b|c)", []string{"--a"}, true},
		{"!!a & !(b|c)", []string{"--a", "--c"}, false},
		{" a \t &\n (b | !c) ", []string{"--a"}, true},
	} {
		func() {
			p, err := kong.New(requirementGrammar(test.expression))
			if err != nil {
				t.Fatal(err)
			}
			_, err = p.Parse(append([]string{"--run"}, test.present...))
			if (err == nil) != test.valid {
				t.Fatalf("valid=%v, error=%v", test.valid, err)
			}
			if err != nil {
				assertRequirementDiagnostic(t, err, "run", test.expression)
			}
		}()
	}
}

func TestRequirementsInvalidGrammarAtNew(t *testing.T) {
	for _, expression := range []string{"", " ", "a b", "a&&b", "a||b", "a|", "&a", "!", "()", "(a", "a)", "a|missing", "a & !missing", "--a", "a \x00"} {
		t.Run(expression, func(t *testing.T) {
			if _, err := kong.New(requirementGrammar(expression)); err == nil {
				t.Fatalf("accepted %q", expression)
			}
		})
	}
}

func TestRequirementsPresenceAndReuse(t *testing.T) {
	var cli struct {
		Run   bool   `requires:"flag & count & text" default:"true"`
		Flag  bool   `short:"f" aliases:"switch" negatable:"" default:"true"`
		Count int    `default:"4"`
		Text  string `default:"default"`
	}
	cli.Run = true
	p, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"--run", "--flag=false", "--count=0", "--text="},
		{"--run", "--no-flag", "--count=0", "--text="},
		{"--run", "--switch=false", "--count=0", "--text="},
		{"--run", "-f", "--count=0", "--text="},
	} {
		ctx, err := p.Parse(args)
		if err != nil {
			t.Fatal(err)
		}
		if err := ctx.Validate(); err != nil {
			t.Fatal(err)
		}
		if _, err := p.Parse([]string{"--run"}); err == nil {
			t.Fatal("defaults or prior parse counted as supplied")
		}
	}
	if _, err := p.Parse(nil); err != nil {
		t.Fatal("absent owner triggered requirement", err)
	}
}

func TestRequirementsEnvironmentAndResolvers(t *testing.T) {
	var cli struct {
		Run   bool   `requires:"token" env:"KONG_REQUIREMENTS_RUN"`
		Token string `env:"KONG_REQUIREMENTS_TOKEN"`
	}
	t.Setenv("KONG_REQUIREMENTS_RUN", "false")
	p, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse(nil); err == nil {
		t.Fatal("environment owner did not trigger")
	}
	t.Setenv("KONG_REQUIREMENTS_TOKEN", "")
	ctx, err := p.Parse(nil)
	if err != nil {
		t.Fatal("empty environment value is supplied", err)
	}
	if err := os.Unsetenv("KONG_REQUIREMENTS_TOKEN"); err != nil {
		t.Fatal(err)
	}
	if err := ctx.Validate(); err != nil {
		t.Fatal("context lost its resolved environment presence", err)
	}
	var resolved struct {
		Run   bool `requires:"count"`
		Count int
	}
	p, err = kong.New(&resolved, kong.Resolvers(kong.ResolverFunc(func(_ *kong.Context, _ *kong.Path, flag *kong.Flag) (any, error) {
		if flag.Name == "run" {
			return false, nil
		}
		if flag.Name == "count" {
			return 0, nil
		}
		return nil, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse(nil); err != nil {
		t.Fatal(err)
	}
	p, err = kong.New(&resolved, kong.Resolvers(kong.ResolverFunc(func(_ *kong.Context, _ *kong.Path, flag *kong.Flag) (any, error) {
		if flag.Name == "run" {
			return false, nil
		}
		return nil, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse(nil); err == nil {
		t.Fatal("nil resolution counted as present")
	}
}

func TestRequirementsCommandScopeAndInactiveBranches(t *testing.T) {
	var cli struct {
		Token bool
		Run   struct {
			Execute bool `requires:"token"`
		} `cmd:""`
		Other struct {
			Execute bool `requires:"secret"`
			Secret  bool
		} `cmd:""`
	}
	p, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse([]string{"run", "--execute"}); err == nil {
		t.Fatal("missing ancestor accepted")
	}
	if _, err = p.Parse([]string{"--token", "run", "--execute"}); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse([]string{"other", "--execute", "--secret"}); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse([]string{"run"}); err != nil {
		t.Fatal("inactive requirement triggered", err)
	}
}

func TestRequirementsInvalidInactiveAndNonFlagDeclarations(t *testing.T) {
	for _, grammar := range []any{
		&struct {
			Run struct {
				Bad bool `requires:"missing"`
			} `cmd:""`
			Other struct{} `cmd:""`
		}{},
		&struct {
			Root bool                 `requires:"child"`
			Run  struct{ Child bool } `cmd:""`
		}{},
		&struct {
			Run struct{} `cmd:"" requires:"run"`
		}{},
		&struct {
			Value string `arg:"" requires:"flag"`
			Flag  bool
		}{},
		&struct {
			Flag bool `aliases:"alias"`
			Run  bool `requires:"alias"`
		}{},
	} {
		if _, err := kong.New(grammar); err == nil {
			t.Fatalf("invalid declaration accepted: %T", grammar)
		}
	}
}

func TestRequirementsCyclesAndRepeatedTags(t *testing.T) {
	var cli struct {
		A bool `kong:"requires='b',requires='!c'"`
		B bool `requires:"a"`
		C bool
	}
	p, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse([]string{"--a"}); err == nil {
		t.Fatal("dependency implicitly supplied")
	}
	if _, err = p.Parse([]string{"--a", "--b"}); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse([]string{"--a", "--b", "--c"}); err == nil {
		t.Fatal("second requires ignored")
	}
}

func TestRequirementsDynamicAndEmbeddedFinalNames(t *testing.T) {
	type Credentials struct {
		Token bool
		Send  bool `requires:"auth-token"`
	}
	var cli struct {
		Credentials `embed:"" prefix:"auth-"`
	}
	var command struct {
		Run bool `requires:"auth-token"`
	}
	p, err := kong.New(&cli, kong.DynamicCommand("push", "", "", &command))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse([]string{"push", "--run"}); err == nil {
		t.Fatal("dynamic requirement ignored")
	}
	if _, err = p.Parse([]string{"--auth-token", "--auth-send", "push", "--run"}); err != nil {
		t.Fatal(err)
	}
	var invalid struct {
		Run bool `requires:"missing"`
	}
	if _, err = kong.New(&struct{}{}, kong.DynamicCommand("bad", "", "", &invalid)); err == nil {
		t.Fatal("invalid dynamic requirement accepted")
	}
	var named struct {
		Token bool
		Send  bool `requires:"opt-Token"`
	}
	p, err = kong.New(&named, kong.FlagNamer(func(name string) string { return "opt-" + name }))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse([]string{"--opt-Send"}); err == nil {
		t.Fatal("custom names ignored")
	}
	if _, err = p.Parse([]string{"--opt-Token", "--opt-Send"}); err != nil {
		t.Fatal(err)
	}
}

func TestRequirementsLastWinsUsesSurvivingFlags(t *testing.T) {
	var cli struct {
		A     bool `lastwins:"mode" requires:"token"`
		B     bool `lastwins:"mode"`
		Run   bool `requires:"b"`
		Token bool
	}
	p, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse([]string{"--a", "--b", "--run"}); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse([]string{"--b", "--a", "--run", "--token"}); err == nil {
		t.Fatal("overridden b counted as present")
	}
}

func TestRequirementsManualContextValidation(t *testing.T) {
	var cli struct {
		Run   bool `requires:"run & token"`
		Token bool
	}
	p, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := kong.Trace(p, []string{"--run"})
	if err != nil || ctx.Error != nil {
		t.Fatalf("trace failed: %v %v", err, ctx.Error)
	}
	if err := ctx.Reset(); err != nil {
		t.Fatal(err)
	}
	if err := ctx.Resolve(); err != nil {
		t.Fatal(err)
	}
	if _, err := ctx.Apply(); err != nil {
		t.Fatal(err)
	}
	if err := ctx.Validate(); err == nil {
		t.Fatal("Context.Validate ignored requires")
	}
}

func TestRequirementsCoexistsWithRequired(t *testing.T) {
	for _, test := range []struct {
		args  []string
		valid bool
	}{
		{[]string{"--run", "--token"}, false},
		{[]string{"--mandatory=x", "--run"}, false},
		{[]string{"--mandatory=x", "--run", "--token"}, true},
	} {
		var cli struct {
			Run       bool `requires:"token"`
			Token     bool
			Mandatory string `required:""`
		}
		p, err := kong.New(&cli)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = p.Parse(test.args); (err == nil) != test.valid {
			t.Fatalf("args=%v valid=%v error=%v", test.args, test.valid, err)
		}
	}
}

func TestRequirementsCoexistsWithXor(t *testing.T) {
	for _, test := range []struct {
		args  []string
		valid bool
	}{
		{[]string{"--run", "--a", "--b"}, false},
		{[]string{"--run"}, false},
		{[]string{"--run", "--a"}, true},
		{[]string{"--run", "--b"}, true},
	} {
		var cli struct {
			Run bool `requires:"a|b"`
			A   bool `xor:"mode"`
			B   bool `xor:"mode"`
		}
		p, err := kong.New(&cli)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = p.Parse(test.args); (err == nil) != test.valid {
			t.Fatalf("args=%v valid=%v error=%v", test.args, test.valid, err)
		}
	}
}

func TestRequirementsCoexistsWithAnd(t *testing.T) {
	for _, test := range []struct {
		args  []string
		valid bool
	}{
		{[]string{"--run", "--a"}, false},
		{[]string{"--run"}, false},
		{[]string{"--run", "--a", "--b"}, true},
	} {
		var cli struct {
			Run bool `requires:"a"`
			A   bool `and:"pair"`
			B   bool `and:"pair"`
		}
		p, err := kong.New(&cli)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = p.Parse(test.args); (err == nil) != test.valid {
			t.Fatalf("args=%v valid=%v error=%v", test.args, test.valid, err)
		}
	}
}

func TestRequirementsDiagnosticIdentity(t *testing.T) {
	var cli struct {
		Dispatch   bool `name:"dispatch-action" kong:"requires='primary',requires='credential | backup'"`
		Primary    bool
		Credential bool
		Backup     bool
	}
	p, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Parse([]string{"--dispatch-action", "--primary"})
	assertRequirementDiagnostic(t, err, "dispatch-action", "credential | backup")
	if _, err = p.Parse([]string{"--dispatch-action", "--primary", "--backup"}); err != nil {
		t.Fatal(err)
	}
}

func TestRequirementsInactiveEnvironment(t *testing.T) {
	t.Setenv("KONG_SCOPE_EXECUTE_TEST", "false")
	var cli struct {
		Run struct {
			Execute bool `requires:"token" env:"KONG_SCOPE_EXECUTE_TEST"`
			Token   bool
		} `cmd:""`
		Other struct {
			Execute bool `requires:"secret" env:"KONG_SCOPE_EXECUTE_TEST"`
			Secret  bool
		} `cmd:""`
	}
	p, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Parse([]string{"run"}); err == nil {
		t.Fatal("active environment owner did not enforce its requirement")
	}
	for _, args := range [][]string{{"other", "--secret"}, {"run", "--token"}} {
		if _, err = p.Parse(args); err != nil {
			t.Fatalf("inactive branch affected %v: %v", args, err)
		}
	}
}

func TestRequirementsBeforeResetEnvironment(t *testing.T) {
	t.Setenv("KONG_LIFECYCLE_TOKEN_TEST", "secret")
	var cli struct {
		Run   bool   `requires:"!token"`
		Token string `env:"KONG_LIFECYCLE_TOKEN_TEST"`
	}
	p, err := kong.New(&cli, kong.WithBeforeReset(func() error {
		return os.Unsetenv("KONG_LIFECYCLE_TOKEN_TEST")
	}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Parse([]string{"--run"}); err != nil {
		t.Fatalf("environment removed before resolution must not supply token: %v", err)
	}
	if cli.Token != "" {
		t.Fatal("token was unexpectedly consumed")
	}
	if _, err := p.Parse([]string{"--run", "--token=explicit"}); err == nil {
		t.Fatal("removing the environment must not remove an explicit command-line token")
	}
}

type consumingRequirementToken string

func (v *consumingRequirementToken) Decode(ctx *kong.DecodeContext) error {
	var token string
	if err := ctx.Scan.PopValueInto("token", &token); err != nil {
		return err
	}
	*v = consumingRequirementToken(token)
	return os.Unsetenv("KONG_CONSUMED_TOKEN_TEST")
}

func TestRequirementsConsumingEnvironmentDecoder(t *testing.T) {
	t.Setenv("KONG_CONSUMED_TOKEN_TEST", "secret")
	var cli struct {
		Run   bool                      `requires:"token"`
		Token consumingRequirementToken `env:"KONG_CONSUMED_TOKEN_TEST"`
	}
	p, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := p.Parse([]string{"--run"})
	if err != nil {
		t.Fatalf("successfully decoded environment token must count: %v", err)
	}
	if cli.Token != "secret" {
		t.Fatal("token was not decoded")
	}
	if err := ctx.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Parse([]string{"--run"}); err == nil {
		t.Fatal("a later parse must not retain the consumed environment token")
	}
}

func TestRequirementsResetRefresh(t *testing.T) {
	t.Setenv("KONG_REFRESH_TOKEN_TEST", "secret")
	var cli struct {
		Run    bool `requires:"anchor & !token"`
		Anchor bool
		Token  string `env:"KONG_REFRESH_TOKEN_TEST"`
	}
	p, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := kong.Trace(p, []string{"--run", "--anchor"})
	if err != nil {
		t.Fatal(err)
	}
	validate := func() error {
		t.Helper()
		if err := ctx.Reset(); err != nil {
			t.Fatal(err)
		}
		if err := ctx.Resolve(); err != nil {
			t.Fatal(err)
		}
		if _, err := ctx.Apply(); err != nil {
			t.Fatal(err)
		}
		return ctx.Validate()
	}
	if err := validate(); err == nil {
		t.Fatal("initial environment must supply token")
	}
	if err := os.Unsetenv("KONG_REFRESH_TOKEN_TEST"); err != nil {
		t.Fatal(err)
	}
	if err := validate(); err != nil {
		t.Fatalf("reset must refresh environment and retain traced argv: %v", err)
	}
	if !cli.Run || !cli.Anchor || cli.Token != "" {
		t.Fatal("reset/apply produced unexpected values")
	}
	t.Setenv("KONG_REFRESH_TOKEN_TEST", "next")
	if err := validate(); err == nil {
		t.Fatal("a later reset must observe a newly supplied environment token")
	}
}

func TestRequirementsApplyDefaults(t *testing.T) {
	t.Setenv("KONG_DEFAULTS_RUN_TEST", "false")
	t.Setenv("KONG_DEFAULTS_TOKEN_TEST", "false")
	if err := os.Unsetenv("KONG_DEFAULTS_TOKEN_TEST"); err != nil {
		t.Fatal(err)
	}
	type defaultsCLI struct {
		Run   bool `requires:"token" env:"KONG_DEFAULTS_RUN_TEST" default:"true"`
		Token bool `env:"KONG_DEFAULTS_TOKEN_TEST"`
	}
	var cli defaultsCLI
	if err := kong.ApplyDefaults(&cli); err == nil {
		t.Error("an environment-supplied false owner must enforce its requirement through ApplyDefaults")
	}
	t.Setenv("KONG_DEFAULTS_TOKEN_TEST", "false")
	cli = defaultsCLI{}
	p, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := kong.Trace(p, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := ctx.Resolve(); err != nil {
		t.Fatal(err)
	}
	if err := ctx.ApplyDefaults(); err != nil {
		t.Fatal(err)
	}
	if err := ctx.Validate(); err != nil {
		t.Fatalf("false environment values must count as supplied: %v", err)
	}
	if err := os.Unsetenv("KONG_DEFAULTS_TOKEN_TEST"); err != nil {
		t.Fatal(err)
	}
	if err := ctx.Validate(); err != nil {
		t.Fatalf("default application must retain consumed environment presence: %v", err)
	}
	if err := os.Unsetenv("KONG_DEFAULTS_RUN_TEST"); err != nil {
		t.Fatal(err)
	}
	cli = defaultsCLI{}
	if err := kong.ApplyDefaults(&cli); err != nil {
		t.Fatalf("a default alone must not activate the owner: %v", err)
	}
	cli = defaultsCLI{Run: true}
	if err := kong.ApplyDefaults(&cli); err != nil {
		t.Fatalf("an initial value alone must not activate the owner: %v", err)
	}
}
