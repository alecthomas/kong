package kong_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/kong"
)

func TestValueMapper(t *testing.T) {
	var cli struct {
		Flag string
	}
	k := mustNew(t, &cli, kong.ValueMapper(&cli.Flag, testMooMapper{}))
	_, err := k.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, "", cli.Flag)
	_, err = k.Parse([]string{"--flag"})
	assert.NoError(t, err)
	assert.Equal(t, "MOO", cli.Flag)
}

type textUnmarshalerValue int

func (m *textUnmarshalerValue) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "hello" {
		*m = 10
	} else {
		return fmt.Errorf("expected \"hello\"")
	}
	return nil
}

func TestTextUnmarshaler(t *testing.T) {
	var cli struct {
		Value textUnmarshalerValue
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--value=hello"})
	assert.NoError(t, err)
	assert.Equal(t, 10, int(cli.Value))
	_, err = p.Parse([]string{"--value=other"})
	assert.Error(t, err)
}

type jsonUnmarshalerValue string

func (j *jsonUnmarshalerValue) UnmarshalJSON(text []byte) error {
	var v string
	err := json.Unmarshal(text, &v)
	if err != nil {
		return err
	}
	*j = jsonUnmarshalerValue(strings.ToUpper(v))
	return nil
}

func TestJSONUnmarshaler(t *testing.T) {
	var cli struct {
		Value jsonUnmarshalerValue
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--value=\"hello\""})
	assert.NoError(t, err)
	assert.Equal(t, "HELLO", string(cli.Value))
}

func TestNamedMapper(t *testing.T) {
	var cli struct {
		Flag string `type:"moo"`
	}
	k := mustNew(t, &cli, kong.NamedMapper("moo", testMooMapper{}))
	_, err := k.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, "", cli.Flag)
	_, err = k.Parse([]string{"--flag"})
	assert.NoError(t, err)
	assert.Equal(t, "MOO", cli.Flag)
}

type testMooMapper struct {
	text string
}

func (t testMooMapper) Decode(ctx *kong.DecodeContext, target reflect.Value) error {
	if t.text == "" {
		target.SetString("MOO")
	} else {
		target.SetString(t.text)
	}
	return nil
}
func (testMooMapper) IsBool() bool { return true }

func TestTimeMapper(t *testing.T) {
	var cli struct {
		Flag time.Time `format:"2006"`
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--flag=2008"})
	assert.NoError(t, err)
	expected, err := time.Parse("2006", "2008")
	assert.NoError(t, err)
	assert.Equal(t, 2008, expected.Year())
	assert.Equal(t, expected, cli.Flag)
}

func TestDurationMapper(t *testing.T) {
	var cli struct {
		Flag time.Duration
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--flag=5s"})
	assert.NoError(t, err)
	assert.Equal(t, time.Second*5, cli.Flag)
}

func TestDurationMapperJSONResolver(t *testing.T) {
	var cli struct {
		Flag time.Duration
	}
	resolver, err := kong.JSON(strings.NewReader(`{"flag": 5000000000}`))
	assert.NoError(t, err)
	k := mustNew(t, &cli, kong.Resolvers(resolver))
	_, err = k.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, time.Second*5, cli.Flag)
}

func TestSplitEscaped(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, kong.SplitEscaped("a,b", ','))
	assert.Equal(t, []string{"a,b", "c"}, kong.SplitEscaped(`a\,b,c`, ','))
	assert.Equal(t, []string{"a,b,c"}, kong.SplitEscaped(`a,b,c`, -1))
}

func TestJoinEscaped(t *testing.T) {
	assert.Equal(t, `a,b`, kong.JoinEscaped([]string{"a", "b"}, ','))
	assert.Equal(t, `a\,b,c`, kong.JoinEscaped([]string{`a,b`, `c`}, ','))
	assert.Equal(t, kong.JoinEscaped(kong.SplitEscaped(`a\,b,c`, ','), ','), `a\,b,c`)
}

func TestMapWithNamedTypes(t *testing.T) {
	var cli struct {
		TypedValue map[string]string `type:":moo"`
		TypedKey   map[string]string `type:"upper:"`
	}
	k := mustNew(t, &cli, kong.NamedMapper("moo", testMooMapper{}), kong.NamedMapper("upper", testUppercaseMapper{}))
	_, err := k.Parse([]string{"--typed-value", "first=5s", "--typed-value", "second=10s"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"first": "MOO", "second": "MOO"}, cli.TypedValue)
	_, err = k.Parse([]string{"--typed-key", "first=5s", "--typed-key", "second=10s"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"FIRST": "5s", "SECOND": "10s"}, cli.TypedKey)
}

func TestMapWithMultipleValues(t *testing.T) {
	var cli struct {
		Value map[string]string
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--value=a=b;c=d"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "b", "c": "d"}, cli.Value)
}

func TestMapWithDifferentSeparator(t *testing.T) {
	var cli struct {
		Value map[string]string `mapsep:","`
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--value=a=b,c=d"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "b", "c": "d"}, cli.Value)
}

func TestMapWithNoSeparator(t *testing.T) {
	var cli struct {
		Slice []string          `sep:"none"`
		Value map[string]string `mapsep:"none"`
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--slice=a,n,c", "--value=a=b;n=d"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a,n,c"}, cli.Slice)
	assert.Equal(t, map[string]string{"a": "b;n=d"}, cli.Value)
}

func TestURLMapper(t *testing.T) {
	var cli struct {
		URL *url.URL `arg:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"http://w3.org"})
	assert.NoError(t, err)
	assert.Equal(t, "http://w3.org", cli.URL.String())
	_, err = p.Parse([]string{":foo"})
	assert.Error(t, err)
}

func TestSliceConsumesRemainingPositionalArgs(t *testing.T) {
	var cli struct {
		Remainder []string `arg:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--", "ls", "-lart"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"ls", "-lart"}, cli.Remainder)
}

func TestPassthroughStopsParsing(t *testing.T) {
	type cli struct {
		Interactive bool     `short:"i"`
		Image       string   `arg:""`
		Argv        []string `arg:"" optional:"" passthrough:""`
	}

	var actual cli
	p := mustNew(t, &actual)

	_, err := p.Parse([]string{"alpine", "sudo", "-i", "true"})
	assert.NoError(t, err)
	assert.Equal(t, cli{
		Interactive: false,
		Image:       "alpine",
		Argv:        []string{"sudo", "-i", "true"},
	}, actual)

	_, err = p.Parse([]string{"alpine", "-i", "sudo", "-i", "true"})
	assert.NoError(t, err)
	assert.Equal(t, cli{
		Interactive: true,
		Image:       "alpine",
		Argv:        []string{"sudo", "-i", "true"},
	}, actual)
}

type mappedValue struct {
	decoded string
}

func (m *mappedValue) Decode(ctx *kong.DecodeContext) error {
	err := ctx.Scan.PopValueInto("mapped", &m.decoded)
	return err
}

func TestMapperValue(t *testing.T) {
	var cli struct {
		Value mappedValue `arg:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"foo"})
	assert.NoError(t, err)
	assert.Equal(t, "foo", cli.Value.decoded)
}

func TestFileContentFlag(t *testing.T) {
	var cli struct {
		File kong.FileContentFlag
	}
	f, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	fmt.Fprint(f, "hello world")
	f.Close()
	_, err = mustNew(t, &cli).Parse([]string{"--file", f.Name()})
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world"), []byte(cli.File))
}

func TestNamedFileContentFlag(t *testing.T) {
	var cli struct {
		File kong.NamedFileContentFlag
	}
	f, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	fmt.Fprint(f, "hello world")
	f.Close()
	_, err = mustNew(t, &cli).Parse([]string{"--file", f.Name()})
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world"), cli.File.Contents)
	assert.Equal(t, f.Name(), cli.File.Filename)
}

func TestNamedSliceTypesDontHaveEllipsis(t *testing.T) {
	var cli struct {
		File kong.FileContentFlag
	}
	b := bytes.NewBuffer(nil)
	parser := mustNew(t, &cli, kong.Writers(b, b), kong.Exit(func(int) { panic("exit") }))
	// Ensure that --help
	assert.Panics(t, func() {
		_, err := parser.Parse([]string{"--help"})
		assert.NoError(t, err)
	})
	assert.NotContains(t, b.String(), `--file=FILE-CONTENT-FLAG,...`)
}

func TestCounter(t *testing.T) {
	var cli struct {
		Int   int     `type:"counter" short:"i"`
		Uint  uint    `type:"counter" short:"u"`
		Float float64 `type:"counter" short:"f"`
	}
	p := mustNew(t, &cli)

	_, err := p.Parse([]string{"--int", "--int", "--int"})
	assert.NoError(t, err)
	assert.Equal(t, 3, cli.Int)

	_, err = p.Parse([]string{"--int=5"})
	assert.NoError(t, err)
	assert.Equal(t, 5, cli.Int)

	_, err = p.Parse([]string{"-iii"})
	assert.NoError(t, err)
	assert.Equal(t, 3, cli.Int)

	_, err = p.Parse([]string{"-uuu"})
	assert.NoError(t, err)
	assert.Equal(t, uint(3), cli.Uint)

	_, err = p.Parse([]string{"-fff"})
	assert.NoError(t, err)
	assert.Equal(t, 3., cli.Float)
}

func TestNumbers(t *testing.T) {
	type CLI struct {
		F32 float32
		F64 float64
		I8  int8
		I16 int16
		I32 int32
		I64 int64
		U8  uint8
		U16 uint16
		U32 uint32
		U64 uint64
	}
	var cli CLI
	p := mustNew(t, &cli)
	t.Run("Max", func(t *testing.T) {
		_, err := p.Parse([]string{
			"--f-32", fmt.Sprintf("%v", math.MaxFloat32),
			"--f-64", fmt.Sprintf("%v", math.MaxFloat64),
			"--i-8", fmt.Sprintf("%v", int8(math.MaxInt8)), //nolint:perfsprint // want int8
			"--i-16", fmt.Sprintf("%v", int16(math.MaxInt16)), //nolint:perfsprint // want int16
			"--i-32", fmt.Sprintf("%v", int32(math.MaxInt32)), //nolint:perfsprint // want int32
			"--i-64", fmt.Sprintf("%v", int64(math.MaxInt64)), //nolint:perfsprint // want int64
			"--u-8", fmt.Sprintf("%v", uint8(math.MaxUint8)), //nolint:perfsprint // want uint8
			"--u-16", fmt.Sprintf("%v", uint16(math.MaxUint16)), //nolint:perfsprint // want uint16
			"--u-32", fmt.Sprintf("%v", uint32(math.MaxUint32)), //nolint:perfsprint // want uint32
			"--u-64", fmt.Sprintf("%v", uint64(math.MaxUint64)), //nolint:perfsprint // want uint64
		})
		assert.NoError(t, err)
		assert.Equal(t, CLI{
			F32: math.MaxFloat32,
			F64: math.MaxFloat64,
			I8:  math.MaxInt8,
			I16: math.MaxInt16,
			I32: math.MaxInt32,
			I64: math.MaxInt64,
			U8:  math.MaxUint8,
			U16: math.MaxUint16,
			U32: math.MaxUint32,
			U64: math.MaxUint64,
		}, cli)
	})
	t.Run("Min", func(t *testing.T) {
		_, err := p.Parse([]string{
			fmt.Sprintf("--i-8=%v", int8(math.MinInt8)),
			fmt.Sprintf("--i-16=%v", int16(math.MinInt16)),
			fmt.Sprintf("--i-32=%v", int32(math.MinInt32)),
			fmt.Sprintf("--i-64=%v", int64(math.MinInt64)),
			fmt.Sprintf("--u-8=%v", 0),
			fmt.Sprintf("--u-16=%v", 0),
			fmt.Sprintf("--u-32=%v", 0),
			fmt.Sprintf("--u-64=%v", 0),
		})
		assert.NoError(t, err)
		assert.Equal(t, CLI{
			I8:  math.MinInt8,
			I16: math.MinInt16,
			I32: math.MinInt32,
			I64: math.MinInt64,
		}, cli)
	})
}

func TestJSONLargeNumber(t *testing.T) {
	// Make sure that large numbers are not internally converted to
	// scientific notation when the mapper parses the values.
	// (Scientific notation is e.g. `1e+06` instead of `1000000`.)

	// Large signed integers:
	{
		var cli struct {
			N int64
		}
		json := `{"n": 1000000}`
		r, err := kong.JSON(strings.NewReader(json))
		assert.NoError(t, err)
		parser := mustNew(t, &cli, kong.Resolvers(r))
		_, err = parser.Parse([]string{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1000000), cli.N)
	}

	// Large unsigned integers:
	{
		var cli struct {
			N uint64
		}
		json := `{"n": 1000000}`
		r, err := kong.JSON(strings.NewReader(json))
		assert.NoError(t, err)
		parser := mustNew(t, &cli, kong.Resolvers(r))
		_, err = parser.Parse([]string{})
		assert.NoError(t, err)
		assert.Equal(t, uint64(1000000), cli.N)
	}

	// Large floats:
	{
		var cli struct {
			N float64
		}
		json := `{"n": 1000000.1}`
		r, err := kong.JSON(strings.NewReader(json))
		assert.NoError(t, err)
		parser := mustNew(t, &cli, kong.Resolvers(r))
		_, err = parser.Parse([]string{})
		assert.NoError(t, err)
		assert.Equal(t, float64(1000000.1), cli.N)
	}
}

func TestFileMapper(t *testing.T) {
	type CLI struct {
		File *os.File `arg:""`
	}
	var cli CLI
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"testdata/file.txt"})
	assert.NoError(t, err)
	assert.NotZero(t, cli.File)
	_ = cli.File.Close()
	_, err = p.Parse([]string{"testdata/missing.txt"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing.txt:")
	assert.IsError(t, err, os.ErrNotExist)
	_, err = p.Parse([]string{"-"})
	assert.NoError(t, err)
	assert.Equal(t, os.Stdin, cli.File)
}

func TestFileContentMapper(t *testing.T) {
	type CLI struct {
		File []byte `type:"filecontent"`
	}
	var cli CLI
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--file", "testdata/file.txt"})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`Hello world.`), cli.File)
	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--file", "testdata/missing.txt"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing.txt:")
	assert.IsError(t, err, os.ErrNotExist)
	p = mustNew(t, &cli)

	_, err = p.Parse([]string{"--file", "testdata"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")
}

func TestPathMapperUsingStringPointer(t *testing.T) {
	type CLI struct {
		Path *string `type:"path"`
	}
	var cli CLI

	t.Run("With value", func(t *testing.T) {
		pwd, err := os.Getwd()
		assert.NoError(t, err)
		p := mustNew(t, &cli)
		_, err = p.Parse([]string{"--path", "."})
		assert.NoError(t, err)
		assert.NotZero(t, cli.Path)
		assert.Equal(t, pwd, *cli.Path)
	})

	t.Run("Zero value", func(t *testing.T) {
		p := mustNew(t, &cli)
		_, err := p.Parse([]string{"--path", ""})
		assert.NoError(t, err)
		assert.NotZero(t, cli.Path)
		wd, err := os.Getwd()
		assert.NoError(t, err)
		assert.Equal(t, wd, *cli.Path)
	})

	t.Run("Without value", func(t *testing.T) {
		p := mustNew(t, &cli)
		_, err := p.Parse([]string{"--"})
		assert.NoError(t, err)
		assert.Equal(t, nil, cli.Path)
	})

	t.Run("Non-string pointer", func(t *testing.T) {
		type CLI struct {
			Path *any `type:"path"`
		}
		var cli CLI
		p := mustNew(t, &cli)
		_, err := p.Parse([]string{"--path", ""})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `"path" type must be applied to a string`)
	})
}

//nolint:dupl
func TestExistingFileMapper(t *testing.T) {
	type CLI struct {
		File string `type:"existingfile"`
	}
	var cli CLI
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--file", "testdata/file.txt"})
	assert.NoError(t, err)
	assert.NotZero(t, cli.File)
	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--file", "testdata/missing.txt"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing.txt:")
	assert.IsError(t, err, os.ErrNotExist)
	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--file", "testdata/"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exists but is a directory")
}

func TestExistingFileMapperSlice(t *testing.T) {
	type CLI struct {
		Files []string `type:"existingfile"`
	}
	var cli CLI
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--files", "testdata/file.txt", "--files", "testdata/file.txt"})
	assert.NoError(t, err)
	assert.NotZero(t, cli.Files)
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(pwd, "testdata", "file.txt"), filepath.Join(pwd, "testdata", "file.txt")}, cli.Files)
}

func TestExistingFileMapperDefaultMissing(t *testing.T) {
	type CLI struct {
		File string `type:"existingfile" default:"testdata/missing.txt"`
	}
	var cli CLI
	p := mustNew(t, &cli)
	file := filepath.Join("testdata", "file.txt")
	_, err := p.Parse([]string{"--file", file})
	assert.NoError(t, err)
	assert.NotZero(t, cli.File)
	assert.Contains(t, cli.File, file)
	p = mustNew(t, &cli)
	_, err = p.Parse([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing.txt:")
	assert.IsError(t, err, os.ErrNotExist)
}

func TestExistingFileMapperDefaultMissingCmds(t *testing.T) {
	type CLI struct {
		CmdA struct {
			FileA string `type:"existingfile" default:"testdata/aaa-missing.txt"`
			FileB string `type:"existingfile" default:"testdata/bbb-missing.txt"`
		} `cmd:""`
		CmdC struct {
			FileC string `type:"existingfile" default:"testdata/ccc-missing.txt"`
		} `cmd:""`
	}
	var cli CLI
	file := filepath.Join("testdata", "file.txt")
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"cmd-a", "--file-a", file, "--file-b", file})
	assert.NoError(t, err)
	assert.NotZero(t, cli.CmdA.FileA)
	assert.Contains(t, cli.CmdA.FileA, file)
	assert.NotZero(t, cli.CmdA.FileB)
	assert.Contains(t, cli.CmdA.FileB, file)
	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"cmd-a", "--file-a", file})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bbb-missing.txt:")
	assert.IsError(t, err, os.ErrNotExist)
}

//nolint:dupl
func TestExistingDirMapper(t *testing.T) {
	type CLI struct {
		Dir string `type:"existingdir"`
	}
	var cli CLI
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--dir", "testdata/"})
	assert.NoError(t, err)
	assert.NotZero(t, cli.Dir)
	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--dir", "missingdata/"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missingdata:")
	assert.IsError(t, err, os.ErrNotExist)
	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--dir", "testdata/file.txt"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exists but is not a directory")
}

func TestExistingDirMapperDefaultMissing(t *testing.T) {
	type CLI struct {
		Dir string `type:"existingdir" default:"missing-dir"`
	}
	var cli CLI
	p := mustNew(t, &cli)
	dir := "testdata"
	_, err := p.Parse([]string{"--dir", dir})
	assert.NoError(t, err)
	assert.NotZero(t, cli.Dir)
	assert.Contains(t, cli.Dir, dir)
	p = mustNew(t, &cli)
	_, err = p.Parse([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing-dir:")
	assert.IsError(t, err, os.ErrNotExist)
}

func TestExistingDirMapperDefaultMissingCmds(t *testing.T) {
	type CLI struct {
		CmdA struct {
			DirA string `type:"existingdir" default:"aaa-missing-dir"`
			DirB string `type:"existingdir" default:"bbb-missing-dir"`
		} `cmd:""`
		CmdC struct {
			DirC string `type:"existingdir" default:"ccc-missing-dir"`
		} `cmd:""`
	}
	var cli CLI
	dir := "testdata"
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"cmd-a", "--dir-a", dir, "--dir-b", dir})
	assert.NoError(t, err)
	assert.NotZero(t, cli.CmdA.DirA)
	assert.NotZero(t, cli.CmdA.DirB)
	assert.Contains(t, cli.CmdA.DirA, dir)
	assert.Contains(t, cli.CmdA.DirB, dir)
	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"cmd-a", "--dir-a", dir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bbb-missing-dir:")
	assert.IsError(t, err, os.ErrNotExist)
}

func TestMapperPlaceHolder(t *testing.T) {
	var cli struct {
		Flag string
	}
	b := bytes.NewBuffer(nil)
	k := mustNew(
		t,
		&cli,
		kong.Writers(b, b),
		kong.ValueMapper(&cli.Flag, testMapperWithPlaceHolder{}),
		kong.Exit(func(int) { panic("exit") }),
	)
	// Ensure that --help
	assert.Panics(t, func() {
		_, err := k.Parse([]string{"--help"})
		assert.NoError(t, err)
	})
	assert.Contains(t, b.String(), "--flag=/a/b/c")
}

type testMapperWithPlaceHolder struct{}

func (t testMapperWithPlaceHolder) Decode(ctx *kong.DecodeContext, target reflect.Value) error {
	target.SetString("hi")
	return nil
}

func (t testMapperWithPlaceHolder) PlaceHolder(flag *kong.Flag) string {
	return "/a/b/c"
}

func TestMapperVarsContributor(t *testing.T) {
	var cli struct {
		Flag string `help:"Some help with ${avar}"`
	}
	b := bytes.NewBuffer(nil)
	k := mustNew(
		t,
		&cli,
		kong.Writers(b, b),
		kong.ValueMapper(&cli.Flag, testMapperVarsContributor{}),
		kong.Exit(func(int) { panic("exit") }),
	)
	// Ensure that --help
	assert.Panics(t, func() {
		_, err := k.Parse([]string{"--help"})
		assert.NoError(t, err)
	})
	assert.Contains(t, b.String(), "--flag=STRING")
	assert.Contains(t, b.String(), "Some help with a var", b.String())
}

type testMapperVarsContributor struct{}

func (t testMapperVarsContributor) Vars(value *kong.Value) kong.Vars {
	return kong.Vars{"avar": "a var"}
}

func (t testMapperVarsContributor) Decode(ctx *kong.DecodeContext, target reflect.Value) error {
	target.SetString("hi")
	return nil
}

func TestValuesThatLookLikeFlags(t *testing.T) {
	var cli struct {
		Slice []string
		Map   map[string]string
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--slice", "-foo"})
	assert.Error(t, err)
	_, err = k.Parse([]string{"--map", "-foo=-bar"})
	assert.Error(t, err)
	_, err = k.Parse([]string{"--slice=-foo", "--slice=-bar"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"-foo", "-bar"}, cli.Slice)
	_, err = k.Parse([]string{"--map=-foo=-bar"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"-foo": "-bar"}, cli.Map)
}
