# An interactive SSH server

In addition to command-lines, Kong can be used interactively. This example 
serves a Kong command-line over SSH.

Run with `go run .` then ssh to it like so:

```
$ ssh -p 6740 127.0.0.1
Welcome!
> ?

Example using Kong for interactive command parsing.

Commands:
  help [<command> ...]
    Show help.

  status
    Show server status.

> status
OK
> help status

Show server status.

Flags:
  -v, --verbose    Show verbose status information.

> status 
OK
> status -v
OK
> status foo
error: unexpected argument foo

Show server status.

Flags:
> ^D
```
