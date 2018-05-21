# Kong is a command-line parser for Go [![CircleCI](https://circleci.com/gh/alecthomas/kong.svg?style=svg&circle-token=477fecac758383bf281453187269b913130f17d2)](https://circleci.com/gh/alecthomas/kong)

It parses a command-line into a struct. eg.

```go
package main

import "github.com/alecthomas/kong"

var CLI struct {
  Rm struct {
    Force     bool `kong:"help='Force removal.'"`
    Recursive bool `kong:"help='Recursively remove files.'"`

    Paths []string `kong:"help='Paths to remove.',type='path'"`
  } `kong:"help='Remove files.'"`

  Ls struct {
    Paths []string `kong:"help='Paths to list.',type='path'"`
  } `kong:"help='List paths.'"`
}

func main() {
  kong.Parse(&CLI)
}
```

## Decoders

Command-line arguments are mapped to Go values via the Decoder interface:

```go
// A Decoder knows how to decode text into a Go value.
type Decoder interface {
	// Decode scan into target.
	//
	// "ctx" contains context about the value being decoded that may be useful
	// to some decoders.
	Decode(ctx *DecoderContext, scan *Scanner, target reflect.Value) error
}
```

All builtin Go types (as well as a bunch of useful stdlib types like `time.Time`) have decoders registered by default. Decoders for custom types can be added using `kong.RegisterDecoder(decoder)`. Decoders are mapped from fields in three ways:

1. By registering a `kong.NamedDecoder` and using the tag `type:"<name>"`.
2. By registering a `kong.KindDecoder` with a `reflect.Kind`.
3. By registering a `kong.TypeDecoder` with a `reflect.Type`.
