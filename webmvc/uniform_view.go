package webmvc

type GenericWebResult struct {
	data     interface{}
	httpCode int
}

func _() {
	var _ WebResult = &GenericWebResult{}
}

func WebOk(data interface{}) WebResult {
	return &GenericWebResult{
		data: data,
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
