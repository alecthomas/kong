package kong

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
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
	kindDecoders  map[reflect.Kind]KindDecoder
)

// DecoderForField finds a decoder for a struct field.
//
func DecoderForField(field reflect.StructField) Decoder {
	name, ok := field.Tag.Lookup("type")
	if ok {
		if decoder, ok := namedDecoders[name]; ok {
			return decoder
		}
	}
	return DecoderForType(field.Type)
}

func DecoderForType(typ reflect.Type) Decoder {
	var decoder Decoder
	var ok bool
	if decoder, ok = typeDecoders[typ]; ok {
		return decoder
	} else if decoder, ok = kindDecoders[typ.Kind()]; ok {
		return decoder
	}
	return missingDecoder
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
			panic("unsupported decoder type " + reflect.TypeOf(decoder).String())
		}
	}
}

func init() {
	intDecoder := NewKindDecoder(reflect.Int, func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
		n, err := strconv.ParseInt(scan.PopValue("int"), 10, 64)
		if err != nil {
			return err
		}
		target.SetInt(n)
		return nil
	})
	uintDecoder := NewKindDecoder(reflect.Uint, func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
		n, err := strconv.ParseUint(scan.PopValue("uint"), 10, 64)
		if err != nil {
			return err
		}
		target.SetUint(n)
		return nil
	})
	kindDecoders = map[reflect.Kind]KindDecoder{
		reflect.Int:    intDecoder,
		reflect.Int8:   intDecoder,
		reflect.Int16:  intDecoder,
		reflect.Int32:  intDecoder,
		reflect.Int64:  intDecoder,
		reflect.Uint:   uintDecoder,
		reflect.Uint8:  uintDecoder,
		reflect.Uint16: uintDecoder,
		reflect.Uint32: uintDecoder,
		reflect.Uint64: uintDecoder,
		reflect.Float32: NewKindDecoder(reflect.Float32, func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
			n, err := strconv.ParseFloat(scan.PopValue("float"), 32)
			if err != nil {
				return err
			}
			target.SetFloat(n)
			return nil
		}),
		reflect.Float64: NewKindDecoder(reflect.Float64, func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
			n, err := strconv.ParseFloat(scan.PopValue("float"), 64)
			if err != nil {
				return err
			}
			target.SetFloat(n)
			return nil
		}),
		reflect.String: NewKindDecoder(reflect.String, func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
			target.SetString(scan.PopValue("string"))
			return nil
		}),
		reflect.Bool: NewKindDecoder(reflect.Bool, func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
			target.SetBool(true)
			return nil
		}),
		reflect.Slice: NewKindDecoder(reflect.Slice, func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
			el := target.Type().Elem()
			sep, ok := ctx.Value.Field.Tag.Lookup("sep")
			if !ok {
				sep = ","
			}
			childScanner := Scan(strings.Split(scan.PopValue("slice"), sep)...)
			childDecoder := DecoderForType(el)
			for childScanner.Peek().Type != EOLToken {
				childValue := reflect.New(el).Elem()
				err := childDecoder.Decode(ctx, childScanner, childValue)
				if err != nil {
					return err
				}
				target.Set(reflect.Append(target, childValue))
			}
			return nil
		}),
	}
}

var missingDecoder DecoderFunc = func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
	return fmt.Errorf("no decoder for %q (of type %T) for field %q", target.String(), target.Type(), ctx.Value.Field.Name)
}
