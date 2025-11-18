package instance

import (
	. "main/aikido_types"
	"main/context"
	"sync"
	"unsafe"
)

// RequestProcessorInstance encapsulates all thread-local/request-scoped globals
type RequestProcessorInstance struct {
	// Per-request state (changes with each request/token update)
	CurrentToken    string
	CurrentServer   *ServerData
	RequestContext  context.RequestContextData
	ContextInstance unsafe.Pointer // Stores instance pointer for context callbacks
	ContextCallback unsafe.Pointer // Callback function pointer (C.ContextCallback) - must be instance-local for ZTS

	// Lock for thread safety (only used/locked in ZTS)
	mu    sync.Mutex
	isZTS bool // Set once at creation time - determines if locking is needed
}

// NewRequestProcessorInstance creates a new instance
// isZTS: true for Franken PHP (ZTS), false for standard PHP (NTS)
func NewRequestProcessorInstance(isZTS bool) *RequestProcessorInstance {
	return &RequestProcessorInstance{
		CurrentToken:   "",
		CurrentServer:  nil,
		RequestContext: context.RequestContextData{},
		isZTS:          isZTS,
	}
}

// SetCurrentServer updates the current server for this instance
// Conditional locking: only locks if ZTS mode
func (i *RequestProcessorInstance) SetCurrentServer(server *ServerData) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.CurrentServer = server
}

// GetCurrentServer returns the current server for this instance
// Conditional locking: only locks if ZTS mode
func (i *RequestProcessorInstance) GetCurrentServer() *ServerData {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.CurrentServer
}

// SetCurrentToken updates the current token for this instance
// Conditional locking: only locks if ZTS mode
func (i *RequestProcessorInstance) SetCurrentToken(token string) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.CurrentToken = token
}

// GetCurrentToken returns the current token for this instance
// Conditional locking: only locks if ZTS mode
func (i *RequestProcessorInstance) GetCurrentToken() string {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.CurrentToken
}

// SetRequestContext updates the request context for this instance
// Conditional locking: only locks if ZTS mode
func (i *RequestProcessorInstance) SetRequestContext(ctx context.RequestContextData) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.RequestContext = ctx
}

// GetRequestContext returns the request context for this instance
// Conditional locking: only locks if ZTS mode
func (i *RequestProcessorInstance) GetRequestContext() *context.RequestContextData {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return &i.RequestContext
}

// IsInitialized checks if this instance has been initialized
func (i *RequestProcessorInstance) IsInitialized() bool {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.CurrentServer != nil
}

// IsZTS returns whether this instance is running in ZTS mode
func (i *RequestProcessorInstance) IsZTS() bool {
	return i.isZTS
}

// SetContextCallback stores the context callback for this instance
// Conditional locking: only locks if ZTS mode
func (i *RequestProcessorInstance) SetContextCallback(callback unsafe.Pointer) {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	i.ContextCallback = callback
}

// GetContextCallback returns the context callback for this instance
// Conditional locking: only locks if ZTS mode
func (i *RequestProcessorInstance) GetContextCallback() unsafe.Pointer {
	if i.isZTS {
		i.mu.Lock()
		defer i.mu.Unlock()
	}
	return i.ContextCallback
}
