package path_traversal

import (
	"main/utils"
	"testing"
)

func TestCheckContextForPathTraversal(t *testing.T) {

	unsafeContext := map[string]interface{}{

		"cookies":       map[string]interface{}{},
		"headers":       map[string]interface{}{},
		"remoteAddress": "ip",
		"method":        "POST",
		"url":           "url",
		"query":         map[string]interface{}{},
		"body":          map[string]interface{}{},
		"routeParams": map[string]interface{}{
			"path": "../file",
		},
		"source": "express",
		"route":  nil,
	}

	t.Run("it detects path traversal from route parameter", func(t *testing.T) {
		operation := "operation"
		result := CheckContextForPathTraversal("../file/test.txt", operation, unsafeContext, true)

		if result == nil {
			t.Errorf("expected result, got nil")
			return
		}
		if result.Operation != operation {
			t.Errorf("expected operation, got %v", result.Operation)
		}
		if result.Kind != utils.Kind("path_traversal") {
			t.Errorf("expected kind, got %v", result.Kind)
		}
		if result.Source != "routeParams" {
			t.Errorf("expected source, got %v", result.Source)
		}
		if result.PathToPayload != ".path" {
			t.Errorf("expected path to payload, got %v", result.PathToPayload)
		}
		if result.Metadata["filename"] != "../file/test.txt" {
			t.Errorf("expected filename, got %v", result.Metadata["filename"])
		}
		if result.Payload != "../file" {
			t.Errorf("expected payload, got %v", result.Payload)
		}

	})

	t.Run("it does not flag safe operation", func(t *testing.T) {
		operation := "path.normalize"
		safeContext := map[string]interface{}{
			"url":    "/_next/static/RjAvHy_jB1ciRT_xBrSyI/_ssgManifest.js",
			"method": "GET",
			"headers": map[string]interface{}{
				"host":               "localhost:3000",
				"connection":         "keep-alive",
				"pragma":             "no-cache",
				"cache-control":      "no-cache",
				"sec-ch-ua":          `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`,
				"sec-ch-ua-mobile":   "?0",
				"user-agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
				"sec-ch-ua-platform": `"macOS"`,
				"accept":             "*/*",
				"sec-fetch-site":     "same-origin",
				"sec-fetch-mode":     "no-cors",
				"sec-fetch-dest":     "script",
				"referer":            "http://localhost:3000/",
				"accept-encoding":    "gzip, deflate, br, zstd",
				"accept-language":    "nl,en;q=0.9,en-US;q=0.8",
				"cookie":             "Phpstorm-8262f4a6=6a1925f9-2f0e-45ea-8336-a6988d56b1aa",
				"x-forwarded-host":   "localhost:3000",
				"x-forwarded-port":   "3000",
				"x-forwarded-proto":  "http",
				"x-forwarded-for":    "127.0.0.1",
			},
			"route":         nil,
			"query":         map[string]interface{}{},
			"source":        "http.createServer",
			"routeParams":   map[string]interface{}{},
			"cookies":       map[string]interface{}{"Phpstorm-8262f4a6": "6a1925f9-2f0e-45ea-8336-a6988d56b1aa"},
			"body":          nil,
			"remoteAddress": "127.0.0.1",
		}

		result := CheckContextForPathTraversal("../../web/spec-extension/cookies", operation, safeContext, true)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})
}
