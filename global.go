package kong

import (
	"os"
)

// App is the default global instance. It is populated by Parse().
var App *Kong

// Parse constructs a new parser and parses the default command-line.
func Parse(cli interface{}, options ...Option) string {
	parser, err := New(cli, options...)
	if err != nil {
		panic(err)
	}
	App = parser
	cmd, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	return cmd
}

func FatalIfErrorf(err error, args ...interface{}) {
	if App == nil {
		panic("call kong.Parse() before using kong.FatalIfErrorf()")
	}
	App.FatalIfErrorf(err, args...)
}

func Errorf(format string, args ...interface{}) {
	if App == nil {
		panic("call kong.Parse() before using kong.Errorf()")
	}
	App.Errorf(format, args...)
}

func Printf(format string, args ...interface{}) {
	if App == nil {
		panic("call kong.Parse() before using kong.Printf()")
	}
	App.Printf(format, args...)
}
