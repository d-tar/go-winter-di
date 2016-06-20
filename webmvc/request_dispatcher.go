package webmvc

import (
//	"net/http"

)
import (
	"fmt"
	"github.com/d-tar/wntr"
	"log"
	"net/http"
	"regexp"
	"strings"
	"unicode"
)

type RequestMapping struct {
	Pattern         string
	Handler         WebController
	CompiledPattern *regexp.Regexp
	NamedParams     []string
}

var _ http.Handler = &RequestDispatcher{}

type RequestDispatcher struct {
	mappingTable map[string]*RequestMapping

	Mvc *WebViewResolver       `inject:"t"`
	Ctx wntr.ConfiguredContext `inject:"t"`
}

func (disp *RequestDispatcher) PostInit() error {
	for _, ctl := range disp.Ctx.FindComponentsByType(gWebControllerType) {

		tag := ctl.Tags()
		if uri := tag.Get("@web-uri"); uri != "" {
			rmap := RequestMapping{
				Handler: ctl.Instance().(WebController),
				Pattern: uri,
			}

			if err := disp.MapRequest(rmap); err != nil {
				return nil
			}

		}
	}
	return nil
}

func (disp *RequestDispatcher) MapRequest(m RequestMapping) error {
	seq, params := ScanMappingPattern(m.Pattern)
	rgx, err := CompileRegex(seq)
	if err != nil {
		return err
	}
	m.CompiledPattern = rgx
	m.NamedParams = params

	if disp.mappingTable == nil {
		disp.mappingTable = make(map[string]*RequestMapping)
	}

	if _, ok := disp.mappingTable[rgx.String()]; ok {
		return fmt.Errorf("Pattern '%v' already mapped")
	}
	log.Printf("RequestDispatcher: Mapped %v on to %v \n", m.Pattern, m.Handler)
	disp.mappingTable[rgx.String()] = &m
	return nil
}

func (disp *RequestDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := r.RequestURI

	m, params := disp.findMappingForUri(uri)

	if m == nil {
		disp.serve404(w, r)
		return
	}

	handler := m.Handler

	webReq := &WebRequest{
		HttpRequest:     r,
		NamedParameters: params,
	}

	result := handler.Serve(webReq)

	if err := disp.Mvc.HandleWebResult(result, w, r); err != nil {
		panic(err)
	}

}

func (disp *RequestDispatcher) serve404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprint(w, "<html><body><h2>404 Not Found</h2><br>No Request Processor for URI:<br><br><h4><u>", r.RequestURI, "</h4></u><br><br><i>Faithfully yours, WebMVC RequestDispatcher</i></body></html>")
}

func (disp *RequestDispatcher) findMappingForUri(uri string) (*RequestMapping, map[string]string) {
	for _, mapping := range disp.mappingTable {
		match := mapping.CompiledPattern.FindStringSubmatch(uri)
		if len(match) == 0 {
			continue
		}

		namedParams := make(map[string]string)

		for i := 1; i < len(match); i++ {
			pName := mapping.NamedParams[i-1]
			namedParams[pName] = match[i]
		}

		return mapping, namedParams
	}
	return nil, nil
}

func CompileRegex(p []string) (*regexp.Regexp, error) {
	result := "^"
	pos := 0
	for _, frag := range p {
		if frag == "" {
			return nil, fmt.Errorf("Bad pattern. Found empty-string fragment at %v", pos)
		} else if frag == "*" {
			result += ".*"
		} else if strings.HasPrefix(frag, ":") {
			result += "(.*?)"
		} else {
			result += regexp.QuoteMeta(frag)
		}
		pos += len(frag)
	}
	result += "$"
	return regexp.Compile(result)
}

func ScanMappingPattern(p string) ([]string, []string) {
	tail := p

	fragments := make([]string, 0)
	namedParams := make([]string, 0)

	for {
		pos := strings.IndexAny(tail, ":*")

		if pos < 0 {
			if len(tail) > 0 {
				fragments = append(fragments, tail)
			}
			break
		}

		head := tail[0:pos]
		tail = tail[pos:]

		fragments = append(fragments, head)

		if tail[0] == '*' {
			fragments = append(fragments, "*")
			tail = tail[1:]
		} else if len(tail) > 0 {
			var param string
			param, tail = extractParam(tail[1:])
			fragments = append(fragments, (":" + param))
			namedParams = append(namedParams, param)
		}

	}

	return fragments, namedParams
}

func extractParam(p string) (string, string) {
	i := strings.IndexFunc(p, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	if i < 0 {
		return p, ""
	}

	return p[0:i], p[i:]
}
