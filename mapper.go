package kong

import (
	"fmt"
	"io/ioutil"
	"math/bits"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	mapperValueType = reflect.TypeOf((*MapperValue)(nil)).Elem()
	boolMapperType  = reflect.TypeOf((*BoolMapper)(nil)).Elem()
)

// DecodeContext is passed to a Mapper's Decode().
//
// It contains the Value being decoded into and the Scanner to parse from.
type DecodeContext struct {
	// Value being decoded into.
	Value *Value
	// Scan contains the input to scan into Target.
	Scan *Scanner
}

// WithScanner creates a clone of this context with a new Scanner.
func (r *DecodeContext) WithScanner(scan *Scanner) *DecodeContext {
	return &DecodeContext{
		Value: r.Value,
		Scan:  scan,
	}
}

// MapperValue may be implemented by fields in order to provide custom mapping.
type MapperValue interface {
	Decode(ctx *DecodeContext) error
}

type mapperValueAdapter struct {
	isBool bool
}

func (m *mapperValueAdapter) Decode(ctx *DecodeContext, target reflect.Value) error {
	if target.Type().Implements(mapperValueType) {
		return target.Interface().(MapperValue).Decode(ctx)
	}
	return target.Addr().Interface().(MapperValue).Decode(ctx)
}

func (m *mapperValueAdapter) IsBool() bool {
	return m.isBool
}

// A Mapper represents how a field is mapped from command-line values to Go.
//
// Mappers can be associated with concrete fields via pointer, reflect.Type, reflect.Kind, or via a "type" tag.
//
// Additionally, if a type implements the MappverValue interface, it will be used.
type Mapper interface {
	// Decode ctx.Value with ctx.Scanner into target.
	Decode(ctx *DecodeContext, target reflect.Value) error
}

// A BoolMapper is a Mapper to a value that is a boolean.
//
// This is used solely for formatting help.
type BoolMapper interface {
	IsBool() bool
}

// A MapperFunc is a single function that complies with the Mapper interface.
type MapperFunc func(ctx *DecodeContext, target reflect.Value) error

func (m MapperFunc) Decode(ctx *DecodeContext, target reflect.Value) error { //nolint: golint
	return m(ctx, target)
}

// A Registry contains a set of mappers and supporting lookup methods.
type Registry struct {
	names  map[string]Mapper
	types  map[reflect.Type]Mapper
	kinds  map[reflect.Kind]Mapper
	values map[reflect.Value]Mapper
}

// NewRegistry creates a new (empty) Registry.
func NewRegistry() *Registry {
	return &Registry{
		names:  map[string]Mapper{},
		types:  map[reflect.Type]Mapper{},
		kinds:  map[reflect.Kind]Mapper{},
		values: map[reflect.Value]Mapper{},
	}
}

// ForNamedValue finds a mapper for a value with a user-specified name.
//
// Will return nil if a mapper can not be determined.
func (r *Registry) ForNamedValue(name string, value reflect.Value) Mapper {
	if mapper, ok := r.names[name]; ok {
		return mapper
	}
	return r.ForValue(value)
}

// ForValue looks up the Mapper for a reflect.Value.
func (r *Registry) ForValue(value reflect.Value) Mapper {
	if mapper, ok := r.values[value]; ok {
		return mapper
	}
	return r.ForType(value.Type())
}

// ForNamedType finds a mapper for a type with a user-specified name.
//
// Will return nil if a mapper can not be determined.
func (r *Registry) ForNamedType(name string, typ reflect.Type) Mapper {
	if mapper, ok := r.names[name]; ok {
		return mapper
	}
	return r.ForType(typ)
}

// ForType finds a mapper from a type, by type, then kind.
//
// Will return nil if a mapper can not be determined.
func (r *Registry) ForType(typ reflect.Type) Mapper {
	// Check if the type implements MapperValue.
	for _, impl := range []reflect.Type{typ, reflect.PtrTo(typ)} {
		if impl.Implements(mapperValueType) {
			return &mapperValueAdapter{impl.Implements(boolMapperType)}
		}
	}
	var mapper Mapper
	var ok bool
	if mapper, ok = r.types[typ]; ok {
		return mapper
	} else if mapper, ok = r.kinds[typ.Kind()]; ok {
		return mapper
	}
	return nil
}

// RegisterKind registers a Mapper for a reflect.Kind.
func (r *Registry) RegisterKind(kind reflect.Kind, mapper Mapper) *Registry {
	r.kinds[kind] = mapper
	return r
}

// RegisterName registers a mapper to be used if the value mapper has a "type" tag matching name.
//
// eg.
//
// 		Mapper string `kong:"type='colour'`
//   	registry.RegisterName("colour", ...)
func (r *Registry) RegisterName(name string, mapper Mapper) *Registry {
	r.names[name] = mapper
	return r
}

// RegisterType registers a Mapper for a reflect.Type.
func (r *Registry) RegisterType(typ reflect.Type, mapper Mapper) *Registry {
	r.types[typ] = mapper
	return r
}

// RegisterValue registers a Mapper by pointer to the field value.
func (r *Registry) RegisterValue(ptr interface{}, mapper Mapper) *Registry {
	key := reflect.ValueOf(ptr)
	if key.Kind() != reflect.Ptr {
		panic("expected a pointer")
	}
	key = key.Elem()
	r.values[key] = mapper
	return r
}

// RegisterDefaults registers Mappers for all builtin supported Go types and some common stdlib types.
func (r *Registry) RegisterDefaults() *Registry {
	return r.RegisterKind(reflect.Int, intDecoder(bits.UintSize)).
		RegisterKind(reflect.Int8, intDecoder(8)).
		RegisterKind(reflect.Int16, intDecoder(16)).
		RegisterKind(reflect.Int32, intDecoder(32)).
		RegisterKind(reflect.Int64, intDecoder(64)).
		RegisterKind(reflect.Uint, uintDecoder(64)).
		RegisterKind(reflect.Uint8, uintDecoder(bits.UintSize)).
		RegisterKind(reflect.Uint16, uintDecoder(16)).
		RegisterKind(reflect.Uint32, uintDecoder(32)).
		RegisterKind(reflect.Uint64, uintDecoder(64)).
		RegisterKind(reflect.Float32, floatDecoder(32)).
		RegisterKind(reflect.Float64, floatDecoder(64)).
		RegisterKind(reflect.String, MapperFunc(func(ctx *DecodeContext, target reflect.Value) error {
			token := ctx.Scan.PopValue("string")
			target.SetString(token)
			return nil
		})).
		RegisterKind(reflect.Bool, boolMapper{}).
		RegisterKind(reflect.Slice, sliceDecoder(r)).
		RegisterKind(reflect.Map, mapDecoder(r)).
		RegisterType(reflect.TypeOf(time.Time{}), timeDecoder()).
		RegisterType(reflect.TypeOf(time.Duration(0)), durationDecoder()).
		RegisterType(reflect.TypeOf(&url.URL{}), urlMapper()).
		RegisterName("path", pathMapper(r)).
		RegisterName("existingfile", existingFileMapper(r)).
		RegisterName("existingdir", existingDirMapper(r))
}

type boolMapper struct{}

func (boolMapper) Decode(ctx *DecodeContext, target reflect.Value) error {
	if ctx.Scan.Peek().Type == FlagValueToken {
		token := ctx.Scan.Pop()
		target.SetBool(token.Value == "true")
	} else {
		target.SetBool(true)
	}
	return nil
}
func (boolMapper) IsBool() bool { return true }

func durationDecoder() MapperFunc {
	return func(ctx *DecodeContext, target reflect.Value) error {
		r, err := time.ParseDuration(ctx.Scan.PopValue("duration"))
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(r))
		return nil
	}
}

func timeDecoder() MapperFunc {
	return func(ctx *DecodeContext, target reflect.Value) error {
		fmt := time.RFC3339
		if ctx.Value.Format != "" {
			fmt = ctx.Value.Format
		}
		t, err := time.Parse(fmt, ctx.Scan.PopValue("time"))
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(t))
		return nil
	}
}

func intDecoder(bits int) MapperFunc {
	return func(ctx *DecodeContext, target reflect.Value) error {
		value := ctx.Scan.PopValue("int")
		n, err := strconv.ParseInt(value, 10, bits)
		if err != nil {
			return fmt.Errorf("invalid int %q", value)
		}
		target.SetInt(n)
		return nil
	}
}

func uintDecoder(bits int) MapperFunc {
	return func(ctx *DecodeContext, target reflect.Value) error {
		value := ctx.Scan.PopValue("uint")
		n, err := strconv.ParseUint(value, 10, bits)
		if err != nil {
			return fmt.Errorf("invalid uint %q", value)
		}
		target.SetUint(n)
		return nil
	}
}

func floatDecoder(bits int) MapperFunc {
	return func(ctx *DecodeContext, target reflect.Value) error {
		value := ctx.Scan.PopValue("float")
		n, err := strconv.ParseFloat(value, bits)
		if err != nil {
			return fmt.Errorf("invalid float %q", value)
		}
		target.SetFloat(n)
		return nil
	}
}

func mapDecoder(r *Registry) MapperFunc {
	return func(ctx *DecodeContext, target reflect.Value) error {
		if target.IsNil() {
			target.Set(reflect.MakeMap(target.Type()))
		}
		el := target.Type()
		var childScanner *Scanner
		if ctx.Value.Flag != nil {
			// If decoding a flag, we need an argument.
			childScanner = Scan(SplitEscaped(ctx.Scan.PopValue("map"), ';')...)
		} else {
			tokens := ctx.Scan.PopWhile(func(t Token) bool { return t.IsValue() })
			childScanner = ScanFromTokens(tokens...)
		}
		for !childScanner.Peek().IsEOL() {
			token := childScanner.PopValue("map")
			parts := strings.SplitN(token, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("expected \"<key>=<value>\" but got %q", token)
			}
			key, value := parts[0], parts[1]

			keyTypeName, valueTypeName := "", ""
			if typ := ctx.Value.Tag.Type; typ != "" {
				parts := strings.Split(typ, ":")
				if len(parts) != 2 {
					return fmt.Errorf("type:\"\" on map field must be in the form \"[<keytype>]:[<valuetype>]\"")
				}
				keyTypeName, valueTypeName = parts[0], parts[1]
			}

			keyScanner := Scan(key)
			keyDecoder := r.ForNamedType(keyTypeName, el.Key())
			keyValue := reflect.New(el.Key()).Elem()
			if err := keyDecoder.Decode(ctx.WithScanner(keyScanner), keyValue); err != nil {
				return fmt.Errorf("invalid map key %q", key)
			}

			valueScanner := Scan(value)
			valueDecoder := r.ForNamedType(valueTypeName, el.Elem())
			valueValue := reflect.New(el.Elem()).Elem()
			if err := valueDecoder.Decode(ctx.WithScanner(valueScanner), valueValue); err != nil {
				return fmt.Errorf("invalid map value %q", value)
			}

			target.SetMapIndex(keyValue, valueValue)
		}
		return nil
	}
}

func sliceDecoder(r *Registry) MapperFunc {
	return func(ctx *DecodeContext, target reflect.Value) error {
		el := target.Type().Elem()
		sep := ctx.Value.Tag.Sep
		var childScanner *Scanner
		if ctx.Value.Flag != nil {
			// If decoding a flag, we need an argument.
			childScanner = Scan(SplitEscaped(ctx.Scan.PopValue("list"), sep)...)
		} else {
			tokens := ctx.Scan.PopWhile(func(t Token) bool { return t.IsValue() })
			childScanner = ScanFromTokens(tokens...)
		}
		childDecoder := r.ForNamedType(ctx.Value.Tag.Type, el)
		if childDecoder == nil {
			return fmt.Errorf("no mapper for element type of %s", target.Type())
		}
		for !childScanner.Peek().IsEOL() {
			childValue := reflect.New(el).Elem()
			err := childDecoder.Decode(ctx.WithScanner(childScanner), childValue)
			if err != nil {
				return err
			}
			target.Set(reflect.Append(target, childValue))
		}
		return nil
	}
}

func pathMapper(r *Registry) MapperFunc {
	return func(ctx *DecodeContext, target reflect.Value) error {
		if target.Kind() == reflect.Slice {
			return sliceDecoder(r)(ctx, target)
		}
		if target.Kind() != reflect.String {
			return fmt.Errorf("\"path\" type must be applied to a string not %s", target.Type())
		}
		path := ctx.Scan.PopValue("file")
		path = expandPath(path)
		target.SetString(path)
		return nil
	}
}

func existingFileMapper(r *Registry) MapperFunc {
	return func(ctx *DecodeContext, target reflect.Value) error {
		if target.Kind() == reflect.Slice {
			return sliceDecoder(r)(ctx, target)
		}
		if target.Kind() != reflect.String {
			return fmt.Errorf("\"existingfile\" type must be applied to a string not %s", target.Type())
		}
		path := ctx.Scan.PopValue("file")
		path = expandPath(path)
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}
		if stat.IsDir() {
			return fmt.Errorf("%q exists but is a directory", path)
		}
		target.SetString(path)
		return nil
	}
}

func existingDirMapper(r *Registry) MapperFunc {
	return func(ctx *DecodeContext, target reflect.Value) error {
		if target.Kind() == reflect.Slice {
			return sliceDecoder(r)(ctx, target)
		}
		if target.Kind() != reflect.String {
			return fmt.Errorf("\"existingdir\" must be applied to a string not %s", target.Type())
		}
		path := ctx.Scan.PopValue("file")
		path = expandPath(path)
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !stat.IsDir() {
			return fmt.Errorf("%q exists but is not a directory", path)
		}
		target.SetString(path)
		return nil
	}
}

func urlMapper() MapperFunc {
	return func(ctx *DecodeContext, target reflect.Value) error {
		url, err := url.Parse(ctx.Scan.PopValue("url"))
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(url))
		return nil
	}
}

// SplitEscaped splits a string on a separator.
//
// It differs from strings.Split() in that the separator can exist in a field by escaping it with a \. eg.
//
//     SplitEscaped(`hello\,there,bob`, ',') == []string{"hello,there", "bob"}
func SplitEscaped(s string, sep rune) (out []string) {
	escaped := false
	token := ""
	for _, ch := range s {
		switch {
		case escaped:
			token += string(ch)
			escaped = false
		case ch == '\\':
			escaped = true
		case ch == sep && !escaped:
			out = append(out, token)
			token = ""
			escaped = false
		default:
			token += string(ch)
		}
	}
	if token != "" {
		out = append(out, token)
	}
	return
}

// JoinEscaped joins a slice of strings on sep, but also escapes any instances of sep in the fields with \. eg.
//
//     JoinEscaped([]string{"hello,there", "bob"}, ',') == `hello\,there,bob`
func JoinEscaped(s []string, sep rune) string {
	escaped := []string{}
	for _, e := range s {
		escaped = append(escaped, strings.Replace(e, string(sep), `\`+string(sep), -1))
	}
	return strings.Join(escaped, string(sep))
}

// FileContentFlag is a flag value that loads a file's contents into its value.
type FileContentFlag struct {
	Path string
	Data []byte
}

func (f *FileContentFlag) Decode(ctx *DecodeContext) error { // nolint: golint
	filename := ctx.Scan.PopValue("filename")
	data, err := ioutil.ReadFile(filename) // nolint: gosec
	if err != nil {
		return fmt.Errorf("failed to open %q: %s", filename, err)
	}
	f.Path = filename
	f.Data = data
	return nil
}
