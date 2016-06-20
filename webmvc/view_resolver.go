package webmvc

import (
	"fmt"
	"github.com/d-tar/wntr"
	"net/http"
)

var _ wntr.PreInitable = &WebViewResolver{}

type WebViewResolver struct {
	viewTable map[string]WebView
}

func (this *WebViewResolver) PreInit() error {
	if this.viewTable != nil {
		return nil
	}

	this.viewTable = make(map[string]WebView)

	this.viewTable["JSON"] = NewJsonView()

	this.viewTable[""] = this.viewTable["JSON"]

	return nil
}

func (this *WebViewResolver) SetWebViews(vws map[string]WebView) {
	this.viewTable = vws
}

func (h *WebViewResolver) HandleWebResult(r WebResult, w http.ResponseWriter, req *http.Request) error {
	view, ok := h.viewTable[r.ViewName()]
	if !ok {
		return fmt.Errorf("No view named %v", r.ViewName())
	}

	return view.Render(r, w, req)
}

/*
************************************************************************************
Default Model Stage Implementation
************************************************************************************
*/

type GenericWebResult struct {
	data     interface{}
	httpCode int
}

func _() {
	var _ WebResult = &GenericWebResult{}
}

func WebOk(data interface{}) WebResult {
	return &GenericWebResult{
		data:     data,
		httpCode: 200,
	}
}

func WebErr(error interface{}) WebResult {
	return &GenericWebResult{
		data:     error,
		httpCode: 500,
	}
}

func (*GenericWebResult) ViewName() string {
	return ""
}

func (this *GenericWebResult) Model() interface{} {
	return this.data
}

func (this *GenericWebResult) HttpCode() int {
	return this.httpCode
}
