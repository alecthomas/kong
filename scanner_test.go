package kong

import (
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
)

func TestScannerTake(t *testing.T) {
	s := Scan("a", "b", "c")
	assert.Assert(t, s.Pop().Value == "a")
	assert.Assert(t, s.Pop().Value == "b")
	assert.Assert(t, s.Pop().Value == "c")
	assert.Assert(t, s.Pop().Type == EOLToken)
}

func TestScannerPeek(t *testing.T) {
	s := Scan("a", "b", "c")
	assert.Assert(t, s.Peek().Value == "a")
	assert.Assert(t, s.Pop().Value == "a")
	assert.Assert(t, s.Peek().Value == "b")
	assert.Assert(t, s.Pop().Value == "b")
	assert.Assert(t, s.Peek().Value == "c")
	assert.Assert(t, s.Pop().Value == "c")
	assert.Assert(t, s.Peek().Type == EOLToken)
}
