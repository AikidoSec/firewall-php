package globals

import (
	. "main/aikido_types"
	"sync"
)

var EnvironmentConfig EnvironmentConfigData

var AikidoConfig AikidoConfigData

var CloudConfig CloudConfigData
var CloudConfigMutex sync.Mutex
var MiddlewareInstalled bool

const (
	Version = "1.3.7"
)
