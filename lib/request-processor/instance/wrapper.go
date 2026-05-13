package instance

import (
	. "main/aikido_types"
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

	// Stored as interface{} to avoid circular imports with context package
	// RequestContext is *RequestContextData, EventContext is *EventContextData
	// (event_context.go, request_context.go)
	RequestContext interface{}
	EventContext   interface{}
}

func NewRequestProcessorInstance(threadID uint64) *RequestProcessorInstance {
	return &RequestProcessorInstance{
		CurrentToken:  "",
		CurrentServer: nil,
		threadID:      threadID,
	}
}

func (i *RequestProcessorInstance) SetCurrentServer(server *ServerData) {
	i.CurrentServer = server
}

func (i *RequestProcessorInstance) GetCurrentServer() *ServerData {
	return i.CurrentServer
}

func (i *RequestProcessorInstance) SetCurrentToken(token string) {
	i.CurrentToken = token
}

func (i *RequestProcessorInstance) GetCurrentToken() string {
	return i.CurrentToken
}

func (i *RequestProcessorInstance) IsInitialized() bool {
	return i.CurrentServer != nil
}

func (i *RequestProcessorInstance) SetContextCallback(callback unsafe.Pointer) {
	i.ContextCallback = callback
}

func (i *RequestProcessorInstance) GetContextCallback() unsafe.Pointer {
	return i.ContextCallback
}

func (i *RequestProcessorInstance) SetThreadID(tid uint64) {
	i.threadID = tid
}

func (i *RequestProcessorInstance) GetThreadID() uint64 {
	return i.threadID
}

func (i *RequestProcessorInstance) SetRequestContext(ctx interface{}) {
	i.RequestContext = ctx
}

func (i *RequestProcessorInstance) GetRequestContext() interface{} {
	return i.RequestContext
}

func (i *RequestProcessorInstance) SetEventContext(ctx interface{}) {
	i.EventContext = ctx
}

func (i *RequestProcessorInstance) GetEventContext() interface{} {
	return i.EventContext
}

func (i *RequestProcessorInstance) SetContextInstance(ptr unsafe.Pointer) {
	i.ContextInstance = ptr
}

func (i *RequestProcessorInstance) GetContextInstance() unsafe.Pointer {
	return i.ContextInstance
}
