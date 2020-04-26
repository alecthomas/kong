package kong

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScannerTake(t *testing.T) {
	s := Scan("a", "b", "c", "-")
	require.Equal(t, "a", s.Pop().Value)
	require.Equal(t, "b", s.Pop().Value)
	require.Equal(t, "c", s.Pop().Value)
	hyphen := s.Pop()
	require.Equal(t, PositionalArgumentToken, hyphen.InferredType())
	require.Equal(t, EOLToken, s.Pop().Type)
}

func TestScannerPeek(t *testing.T) {
	s := Scan("a", "b", "c")
	require.Equal(t, s.Peek().Value, "a")
	require.Equal(t, s.Pop().Value, "a")
	require.Equal(t, s.Peek().Value, "b")
	require.Equal(t, s.Pop().Value, "b")
	require.Equal(t, s.Peek().Value, "c")
	require.Equal(t, s.Pop().Value, "c")
	require.Equal(t, s.Peek().Type, EOLToken)
}
