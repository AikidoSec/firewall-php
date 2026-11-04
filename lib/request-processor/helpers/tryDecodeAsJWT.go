package helpers

import (
	"encoding/base64"
	"strings"
)

type JWTDecodeResult struct {
	JWT    bool
	Object interface{}
}

func removePadding(s string) string {
	return strings.TrimRight(strings.TrimLeft(s, "="), "=")
}

// decodeBase64Segment decodes a JWT segment as leniently as common application
// decoders. Application-side JWT libraries frequently accept both URL-safe
// (-_) and standard (+/) base64 alphabets, so the firewall must too, otherwise
// claims can be smuggled past scanning by switching the alphabet.
func decodeBase64Segment(segment string) ([]byte, bool) {
	segment = removePadding(segment)
	encodings := []*base64.Encoding{
		base64.RawURLEncoding,
		base64.RawStdEncoding,
	}
	for _, enc := range encodings {
		if payload, err := enc.DecodeString(segment); err == nil {
			return payload, true
		}
	}
	return nil, false
}

func tryDecodeAsJWT(jwt string) JWTDecodeResult {
	if !strings.Contains(jwt, ".") {
		return JWTDecodeResult{JWT: false}
	}
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return JWTDecodeResult{JWT: false}
	}

	payload, ok := decodeBase64Segment(parts[1])
	if !ok {
		return JWTDecodeResult{JWT: false}
	}

	var object interface{}
	err := ParseJSON(payload, &object)

	if err != nil {
		return JWTDecodeResult{JWT: false}
	}

	return JWTDecodeResult{JWT: true, Object: object}
}
