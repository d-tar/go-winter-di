package wntr

import (
	"log"
	"testing"
)

func TestComponentDefinition(t *testing.T) {
	ctx, err := NewContext()

	if err != nil {
		t.Fatal(err)
	}

	if err := ctx.Start(); err != nil {
		t.Fatal("Failed to start context", err)
	}

	if err := ctx.Stop(); err != nil {
		t.Fatal("Failed to stop context", err)
	}
}

/*
   Two-Phase Component Instantiation
*/

type TwoPhaseService struct {
	phase1done    bool
	phase2done    bool
	badPhaseOrder bool
}

func (s *TwoPhaseService) PreInit() error {
	log.Print("Pre init")
	s.phase1done = true
	return nil
}

func (s *TwoPhaseService) PostInit() error {
	log.Print("Post init")
	if !s.phase1done {
		s.badPhaseOrder = true
	}
	s.phase2done = true
	return nil
}

func TestTwoPhaseLifecycle(t *testing.T) {
	ctx, err := NewContext()

	if err != nil {
		t.Fatal(err)
	}

	service := &TwoPhaseService{}

	ctx.RegisterComponent(new(TwoPhaseInitializer))
	ctx.RegisterComponent(service)

	if e := ctx.Start(); e != nil {
		t.Fatal(e)
	}

	if !service.phase1done || !service.phase2done || service.badPhaseOrder {
		t.Fatal("Bad test results", service)
	}
}

/*
Autowiring test
*/

type AutowiringService struct {
	Test *testing.T `inject:"type"` //Inject pointer by type
}

func TestAutowiring(t *testing.T) {
	ctx, err := NewContext()

	if err != nil {
		t.Fatal(err)
	}

	service := &AutowiringService{}

	ctx.RegisterComponent(NewAutowiringProcessor())
	ctx.RegisterComponent(service)

	//Let us inject T service
	ctx.RegisterComponent(t)

	ctx.Start()

	if service.Test != t {
		t.Fatal("Autowiring by type does not working", service)
	}
}

/*
Autowiring by interface
*/

type DaoInterface interface {
	DoCrud()
}

type DaoImpl struct {
	crudDone bool
}

func (i *DaoImpl) DoCrud() {
	i.crudDone = true
}

type CrudService struct {
	//NOTE: Interface is injected 'by-value' not 'by-pointer'
	Dao DaoInterface `inject:"type"`
}

func TestInterfaceAutowiring(t *testing.T) {
	ctx, _ := NewContext()
	ctx.RegisterComponent(NewAutowiringProcessor())

	dao := &DaoImpl{}
	service := &CrudService{}

	ctx.RegisterComponent(dao)
	ctx.RegisterComponent(service)

	ctx.Start() //Setup context

	if service.Dao != dao {
		t.Fatal("Bad autowiring", service)
	}

	//Assert service is mutable
	service.Dao.DoCrud()
	if !dao.crudDone {
		t.Fatal("Crud was not done!", dao)
	}
}

/*
Test circular dependency detection
*/

type ClassA struct {
	B *ClassB `inject:"t"`
}

type ClassB struct {
	C *ClassC `inject:"t"`
}

type ClassC struct {
	A *ClassA `inject:"t"`
}

func TestCircularDependency(t *testing.T) {

	a, b, c := new(ClassA), new(ClassB), new(ClassC)

	ctx, _ := FastDefaultContext(a, b, c)

	e := ctx.Start()

	if e == nil {
		t.Fatal("Curcular dependency was resolved")
	}

	t.Log("Ok:", e)

}

/*
Test PreDestroyable
*/

type DisposableComp struct {
	disposed bool
}

func (this *DisposableComp) PreDestroy() {
	this.disposed = true
	log.Println("Component disposed")
}

func TestPredestroyableInterface(t *testing.T) {
	var test struct {
		C DisposableComp
	}

	ctx, _ := FastBoot(&test)
	ctx.Stop()

	if !test.C.disposed {
		t.Fatal("Component was not disposed")
	}
}

/*
Test destruction order

Here we expect that destruction order is inverted to configuration order
Note: it's dependency aware, so dependent component SHOULD be
destroyed before it's dependency
*/

type T1 struct {
}

type T2 struct {
	Tptr *T1 `inject:"t"`
}

type T3 struct {
	Tptr *T2 `inject:"t"`
}

func TestPredictedDestructionOrder(t *testing.T) {
	var test struct {
		I3 T1
		I2 T2
		I1 T3
	}

	ctx := ContextOrPanic(&test)

	ctx.Stop()
}
