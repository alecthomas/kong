package kong

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := ApplyDefaults(&tt.target)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}
