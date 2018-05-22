package kong

import (
	"os"
)

// Parse constructs a new parser and parses the default command-line.
func Parse(cli interface{}, options ...Option) string {
	parser, err := New(cli, options...)
	if err != nil {
		panic(err)
	}
	cmd, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	return cmd
}
