package shell_injection

import (
	"main/context"
	"main/utils"
	zen_internals "main/vulnerabilities/zen-internals"
	"testing"
)

func TestCheckContextForShellInjection(t *testing.T) {
	zen_internals.InitZenInternals()
	t.Run("it detects shell injection", func(t *testing.T) {
		context.LoadForUnitTests(map[string]string{
			"remoteAddress": "ip",
			"method":        "POST",
			"url":           "url",
			"body": context.GetJsonString(map[string]interface{}{
				"domain": "www.example`whoami`.com",
			}),
			"source": "express",
			"route":  "/",
		})
		operation := "child_process.exec"
		result := CheckContextForShellInjection("binary --domain www.example`whoami`.com", operation)

		if result == nil {
			t.Errorf("expected result, got nil")
			return
		}
		if result.Operation != operation {
			t.Errorf("expected operation, got %v", result.Operation)
		}
		if result.Kind != utils.Kind("shell_injection") {
			t.Errorf("expected kind, got %v", result.Kind)
		}
		if result.Source != "body" {
			t.Errorf("expected source, got %v", result.Source)
		}
		if result.PathToPayload != ".domain" {
			t.Errorf("expected path to payload, got %v", result.PathToPayload)
		}
		if result.Metadata["command"] != "binary --domain www.example`whoami`.com" {
			t.Errorf("expected command, got %v", result.Metadata["command"])
		}
		if result.Payload != "www.example`whoami`.com" {
			t.Errorf("expected payload, got %v", result.Payload)
		}

	})

	t.Run("it detects shell injection from route params", func(t *testing.T) {
		operation := "child_process.exec"
		context.LoadForUnitTests(map[string]string{
			"remoteAddress": "ip",
			"method":        "POST",
			"url":           "url",
			"body": context.GetJsonString(map[string]interface{}{
				"domain": "www.example`whoami`.com",
			}),
			"source": "express",
			"route":  "/",
		})

		result := CheckContextForShellInjection("binary --domain www.example`whoami`.com", operation)

		if result == nil {
			t.Errorf("expected result, got nil")
			return
		}
		if result.Operation != operation {
			t.Errorf("expected operation, got %v", result.Operation)
		}
		if result.Kind != utils.Kind("shell_injection") {
			t.Errorf("expected kind, got %v", result.Kind)
		}
		if result.Source != "body" {
			t.Errorf("expected source, got %v", result.Source)
		}
		if result.PathToPayload != ".domain" {
			t.Errorf("expected path to payload, got %v", result.PathToPayload)
		}
		if result.Metadata["command"] != "binary --domain www.example`whoami`.com" {
			t.Errorf("expected command, got %v", result.Metadata["command"])
		}
		if result.Payload != "www.example`whoami`.com" {
			t.Errorf("expected payload, got %v", result.Payload)
		}
	})

	zen_internals.CloseZenInternals()
}
