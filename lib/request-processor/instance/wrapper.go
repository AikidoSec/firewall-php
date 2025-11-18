package instance

import (
	. "main/aikido_types"
	"main/context"
	"sync"
	"unsafe"
)

// RequestProcessorInstance holds per-request state for each PHP thread.
// In NTS mode (standard PHP), there's one global instance.
// In ZTS mode (FrankenPHP), each thread gets its own instance with locking.
type RequestProcessorInstance struct {
	CurrentToken    string
	CurrentServer   *ServerData
	RequestContext  context.RequestContextData
	ContextInstance unsafe.Pointer // For context callbacks
	ContextCallback unsafe.Pointer // C function pointer, must be per-instance in ZTS

	mu    sync.Mutex // Only used when isZTS is true
	isZTS bool
}

// NewRequestProcessorInstance creates an instance. Pass isZTS=true for FrankenPHP.
func NewRequestProcessorInstance(isZTS bool) *RequestProcessorInstance {
	return &RequestProcessorInstance{
		CurrentToken:   "",
		CurrentServer:  nil,
		RequestContext: context.RequestContextData{},
		isZTS:          isZTS,
	}
}

func (i *RequestProcessorInstance) SetCurrentServer(server *ServerData) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.CurrentServer = server
}

func (i *RequestProcessorInstance) GetCurrentServer() *ServerData {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.CurrentServer
}

func (i *RequestProcessorInstance) SetCurrentToken(token string) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.CurrentToken = token
}

func (i *RequestProcessorInstance) GetCurrentToken() string {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.CurrentToken
}

func (i *RequestProcessorInstance) SetRequestContext(ctx context.RequestContextData) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.RequestContext = ctx
}

func (i *RequestProcessorInstance) GetRequestContext() *context.RequestContextData {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return &i.RequestContext
}

func (i *RequestProcessorInstance) IsInitialized() bool {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.CurrentServer != nil
}

func (i *RequestProcessorInstance) IsZTS() bool {
	return i.isZTS
}

func (i *RequestProcessorInstance) SetContextCallback(callback unsafe.Pointer) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.ContextCallback = callback
}

func (i *RequestProcessorInstance) GetContextCallback() unsafe.Pointer {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.ContextCallback
}
