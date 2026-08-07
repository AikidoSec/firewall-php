package path_traversal

import (
	"main/context"
	"main/utils"
	"strings"
	"testing"
	"time"
)

func TestCheckContextForPathTraversal(t *testing.T) {

	t.Run("it detects path traversal from body parameter", func(t *testing.T) {
		instance := context.LoadForUnitTests(map[string]string{
			"remoteAddress": "ip",
			"method":        "POST",
			"url":           "url",
			"body":          context.GetJsonString(map[string]interface{}{"path": "../file"}),
			"source":        "express",
		})

		operation := "operation"
		result := CheckContextForPathTraversal(instance, "../file/test.txt", operation, true)

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
		operation := "path.normalize"
		instance := context.LoadForUnitTests(map[string]string{
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

		result := CheckContextForPathTraversal(instance, "../../web/spec-extension/cookies", operation, true)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})
}

func TestSanitizePath(t *testing.T) {
	t.Run("file://test.txt", func(t *testing.T) {
		if SanitizePath("file://test.txt") != "test.txt" {
			t.Error("expected test.txt")
		}
	})

	t.Run("FILE://test.txt", func(t *testing.T) {
		if SanitizePath("FILE://test.txt") != "test.txt" {
			t.Error("expected test.txt")
		}
	})

	t.Run("fIle://test.txt", func(t *testing.T) {
		if SanitizePath("fIle://test.txt") != "test.txt" {
			t.Error("expected test.txt")
		}
	})
	t.Run("test.txt", func(t *testing.T) {
		if SanitizePath("test.txt") != "test.txt" {
			t.Error("expected test.txt")
		}
	})
	t.Run("One character", func(t *testing.T) {
		if SanitizePath("t") != "t" {
			t.Error("expected t")
		}
	})

}

func TestCheckContextForPathTraversal_ManyCallsOnLargeBody(t *testing.T) {
	instance := context.LoadForUnitTests(map[string]string{
		"remoteAddress": "127.0.0.1",
		"method":        "POST",
		"url":           "/mail/inbox",
		"body":          context.GetJsonString(map[string]interface{}{"rawMime": strings.Repeat("A", 3*1024*1024)}),
		"source":        "laravel",
	})
	defer context.UnloadForUnitTests()

	start := time.Now()
	for i := 0; i < 200; i++ {
		if result := CheckContextForPathTraversal(instance, "/tmp/php_mime_tmp.eml", "file_get_contents", true); result != nil {
			t.Fatalf("unexpected path traversal flagged: %+v", result)
		}
	}
	elapsed := time.Since(start)
	if elapsed > time.Second {
		t.Fatalf("200 calls took %v, expected well under 1s - path traversal candidates are not being cached per request", elapsed)
	}
}

func TestCheckContextForPathTraversal_CacheInvalidatedBetweenRequests(t *testing.T) {
	instance := context.LoadForUnitTests(map[string]string{
		"remoteAddress": "127.0.0.1",
		"method":        "POST",
		"url":           "/mail/inbox",
		"body":          context.GetJsonString(map[string]interface{}{"attack": "../../etc/shadow"}),
		"source":        "laravel",
	})
	defer context.UnloadForUnitTests()

	if result := CheckContextForPathTraversal(instance, "/var/www/../../etc/shadow", "file_get_contents", true); result == nil {
		t.Fatalf("expected first request's payload to be detected")
	}

	context.Clear(instance)
	context.TestContext = map[string]string{
		"remoteAddress": "127.0.0.1",
		"method":        "POST",
		"url":           "/mail/inbox",
		"body":          context.GetJsonString(map[string]interface{}{"attack": "../../var/log/app.log"}),
		"source":        "laravel",
	}

	if result := CheckContextForPathTraversal(instance, "/var/www/../../etc/shadow", "file_get_contents", true); result != nil {
		t.Fatalf("expected no match, got %+v - stale candidates from the previous request leaked in", result)
	}

	if result := CheckContextForPathTraversal(instance, "/opt/app/../../var/log/app.log", "file_get_contents", true); result == nil {
		t.Fatalf("expected second request's own payload to still be detected")
	}
}
