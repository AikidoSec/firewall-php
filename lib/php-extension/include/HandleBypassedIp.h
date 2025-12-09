#pragma once

// Initialize the IP bypass check at request start.
// Resets state and checks if the current IP should be bypassed.
// This should be called during request initialization.
void InitIpBypassCheck();

// Check if the current IP is bypassed.
bool IsCurrentIpBypassed();

// Set the IP bypass state to true.
void SetIpBypassed();

