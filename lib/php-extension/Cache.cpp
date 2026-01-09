#include "Includes.h"

RequestCache requestCache;
EventCacheStack eventCacheStack;

void RequestCache::Reset() {
    *this = RequestCache();
}

void EventCache::Reset() {
    *this = EventCache();
}

void EventCacheStack::Push() {
    contexts.push(EventCache());
}

void EventCacheStack::Pop() {
    if (!contexts.empty()) {
        contexts.pop();
    }
}

EventCache& EventCacheStack::Current() {
    return contexts.top();
}

bool EventCacheStack::Empty() {
    return contexts.empty();
}

ScopedEventContext::ScopedEventContext() {
    eventCacheStack.Push();
}

ScopedEventContext::~ScopedEventContext() {
    eventCacheStack.Pop();
}
