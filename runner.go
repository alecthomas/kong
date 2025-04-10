package kong

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Runner struct {
	Context *Context
	Node    *Node
	Next    *Runner

	method reflect.Value
}

func (e *Runner) prepareBindings(binds ...any) (bindings, error) {
	resultBindings := e.Context.Kong.bindings.clone().merge(e.Context.bindings)
	for p := e.Node; p != nil; p = p.Parent {
		resultBindings = resultBindings.add(p.Target.Addr().Interface())
		// Try value and pointer to value.
		for _, p := range []reflect.Value{p.Target, p.Target.Addr()} {
			t := p.Type()
			for i := 0; i < p.NumMethod(); i++ {
				methodt := t.Method(i)
				if strings.HasPrefix(methodt.Name, "Provide") {
					method := p.Method(i)
					if err := resultBindings.addProvider(method.Interface(), false /* singleton */); err != nil {
						return nil, fmt.Errorf("%s.%s: %w", t.Name(), methodt.Name, err)
					}
				}
			}
		}
	}

	resultBindings.add(binds...)
	resultBindings.add(e.Context)
	resultBindings.add(e)

	return resultBindings, nil
}

func (e *Runner) runCurrent(binds ...any) (err error) {
	resultBindings, err := e.prepareBindings(binds...)
	if err != nil {
		return err
	}
	return callFunction(e.method, resultBindings)
}

func (e *Runner) RunNext(binds ...any) (err error) {
	if e == nil || e.Next == nil {
		return errors.New("no next executor")
	}
	return e.Next.runCurrent(binds...)
}
