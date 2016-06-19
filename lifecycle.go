package wntr

import (
	"log"
	"errors"
	"reflect"
	"fmt"
)

//Marker interface to be implemented to
//control lifecycle for each component in context
type ComponentLifecycle interface {
	//1st initialization phase: configure as type, before context configuration
	OnPrepareComponent(c *ComponentImpl) error
	//2nd initialization phase: configure as component of context
	OnComponentReady(c *ComponentImpl) error
	//3rd phase: dispose component
	OnDestroyComponent(c *ComponentImpl) error
}

type ComponentConfigurer interface {
	ConfigureComponent(c *ComponentImpl) error
}

type ConfiguredContext interface {
	FindComponentsByType(reflect.Type) []Component
}

//Interface to be implemented by component
//that needs some initialization before configuring by context
type PreInitable interface {
	PreInit() error
}

//Interface to be implemented by component that needs to
//perform post-initialization after configuring by context
type PostInitable interface {
	PostInit() error
}

type PreDestroyable interface {
	PreDestroy()
}

//Default TwoPhase lifecycle implementation
// 1st phase - is 'before component configured'
// 2nd phase - is 'after component configured'
//
// Required to add ComponentLifecycle components to context
type StandardLifecycle struct {
	lifecycleProcessors []ComponentLifecycle
	//table of current component states
	componentStates     map[*ComponentImpl]uint32
	//list of components in acquisition order
	//  * this is an inversion of disposition order
	componentOrder      []*ComponentImpl

	ctx 			*MutableContext
}

//Component that implements PreInitable and PostInitable interfaces behaviour
type TwoPhaseInitializer struct {
}

func assertTypeValid() {
	var _ ComponentLifecycle = &TwoPhaseInitializer{}
}

/* Implementation */

func NewStandardLifecycle() *StandardLifecycle{
	return &StandardLifecycle{
		componentStates:make(map[*ComponentImpl]uint32),
	}
}

const(
	stateNotWired = iota
	stateResolving
	stateResolved
)

func (this *StandardLifecycle) SetContext(c Context) error{
	if v, ok := c.(*MutableContext); ok {
		this.ctx = v
		return nil
	}
	return fmt.Errorf("Unsupported context type")
}

func  (h *StandardLifecycle) OnComponentRegistered(c *ComponentImpl){
	if p,ok:=c.inst.(ComponentLifecycle);ok{
		h.lifecycleProcessors = append(h.lifecycleProcessors,p)
	}

}

func (h *StandardLifecycle) OnStartContext(ctx *MutableContext) error {

	for _, comp := range ctx.components {
		if err := h.ConfigureComponent(comp);err!=nil{
			return err
		}
	}

	log.Println("StandardLifecycle: ", len(ctx.components), "components", len(h.lifecycleProcessors), "processors")

	return nil
}

func (h *StandardLifecycle) ConfigureComponent(c *ComponentImpl) error{
	log.Println("Configuring component",c.ty)
	if s, ok := h.componentStates[c]; !ok {
		h.componentStates[c] = stateResolving
		log.Println("Start configuring",c,ok,s)
	} else if s == stateResolving {
		return errors.New("Circular dependency")
	} else { //Configured already
		return nil
	}

	for _,p := range h.lifecycleProcessors{
		log.Println(reflect.TypeOf(p).Elem().Name(),c.ty)
		if err := p.OnPrepareComponent(c);err!=nil{
			return err
		}
	}

	for _,p := range h.lifecycleProcessors {
		if err := p.OnComponentReady(c); err != nil {
			return err
		}
	}

	h.componentStates[c] = stateResolved
	log.Println("Component configured",c,h.componentStates)
	h.componentOrder = append(h.componentOrder,c)

	return nil
}

func (h *StandardLifecycle) OnStopContext(ctx *MutableContext) error {
	eIdx := len(h.componentOrder)-1

	for i,_ := range h.componentOrder{
		c:=h.componentOrder[eIdx-i]
		h.deconstructComponent(c)
	}

	return nil
}

func (h *StandardLifecycle) deconstructComponent(c *ComponentImpl){
	log.Println("Deconstructing component",c.ty)
	for _,p:= range h.lifecycleProcessors{
		p.OnDestroyComponent(c)
	}
}

func (h*StandardLifecycle) FindComponentsByType(t reflect.Type) []Component{
	comps:=h.ctx.FindComponentsByType(t)

	r := make([]Component,len(comps),len(comps))

	for i,c := range comps{
		r[i] = c
	}

	return r
}

func (h *TwoPhaseInitializer) OnComponentReady(c *ComponentImpl) error {

	if v, ok := c.inst.(PostInitable); ok {
		v.PostInit()
	}

	return nil
}

func (h *TwoPhaseInitializer) OnPrepareComponent(c *ComponentImpl) error {
	if v, ok := c.inst.(PreInitable); ok {
		v.PreInit()
	}
	return nil
}

func (h *TwoPhaseInitializer) OnDestroyComponent(c *ComponentImpl) error {
	if v, ok := c.inst.(PreDestroyable); ok {
		v.PreDestroy()
	}
	return nil
}
