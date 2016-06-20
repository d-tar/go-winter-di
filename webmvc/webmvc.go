package webmvc

import "net/http"

//Model And View interface
type WebResult interface {
	ViewName() string
	Model() interface{}
	HttpCode() int
}

type WebRequest struct {
	HttpRequest     *http.Request
	NamedParameters map[string]string
}

//Base web request processing interface
type WebController interface {
	Serve(*WebRequest) WebResult
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
type HandlerFunc func(*WebRequest) WebResult

func _() {
	var h HandlerFunc = nil
	var _ WebController = h
}

func (this HandlerFunc) Serve(r *WebRequest) WebResult {
	return this(r)
}

type EnableDefaultWebMvc struct {
	//Web Server Component
	//  Serving HTTP commands and routes them to request dispatcher
	Web WebServerComponent
	//Request Dispatcher
	//   Finds request handler
	//   Invokes handler with WebRequest and retrieves WebResult
	//   Passes WebResult to View Resolver to render answer
	Dispatcher RequestDispatcher
	//View resovler component
	//  Accepts WebResults and finds appropriate WebView to render it
	Mvc WebViewResolver
}
