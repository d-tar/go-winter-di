package webmvc

import (
	"fmt"
	"github.com/d-tar/wntr"
	"net/http"
)

//Model And View interface
type WebResult interface {
	ViewName() string
	Model() interface{}
	HttpCode() int
}

//Base web request processing interface
type WebController interface {
	Serve(*http.Request) WebResult
}

type WebView interface {
	Render(WebResult, http.ResponseWriter, *http.Request) error
}

type WebViewFunc func(WebResult, http.ResponseWriter, *http.Request) error

func (t WebViewFunc) Render(mav WebResult, w http.ResponseWriter, r *http.Request) error {
	return t(mav, w, r)
}

//Single function request processor interface
//
//  use it to create single-method-request-handlers'
//
//   var routes struct{
//           MyHandler  HandlerFunc `@web-uri:"/my-uri"`
//   }
//
//   routes.MyHandler = functionalHandler
//
//
type HandlerFunc func(*http.Request) WebResult

func _() {
	var h HandlerFunc = nil
	var _ WebController = h
}

func (this HandlerFunc) Serve(r *http.Request) WebResult {
	return this(r)
}

var _ wntr.PreInitable = &MvcHandler{}

type MvcHandler struct {
	viewTable map[string]WebView
}

func (this *MvcHandler) PreInit() error {
	if this.viewTable != nil {
		return nil
	}

	this.viewTable = make(map[string]WebView)

	this.viewTable["JSON"] = NewJsonView()

	this.viewTable[""] = this.viewTable["JSON"]

	return nil
}

func (this *MvcHandler) SetWebViews(vws map[string]WebView) {
	this.viewTable = vws
}

func (h *MvcHandler) HandleWebResult(r WebResult, w http.ResponseWriter, req *http.Request) error {
	view, ok := h.viewTable[r.ViewName()]
	if !ok {
		return fmt.Errorf("No view named %v", r.ViewName())
	}

	return view.Render(r, w, req)
}
