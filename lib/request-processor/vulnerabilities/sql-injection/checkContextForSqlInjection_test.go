package sql_injection

import (
	"main/context"
	"main/utils"
	zen_internals "main/vulnerabilities/zen-internals"
	"testing"
)

func TestCheckContextForSqlInjection(t *testing.T) {
	zen_internals.InitZenInternals()

	sql := "SELECT * FROM users WHERE id = '1' OR 1=1; -- '"
	operation := "mysql.query"
	context.LoadForUnitTests(map[string]string{
		"remoteAddress": "ip",
		"method":        "POST",
		"url":           "url",
		"body":          context.GetJsonString(map[string]interface{}{"id": "1' OR 1=1; --"}),
		"source":        "express",
		"route":         "/",
	})

	result := CheckContextForSqlInjection(sql, operation, "mysql")

	if result == nil {
		t.Errorf("Expected result, got nil")
		return
	}
	if result.Operation != operation {
		t.Errorf("Expected operation %s, got %s", operation, result.Operation)
	}

	if result.Kind != utils.Kind("sql_injection") {
		t.Errorf("Expected kind %s, got %s", utils.Kind("sql_injection"), result.Kind)
	}
	if result.Source != "body" {
		t.Errorf("Expected source body, got %s", result.Source)
	}
	if result.PathToPayload != ".id" {
		t.Errorf("Expected pathToPayload .id, got %s", result.PathToPayload)
	}
	if result.Payload != "1' OR 1=1; --" {
		t.Errorf("Expected payload 1' OR 1=1; --, got %s", result.Payload)
	}

	zen_internals.CloseZenInternals()
}
