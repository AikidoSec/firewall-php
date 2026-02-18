package context

import "main/instance"

type Source struct {
	Name     string
	CacheGet func(*instance.RequestProcessorInstance) map[string]string
}

var SOURCES = []Source{
	{"body", GetBodyParsedFlattened},
	{"query", GetQueryParsedFlattened},
	{"headers", GetHeadersParsedFlattened},
	{"cookies", GetCookiesParsedFlattened},
	{"routeParams", GetRouteParamsParsedFlattened},
}
