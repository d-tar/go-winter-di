package wntr

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

//Default component that enables `autowire` tag on fields
type AutowiringProcessor struct {
	//points to component context served by this prcessor
	ctx             *MutableContext
	configurer	ComponentConfigurer
}

func __test_iface_at_compile_time() {
	var _ ComponentLifecycle = &AutowiringProcessor{}
}

//Default contructor for AutowiringProcessor
func NewAutowiringProcessor() *AutowiringProcessor {
	return &AutowiringProcessor{}
}

/* implementation */

//Get current module context
func (this *AutowiringProcessor) SetContext(c Context) error {
	if v, ok := c.(*MutableContext); ok {
		this.ctx = v
		ty := reflect.TypeOf((*ComponentConfigurer)(nil)).Elem()

		components:=v.FindComponentsByType(ty)

		if len(components)!=1{
			return fmt.Errorf("Bad context setup. Required 1 instance that implements ComponentConfigurer iface. Actual count: %v",len(components))
		}

		this.configurer = components[0].Inst.(ComponentConfigurer)

		return nil
	}

	return errors.New("Unsupported context type")
}

func (this *AutowiringProcessor) OnPrepareComponent(c *Component) error {
	return this.autowireInstance(c.Inst)
}

func (this *AutowiringProcessor) OnComponentReady(c *Component) error {
	return nil
}

func (this *AutowiringProcessor) OnDestroyComponent(c *Component) error{
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
	t := reflect.ValueOf(c).Elem()

	//We can autowire just structs
	if t.Kind()!=reflect.Struct{
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		fldAccessor := t.Field(i)
		fld := t.Type().Field(i)

		if v := fld.Tag.Get("inject"); v == "type" || v=="t" {
			if err := this.injectFieldByType(fldAccessor, fld.Type); err != nil {
				return err
			}
		}

		if v:= fld.Tag.Get("inject"); v=="all" || v=="a"{
			if err := this.injectAllComponentsByType(fldAccessor, fld.Type); err != nil {
				return err
			}
		}
	}

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

	r := candidates[0]

	if err:=this.configurer.ConfigureComponent(r);err!=nil{
		return err
	}

	fld.Set(reflect.ValueOf(r.Inst))

	return nil
}

func (this *AutowiringProcessor) injectAllComponentsByType(fld reflect.Value, t reflect.Type) error {
	log.Println("Injecting field", fld, "by all type instances")

	if t.Kind()!=reflect.Slice{
		return fmt.Errorf("Bad inject:all field. Slice expected, got: %v",t.Kind())
	}

	sliceType := t

	t = t.Elem() //Get slice's type

	candidates := this.ctx.FindComponentsByType(t)

	if len(candidates) == 0 {
		return errors.New(fmt.Sprint("Component not found. Type", t))
	}

	if !fld.CanSet() {
		return errors.New(fmt.Sprint("Field", fld, " cannot be set. Is it declared public?"))
	}

	target := reflect.MakeSlice(sliceType,len(candidates),len(candidates))


	for i,c := range candidates{
		r:=c

		if err:=this.configurer.ConfigureComponent(r);err!=nil{
			return err
		}

		target.Index(i).Set(reflect.ValueOf(r.Inst))
	}


	fld.Set(target)

	return nil
}