package kong

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func Test_helpWriterNoDocToText(t *testing.T) {
	// NOTE: the space after the second break to check if noDocToTextMode
	// respects the user expectation. The go doc.ToText method returns
	// different results if a space is passed after a newline.
	subject := "I am a short string\nand have my own line breaks\n and some info"
	defaultMode := &helpWriter{
		lines:       &([]string{}),
		width:       80,
		HelpOptions: HelpOptions{},
	}
	defaultMode.Wrap(subject)
	require.NotNil(t, defaultMode.lines)
	require.Equal(t, 3, len(*defaultMode.lines))
	actual := *defaultMode.lines
	assert.Equal(t, "I am a short string and have my own line breaks", actual[0])
	assert.Equal(t, "", actual[1])
	assert.Equal(t, "    and some info", actual[2])

	noDocToTextMode := &helpWriter{
		lines: &([]string{}),
		width: 80,
		HelpOptions: HelpOptions{
			NoDocToText: true,
		},
	}
	noDocToTextMode.Wrap(subject)
	require.NotNil(t, noDocToTextMode.lines)
	assert.Equal(t, 3, len(*noDocToTextMode.lines))
	actual = *noDocToTextMode.lines
	assert.Equal(t, "I am a short string", actual[0])
	assert.Equal(t, "and have my own line breaks", actual[1])
	assert.Equal(t, " and some info", actual[2])
}
