package wntr

import (
	"log"
)

//Marker interface to be implemented to
//control lifecycle for each component in context
type ComponentLifecycle interface {
	//1st initialization phase: configure as type, before context configuration
	OnPrepareComponent(c Component) error
	//2nd initialization phase: configure as component of context
	OnComponentReady(c Component) error
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

//Default TwoPhase lifecycle implementation
// 1st phase - is 'before component configured'
// 2nd phase - is 'after component configured'
//
// Required to add ComponentLifecycle components to context
type StandardLifecycle struct {
}

//Component that implements PreInitable and PostInitable interfaces behaviour
type TwoPhaseInitializer struct {
}

func assertTypeValid() {
	var _ ComponentLifecycle = &TwoPhaseInitializer{}
}

/* Implementation */

func (h *StandardLifecycle) OnStartContext(ctx *MutableContext) error {

	p, c := make([]ComponentLifecycle, 0), make([]Component, 0)

	for _, comp := range ctx.components {
		c = append(c, comp)
		if v, ok := comp.inst.(ComponentLifecycle); ok {
			p = append(p, v)
		}
	}

	for _, proc := range p {
		for _, comp := range c {
			if err := proc.OnPrepareComponent(comp); err != nil {
				return err
			}
		}
	}

	for _, proc := range p {
		for _, comp := range c {
			if err := proc.OnComponentReady(comp); err != nil {
				return err
			}
		}
	}

	log.Println("StandardLifecycle: ", len(c), "components", len(p), "processors")

	return nil
}

func (h *StandardLifecycle) OnStopContext(ctx *MutableContext) error {
	return nil
}

func (h *TwoPhaseInitializer) OnComponentReady(c Component) error {

	if v, ok := c.inst.(PostInitable); ok {
		v.PostInit()
	}

	return nil
}

func (h *TwoPhaseInitializer) OnPrepareComponent(c Component) error {
	if v, ok := c.inst.(PreInitable); ok {
		v.PreInit()
	}
	return nil
}
