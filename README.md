<!-- markdownlint-disable MD013 MD033 -->
<p align="center"><img width="90%" src="kong.png" /></p>

# Kong is a command-line parser for Go [![CircleCI](https://img.shields.io/circleci/project/github/alecthomas/kong.svg)](https://circleci.com/gh/alecthomas/kong)

<!-- MarkdownTOC -->

1. [Introduction](#introduction)
1. [Help](#help)
1. [Flags](#flags)
1. [Commands and sub-commands](#commands-and-sub-commands)
1. [Branching positional arguments](#branching-positional-arguments)
1. [Terminating positional arguments](#terminating-positional-arguments)
1. [Slices](#slices)
1. [Maps](#maps)
1. [Custom named types](#custom-named-types)
1. [Supported tags](#supported-tags)
1. [Modifying Kong's behaviour](#modifying-kongs-behaviour)
    1. [`Name(help)` and `Description(help)` - set the application name description](#namehelp-and-descriptionhelp---set-the-application-name-description)
    1. [`Configuration(loader, paths...)` - load defaults from configuration files](#configurationloader-paths---load-defaults-from-configuration-files)
    1. [`Resolver(...)` - support for default values from external sources](#resolver---support-for-default-values-from-external-sources)
    1. [`*Mapper(...)` - customising how the command-line is mapped to Go values](#mapper---customising-how-the-command-line-is-mapped-to-go-values)
    1. [`HelpOptions(HelpPrinterOptions)` and `Help(HelpFunc)` - customising help](#helpoptionshelpprinteroptions-and-helphelpfunc---customising-help)
    1. [`Hook(&field, HookFunc)` - callback hooks to execute when the command-line is parsed](#hookfield-hookfunc---callback-hooks-to-execute-when-the-command-line-is-parsed)
    1. [Other options](#other-options)

<!-- /MarkdownTOC -->

## Introduction

Kong aims to support arbitrarily complex command-line structures with as little developer effort as possible.

To achieve that, command-lines are expressed as Go types, with the structure and tags directing how the command line is mapped onto the struct.

For example, the following command-line:

    shell rm [-f] [-r] <paths> ...
    shell ls [<paths> ...]

Can be represented by the following command-line structure:

```go
package main

import "github.com/alecthomas/kong"

var CLI struct {
  Rm struct {
    Force     bool `help:"Force removal."`
    Recursive bool `help:"Recursively remove files."`

    Paths []string `arg help:"Paths to remove." type:"path"`
  } `cmd help:"Remove files."`

  Ls struct {
    Paths []string `arg optional help:"Paths to list." type:"path"`
  } `cmd help:"List paths."`
}

func main() {
  kong.Parse(&CLI)
}
```

## Help

Help is automatically generated. With no other arguments provided, help will display a full summary of all available commands.

eg.

    $ shell --help
    usage: shell <command>

    A shell-like example app.

    Flags:
      --help   Show context-sensitive help.
      --debug  Debug mode.

    Commands:
      rm <paths> ...
        Remove files.

      ls [<paths> ...]
        List paths.

If a command is provided, the help will show full detail on the command including all available flags.

eg.

    $ shell --help rm
    usage: shell rm <paths> ...

    Remove files.

    Arguments:
      <paths> ...  Paths to remove.

    Flags:
          --debug        Debug mode.

      -f, --force        Force removal.
      -r, --recursive    Recursively remove files.

## Flags

Any [mapped](#mapper---customising-how-the-command-line-is-mapped-to-go-values) field in the command structure *not* tagged with `cmd` or `arg` will be a flag. Flags are optional by default.

eg. The command-line `app [--flag="foo"]` can be represented by the following.

```go
type CLI struct {
  Flag string
}
```

## Commands and sub-commands

Sub-commands are specified by tagging a struct field with `cmd`. Kong supports arbitrarily nested commands.

eg. The following struct represents the CLI structure `command [--flag="str"] sub-command`.

```go
type CLI struct {
  Command struct {
    Flag string

    SubCommand struct {
    } `cmd`
  } `cmd`
}
```

## Branching positional arguments

In addition to sub-commands, structs can also be configured as branching positional arguments.

This is achieved by tagging an [unmapped](#mapper---customising-how-the-command-line-is-mapped-to-go-values) nested struct field with `arg`, then including a positional argument field inside that struct _with the same name_. For example, the following command structure:

    app rename <name> to <name>

Can be represented with the following:

```go
var CLI struct {
  Rename struct {
    Name struct {
      Name string `arg` // <-- NOTE: identical name to enclosing struct field.
      To struct {
        Name struct {
          Name string `arg`
        } `arg`
      } `cmd`
    } `arg`
  } `cmd`
}
```

This looks a little verbose in this contrived example, but typically this will not be the case.

## Terminating positional arguments

If a [mapped type](#mapper---customising-how-the-command-line-is-mapped-to-go-values) is tagged with `arg` it will be treated as the final positional values to be parsed on the command line.

If a positional argument is a slice, all remaining arguments will be appended to that slice.

## Slices

Slice values are treated specially. First the input is split on the `sep:"<rune>"` tag (defaults to `,`), then each element is parsed by the slice element type and appended to the slice. If the same value is encountered multiple times, elements continue to be appended.

To represent the following command-line:

    cmd ls <file> <file> ...

You would use the following:

```go
var CLI struct {
  Ls struct {
    Files []string `arg type:"existingfile"`
  } `cmd`
}
```

## Maps

Maps are similar to slices except that only one key/value pair can be assigned per value, and the `sep` tag denotes the assignment character and defaults to `=`.

To represent the following command-line:

    cmd config set <key>=<value> <key>=<value> ...

You would use the following:

```go
var CLI struct {
  Config struct {
    Set struct {
      Config map[string]float64 `arg type:"file:"`
    } `cmd`
  } `cmd`
}
```

## Custom named types

Kong includes a number of builtin custom type mappers. These can be used by
specifying the tag `type:"<type>"`. They are registered with the option
function `NamedMapper(name, mapper)`.

| Name              | Description                                       |
|-------------------|---------------------------------------------------|
| `file`            | A path. ~ expansion is applied.                   |
| `existingfile`    | An existing path. ~ expansion is applied.         |
| `existingdir`     | An existing directory. ~ expansion is applied.    |


Slices and maps treat type tags specially. For slices, the `type:""` tag
specifies the element type. For maps, the tag has the format
`tag:"[<key>]:[<value>]"` where either may be omitted.


## Supported tags

Tags can be in two forms:

1. Standard Go syntax, eg. `kong:"required,name='foo'"`.
2. Bare tags, eg. `required name:"foo"`

Both can coexist with standard Tag parsing.

| Tag                    | Description                                 |
| -----------------------| ------------------------------------------- |
| `cmd`                  | If present, struct is a command.            |
| `arg`                  | If present, field is an argument.           |
| `env:"X"`              | Specify envar to use for default value.
| `name:"X"`             | Long name, for overriding field name.       |
| `help:"X"`             | Help text.                                  |
| `type:"X"`             | Specify [named types](#custom-named-types) to use.                |
| `placeholder:"X"`      | Placeholder text.                           |
| `default:"X"`          | Default value.                              |
| `short:"X"`            | Short name, if flag.                        |
| `required`             | If present, flag/arg is required.           |
| `optional`             | If present, flag/arg is optional.           |
| `hidden`               | If present, flag is hidden.                 |
| `format:"X"`           | Format for parsing input, if supported.     |
| `sep:"X"`              | Separator for sequences (defaults to ","). May be `none` to disable splitting. |

## Modifying Kong's behaviour

Each Kong parser can be configured via functional options passed to `New(cli interface{}, options...Option)`.

The full set of options can be found [here](https://godoc.org/github.com/alecthomas/kong#Option).

### `Name(help)` and `Description(help)` - set the application name description

Set the application name and/or description.

The name of the application will default to the binary name, but can be overridden with `Name(name)`.

As with all help in Kong, text will be wrapped to the terminal.

### `Configuration(loader, paths...)` - load defaults from configuration files

This option provides Kong with support for loading defaults from a set of configuration files. Each file is opened, if possible, and the loader called to create a resolver for that file.

eg.

```go
kong.Parse(&cli, kong.Configuration(kong.JSON, "/etc/myapp.json", "~/.myapp.json"))
```

### `Resolver(...)` - support for default values from external sources

Resolvers are Kong's extension point for providing default values from external sources. As an example, support for environment variables via the `env` tag is provided by a resolver. There's also a builtin resolver for JSON configuration files.

Example resolvers can be found in [resolver.go](https://github.com/alecthomas/kong/blob/master/resolver.go).

### `*Mapper(...)` - customising how the command-line is mapped to Go values

Command-line arguments are mapped to Go values via the Mapper interface:

```go
// A Mapper knows how to map command-line input to Go.
type Mapper interface {
  // Decode scan into target.
  //
  // "ctx" contains context about the value being decoded that may be useful
  // to some mapperss.
  Decode(ctx *MapperContext, scan *Scanner, target reflect.Value) error
}
```

All builtin Go types (as well as a bunch of useful stdlib types like `time.Time`) have mappers registered by default. Mappers for custom types can be added using `kong.??Mapper(...)` options. Mappers are applied to fields in four ways:

1. `NamedMapper(string, Mapper)` and using the tag key `type:"<name>"`.
2. `KindMapper(reflect.Kind, Mapper)`.
3. `TypeMapper(reflect.Type, Mapper)`.
4. `ValueMapper(interface{}, Mapper)`, passing in a pointer to a field of the grammar.

### `HelpOptions(HelpPrinterOptions)` and `Help(HelpFunc)` - customising help

The default help output is usually sufficient, but if not there are two solutions.

1. Use `HelpOptions(HelpPrinterOptions)` to configure how help is formatted (see [HelpPrinterOptions](https://godoc.org/github.com/alecthomas/kong#HelpPrinterOptions) for details).
2. Custom help can be wired into Kong via the `Help(HelpFunc)` option. The `HelpFunc` is passed a `Context`, which contains the parsed context for the current command-line. See the implementation of `PrintHelp` for an example.

### `Hook(&field, HookFunc)` - callback hooks to execute when the command-line is parsed

Hooks are callback functions that are bound to a node in the command-line and executed at parse time, before structural validation and assignment.

eg.

```go
app := kong.Must(&CLI, kong.Hook(&CLI.Debug, func(ctx *Context, path *Path) error {
  log.SetLevel(DEBUG)
  return nil
}))
```

Note: it is generally more advisable to use an imperative approach to building command-lines, eg.

```go
if CLI.Debug {
  log.SetLevel(DEBUG)
}
```

But under some circumstances, hooks are the right choice.

### Other options

The full set of options can be found [here](https://godoc.org/github.com/alecthomas/kong#Option).
