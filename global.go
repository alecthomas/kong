package kong

import "os"

// Parse constructs a new parser and parses the default command-line.
func Parse(cli interface{}, options ...Option) {
	parser, err := New(cli, options...)
	parser.FatalIfErrorf(err)
	_, err = parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
}
