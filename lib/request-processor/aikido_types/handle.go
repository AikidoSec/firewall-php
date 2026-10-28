package aikido_types

import "main/ipc/protos"

type Method struct {
	ClassName  string
	MethodName string
}

type RequestShutdownParams struct {
	ThreadID               uint64
	Token                  string
	Method                 string
	Route                  string
	RouteParsed            string
	StatusCode             int
	User                   string
	UserAgent              string
	IP                     string
	Url                    string
	RateLimitGroup         string
	APISpec                *protos.APISpec
	RateLimited            bool
	QueryParsed            map[string]interface{}
	IsWebScanner           bool
	ShouldDiscoverRoute    bool
	PhpFatalErrorKind      string
	HasUploadProgressField bool
}

// ShouldReport tells whether the agent needs the request metadata.
func (p RequestShutdownParams) ShouldReport() bool {
	return p.RateLimited ||
		p.ShouldDiscoverRoute ||
		p.IsWebScanner ||
		p.PhpFatalErrorKind != "" ||
		p.HasUploadProgressField
}
