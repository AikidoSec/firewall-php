package aikido_types

import (
	"regexp"
	"sync"
)

type ServerData struct {
	AikidoConfig        AikidoConfigData
	CloudConfig         CloudConfigData
	CloudConfigMutex    sync.Mutex
	MiddlewareInstalled bool
	ParamMatchers       map[string]*regexp.Regexp
}
