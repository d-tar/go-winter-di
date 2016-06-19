package wntr

import (
	"log"
	"errors"
	"reflect"
)

//Marker interface to be implemented to
//control lifecycle for each component in context
type ComponentLifecycle interface {
	//1st initialization phase: configure as type, before context configuration
	OnPrepareComponent(c *Component) error
	//2nd initialization phase: configure as component of context
	OnComponentReady(c *Component) error
	//3rd phase: dispose component
	OnDestroyComponent(c *Component) error
}

type ComponentConfigurer interface {
	ConfigureComponent(c *Component) error
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
	componentStates     map[*Component]uint32
	//list of components in acquisition order
	//  * this is an inversion of disposition order
	componentOrder      []*Component
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
		componentStates:make(map[*Component]uint32),
	}
}

const(
	stateNotWired = iota
	stateResolving
	stateResolved
)

func  (h *StandardLifecycle) OnComponentRegistered(c *Component){
	if p,ok:=c.Inst.(ComponentLifecycle);ok{
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

func (h *StandardLifecycle) ConfigureComponent(c *Component) error{
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

func (h *StandardLifecycle) deconstructComponent(c * Component){
	log.Println("Deconstructing component",c.ty)
	for _,p:= range h.lifecycleProcessors{
		p.OnDestroyComponent(c)
	}
}

func (h *TwoPhaseInitializer) OnComponentReady(c *Component) error {

	if v, ok := c.Inst.(PostInitable); ok {
		v.PostInit()
	}

	return nil
}

func (h *TwoPhaseInitializer) OnPrepareComponent(c *Component) error {
	if v, ok := c.Inst.(PreInitable); ok {
		v.PreInit()
	}
	return nil
}

func (h *TwoPhaseInitializer) OnDestroyComponent(c *Component) error {
	if v, ok := c.Inst.(PreDestroyable); ok {
		v.PreDestroy()
	}
	return nil
}
