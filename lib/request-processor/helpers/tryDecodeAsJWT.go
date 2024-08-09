package helpers

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

type JWTDecodeResult struct {
	JWT    bool
	Object interface{}
}

func removePadding(s string) string {
	return strings.TrimRight(strings.TrimLeft(s, "="), "=")
}

func tryDecodeAsJWT(jwt string) JWTDecodeResult {
	if !strings.Contains(jwt, ".") {
		return JWTDecodeResult{JWT: false}
	}
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return JWTDecodeResult{JWT: false}
	}
	//remove padding

	payload, err := base64.RawURLEncoding.DecodeString(removePadding(parts[1]))
	if err != nil {
		return JWTDecodeResult{JWT: false}
	}

	var object interface{}
	err = json.Unmarshal(payload, &object)
	if err != nil {
		return JWTDecodeResult{JWT: false}
	}

	return JWTDecodeResult{JWT: true, Object: object}
}
