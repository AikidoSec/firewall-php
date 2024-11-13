package path_traversal

import (
	. "main/aikido_types"
	"main/context"
	"main/utils"
	"testing"
)

func TestCheckContextForPathTraversal(t *testing.T) {

	t.Run("it detects path traversal from body parameter", func(t *testing.T) {
		context.LoadForUnitTests(map[string]string{
			"remoteAddress": "ip",
			"method":        "POST",
			"url":           "url",
			"body":          context.GetJsonString(map[string]interface{}{"path": "../file"}),
			"source":        "express",
		})

		operation := "operation"
		result := CheckContextForPathTraversal(&FileAccessed{Filename: "../file/test.txt", Operation: operation})

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
		if result.Source != "body" {
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
		context.LoadForUnitTests(map[string]string{
			"remoteAddress": "ip",
			"method":        "POST",
			"url":           "url",
		})

		operation := "path.normalize"
		context.LoadForUnitTests(map[string]string{
			"url":    "/_next/static/RjAvHy_jB1ciRT_xBrSyI/_ssgManifest.js",
			"method": "GET",
			"headers": context.GetJsonString(map[string]interface{}{
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
			}),
			"source":        "http.createServer",
			"cookies":       context.GetJsonString(map[string]interface{}{"Phpstorm-8262f4a6": "6a1925f9-2f0e-45ea-8336-a6988d56b1aa"}),
			"remoteAddress": "127.0.0.1",
		})

		result := CheckContextForPathTraversal(&FileAccessed{Filename: "../../web/spec-extension/cookies", Operation: operation})
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})
}
