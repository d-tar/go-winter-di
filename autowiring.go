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
	ctx        ConfiguredContext
	configurer ComponentConfigurer
}

func _() {
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
		if err := v.FindSingleComponent(&this.configurer); err != nil {
			return fmt.Errorf("Bad context setup. Failed to FindSingleComponent ComponentConfigurer: %v", err)
		}

		if err := v.FindSingleComponent(&this.ctx); err != nil {
			return fmt.Errorf("Bad context setup. Failed to FindSingleComponent ConfiguredContext: %v", err)
		}

		//this.ctx = v
		return nil
	}

	return errors.New("Unsupported context type")
}

func (this *AutowiringProcessor) OnPrepareComponent(c *ComponentImpl) error {
	return this.autowireInstance(c.inst)
}

func (this *AutowiringProcessor) OnComponentReady(c *ComponentImpl) error {
	return nil
}

func (this *AutowiringProcessor) OnDestroyComponent(c *ComponentImpl) error {
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
	if t.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		fldAccessor := t.Field(i)
		fld := t.Type().Field(i)

		if v := fld.Tag.Get("inject"); v == "type" || v == "t" {
			if err := this.injectFieldByType(fldAccessor, fld); err != nil {
				//return fmt.Errorf("Unable to process field %v, error %v",fld.Name,err)
				return typeConstructError(t.Type(), fld, err)
			}
		}

		if v := fld.Tag.Get("inject"); v == "all" || v == "a" {
			if err := this.injectAllComponentsByType(fldAccessor, fld); err != nil {
				//return err
				return typeConstructError(t.Type(), fld, err)
			}
		}
	}

	return nil
}

func (this *AutowiringProcessor) injectFieldByType(fld reflect.Value, f reflect.StructField) error {
	t := f.Type
	log.Println("Injecting field", f.Name, "by type")

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

	if err := this.configurer.ConfigureComponent(r.(*ComponentImpl)); err != nil {
		return err
	}

	fld.Set(reflect.ValueOf(r.Instance()))

	return nil
}

func (this *AutowiringProcessor) injectAllComponentsByType(fld reflect.Value, f reflect.StructField) error {
	t := f.Type
	log.Println("Injecting field", f.Name, "by all type instances")

	if t.Kind() != reflect.Slice {
		return fmt.Errorf("Bad inject:all field. Slice expected, got: %v", t.Kind())
	}

	sliceType := t

	t = t.Elem() //Get slice's type

	candidates := this.ctx.FindComponentsByType(t)

	if len(candidates) == 0 {
		return nil //errors.New(fmt.Sprint("Component not found. Type", t))
	}

	if !fld.CanSet() {
		return errors.New(fmt.Sprint("Field", fld, " cannot be set. Is it declared public?"))
	}

	target := reflect.MakeSlice(sliceType, len(candidates), len(candidates))

	for i, c := range candidates {
		r := c

		if err := this.configurer.ConfigureComponent(r.(*ComponentImpl)); err != nil {
			return err
		}

		target.Index(i).Set(reflect.ValueOf(r.Instance()))
	}

	fld.Set(target)

	return nil
}

func typeConstructError(t reflect.Type, f reflect.StructField, cause error) error {
	return fmt.Errorf("Unable to costruct type %v:  Failed to fill field %v: %v", t, f, cause)
}
