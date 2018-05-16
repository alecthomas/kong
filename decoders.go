package kong

import (
	"fmt"
	"reflect"
	"strconv"
)

type Decoder interface {
	Decode(scan *Scanner, target reflect.Value) error
}

type DecoderFunc func(scan *Scanner, target reflect.Value) error

func (d DecoderFunc) Decode(scan *Scanner, target reflect.Value) error { return d(scan, target) }

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
	intDecoder := NewKindDecoder(reflect.Int, func(scan *Scanner, target reflect.Value) error {
		n, err := strconv.ParseInt(scan.PopValue("int"), 10, 64)
		if err != nil {
			return err
		}
		target.SetInt(n)
		return nil
	})
	uintDecoder := NewKindDecoder(reflect.Uint, func(scan *Scanner, target reflect.Value) error {
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
		reflect.Float32: NewKindDecoder(reflect.Float32, func(scan *Scanner, target reflect.Value) error {
			n, err := strconv.ParseFloat(scan.PopValue("float"), 32)
			if err != nil {
				return err
			}
			target.SetFloat(n)
			return nil
		}),
		reflect.Float64: NewKindDecoder(reflect.Float64, func(scan *Scanner, target reflect.Value) error {
			n, err := strconv.ParseFloat(scan.PopValue("float"), 64)
			if err != nil {
				return err
			}
			target.SetFloat(n)
			return nil
		}),
		reflect.String: NewKindDecoder(reflect.String, func(scan *Scanner, target reflect.Value) error {
			target.SetString(scan.PopValue("string"))
			return nil
		}),
		reflect.Bool: NewKindDecoder(reflect.Bool, func(scan *Scanner, target reflect.Value) error {
			target.SetBool(true)
			return nil
		}),
	}
}

var missingDecoder DecoderFunc = func(scan *Scanner, target reflect.Value) error {
	return fmt.Errorf("no decoder for %q (of type %T)", target.String(), target.Type())
}
