#include "Includes.h"

/*
    Global cache instances for request and event-level context.
    
    RequestCache: Per-request context that persists across the entire PHP request lifecycle.
    EventCacheStack: Stack-based event context that handles nested hook invocations safely.
*/
RequestCache requestCache;
EventCacheStack eventCacheStack;

void RequestCache::Reset() {
    *this = RequestCache();
}

void EventCache::Reset() {
    *this = EventCache();
}

/*
    EventCacheStack implementation:

    The stack holds per-hook event context. Each hook invocation pushes a new
    EventCache onto the stack, and pops it when the hook scope ends.
    
    This allows nested hooks (one hooked function calling another) to each have
    their own independent context without interfering with each other. Code that
    needs the current event context always reads from Top().
    
    Example flow:
    1. PRE handler: Push() -> Top().outgoingRequestUrl = "http://example.com"
    2. curl_exec() runs, follows redirect
    3. Callback invokes file_put_contents() -> Push() (new context on stack)
    4. Nested hook completes -> Pop() (outer context restored)
    5. POST handler: Top().outgoingRequestUrl still valid -> SSRF check runs
*/

void EventCacheStack::Push() {
    contexts.push(EventCache());
}

void EventCacheStack::Pop() {
    if (!contexts.empty()) {
        contexts.pop();
    }
}

EventCache& EventCacheStack::Top() {
    return contexts.top();
}

bool EventCacheStack::Empty() {
    return contexts.empty();
}

/*
    RAII wrapper for automatic context management.
    
    Ensures proper push/pop even on exceptions. Used at the start of every hook handler:
    - Constructor: Pushes new context onto stack
    - Destructor: Pops context when leaving scope (automatic cleanup)
    
    This prevents context leaks and ensures stack integrity.
*/
ScopedEventContext::ScopedEventContext() {
    eventCacheStack.Push();
}

ScopedEventContext::~ScopedEventContext() {
    eventCacheStack.Pop();
}
