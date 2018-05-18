package kong

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScannerTake(t *testing.T) {
	s := Scan("a", "b", "c")
	require.Equal(t, s.Pop().Value, "a")
	require.Equal(t, s.Pop().Value, "b")
	require.Equal(t, s.Pop().Value, "c")
	require.Equal(t, s.Pop().Type, EOLToken)
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
