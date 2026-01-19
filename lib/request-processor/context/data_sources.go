package context

type Source struct {
	Name     string
	CacheGet func() map[string]string
}

var SOURCES = []Source{
	{"body", GetBodyParsedFlattened},
	{"query", GetQueryParsedFlattened},
	{"headers", GetHeadersParsedFlattened},
	{"cookies", GetCookiesParsedFlattened},
	{"routeParams", GetRouteParamsParsedFlattened},
	{"graphql", GetGraphQLParsedFlattened},
}
