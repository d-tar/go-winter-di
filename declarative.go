package wntr

import (
	"fmt"
	"log"
	"reflect"
)

func ContextOrPanic(definitions ...interface{}) Context {
	ctx, err := FastBoot(definitions...)

	if err != nil {
		panic(err)
	}

	return ctx
}

func FastBoot(definitions ...interface{}) (Context, error) {
	ctx, err := CreateComplexContext(definitions...) // (headbang)

	if err != nil {
		return nil, err
	}

	if err := ctx.Start(); err != nil {
		return nil, err
	}

	return ctx, nil
}

func CreateComplexContext(definitions ...interface{}) (Context, error) {
	ctx, err := FastDefaultContext()

	if err != nil {
		return nil, err
	}

	return ctx, PopulateContextFromDefinitions(ctx, definitions...)
}

func PopulateContextFromDefinitions(ctx Context, definitions ...interface{}) error {
	for _, configuration := range definitions {
		if err := populateComponents(ctx, configuration); err != nil {
			return err
		}
	}
	return nil
}

func populateComponents(ctx Context, def interface{}) error {

	v := reflect.ValueOf(def)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		t = t.Elem() //Dereference pointer to real type
	} else {
		return fmt.Errorf("Cannot create context from value-of-config. Pass pointer to config instead")
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("Cannot create context from non pointer-to-configuration structures. Expected kind Struct, but got %v", t.Kind())
	}

	if t.Kind() == reflect.Slice {
		return nil
	}

	ctx.RegisterComponent(v.Interface())

	log.Println("StructContext: Registering definitions from ", t, "...")

	for i := 0; i < t.NumField(); i++ {
		fld := t.Field(i)

		//If we have interface - let's expose it
		if fld.Type.Kind() == reflect.Interface {
			fldVal := v.Elem().Field(i)
			ptrToFld := fldVal.Interface()
			if ptrToFld != nil {
				ctx.RegisterComponentWithTags(ptrToFld, string(fld.Tag))
			}
		}

		if fld.Type.Kind() != reflect.Struct && fld.Type.Kind() != reflect.Func {
			continue
		}

		fldVal := v.Elem().Field(i)
		ptrToFld := fldVal.Addr().Interface()

		if fld.Anonymous {
			if err := populateComponents(ctx, ptrToFld); err != nil {
				return err
			}
			continue
		}

		ctx.RegisterComponentWithTags(ptrToFld, string(fld.Tag))
	}

	return nil
}
