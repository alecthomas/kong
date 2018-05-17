package kong

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type DecoderContext struct {
	// Value being decoded into.
	Value *Value
}

type Decoder interface {
	Decode(ctx *DecoderContext, scan *Scanner, target reflect.Value) error
}

type DecoderFunc func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error

func (d DecoderFunc) Decode(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
	return d(ctx, scan, target)
}

var _ Decoder = DecoderFunc(nil)

type TypeDecoder interface {
	Type() reflect.Type
	Decoder
}

func NewTypeDecoder(typ reflect.Type, decoder DecoderFunc) TypeDecoder {
	return &typeDecoder{typ, decoder}
}

type typeDecoder struct {
	typ reflect.Type
	DecoderFunc
}

func (t *typeDecoder) Type() reflect.Type { return t.typ }

var _ TypeDecoder = &typeDecoder{}

type KindDecoder interface {
	Kind() reflect.Kind
	Decoder
}

func NewKindDecoder(kind reflect.Kind, decoder DecoderFunc) KindDecoder {
	return &kindDecoder{kind, decoder}
}

type kindDecoder struct {
	kind reflect.Kind
	DecoderFunc
}

func (k *kindDecoder) Kind() reflect.Kind { return k.kind }

var _ KindDecoder = &kindDecoder{}

type NamedDecoder interface {
	Name() string
	Decoder
}

func NewNamedDecoder(name string, decoder DecoderFunc) NamedDecoder {
	return &namedDecoder{name, decoder}
}

type namedDecoder struct {
	name string
	DecoderFunc
}

func (n *namedDecoder) Name() string { return n.name }

var _ NamedDecoder = &namedDecoder{}

var (
	namedDecoders = map[string]NamedDecoder{}
	typeDecoders  = map[reflect.Type]TypeDecoder{}
	kindDecoders  = map[reflect.Kind]KindDecoder{}
)

// DecoderForField finds a decoder for a struct field.
//
// Will return nil if a decoder can not be determined.
func DecoderForField(field reflect.StructField) Decoder {
	name, ok := field.Tag.Lookup("type")
	if ok {
		if decoder, ok := namedDecoders[name]; ok {
			return decoder
		}
	}
	return DecoderForType(field.Type)
}

// DecoderForType finds a decoder via a type or kind.
//
// Will return nil if a decoder can not be determined.
func DecoderForType(typ reflect.Type) Decoder {
	var decoder Decoder
	var ok bool
	if decoder, ok = typeDecoders[typ]; ok {
		return decoder
	} else if decoder, ok = kindDecoders[typ.Kind()]; ok {
		return decoder
	}
	return nil
}

// RegisterDecoder registers decoders.
//
// Decoders must be one of TypeDecoder, KindDecoder or NamedDecoder.
func RegisterDecoder(decoders ...Decoder) {
	for _, decoder := range decoders {
		switch decoder := decoder.(type) {
		case TypeDecoder:
			typeDecoders[decoder.Type()] = decoder
		case KindDecoder:
			kindDecoders[decoder.Kind()] = decoder
		case NamedDecoder:
			namedDecoders[decoder.Name()] = decoder
		default:
			fail("unsupported decoder type " + reflect.TypeOf(decoder).String())
		}
	}
}

func init() {
	RegisterDecoder(
		NewKindDecoder(reflect.Int, intDecoder),
		NewKindDecoder(reflect.Int8, intDecoder),
		NewKindDecoder(reflect.Int16, intDecoder),
		NewKindDecoder(reflect.Int32, intDecoder),
		NewKindDecoder(reflect.Int64, intDecoder),
		NewKindDecoder(reflect.Uint, uintDecoder),
		NewKindDecoder(reflect.Uint8, uintDecoder),
		NewKindDecoder(reflect.Uint16, uintDecoder),
		NewKindDecoder(reflect.Uint32, uintDecoder),
		NewKindDecoder(reflect.Uint64, uintDecoder),
		NewKindDecoder(reflect.Float32, floatDecoder(32)),
		NewKindDecoder(reflect.Float64, floatDecoder(64)),
		NewKindDecoder(reflect.String, func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
			target.SetString(scan.PopValue("string"))
			return nil
		}),
		NewKindDecoder(reflect.Bool, func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
			target.SetBool(true)
			return nil
		}),
		NewKindDecoder(reflect.Slice, sliceDecoder),
		NewTypeDecoder(reflect.TypeOf(time.Time{}), timeDecoder),
		NewTypeDecoder(reflect.TypeOf(time.Duration(0)), durationDecoder),
	)
}

func durationDecoder(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
	d, err := time.ParseDuration(scan.PopValue("duration"))
	if err != nil {
		return err
	}
	target.Set(reflect.ValueOf(d))
	return nil
}

func timeDecoder(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
	fmt := time.RFC3339
	if ctx.Value.Format != "" {
		fmt = ctx.Value.Format
	}
	t, err := time.Parse(fmt, scan.PopValue("time"))
	if err != nil {
		return err
	}
	target.Set(reflect.ValueOf(t))
	return nil
}

func intDecoder(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
	n, err := strconv.ParseInt(scan.PopValue("int"), 10, 64)
	if err != nil {
		return err
	}
	target.SetInt(n)
	return nil
}

func uintDecoder(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
	n, err := strconv.ParseUint(scan.PopValue("uint"), 10, 64)
	if err != nil {
		return err
	}
	target.SetUint(n)
	return nil
}

func floatDecoder(bits int) DecoderFunc {
	return func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
		n, err := strconv.ParseFloat(scan.PopValue("float"), bits)
		if err != nil {
			return err
		}
		target.SetFloat(n)
		return nil
	}
}

func sliceDecoder(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
	el := target.Type().Elem()
	sep, ok := ctx.Value.Field.Tag.Lookup("sep")
	if !ok {
		sep = ","
	}
	childScanner := Scan(strings.Split(scan.PopValue("list"), sep)...)
	childDecoder := DecoderForType(el)
	if childDecoder == nil {
		return fmt.Errorf("no decoder for element type of %s", target.Type())
	}
	for childScanner.Peek().Type != EOLToken {
		childValue := reflect.New(el).Elem()
		err := childDecoder.Decode(ctx, childScanner, childValue)
		if err != nil {
			return err
		}
		target.Set(reflect.Append(target, childValue))
	}
	return nil
}
