package webmvc

import (
	"encoding/json"
	"net/http"
)

func NewJsonView() WebView {
	var q WebViewFunc = RenderJsonView
	return q
}

func RenderJsonView(mav WebResult, w http.ResponseWriter, r *http.Request) error {
	b, err := json.Marshal(mav.Model())
	if err != nil {
		return err
	}
	w.Header().Add("Content-Type", "application/json;charset=UTF-8")

	if mav.HttpCode() != 0 {
		w.WriteHeader(mav.HttpCode())
	}

	w.Write(b)
	return nil
}

/*
import (
	"encoding/json"
	"net/http"
)

type JsonWebResult struct {
	data     interface{}
	httpCode int
}

func _() {
	//var _ WebResult = &JsonWebResult{}
}

func (v *JsonWebResult) Render(w http.ResponseWriter) error {
	b, err := json.Marshal(v.data)
	if err != nil {
		return err
	}
	w.Header().Add("Content-Type", "application/json;charset=UTF-8")
	if v.httpCode != 0 {
		w.WriteHeader(v.httpCode)
	}
	w.Write(b)
	return nil
}

type BaseJsonResponse struct {
	Data   interface{} `json:"data"`
	Error  interface{} `json:"error"`
	Status string      `json:"status"`
}

func JsonOk(data interface{}) WebResult {
	return &JsonWebResult{
		data: &BaseJsonResponse{
			Data:   data,
			Status: "ok",
		},
	}
}

func JsonErr(error interface{}) WebResult {
	return &JsonWebResult{
		data: &BaseJsonResponse{
			Error:  error,
			Status: "err",
		},
		httpCode: 500,
	}
}
*/
