package wntr

import (
	"reflect"
	"log"
)

func FastBoot(definitions...interface{}) (Context,error){
	ctx,err := CreateComplexContext(definitions...) // (headbang)

	if err != nil {
		return nil,err
	}

	if err:=ctx.Start();err!=nil{
		return nil,err
	}

	return ctx,nil
}

func CreateComplexContext(definitions...interface{}) (Context,error){
	ctx,err := FastDefaultContext();

	if err != nil {
		return nil,err
	}

	for _,beans := range definitions{
		if err := populateComponents(ctx,beans); err != nil{
			return nil,err
		}
	}

	return ctx,nil
}


func populateComponents(ctx Context, def interface{}) error{



	v:=reflect.ValueOf(def)
	t:=v.Type()

	if t.Kind()==reflect.Ptr{
		t = t.Elem() //Dereference pointer to real type
	}

	if t.Kind()==reflect.Slice{
		return nil
	}

	ctx.RegisterComponent(v.Interface())


	log.Println("Populating definition from",t,t.Kind())

	for i:=0;i<t.NumField();i++{
		fld:=t.Field(i)


		log.Println(fld.Name,fld.Type.Kind())

		//If we have interface - let's expose it
		if fld.Type.Kind()==reflect.Interface{
			fldVal := v.Elem().Field(i)
			ptrToFld := fldVal.Interface()
			if ptrToFld!=nil{
				ctx.RegisterComponentWithTags(ptrToFld,string(fld.Tag))
			}
		}

		if(fld.Type.Kind()!=reflect.Struct&&fld.Type.Kind()!=reflect.Func){
			continue
		}

		fldVal := v.Elem().Field(i)
		ptrToFld := fldVal.Addr().Interface()

		if fld.Anonymous{
			if err := populateComponents(ctx,ptrToFld);err!=nil{
				return err
			}
			continue
		}

		log.Println(ptrToFld)

		ctx.RegisterComponentWithTags(ptrToFld,string(fld.Tag))
	}

	return nil
}