package context

// #include "../../API.h"
import "C"
import (
	"main/instance"
	"main/utils"
)

/*
Event level context cache below (changes on each hooked PHP function)
It caries the state from pre-function hook to post-function hook.
It is cleared after.
*/
type EventContextData struct {
	CurrentSsrfInterceptorResult *utils.InterceptorResult
}

func getEventContext(instance *instance.RequestProcessorInstance) *EventContextData {
	if instance == nil {
		return nil
	}

	ctx := instance.GetEventContext()
	if ctx == nil {
		return nil
	}
	return ctx.(*EventContextData)
}

func ResetEventContext(instance *instance.RequestProcessorInstance) bool {
	if instance == nil {
		return false
	}
	instance.SetEventContext(&EventContextData{})
	return true
}

/*
A partial interceptor result in stored when user-provided information was matched in the content
of the currently called PHP function.
We store this information because we cannot emit a detection at this point.
We will use this after the PHP function call ends, because at that point we have more information
that could help us emit a detection, combined with the partial interceptor result that was stored
before the function call.
A partial interceptor result stores the payload that matched the user input, the path to it, the
PHP function that was called, ..., basically the data needed for reporting if this actually turns into
a detection at a later stage.
*/
func EventContextSetCurrentSsrfInterceptorResult(instance *instance.RequestProcessorInstance, interceptorResult *utils.InterceptorResult) {
	ctx := getEventContext(instance)
	if ctx != nil {
		ctx.CurrentSsrfInterceptorResult = interceptorResult
	}
}

func GetCurrentSsrfInterceptorResult(instance *instance.RequestProcessorInstance) *utils.InterceptorResult {
	ctx := getEventContext(instance)
	if ctx == nil {
		return nil
	}
	return ctx.CurrentSsrfInterceptorResult
}
