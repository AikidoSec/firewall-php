#include "Includes.h"

RequestCache requestCache;
EventCache eventCache;

void RequestCache::Reset() {
    *this = RequestCache();
}

void EventCache::Reset() {
    *this = EventCache();
}
