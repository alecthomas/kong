package kong

import "os"

func Parse(cli interface{}) {
	parser, err := New("", "", cli)
	parser.FatalIfErrorf(err)
	err = parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
}
