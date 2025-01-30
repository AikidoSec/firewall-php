package globals

import (
	. "main/aikido_types"
	"sync"
)

var EnvironmentConfig EnvironmentConfigData

var AikidoConfig AikidoConfigData

var CloudConfig CloudConfigData
var CloudConfigMutex sync.Mutex

const (
	Version = "1.0.108"
)
