package kong_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/alecthomas/kong"
)

func TestValueMapper(t *testing.T) {
	var cli struct {
		Flag string
	}
	k := mustNew(t, &cli, kong.ValueMapper(&cli.Flag, testMooMapper{}))
	_, err := k.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "", cli.Flag)
	_, err = k.Parse([]string{"--flag"})
	require.NoError(t, err)
	require.Equal(t, "MOO", cli.Flag)
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
	require.NoError(t, err)
	require.Equal(t, 10, int(cli.Value))
	_, err = p.Parse([]string{"--value=other"})
	require.Error(t, err)
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
	require.NoError(t, err)
	require.Equal(t, "HELLO", string(cli.Value))
}

func TestNamedMapper(t *testing.T) {
	var cli struct {
		Flag string `type:"moo"`
	}
	k := mustNew(t, &cli, kong.NamedMapper("moo", testMooMapper{}))
	_, err := k.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "", cli.Flag)
	_, err = k.Parse([]string{"--flag"})
	require.NoError(t, err)
	require.Equal(t, "MOO", cli.Flag)
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
	require.NoError(t, err)
	expected, err := time.Parse("2006", "2008")
	require.NoError(t, err)
	require.Equal(t, 2008, expected.Year())
	require.Equal(t, expected, cli.Flag)
}

func TestDurationMapper(t *testing.T) {
	var cli struct {
		Flag time.Duration
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--flag=5s"})
	require.NoError(t, err)
	require.Equal(t, time.Second*5, cli.Flag)
}

func TestSplitEscaped(t *testing.T) {
	require.Equal(t, []string{"a", "b"}, kong.SplitEscaped("a,b", ','))
	require.Equal(t, []string{"a,b", "c"}, kong.SplitEscaped(`a\,b,c`, ','))
	require.Equal(t, []string{"a,b,c"}, kong.SplitEscaped(`a,b,c`, -1))
}

func TestJoinEscaped(t *testing.T) {
	require.Equal(t, `a,b`, kong.JoinEscaped([]string{"a", "b"}, ','))
	require.Equal(t, `a\,b,c`, kong.JoinEscaped([]string{`a,b`, `c`}, ','))
	require.Equal(t, kong.JoinEscaped(kong.SplitEscaped(`a\,b,c`, ','), ','), `a\,b,c`)
}

func TestMapWithNamedTypes(t *testing.T) {
	var cli struct {
		TypedValue map[string]string `type:":moo"`
		TypedKey   map[string]string `type:"upper:"`
	}
	k := mustNew(t, &cli, kong.NamedMapper("moo", testMooMapper{}), kong.NamedMapper("upper", testUppercaseMapper{}))
	_, err := k.Parse([]string{"--typed-value", "first=5s", "--typed-value", "second=10s"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"first": "MOO", "second": "MOO"}, cli.TypedValue)
	_, err = k.Parse([]string{"--typed-key", "first=5s", "--typed-key", "second=10s"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"FIRST": "5s", "SECOND": "10s"}, cli.TypedKey)
}

func TestMapWithMultipleValues(t *testing.T) {
	var cli struct {
		Value map[string]string
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--value=a=b;c=d"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"a": "b", "c": "d"}, cli.Value)
}

func TestMapWithDifferentSeparator(t *testing.T) {
	var cli struct {
		Value map[string]string `mapsep:","`
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--value=a=b,c=d"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"a": "b", "c": "d"}, cli.Value)
}

func TestMapWithNoSeparator(t *testing.T) {
	var cli struct {
		Slice []string          `sep:"none"`
		Value map[string]string `mapsep:"none"`
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--slice=a,n,c", "--value=a=b;n=d"})
	require.NoError(t, err)
	require.Equal(t, []string{"a,n,c"}, cli.Slice)
	require.Equal(t, map[string]string{"a": "b;n=d"}, cli.Value)
}

func TestURLMapper(t *testing.T) {
	var cli struct {
		URL *url.URL `arg:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"http://w3.org"})
	require.NoError(t, err)
	require.Equal(t, "http://w3.org", cli.URL.String())
	_, err = p.Parse([]string{":foo"})
	require.Error(t, err)
}

func TestSliceConsumesRemainingPositionalArgs(t *testing.T) {
	var cli struct {
		Remainder []string `arg:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--", "ls", "-lart"})
	require.NoError(t, err)
	require.Equal(t, []string{"ls", "-lart"}, cli.Remainder)
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
	require.NoError(t, err)
	require.Equal(t, "foo", cli.Value.decoded)
}

func TestFileContentFlag(t *testing.T) {
	var cli struct {
		File kong.FileContentFlag
	}
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	fmt.Fprint(f, "hello world")
	f.Close()
	_, err = mustNew(t, &cli).Parse([]string{"--file", f.Name()})
	require.NoError(t, err)
	require.Equal(t, []byte("hello world"), []byte(cli.File))
}

func TestNamedFileContentFlag(t *testing.T) {
	var cli struct {
		File kong.NamedFileContentFlag
	}
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	fmt.Fprint(f, "hello world")
	f.Close()
	_, err = mustNew(t, &cli).Parse([]string{"--file", f.Name()})
	require.NoError(t, err)
	require.Equal(t, []byte("hello world"), cli.File.Contents)
	require.Equal(t, f.Name(), cli.File.Filename)
}

func TestNamedSliceTypesDontHaveEllipsis(t *testing.T) {
	var cli struct {
		File kong.FileContentFlag
	}
	b := bytes.NewBuffer(nil)
	parser := mustNew(t, &cli, kong.Writers(b, b), kong.Exit(func(int) { panic("exit") }))
	// Ensure that --help
	require.Panics(t, func() {
		_, err := parser.Parse([]string{"--help"})
		require.NoError(t, err)
	})
	require.NotContains(t, b.String(), `--file=FILE-CONTENT-FLAG,...`)
}

func TestCounter(t *testing.T) {
	var cli struct {
		Int   int     `type:"counter" short:"i"`
		Uint  uint    `type:"counter" short:"u"`
		Float float64 `type:"counter" short:"f"`
	}
	p := mustNew(t, &cli)

	_, err := p.Parse([]string{"--int", "--int", "--int"})
	require.NoError(t, err)
	require.Equal(t, 3, cli.Int)

	_, err = p.Parse([]string{"--int=5"})
	require.NoError(t, err)
	require.Equal(t, 5, cli.Int)

	_, err = p.Parse([]string{"-iii"})
	require.NoError(t, err)
	require.Equal(t, 3, cli.Int)

	_, err = p.Parse([]string{"-uuu"})
	require.NoError(t, err)
	require.Equal(t, uint(3), cli.Uint)

	_, err = p.Parse([]string{"-fff"})
	require.NoError(t, err)
	require.Equal(t, 3., cli.Float)
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
			"--i-8", fmt.Sprintf("%v", int8(math.MaxInt8)),
			"--i-16", fmt.Sprintf("%v", int16(math.MaxInt16)),
			"--i-32", fmt.Sprintf("%v", int32(math.MaxInt32)),
			"--i-64", fmt.Sprintf("%v", int64(math.MaxInt64)),
			"--u-8", fmt.Sprintf("%v", uint8(math.MaxUint8)),
			"--u-16", fmt.Sprintf("%v", uint16(math.MaxUint16)),
			"--u-32", fmt.Sprintf("%v", uint32(math.MaxUint32)),
			"--u-64", fmt.Sprintf("%v", uint64(math.MaxUint64)),
		})
		require.NoError(t, err)
		require.Equal(t, CLI{
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
		require.NoError(t, err)
		require.Equal(t, CLI{
			I8:  math.MinInt8,
			I16: math.MinInt16,
			I32: math.MinInt32,
			I64: math.MinInt64,
		}, cli)
	})
}

func TestFileMapper(t *testing.T) {
	type CLI struct {
		File *os.File `arg:""`
	}
	var cli CLI
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"testdata/file.txt"})
	require.NoError(t, err)
	require.NotNil(t, cli.File)
	_ = cli.File.Close()
	_, err = p.Parse([]string{"testdata/missing.txt"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing.txt: no such file or directory")
	_, err = p.Parse([]string{"-"})
	require.NoError(t, err)
	require.Equal(t, os.Stdin, cli.File)
}
