package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"math/big"
	"testing"
)

var (
	lower    = "abcdefghijklmnopqrstuvwxyz"
	upper    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers  = "0123456789"
	specials = "!#$%^&*|;:<>"
)

func secretFromCharset(length int, charset string) string {
	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

func TestLooksLikeASecret(t *testing.T) {
	t.Run("it returns false for empty string", func(t *testing.T) {
		if LooksLikeASecret("") {
			t.Errorf("expected false for empty string")
		}
	})

	t.Run("it returns false for short strings", func(t *testing.T) {
		shortStrings := []string{"c", "NR", "7t3", "4qEK", "KJr6s", "KXiW4a", "Fupm2Vi", "jiGmyGfg", "SJPLzVQ8t", "OmNf04j6mU"}
		for _, s := range shortStrings {
			if LooksLikeASecret(s) {
				t.Errorf("expected false for short string %s", s)
			}
		}
	})

	t.Run("it returns true for long strings", func(t *testing.T) {
		longStrings := []string{"rsVEExrR2sVDONyeWwND", ":2fbg;:qf$BRBc<2AG8&"}
		for _, s := range longStrings {
			if !LooksLikeASecret(s) {
				t.Errorf("expected true for long string %s", s)
			}
		}
	})

	t.Run("it flags very long strings", func(t *testing.T) {
		veryLongString := "efDJHhzvkytpXoMkFUgag6shWJktYZ5QUrUCTfecFELpdvaoAT3tekI4ZhpzbqLt"
		if !LooksLikeASecret(veryLongString) {
			t.Errorf("expected true for very long string")
		}
	})

	t.Run("it flags very very long strings", func(t *testing.T) {
		veryVeryLongString := "XqSwF6ySwMdTomIdmgFWcMVXWf5L0oVvO5sIjaCPI7EjiPvRZhZGWx3A6mLl1HXPOHdUeabsjhngW06JiLhAchFwgtUaAYXLolZn75WsJVKHxEM1mEXhlmZepLCGwRAM"
		if !LooksLikeASecret(veryVeryLongString) {
			t.Errorf("expected true for very very long string")
		}
	})

	t.Run("it returns false if contains white space", func(t *testing.T) {
		if LooksLikeASecret("rsVEExrR2sVDONyeWwND ") {
			t.Errorf("expected false for string with white space")
		}
	})

	t.Run("it returns false if it has less than 2 charsets", func(t *testing.T) {
		if LooksLikeASecret(secretFromCharset(10, lower)) {
			t.Errorf("expected false for string with only lower case letters")
		}
		if LooksLikeASecret(secretFromCharset(10, upper)) {
			t.Errorf("expected false for string with only upper case letters")
		}
		if LooksLikeASecret(secretFromCharset(10, numbers)) {
			t.Errorf("expected false for string with only numbers")
		}
		if LooksLikeASecret(secretFromCharset(10, specials)) {
			t.Errorf("expected false for string with only special characters")
		}
	})

	urlTerms := []string{
		"development", "programming", "applications", "implementation", "environment", "technologies",
		"documentation", "demonstration", "configuration", "administrator", "visualization",
		"international", "collaboration", "opportunities", "functionality", "customization",
		"specifications", "optimization", "contributions", "accessibility", "subscription",
		"subscriptions", "infrastructure", "architecture", "authentication", "sustainability",
		"notifications", "announcements", "recommendations", "communication", "compatibility",
		"enhancement", "integration", "performance", "improvements", "introduction", "capabilities",
		"communities", "credentials", "integration", "permissions", "validation", "serialization",
		"deserialization", "rate-limiting", "throttling", "load-balancer", "microservices",
		"endpoints", "data-transfer", "encryption", "authorization", "bearer-token", "multipart",
		"urlencoded", "api-docs", "postman", "json-schema", "serialization", "deserialization",
		"rate-limiting", "throttling", "load-balancer", "api-gateway", "microservices", "endpoints",
		"data-transfer", "encryption", "signature", "poppins-bold-webfont.woff2", "karla-bold-webfont.woff2",
		"startEmailBasedLogin", "jenkinsFile", "ConnectionStrings.config", "coach", "login", "payment_methods",
		"activity_logs", "feedback_responses", "balance_transactions", "customer_sessions", "payment_intents",
		"billing_portal", "subscription_items", "namedLayouts", "PlatformAction", "quickActions", "queryLocator",
		"relevantItems", "parameterizedSearch",
	}

	t.Run("it returns false for common url terms", func(t *testing.T) {
		for _, term := range urlTerms {
			if LooksLikeASecret(term) {
				t.Errorf("expected false for common url term %s", term)
			}
		}
	})

	t.Run("it returns false for known word separators", func(t *testing.T) {
		if LooksLikeASecret("this-is-a-secret-1") {
			t.Errorf("expected false for string with known word separators")
		}
	})

	t.Run("a number is not a secret", func(t *testing.T) {
		if LooksLikeASecret("1234567890") {
			t.Errorf("expected false for number string 1234567890")
		}
		if LooksLikeASecret("12345678901234567890") {
			t.Errorf("expected false for number string 12345678901234567890")
		}
	})

	secrets := []string{
		"yqHYTS<agpi^aa1",
		"hIofuWBifkJI5iVsSNKKKDpBfmMqJJwuXMxau6AS8WZaHVLDAMeJXo3BwsFyrIIm",
		"AG7DrGi3pDDIUU1PrEsj",
		"CnJ4DunhYfv2db6T1FRfciRBHtlNKOYrjoz",
		"Gic*EfMq:^MQ|ZcmX:yW1",
		"AG7DrGi3pDDIUU1PrEsj",
	}

	t.Run("it returns true for known secrets", func(t *testing.T) {
		for _, secret := range secrets {
			if !LooksLikeASecret(secret) {
				t.Errorf("expected true for known secret %s", secret)
			}
		}
	})
}

func generateHash(algorithm string) string {
	data := []byte("test")

	switch algorithm {
	case "md5":
		hash := md5.Sum(data)
		return hex.EncodeToString(hash[:])
	case "sha1":
		hash := sha1.Sum(data)
		return hex.EncodeToString(hash[:])
	case "sha256":
		hash := sha256.Sum256(data)
		return hex.EncodeToString(hash[:])
	case "sha512":
		hash := sha512.Sum512(data)
		return hex.EncodeToString(hash[:])
	default:
		return ""
	}
}

func TestBuildRouteFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"", ""},
		{"http", "http"},
		{"/", "/"},
		{"http://localhost/", "/"},
		{"/posts/3", "/posts/:number"},
		{"http://localhost/posts/3", "/posts/:number"},
		{"http://localhost/posts/3/", "/posts/:number"},
		{"http://localhost/posts/3/comments/10", "/posts/:number/comments/:number"},
		{"/blog/2023/05/great-article", "/blog/:number/:number/great-article"},
		{"/posts/2023-05-01", "/posts/:date"},
		{"/posts/2023-05-01/", "/posts/:date"},
		{"/posts/2023-05-01/comments/2023-05-01", "/posts/:date/comments/:date"},
		{"/posts/01-05-2023", "/posts/:date"},
		{"/posts/3,000", "/posts/3,000"},
		{"/v1/posts/3", "/v1/posts/:number"},
		{"/posts/d9428888-122b-11e1-b85c-61cd3cbb3210", "/posts/:uuid"},
		{"/posts/000003e8-2363-21ef-b200-325096b39f47", "/posts/:uuid"},
		{"/posts/a981a0c2-68b1-35dc-bcfc-296e52ab01ec", "/posts/:uuid"},
		{"/posts/109156be-c4fb-41ea-b1b4-efe1671c5836", "/posts/:uuid"},
		{"/posts/90123e1c-7512-523e-bb28-76fab9f2f73d", "/posts/:uuid"},
		{"/posts/1ef21d2f-1207-6660-8c4f-419efbd44d48", "/posts/:uuid"},
		{"/posts/017f22e2-79b0-7cc3-98c4-dc0c0c07398f", "/posts/:uuid"},
		{"/posts/0d8f23a0-697f-83ae-802e-48f3756dd581", "/posts/:uuid"},
		{"/posts/00000000-0000-1000-6000-000000000000", "/posts/00000000-0000-1000-6000-000000000000"},
		{"/posts/abc", "/posts/abc"},
		{"/login/john.doe@acme.com", "/login/:email"},
		{"/login/john.doe+alias@acme.com", "/login/:email"},
		{"/block/1.2.3.4", "/block/:ip"},
		{"/block/2001:2:ffff:ffff:ffff:ffff:ffff:ffff", "/block/:ip"},
		{"/block/64:ff9a::255.255.255.255", "/block/:ip"},
		{"/block/100::", "/block/:ip"},
		{"/block/fec0::", "/block/:ip"},
		{"/block/227.202.96.196", "/block/:ip"},
		{"/files/" + generateHash("md5"), "/files/:hash"},
		{"/files/" + generateHash("sha1"), "/files/:hash"},
		{"/files/" + generateHash("sha256"), "/files/:hash"},
		{"/files/" + generateHash("sha512"), "/files/:hash"},
		{"/confirm/CnJ4DunhYfv2db6T1FRfciRBHtlNKOYrjoz", "/confirm/:secret"},
	}

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			result := BuildRouteFromURL(test.url)
			if result != test.expected {
				t.Errorf("expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestParseFormData(t *testing.T) {
	data := "a=1&b=2"
	result := ParseFormData(data, "&")
	if result["a"] != "1" {
		t.Errorf("Expected 1, got %v", result["a"])
	}
	if result["b"] != "2" {
		t.Errorf("Expected 2, got %v", result["b"])
	}
}

func TestParseContext(t *testing.T) {
	context := map[string]interface{}{"body": "{\"a\":\"1\", \"b\":\"2\"}", "query": "c=3&d=4"}
	result := ParseContext(context)

	if result["body"].(map[string]interface{})["a"] != "1" {
		t.Errorf("Expected 1, got %v", result["body"].(map[string]interface{})["a"])
	}
	if result["body"].(map[string]interface{})["b"] != "2" {
		t.Errorf("Expected 2, got %v", result["body"].(map[string]interface{})["b"])
	}
	if result["query"].(map[string]interface{})["c"] != "3" {
		t.Errorf("Expected 3, got %v", result["query"].(map[string]interface{})["c"])
	}
	if result["query"].(map[string]interface{})["d"] != "4" {
		t.Errorf("Expected 4, got %v", result["query"].(map[string]interface{})["d"])
	}
}
