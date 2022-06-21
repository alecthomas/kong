package kong

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestScannerTake(t *testing.T) {
	s := Scan("a", "b", "c", "-")
	assert.Equal(t, "a", s.Pop().Value)
	assert.Equal(t, "b", s.Pop().Value)
	assert.Equal(t, "c", s.Pop().Value)
	hyphen := s.Pop()
	assert.Equal(t, PositionalArgumentToken, hyphen.InferredType())
	assert.Equal(t, EOLToken, s.Pop().Type)
}

func TestScannerPeek(t *testing.T) {
	s := Scan("a", "b", "c")
	assert.Equal(t, s.Peek().Value, "a")
	assert.Equal(t, s.Pop().Value, "a")
	assert.Equal(t, s.Peek().Value, "b")
	assert.Equal(t, s.Pop().Value, "b")
	assert.Equal(t, s.Peek().Value, "c")
	assert.Equal(t, s.Pop().Value, "c")
	assert.Equal(t, s.Peek().Type, EOLToken)
}
