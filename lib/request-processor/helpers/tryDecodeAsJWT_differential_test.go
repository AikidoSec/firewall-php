package helpers

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

// buildDifferentialTokens returns two JWTs carrying the identical claim, differing
// only in the base64 alphabet of the payload segment (url-safe vs standard).
func buildDifferentialTokens(t *testing.T) (urlSafe, standard, claim string) {
	claim = "x' OR '1'='1"
	var std string
	for i := 1; i < 64; i++ {
		obj := map[string]interface{}{"z": strings.Repeat("\u00ff", i), "name": claim}
		payload, _ := json.Marshal(obj)
		cand := strings.TrimRight(base64.StdEncoding.EncodeToString(payload), "=")
		if strings.ContainsAny(cand, "+/") {
			std = cand
			break
		}
	}
	if std == "" {
		t.Fatal("could not build a payload with + or / in std base64")
	}
	url := strings.NewReplacer("+", "-", "/", "_").Replace(std)
	hdr := strings.TrimRight(base64.URLEncoding.EncodeToString([]byte(`{"alg":"none"}`)), "=")
	return hdr + "." + url + ".sig", hdr + "." + std + ".sig", claim
}

// A JWT whose payload uses standard base64 (+/) must be decoded and its claims
// scanned identically to the url-safe form.
func TestTryDecodeAsJWTStandardBase64(t *testing.T) {
	urlSafe, standard, claim := buildDifferentialTokens(t)

	if r := tryDecodeAsJWT(urlSafe); !r.JWT {
		t.Fatalf("url-safe JWT should decode")
	}
	if r := tryDecodeAsJWT(standard); !r.JWT {
		t.Fatalf("standard-base64 JWT should decode (differential bypass)")
	}

	urlStrings := ExtractStringsFromUserInput(urlSafe, []PathPart{}, 0)
	stdStrings := ExtractStringsFromUserInput(standard, []PathPart{}, 0)

	if _, ok := urlStrings[claim]; !ok {
		t.Errorf("url-safe token: claim %q not extracted", claim)
	}
	if _, ok := stdStrings[claim]; !ok {
		t.Errorf("standard-base64 token: claim %q not extracted -> WAF bypass", claim)
	}
}
