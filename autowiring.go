package wntr

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

//Default component that enables `autowire` tag on fields
type AutowiringProcessor struct {
	ctx             *MutableContext
	componentStates map[interface{}]uint32
}

func __test_iface_at_compile_time() {
	var _ ComponentLifecycle = &AutowiringProcessor{}
}

//Default contructor for AutowiringProcessor
func NewAutowiringProcessor() *AutowiringProcessor {
	return &AutowiringProcessor{
		componentStates: make(map[interface{}]uint32),
	}
}

/* implementation */

const(
	stateNotWired = iota
	stateResolving
	stateResolved
)

//Get current module context
func (this *AutowiringProcessor) SetContext(c Context) error {
	if v, ok := c.(*MutableContext); ok {
		this.ctx = v
		return nil
	}

	return errors.New("Unsupported context type")
}

func (this *AutowiringProcessor) OnPrepareComponent(c Component) error {
	return this.autowireInstance(c.inst)
}

func (this *AutowiringProcessor) OnComponentReady(c Component) error {
	return nil
}

//Processes every autowire'ed field for instance
//Implementation is too specific, with lot of assumptions
//   1) components are 'pointers-to-type' only
//   2) autowire marked fields have types:
//              - pointer to struct type
//		- direct interface type
//
//  Other cases may cause unpredictable exceptions
func (this *AutowiringProcessor) autowireInstance(c interface{}) error {

	if s, ok := this.componentStates[c]; !ok {
		this.componentStates[c] = stateResolving
	} else if s == stateResolving {
		return errors.New("Circular dependency")
	} else { //Autowired already
		return nil
	}

	t := reflect.ValueOf(c).Elem()

	for i := 0; i < t.NumField(); i++ {
		fldAccessor := t.Field(i)
		fld := t.Type().Field(i)

		if v := fld.Tag.Get("autowire"); v == "type" {
			if err := this.injectFieldByType(fldAccessor, fld.Type); err != nil {
				return err
			}
		}
	}
	this.componentStates[c] = stateResolved
	return nil
}

func (this *AutowiringProcessor) injectFieldByType(fld reflect.Value, t reflect.Type) error {

	log.Println("Injecting field", fld, "by type")

	candidates := this.ctx.FindComponentsByType(t)

	if len(candidates) == 0 {
		return errors.New(fmt.Sprint("Component not found. Type", t))
	}
	if len(candidates) > 1 {
		return errors.New(fmt.Sprint("Too many components with type. Type", ". Expected 1, Got: ", len(candidates)))
	}

	if !fld.CanSet() {
		return errors.New(fmt.Sprint("Field", fld, " cannot be set. Is it declared public?"))
	}

	r := candidates[0].inst

	if s, ok := this.componentStates[r]; !ok || s != 2 {
		if err := this.autowireInstance(r); err != nil {
			return err
		}
	}

	fld.Set(reflect.ValueOf(r))

	return nil
}
