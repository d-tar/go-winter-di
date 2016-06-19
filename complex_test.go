package wntr

import (
	"testing"
	"log"
)


type Dao2 interface {
	DoJob()
}

type DaoImpl2 struct {
	job string
}

type Controller struct {
	Dao Dao2 `autowire:"type"`
}

func (c*Controller) DoController(){
	if c.Dao == nil{
		panic("No dao injected")
	}else{
		log.Println("OK, got dao")
	}
}

func (d*DaoImpl2) DoJob(){

}

type Baseapp struct{
	Dao DaoImpl2
	Ctrl Controller
}

func TestComplexConfiugration(t *testing.T){



	var app struct{
		Baseapp
	}

	ctx,err := CreateComplexContext(&app)


	if err != nil {
		t.Fatal(err)
	}

	ctx.Start()

	app.Ctrl.DoController()

	ctx.Stop()

}


func TestSmallConfiguration(t *testing.T){
	var app struct{
		Baseapp
		Ctx Context `autowire:"type"`
	}

	if _,err := FastBoot(&app);err!=nil{
		t.Fatal(err)
	}

	app.Ctx.Stop()
}
