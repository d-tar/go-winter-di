package wntr

import (
	"fmt"
	"log"
	"reflect"
)

type ForkedLifecycle struct {
	*StandardLifecycle
	Parent ConfiguredContext
}

var _ ConfiguredContext = (*ForkedLifecycle)(nil)
var _ CtxEventHandler = (*ForkedLifecycle)(nil)

func (this *ForkedLifecycle) FindComponentsByType(t reflect.Type) []Component {
	comps := this.StandardLifecycle.FindComponentsByType(t)

	//If we've resolved components by our lifecyle - well
	if len(comps) > 0 {
		return comps
	}

	log.Println("Falling back to parent")

	//Otherwise, let's ask for components our parent
	return this.Parent.FindComponentsByType(t)
}

func NewForkedLifecycle(Parent ConfiguredContext) *ForkedLifecycle {
	return &ForkedLifecycle{
		Parent:            Parent,
		StandardLifecycle: NewStandardLifecycle(),
	}
}

var gConfiguredContextType reflect.Type = reflect.TypeOf((*ConfiguredContext)(nil)).Elem()

func ForkContext(ctxToFork Context) (Context, error) {
	ctx := newMutableContext()

	mutCtx := ctxToFork.(*MutableContext)

	comps := mutCtx.FindComponentsByType(gConfiguredContextType)

	if len(comps) != 1 {
		return nil, fmt.Errorf("Strage parent context: Expected 1 instance of ConfiguredContext, got: %v", len(comps))
	}

	ctx.RegisterComponent(NewForkedLifecycle(comps[0].Instance().(ConfiguredContext)))

	//Enable PreInit & PostInit
	ctx.RegisterComponent(&TwoPhaseInitializer{})
	//Enable autowired
	ctx.RegisterComponent(NewAutowiringProcessor())

	return ctx, nil
}
