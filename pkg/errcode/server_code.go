package errcode

var (
	Success     = NewError(200, "success")
	NotModified = NewError(304, "app not modified")
	ParamError  = NewError(400, "request param error")
	NotFound    = NewError(404, "not found")
	Conflict    = NewError(409, "conflict")
	ServerError = NewError(500, "service internal error")
)
