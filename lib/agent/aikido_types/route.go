package aikido_types

import (
	"main/ipc/protos"
	"slices"
	"strings"
)

func (p *CloudPolicy) ExcludesRequestFromCount(route string) bool {
	p.ConfigCacheMutex.Lock()
	rules := p.CloudConfig.RequestCountExclusions
	p.ConfigCacheMutex.Unlock()

	if slices.Contains(rules.Exact, route) {
		return true
	}
	for _, suffix := range rules.Suffix {
		if suffix != "" && strings.HasSuffix(route, suffix) {
			return true
		}
	}
	return false
}

type RouteData struct {
	Method  string
	Path    string
	APISpec protos.APISpec
}

// BodyDataType represents the type of the body data.
type BodyDataType string

const (
	JSON           BodyDataType = "json"
	FormURLEncoded BodyDataType = "form-urlencoded"
	FormData       BodyDataType = "form-data"
	XML            BodyDataType = "xml"
	Undefined      BodyDataType = ""
)
