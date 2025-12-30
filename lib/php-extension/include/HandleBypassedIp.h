#pragma once

// Check if Aikido is disabled or the current IP is bypassed.
// The IP bypass check is performed lazily on first call.
bool IsAikidoDisabledOrBypassed();
