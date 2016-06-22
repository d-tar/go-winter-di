package wntr

import (
	"fmt"
	"log"
	"reflect"
)

type Converter interface {
	Convert(interface{}) (interface{}, error)
	Type() (reflect.Type, reflect.Type)
}

type ConversionService interface {
	Convert(interface{}, reflect.Type) (interface{}, error)
}

type GenericConversionService struct {
	Converters         []Converter `inject:"a"`
	StandardConverters []Converter
}

var _ ConversionService = (*GenericConversionService)(nil)

func (this *GenericConversionService) Convert(src interface{}, dstTy reflect.Type) (interface{}, error) {
	srcTy := reflect.TypeOf(src)

	for _, converter := range this.StandardConverters {
		fromTy, toTy := converter.Type()
		if srcTy.AssignableTo(fromTy) && dstTy.AssignableTo(toTy) {
			return converter.Convert(src)
		}
	}

	for _, converter := range this.Converters {
		fromTy, toTy := converter.Type()
		if fromTy.AssignableTo(srcTy) && dstTy.AssignableTo(toTy) {
			return converter.Convert(src)
		}
	}

	return nil, fmt.Errorf("No converter found from %T to %v", src, dstTy)
}

func ConverterBridge(converterFunc interface{}) Converter {
	pValue := reflect.ValueOf(converterFunc)
	tValue := reflect.TypeOf(converterFunc)

	if tValue.NumIn() != 1 {
		panic(fmt.Errorf("Bad func interface, expected single in, got: %v", tValue.NumIn()))
	}
	if tValue.NumOut() != 2 {
		panic(fmt.Errorf("Bad func interface, expected two out, got: %v", tValue.NumOut()))
	}

	srcTy := tValue.In(0)
	dstTy := tValue.Out(0)

	log.Println("Created converter from ", srcTy, "to", dstTy)

	return &converterBridgeImpl{
		pfunc: pValue,
		srcTy: srcTy,
		dstTy: dstTy,
	}
}

type converterBridgeImpl struct {
	pfunc reflect.Value
	srcTy reflect.Type
	dstTy reflect.Type
}

func (b *converterBridgeImpl) Type() (reflect.Type, reflect.Type) {
	return b.srcTy, b.dstTy
}

func (b *converterBridgeImpl) Convert(from interface{}) (interface{}, error) {
	vFrom := reflect.ValueOf(from)
	result := b.pfunc.Call([]reflect.Value{vFrom})

	var p1 interface{}
	var p2 error

	if !result[0].IsNil() {
		p1 = result[0].Interface()
	}
	if !result[1].IsNil() {
		p2 = result[1].Interface().(error)
	}

	return p1, p2
}
