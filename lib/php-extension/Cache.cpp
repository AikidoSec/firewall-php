#include "Includes.h"

static void* GetCurrentFiberContext() {
#if PHP_VERSION_ID >= 80100
    return EG(current_fiber_context);
#else
    return nullptr;
#endif
}

void RequestCache::Reset() {
    *this = RequestCache();
}

void RequestCache::EnterIdorIgnoredScope() {
    if (void* fiberContext = GetCurrentFiberContext()) {
        idorIgnoredDepthByFiber[fiberContext]++;
        return;
    }
    idorIgnoredDepth++;
}

void RequestCache::LeaveIdorIgnoredScope() {
    if (void* fiberContext = GetCurrentFiberContext()) {
        auto it = idorIgnoredDepthByFiber.find(fiberContext);
        if (it == idorIgnoredDepthByFiber.end()) {
            return;
        }
        if (it->second <= 1) {
            idorIgnoredDepthByFiber.erase(it);
        } else {
            it->second--;
        }
        return;
    }
    if (idorIgnoredDepth > 0) {
        idorIgnoredDepth--;
    }
}

bool RequestCache::IsIdorIgnored() const {
    if (void* fiberContext = GetCurrentFiberContext()) {
        auto it = idorIgnoredDepthByFiber.find(fiberContext);
        return it != idorIgnoredDepthByFiber.end() && it->second > 0;
    }
    return idorIgnoredDepth > 0;
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
    AIKIDO_GLOBAL(eventCacheStack).Push();
}

ScopedEventContext::~ScopedEventContext() {
    AIKIDO_GLOBAL(eventCacheStack).Pop();
}
