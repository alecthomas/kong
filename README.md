<p align="center"><img src="kong.png" /></p>

# Kong is a command-line parser for Go [![CircleCI](https://circleci.com/gh/alecthomas/kong.svg?style=svg&circle-token=477fecac758383bf281453187269b913130f17d2)](https://circleci.com/gh/alecthomas/kong)

<!-- MarkdownTOC -->

1. [Introduction](#introduction)
1. [Help](#help)
1. [Flags](#flags)
1. [Commands and sub-commands](#commands-and-sub-commands)
1. [Supported tags](#supported-tags)
1. [Configuring Kong](#configuring-kong)
    1. [`*Mapper(...)` - customising how the command-line is mapped to Go values](#mapper---customising-how-the-command-line-is-mapped-to-go-values)
    1. [`Help(HelpFunc)` - customising help](#helphelpfunc---customising-help)
    1. [`Hook(&field, HookFunc)` - callback hooks to execute when the command-line is parsed](#hookfield-hookfunc---callback-hooks-to-execute-when-the-command-line-is-parsed)

<!-- /MarkdownTOC -->


## Introduction

Kong aims to support arbitrarily complex command-line structures with as little developer effort as possible.

To achieve that, command-lines are expressed as Go types, with the structure and tags directing how the command line is mapped onto the struct.

For example, the following command-line:

```
shell rm [-f] [-r] <paths> ...
shell ls [<paths> ...]
```

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

```
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
```

If a command is provided, the help will show full detail on the command including all available flags.

eg.

```
$ shell --help rm
usage: shell rm <paths> ...

Remove files.

Arguments:
  <paths> ...  Paths to remove.

Flags:
      --debug        Debug mode.

  -f, --force        Force removal.
  -r, --recursive    Recursively remove files.
```

## Flags

Any field in the command structure *not* tagged with `cmd` or `arg` will be a flag. Flags are optional by default.

## Commands and sub-commands

Kong supports arbitrarily nested commands and positional arguments. Nested structs tagged with `cmd` will be treated as commands.

Arguments can also optionally have children, in order to support commands like the following:

```
app rename <name> to <name>
```

This is achieved by tagging a nested struct with `arg`, then including a positional argument field inside that struct with the same name. For example:

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

## Supported tags

Tags can be in two forms:

1. Standard Go syntax, eg. `kong:"required,name='foo'"`.
2. Bare tags, eg. `required name:"foo"`

Both can coexist with standard Tag parsing.

| Tag                    | Description                                 |
| -----------------------| ------------------------------------------- |
| `cmd`                  | If present, struct is a command.            |
| `arg`                  | If present, field is an argument.           |
| `type:"X"`             | Specify named Mapper to use.                |
| `help:"X"`             | Help text.                                  |
| `placeholder:"X"`      | Placeholder text.                           |
| `default:"X"`          | Default value.                              |
| `short:"X"`            | Short name, if flag.                        |
| `name:"X"`             | Long name, for overriding field name.       |
| `required`             | If present, flag/arg is required.           |
| `optional`             | If present, flag/arg is optional.           |
| `hidden`               | If present, flag is hidden.                 |
| `format:"X"`           | Format for parsing input, if supported.     |
| `sep:"X"`              | Separator for sequences (defaults to ",")   |

## Configuring Kong

Each Kong parser can be configured via functional options passed to `New(cli interface{}, options...Option)`. The full set of options can be found in `options.go`.

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

All builtin Go types (as well as a bunch of useful stdlib types like `time.Time`) have mapperss registered by default. Mappers for custom types can be added using `kong.??Mapper(...)` options. Mappers are applied to fields in four ways:

1. `NamedMapper(string, Mapper)` and using the tag key `type:"<name>"`.
2. `KindMapper(reflect.Kind, Mapper)`.
3. `TypeMapper(reflect.Type, Mapper)`.
4.  `ValueMapper(interface{}, Mapper)`, passing in a pointer to a field of the grammar.


### `Help(HelpFunc)` - customising help

Custom help can be wired into Kong via the `Help(HelpFunc)` option. The `HelpFunc` is passed a `Context`, which contains the parsed context for the current command-line. See the implementation of `PrintHelp` for an example.

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
