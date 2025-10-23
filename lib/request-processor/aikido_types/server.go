package aikido_types

import "sync"

type ServerData struct {
	AikidoConfig        AikidoConfigData
	CloudConfig         CloudConfigData
	CloudConfigMutex    sync.Mutex
	MiddlewareInstalled bool
}
