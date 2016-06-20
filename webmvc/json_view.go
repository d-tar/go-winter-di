package webmvc

import (
	"encoding/json"
	"net/http"
)

/*
Simple Json View Implementation
*/

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
