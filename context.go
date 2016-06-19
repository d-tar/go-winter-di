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

type ComponentRegisterAware interface {
	OnComponentRegistered(*ComponentImpl)
}

//Public default context constructor
//By default uses StandardLifecycle
func NewContext() (c Context, e error) {
	c = &MutableContext{
		components: make([]*ComponentImpl, 0),
	}

	c.RegisterComponent(c)
	c.RegisterComponent(NewStandardLifecycle())

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
	components []*ComponentImpl //List of registered components
	registrationHandlers []ComponentRegisterAware
}

//Simple holder for registered components
type ComponentImpl struct {
	inst interface{}
	ty   reflect.Type
	tags string
}

type Component interface {
	Instance() interface{}
	Type() reflect.Type
	Tags() reflect.StructTag
}

func (c *MutableContext) RegisterComponent(value interface{}) {
	c.RegisterComponentWithTags(value,"")
}

func (c *MutableContext) RegisterComponentWithTags(value interface{},tags string) {
	t := reflect.TypeOf(value)
	log.Println("Registering component ", t,"tags",tags)

	comp:=&ComponentImpl{value, t,tags}

	c.components = append(c.components, comp)

	if v, ok := value.(ContextAware); ok {
		v.SetContext(c)
	}

	for _,handler := range c.registrationHandlers{
		handler.OnComponentRegistered(comp)
	}

	if v, ok := value.(ComponentRegisterAware); ok {
		c.registrationHandlers = append(c.registrationHandlers,v)
	}


}

func (c *MutableContext) Start() error {
	cnt := 0
	for _, i := range c.components {
		if v, ok := i.inst.(CtxEventHandler); ok {
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
		if v, ok := i.inst.(CtxEventHandler); ok {
			if err := v.OnStopContext(c); err != nil {
				return err
			}

			cnt++
		}
	}
	log.Println("Context stopped", cnt, "processors called")
	return nil
}

func (c *MutableContext) FindComponentsByType(t reflect.Type) []*ComponentImpl {
	r := make([]*ComponentImpl, 0)

	for _, v := range c.components {
		if v.ty.AssignableTo(t) {
			r = append(r, v)
		}
	}
	return r
}


func (t *ComponentImpl)Instance() interface{}{
	return t.inst;
}

func (t *ComponentImpl) Type() reflect.Type{
	return t.ty
}

func (t *ComponentImpl) Tags() reflect.StructTag{
	return reflect.StructTag(t.tags)
}