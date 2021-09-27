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

func TestDurationMapperJSONResolver(t *testing.T) {
	var cli struct {
		Flag time.Duration
	}
	resolver, err := kong.JSON(strings.NewReader(`{"flag": 5000000000}`))
	require.NoError(t, err)
	k := mustNew(t, &cli, kong.Resolvers(resolver))
	_, err = k.Parse(nil)
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

func TestPassthroughStopsParsing(t *testing.T) {
	type cli struct {
		Interactive bool     `short:"i"`
		Image       string   `arg:""`
		Argv        []string `arg:"" optional:"" passthrough:""`
	}

	var actual cli
	p := mustNew(t, &actual)

	_, err := p.Parse([]string{"alpine", "sudo", "-i", "true"})
	require.NoError(t, err)
	require.Equal(t, cli{
		Interactive: false,
		Image:       "alpine",
		Argv:        []string{"sudo", "-i", "true"},
	}, actual)

	_, err = p.Parse([]string{"alpine", "-i", "sudo", "-i", "true"})
	require.NoError(t, err)
	require.Equal(t, cli{
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

func TestPathMapper(t *testing.T) {
	var cli struct {
		Path string `arg:"" type:"path"`
	}
	p := mustNew(t, &cli)

	_, err := p.Parse([]string{"/an/absolute/path"})
	require.NoError(t, err)
	require.Equal(t, "/an/absolute/path", cli.Path)

	_, err = p.Parse([]string{"-"})
	require.NoError(t, err)
	require.Equal(t, "-", cli.Path)
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
	require.Panics(t, func() {
		_, err := k.Parse([]string{"--help"})
		require.NoError(t, err)
	})
	require.Contains(t, b.String(), "--flag=/a/b/c")
}

type testMapperWithPlaceHolder struct {
}

func (t testMapperWithPlaceHolder) Decode(ctx *kong.DecodeContext, target reflect.Value) error {
	target.SetString("hi")
	return nil
}

func (t testMapperWithPlaceHolder) PlaceHolder(flag *kong.Flag) string {
	return "/a/b/c"
}

type Mode int32

// Contains MODE_UNSET, MODE_FOO and MODE_BAR (0, 1, 2). This data is pulled
// from the .pb.go file that protoc generates.
func (m Mode) EnumDescriptor() ([]byte, []int) {
	return []byte{
		// 211 bytes of a gzipped FileDescriptorProto
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0xcf, 0xcd, 0x4a, 0x03, 0x31,
		0x14, 0x05, 0x60, 0x53, 0xab, 0xd6, 0x8b, 0x96, 0x21, 0xb8, 0x98, 0xe5, 0xe8, 0x6a, 0x10, 0x9a,
		0x60, 0xfb, 0x04, 0x56, 0x2b, 0x08, 0xd6, 0x40, 0xd4, 0x8d, 0x2e, 0x24, 0x7f, 0x74, 0x02, 0x93,
		0x49, 0x68, 0x92, 0xf7, 0x97, 0xf9, 0xc1, 0xe5, 0x77, 0x2f, 0x9c, 0xc3, 0x01, 0x70, 0xc2, 0x76,
		0x24, 0x1c, 0x7d, 0xf2, 0xf8, 0x22, 0xaa, 0xc6, 0x38, 0x11, 0xef, 0x7e, 0x00, 0x9e, 0xde, 0x5e,
		0x59, 0x48, 0xd6, 0x77, 0x11, 0xdf, 0xc0, 0x99, 0x36, 0x32, 0x1f, 0x4a, 0x54, 0xa1, 0x7a, 0xc1,
		0x47, 0xf4, 0x57, 0xdb, 0x85, 0x9c, 0xca, 0x59, 0x85, 0xea, 0x4b, 0x3e, 0x02, 0xdf, 0xc2, 0xdc,
		0x79, 0x6d, 0xca, 0xd3, 0x0a, 0xd5, 0xcb, 0xf5, 0x35, 0x99, 0x12, 0xc9, 0xde, 0x6b, 0xc3, 0x87,
		0xd7, 0xfd, 0x1a, 0xe6, 0xbd, 0xf0, 0x12, 0x60, 0xcf, 0x9e, 0x77, 0xbf, 0x5f, 0xef, 0x1f, 0xbb,
		0xcf, 0xe2, 0x04, 0x5f, 0xc1, 0x62, 0xf0, 0x0b, 0x63, 0x05, 0xfa, 0xd7, 0xf6, 0x91, 0x17, 0xb3,
		0xed, 0xe6, 0xfb, 0xe1, 0x60, 0x53, 0x93, 0x25, 0x51, 0xde, 0x51, 0x29, 0x92, 0x6a, 0x94, 0x3f,
		0x06, 0x1a, 0x85, 0x0b, 0xad, 0x59, 0xa9, 0xd6, 0xae, 0x44, 0x08, 0x54, 0x66, 0xdb, 0x6a, 0x3a,
		0x75, 0xca, 0xf3, 0x61, 0xd5, 0xe6, 0x2f, 0x00, 0x00, 0xff, 0xff, 0x42, 0x96, 0xa0, 0x13, 0xe3,
		0x00, 0x00, 0x00,
	}, []int{}
}

func TestPBEnumMapper(t *testing.T) {
	var cli struct {
		RunMode Mode `type:"pbenum"`
	}

	// Enum should exist
	k := mustNew(t, &cli, kong.IgnoreFieldsRegex(".*XXX_"))
	ctx, err := k.Parse([]string{"--run-mode=MODE_FOO"})
	require.NoError(t, err)
	require.NotNil(t, ctx)

	// Enum should not exist
	_, shouldErr := k.Parse([]string{"--run-mode=MODE_BAX"})
	require.Error(t, shouldErr)
	require.Contains(t, shouldErr.Error(), "'MODE_BAX' not available in proto enum map")
}

func TestPBEnumMapperLowercase(t *testing.T) {
	var cli struct {
		RunMode Mode `type:"pbenum" pbenum_lowercase:""`
	}

	// Enum should exist
	k := mustNew(t, &cli, kong.IgnoreFieldsRegex(".*XXX_"))
	ctx, err := k.Parse([]string{"--run-mode=mode_foo"})
	require.NoError(t, err)
	require.NotNil(t, ctx)

	// Enum should not exist
	_, shouldErr := k.Parse([]string{"--run-mode=MODE_FOO"})
	require.Error(t, shouldErr)
	require.Contains(t, shouldErr.Error(), "'MODE_FOO' not available in proto enum map")
}

func TestPBEnumStripPrefix(t *testing.T) {
	var cli struct {
		RunMode Mode `type:"pbenum" pbenum_strip_prefix:"MODE_" pbenum_lowercase:""`
	}

	// Enum should exist
	k := mustNew(t, &cli, kong.IgnoreFieldsRegex(".*XXX_"))
	ctx, err := k.Parse([]string{"--run-mode=foo"})
	require.NoError(t, err)
	require.NotNil(t, ctx)

	// Enum should not exist
	_, shouldErr := k.Parse([]string{"--run-mode=mode_foo"})
	require.Error(t, shouldErr)
	require.Contains(t, shouldErr.Error(), "'mode_foo' not available in proto enum map")
}
