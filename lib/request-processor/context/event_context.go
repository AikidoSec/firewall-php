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

func getEventContext(inst *instance.RequestProcessorInstance) *EventContextData {
	if inst == nil {
		return nil
	}

	ctx := inst.GetEventContext()
	if ctx == nil {
		return nil
	}
	return ctx.(*EventContextData)
}

func ResetEventContext(inst *instance.RequestProcessorInstance) bool {
	if inst == nil {
		return false
	}
	inst.SetEventContext(&EventContextData{})
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
func EventContextSetCurrentSsrfInterceptorResult(inst *instance.RequestProcessorInstance, interceptorResult *utils.InterceptorResult) {
	ctx := getEventContext(inst)
	if ctx != nil {
		ctx.CurrentSsrfInterceptorResult = interceptorResult
	}
}

func GetCurrentSsrfInterceptorResult(inst *instance.RequestProcessorInstance) *utils.InterceptorResult {
	ctx := getEventContext(inst)
	if ctx == nil {
		return nil
	}
	return ctx.CurrentSsrfInterceptorResult
}
