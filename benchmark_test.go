package kong

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func BenchmarkKong_interpolate(b *testing.B) {
	prepareKong := func(t testing.TB, count int) *Kong {
		t.Helper()
		k := &Kong{
			vars:     make(Vars, count),
			registry: NewRegistry().RegisterDefaults(),
		}
		for i := 0; i < count; i++ {
			helpVar := fmt.Sprintf("help_param%d", i)
			k.vars[helpVar] = strconv.Itoa(i)
		}
		grammar := &struct {
			Param0 string `help:"${help_param0}"`
		}{}
		model, err := build(k, grammar)
		assert.NoError(t, err)
		for i := 0; i < count; i++ {
			model.Node.Flags = append(model.Node.Flags, &Flag{
				Value: &Value{
					Help: fmt.Sprintf("${help_param%d}", i),
					Tag:  newEmptyTag(),
				},
			})
		}
		k.Model = model
		return k
	}

	for _, count := range []int{5, 500, 5000} {
		count := count
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			var err error
			k := prepareKong(b, count)
			for i := 0; i < b.N; i++ {
				err = k.interpolate(k.Model.Node)
			}
			assert.NoError(b, err)
			b.ReportAllocs()
		})
	}
}

func Benchmark_interpolateValue(b *testing.B) {
	varsLen := 10000
	k := &Kong{
		vars:     make(Vars, 10000),
		registry: NewRegistry().RegisterDefaults(),
	}
	for i := 0; i < varsLen; i++ {
		helpVar := fmt.Sprintf("help_param%d", i)
		k.vars[helpVar] = strconv.Itoa(i)
	}
	grammar := struct {
		Param9999 string `kong:"cmd,help=${help_param9999}"`
	}{}
	model, err := build(k, &grammar)
	if err != nil {
		b.FailNow()
	}
	k.Model = model
	flag := k.Model.Flags[0]
	for i := 0; i < b.N; i++ {
		err = k.interpolateValue(flag.Value, k.vars)
		if err != nil {
			b.FailNow()
		}
	}
	b.ReportAllocs()
}
