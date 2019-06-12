package kong

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestApplyDefaults(t *testing.T) {
	type CLI struct {
		Str      string        `default:"str"`
		Duration time.Duration `default:"30s"`
	}
	tests := []struct {
		name     string
		target   CLI
		expected CLI
	}{
		{name: "DefaultsWhenNotSet",
			expected: CLI{Str: "str", Duration: time.Second * 30}},
		{name: "PartiallySetDefaults",
			target:   CLI{Duration: time.Second},
			expected: CLI{Str: "str", Duration: time.Second}},
	}
	for _, tt := range tests {
		// nolint: scopelint
		t.Run(tt.name, func(t *testing.T) {
			err := ApplyDefaults(&tt.target)
			require.NoError(t, err)
			require.Equal(t, tt.expected, tt.target)
		})
	}
}
