package kong

import "reflect"

type Decoder interface {
	Decode(input string, target reflect.Value) error
}

type DecoderFunc func(input string, target reflect.Value) error

func (d DecoderFunc) Decode(input string, target reflect.Value) error { return d(input, target) }

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
	RegisterDecoder()
}
