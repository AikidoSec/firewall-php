package aikido_types

import "main/ipc/protos"

type HandlerFunction func() string

type Method struct {
	ClassName  string
	MethodName string
}

type RequestShutdownParams struct {
	Server              *ServerData
	Method              string
	Route               string
	RouteParsed         string
	StatusCode          int
	User                string
	UserAgent           string
	IP                  string
	RateLimitGroup      string
	APISpec             *protos.APISpec
	RateLimited         bool
	QueryParsed         map[string]interface{}
	IsWebScanner        bool
	ShouldDiscoverRoute bool
	IsIpBypassed        bool
}
