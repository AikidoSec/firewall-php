package aikido_types

import (
	"crypto/rand"
	"fmt"
	"time"
)

// The custom Go toolchain we use for Alpine has no uuid package
func newSessionID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "unknown"
	}

	// 48-bit millisecond timestamp, big-endian
	ms := time.Now().UnixMilli()
	for i := range 6 {
		b[i] = byte(ms >> (8 * (5 - i)) & 0xff) //nolint:gosec // masked to a single byte
	}

	b[6] = (b[6] & 0x0f) | 0x70 // version 7
	b[8] = (b[8] & 0x3f) | 0x80 // RFC 9562 variant

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
