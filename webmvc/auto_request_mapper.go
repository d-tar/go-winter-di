package webmvc

import (
	"encoding/json"
	"github.com/d-tar/wntr"
	"io/ioutil"
	"log"
	"reflect"
)

//Each request mapping must route to WebController
var _ WebController

// In: *WebRequest
// Out: WebResult

func AutoHandler(p interface{}) SmartWebHandler {
	return SmartWebHandler{pFunc: reflect.ValueOf(p)}
}

type SmartWebHandler struct {
	pFunc reflect.Value
	Conv  wntr.ConversionService `inject:"t"`
}

var _ WebController = (*SmartWebHandler)(nil)

var WebResultType reflect.Type = reflect.TypeOf((*WebResult)(nil)).Elem()

func (p *SmartWebHandler) Serve(r *WebRequest) WebResult {
	//1. Input value must be allocated, autowired and postproceed by MVC processor
	//2. Handler must be invoked with value (1)
	//3. Return value must be converted to WebResult type

	tFunc := p.pFunc.Type()

	in := []reflect.Value{}

	if tFunc.NumIn() != 0 {
		in = p.setupMapping(tFunc, r)
	}

	out := p.pFunc.Call(in)

	if tFunc.NumOut() == 2 { //if we can have error
		if !out[1].IsNil() { //and error is not null
			err := out[1].Interface().(error)
			res, err := p.Conv.Convert(err, WebResultType)

			if err != nil {
				panic(err)
			}

			return res.(WebResult)
		}
	}

	value := out[0].Interface()
	res, err := p.Conv.Convert(value, WebResultType)

	if err != nil {
		panic(err)
	}

	return res.(WebResult)
}

func (p *SmartWebHandler) setupMapping(tFun reflect.Type, r *WebRequest) []reflect.Value {
	tIn := tFun.In(0) //We expect T here

	in := reflect.New(tIn) // T -> allocate *T

	p.processAnnotations(in.Elem(), r)

	return []reflect.Value{in.Elem()}
}

func (p *SmartWebHandler) processAnnotations(v reflect.Value, r *WebRequest) {
	ty := v.Type()
	log.Println(ty)
	for i := 0; i < ty.NumField(); i++ {
		f := ty.Field(i)

		if name := f.Tag.Get("@path-variable"); name != "" {
			value := r.NamedParameters[name]
			rv := reflect.ValueOf(value)
			v.Field(i).Set(rv)
		} else if name := f.Tag.Get("@request-body"); name != "" {
			b, err := ioutil.ReadAll(r.HttpRequest.Body)
			if err != nil {
				panic(err)
			}

			holder := reflect.New(f.Type)
			err = json.Unmarshal(b, holder.Interface())
			if err != nil {
				panic(err)
			}

			v.Field(i).Set(holder.Elem())
		}
	}

}
