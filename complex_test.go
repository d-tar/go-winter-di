package wntr

import (
	"log"
	"testing"
)

type Dao2 interface {
	DoJob()
}

type DaoImpl2 struct {
	job string
}

type Controller struct {
	Dao Dao2 `inject:"type"`
}

type AllDaoStruct struct {
	Dao []Dao2 `inject:"all"`
}

func (c *Controller) DoController() {
	if c.Dao == nil {
		panic("No dao injected")
	} else {
		log.Println("OK, got dao")
	}
}

func (d *DaoImpl2) DoJob() {

}

type Baseapp struct {
	Dao  DaoImpl2
	Ctrl Controller
}

func TestComplexConfiugration(t *testing.T) {

	var app struct {
		Baseapp
	}

	ctx, err := CreateComplexContext(&app)

	if err != nil {
		t.Fatal(err)
	}

	ctx.Start()

	app.Ctrl.DoController()

	ctx.Stop()

}

func TestSmallConfiguration(t *testing.T) {
	var app struct {
		Baseapp
		Ctx Context `inject:"type"`
	}

	if _, err := FastBoot(&app); err != nil {
		t.Fatal(err)
	}

	app.Ctx.Stop()
}

func TestInjectAllInstances(t *testing.T) {
	var app struct {
		D1 DaoImpl2 `@mvc:"123"`
		D2 DaoImpl2
		D  AllDaoStruct
	}

	if _, err := FastBoot(&app); err != nil {
		t.Fatal("Failed to boot up context", err)
	}

	if app.D.Dao == nil {
		t.Fatal("Inejct all not working", app)
	}

	log.Println(app.D.Dao, app.D.Dao[0])

}

func TestBadConfig(t *testing.T) {
	var app struct {
	}
	_, err := FastBoot(app)

	if err == nil {
		t.Fatal("No error")
	}

	p := &app

	_, err = FastBoot(&p)

	if err == nil {
		t.Fatal("No error")
	}
}
