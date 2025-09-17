#include "Includes.h"

RequestCache requestCache;
EventCache eventCache;

void RequestCache::Reset() {
    *this = RequestCache();
}

void EventCache::Copy(EventCache& other) {
    *this = other;
}

void EventCache::Reset() {
    *this = EventCache();
}
