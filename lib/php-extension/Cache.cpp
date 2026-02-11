#include "Includes.h"

/*
    Cache objects used by the PHP extension to share state with the Go request processor.

    - RequestCache stores data that is scoped to a single incoming HTTP request
      (method, route, user id, rate-limit group, ...). The Go side calls into the
      extension once per request to populate and later clear this structure
      (see context.Init / context.Clear in the Go code).

    - EventCacheStack stores data that is scoped to a single *event*, where an event
      is one hooked PHP function call (e.g. curl_exec, file_get_contents, exec,
      PDO::query, ...). Each hook invocation pushes a new EventCache on the stack
      when it starts and pops it when it finishes, so nested hooks have independent
      contexts.
*/
RequestCache requestCache;
EventCacheStack eventCacheStack;

/*
    Reset helpers:

    These functions re-initialize the cache structs to their default state instead
    of reallocating them. The PHP extension code runs inside long-lived PHP/Apache/FPM
    processes that handle many HTTP requests. Because these cache objects live for
    the lifetime of the process, we must explicitly reset them so that no state
    from one request or event can leak into the next.
*/
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
