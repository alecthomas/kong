package kong

import (
	"fmt"
	"math/bits"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type DecoderContext struct {
	// Value being decoded into.
	Value *Value
}

// A Mapper represents how a field is mapped from command-line values to Go.
//
// Mappers can be associated with concrete fields via pointer, reflect.Type, reflect.Kind, or via a "type" tag.
type Mapper interface {
	Decode(ctx *DecoderContext, scan *Scanner, target reflect.Value) error
}

// A BoolMapper is a Mapper to a value that is a boolean.
type BoolMapper interface {
	Mapper
	IsBool() bool
}

type MapperFunc func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error

func (d MapperFunc) Decode(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
	return d(ctx, scan, target)
}

// A Registry encapsulates a set of fields and lookups to resolve them.
type Registry struct {
	names  map[string]Mapper
	types  map[reflect.Type]Mapper
	kinds  map[reflect.Kind]Mapper
	values map[reflect.Value]Mapper
}

func NewRegistry() *Registry {
	return &Registry{
		names:  map[string]Mapper{},
		types:  map[reflect.Type]Mapper{},
		kinds:  map[reflect.Kind]Mapper{},
		values: map[reflect.Value]Mapper{},
	}
}

// ForNamedType finds a mapper for a value with a user-specified type.
//
// Will return nil if a mapper can not be determined.
func (d *Registry) ForNamedType(name string, value reflect.Value) Mapper {
	if mapper, ok := d.names[name]; ok {
		return mapper
	}
	return d.ForValue(value)
}

func (d *Registry) ForValue(value reflect.Value) Mapper {
	if mapper, ok := d.values[value]; ok {
		return mapper
	}
	return d.ForType(value.Type())
}

// DecoderForType finds a mapper from a type or kind.
//
// Will return nil if a mapper can not be determined.
func (d *Registry) ForType(typ reflect.Type) Mapper {
	var mapper Mapper
	var ok bool
	if mapper, ok = d.types[typ]; ok {
		return mapper
	} else if mapper, ok = d.kinds[typ.Kind()]; ok {
		return mapper
	}
	return nil
}

func (d *Registry) RegisterKind(kind reflect.Kind, mapper Mapper) *Registry {
	d.kinds[kind] = mapper
	return d
}

// RegisterName registeres a mapper to be used if the value mapper has a "type" tag matching name.
//
// eg.
//
// 		Mapper string `kong:"type='colour'`
//   	registry.RegisterName("colour", ...)
func (d *Registry) RegisterName(name string, mapper Mapper) *Registry {
	d.names[name] = mapper
	return d
}

func (d *Registry) RegisterType(typ reflect.Type, mapper Mapper) *Registry {
	d.types[typ] = mapper
	return d
}

// RegisterValue registers a mapper by a pointer to the mapper value.
func (d *Registry) RegisterValue(ptr interface{}, mapper Mapper) *Registry {
	key := reflect.ValueOf(ptr)
	if key.Kind() != reflect.Ptr {
		panic("expected a pointer")
	}
	key = key.Elem()
	d.values[key] = mapper
	return d
}

func (d *Registry) RegisterDefaults() *Registry {
	return d.RegisterKind(reflect.Int, intDecoder(bits.UintSize)).
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
		RegisterKind(reflect.String, MapperFunc(func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
			target.SetString(scan.PopValue("string"))
			return nil
		})).
		RegisterKind(reflect.Bool, boolMapper{}).
		RegisterType(reflect.TypeOf(time.Time{}), timeDecoder()).
		RegisterType(reflect.TypeOf(time.Duration(0)), durationDecoder()).
		RegisterKind(reflect.Slice, sliceDecoder(d))
}

type boolMapper struct{}

func (boolMapper) Decode(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
	target.SetBool(true)
	return nil
}
func (boolMapper) IsBool() bool { return true }

func durationDecoder() MapperFunc {
	return func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
		d, err := time.ParseDuration(scan.PopValue("duration"))
		if err != nil {
			return err
		}
		target.Set(reflect.ValueOf(d))
		return nil
	}
}

func timeDecoder() MapperFunc {
	return func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
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
}

func intDecoder(bits int) MapperFunc {
	return func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
		value := scan.PopValue("int")
		n, err := strconv.ParseInt(value, 10, bits)
		if err != nil {
			return fmt.Errorf("invalid int %q", value)
		}
		target.SetInt(n)
		return nil
	}
}

func uintDecoder(bits int) MapperFunc {
	return func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
		value := scan.PopValue("uint")
		n, err := strconv.ParseUint(value, 10, bits)
		if err != nil {
			return fmt.Errorf("invalid uint %q", value)
		}
		target.SetUint(n)
		return nil
	}
}

func floatDecoder(bits int) MapperFunc {
	return func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
		value := scan.PopValue("float")
		n, err := strconv.ParseFloat(value, bits)
		if err != nil {
			return fmt.Errorf("invalid float %q", value)
		}
		target.SetFloat(n)
		return nil
	}
}

func sliceDecoder(d *Registry) MapperFunc {
	return func(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
		el := target.Type().Elem()
		sep := ctx.Value.Tag.Sep
		var childScanner *Scanner
		if ctx.Value.Flag != nil {
			// If decoding a flag, we need an argument.
			childScanner = Scan(strings.Split(scan.PopValue("list"), sep)...)
		} else {
			tokens := scan.PopUntil(func(t Token) bool { return !t.IsValue() })
			childScanner = Scan(tokens...)
		}
		childDecoder := d.ForType(el)
		if childDecoder == nil {
			return fmt.Errorf("no mapper for element type of %s", target.Type())
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
}
