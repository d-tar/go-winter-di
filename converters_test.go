package wntr

import (
	"reflect"
	"testing"
)

type ConvertedObj struct {
	value string
}

func FromStrConverter(v string) (*ConvertedObj, error) {
	return &ConvertedObj{*v}, nil
}

func TestConversionService(t *testing.T) {

	var app struct {
		Conv ConversionService `inject:"t"`
	}

	ctx, err := FastDefaultContext(&GenericConversionService{}, ConverterBridge(FromStrConverter), &app)

	if err != nil {
		t.Fatal(err)
	}

	if err := ctx.Start(); err != nil {
		t.Fatal(err)
	}

	var b *ConvertedObj
	v, err := app.Conv.Convert("Test 123", reflect.TypeOf(b))

	t.Log(v, err)

}
