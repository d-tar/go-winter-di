package wntr

import (
	"log"
	"reflect"
)

//Public contract for context
//In general:
// 	Context is a holder for components that can be started and stopped
//	See CtxEventHandler interface for context event handling
type Context interface {
	RegisterComponent(interface{})
	RegisterComponentWithTags(interface{},string)
	Start() error
	Stop() error
}

//Private interface for context event handling routines
type CtxEventHandler interface {
	OnStartContext(ctx *MutableContext) error
	OnStopContext(ctx *MutableContext) error
}

//Public contract for components that need to be
//aware of context
//
//  In general CTX can easily be injected by autowiring processor
//  Concrete interface is required by components that cannot be configured
//  by autowiring like autowiring processor itself
type ContextAware interface {
	SetContext(c Context) error
}

//Public default context constructor
//By default uses StandardLifecycle
func NewContext() (c Context, e error) {
	c = &MutableContext{
		components: make([]Component, 0),
	}

	c.RegisterComponent(c)
	c.RegisterComponent(&StandardLifecycle{})

	return c, nil
}

func FastDefaultContext(components ...interface{}) (Context, error) {
	ctx, e := NewContext()
	if e != nil {
		return nil, e
	}

	//Enable PreInit & PostInit
	ctx.RegisterComponent(&TwoPhaseInitializer{})
	//Enable autowired
	ctx.RegisterComponent(NewAutowiringProcessor())

	for _, c := range components {
		ctx.RegisterComponent(c)
	}
	return ctx, e
}

/*  Implementation  */

//Struct implements default context
type MutableContext struct {
	components []Component //List of registered components
}

//Simple holder for registered components
type Component struct {
	Inst interface{}
	ty   reflect.Type
	Tags string
}

func (c *MutableContext) RegisterComponent(value interface{}) {
	c.RegisterComponentWithTags(value,"")
}

func (c *MutableContext) RegisterComponentWithTags(value interface{},tags string) {
	t := reflect.TypeOf(value)
	log.Println("Registering component ", t,"tags",tags)

	c.components = append(c.components, Component{value, t,tags})

	if v, ok := value.(ContextAware); ok {
		v.SetContext(c)
	}
}

func (c *MutableContext) Start() error {
	cnt := 0
	for _, i := range c.components {
		if v, ok := i.Inst.(CtxEventHandler); ok {
			if err := v.OnStartContext(c); err != nil {
				return err
			}

			cnt++
		}
	}
	log.Println("Context started", cnt, "processors called")

	return nil
}

func (c *MutableContext) Stop() error {
	cnt := 0
	for _, i := range c.components {
		if v, ok := i.Inst.(CtxEventHandler); ok {
			if err := v.OnStopContext(c); err != nil {
				return err
			}

			cnt++
		}
	}
	log.Println("Context stopped", cnt, "processors called")
	return nil
}

func (c *MutableContext) FindComponentsByType(t reflect.Type) []Component {
	r := make([]Component, 0)

	for _, v := range c.components {
		if v.ty.ConvertibleTo(t) {
			r = append(r, v)
		}
	}
	return r
}
