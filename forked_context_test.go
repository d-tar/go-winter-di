package wntr

import (
	"fmt"
	"reflect"
	"testing"
)

type Bean struct {
	Value string
}

var BaseContext struct {
	Bean Bean
}

var NestedContext struct {
	Bean *Bean `inject:"t"`
}

func TestFork(t *testing.T) {

	ctx := ContextOrPanic(&BaseContext)

	nestedCtx, err := ForkContext(ctx)

	if err != nil {
		t.Fatal(err)
	}

	if err := PopulateContextFromDefinitions(nestedCtx, &NestedContext); err != nil {
		t.Fatal(err)
	}

	if err := nestedCtx.Start(); err != nil {
		t.Fatal(err)
	}

	if NestedContext.Bean != &BaseContext.Bean {
		t.Fatal("Nested context was not configured")
	}

	nestedCtx.Stop()
}

func Trampoline(d *struct {
	test  string
	test2 *interface{}
}) {
	fmt.Print(d)
}

func TestReflect(t *testing.T) {

	ty := reflect.TypeOf(Trampoline)

	v := reflect.New(ty.In(0).Elem())

	reflect.ValueOf(Trampoline).Call([]reflect.Value{v})
}
