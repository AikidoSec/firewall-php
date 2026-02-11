package instance

import (
	. "main/aikido_types"
	"sync"
	"unsafe"
)

// RequestProcessorInstance holds per-request state for each PHP thread.
// In NTS mode (standard PHP), there's one global instance.
// In ZTS mode (FrankenPHP), each thread gets its own instance.
type RequestProcessorInstance struct {
	CurrentToken    string
	CurrentServer   *ServerData
	threadID        uint64         // CACHED: OS thread ID cached at RINIT for fast context lookups
	ContextInstance unsafe.Pointer // For context callbacks
	ContextCallback unsafe.Pointer // C function pointer, must be per-instance in ZTS

	RequestContext interface{}
	EventContext   interface{}

	mu    sync.Mutex // Only used when isZTS is true
	isZTS bool
}

func NewRequestProcessorInstance(threadID uint64, isZTS bool) *RequestProcessorInstance {
	return &RequestProcessorInstance{
		CurrentToken:  "",
		CurrentServer: nil,
		threadID:      threadID,
		isZTS:         isZTS,
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

func (i *RequestProcessorInstance) SetThreadID(tid uint64) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.threadID = tid
}

func (i *RequestProcessorInstance) GetThreadID() uint64 {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.threadID
}

func (i *RequestProcessorInstance) SetRequestContext(ctx interface{}) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.RequestContext = ctx
}

func (i *RequestProcessorInstance) GetRequestContext() interface{} {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.RequestContext
}

func (i *RequestProcessorInstance) SetEventContext(ctx interface{}) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.EventContext = ctx
}

func (i *RequestProcessorInstance) GetEventContext() interface{} {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.EventContext
}

func (i *RequestProcessorInstance) SetContextInstance(ptr unsafe.Pointer) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.ContextInstance = ptr
}

func (i *RequestProcessorInstance) GetContextInstance() unsafe.Pointer {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.ContextInstance
}
