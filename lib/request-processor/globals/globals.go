package globals

import (
	. "main/aikido_types"
	"sync"
)

var InitData InitConfigData

var CloudConfig CloudConfigData
var CloudConfigMutex sync.Mutex

const (
	Version = "1.0.41"
)
