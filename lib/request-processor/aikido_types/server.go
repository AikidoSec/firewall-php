package aikido_types

import (
	"regexp"
	"sync"
)

type ServerData struct {
	AikidoConfig        AikidoConfigData
	TrustProxyConfig    TrustProxyConfig
	CloudConfig         CloudConfigData
	CloudConfigMutex    sync.Mutex
	MiddlewareInstalled bool
	ServerInitialized   bool
	ServerInitMutex     sync.Mutex
	ParamMatchers       map[string]*regexp.Regexp
}
