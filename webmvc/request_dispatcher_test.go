package webmvc

import (
	"log"
	"regexp"
	"testing"
)

func Test(t *testing.T) {

	log.Println(ScanMappingPattern("/qq"))
	log.Println(ScanMappingPattern(""))
	log.Println(ScanMappingPattern("/qq/:12"))
	log.Println(ScanMappingPattern("/index/:uname/qqqwe/:qqq/*"))
	log.Println(ScanMappingPattern("/index/*"))
}

func Test2(t *testing.T) {
	rx, err := scanAndCompile("/q?/:11/*/:bza")
	if err != nil {
		t.Fatal(err)
	}

	log.Println(rx.FindStringSubmatch("/q?/param1/asdasd/asd/param2"))
}

func scanAndCompile(p string) (*regexp.Regexp, error) {
	a, _ := ScanMappingPattern(p)
	return CompileRegex(a)
}
