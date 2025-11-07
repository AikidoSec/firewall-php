package helpers

import (
	"reflect"
	"testing"
)

func TestExtractStringsFromUserInput(t *testing.T) {
	t.Run("empty object returns empty array", func(t *testing.T) {
		obj := map[string]interface{}{}
		pathToPayload := []PathPart{}
		expected := map[string]string{}
		actual := ExtractStringsFromUserInput(obj, pathToPayload, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}
	})

	t.Run("it can extract query objects", func(t *testing.T) {
		obj := map[string]interface{}{
			"age": map[string]interface{}{
				"$gt": "21",
			},
		}

		expected := map[string]string{
			"age": ".",
			"$gt": ".age",
			"21":  ".age.$gt",
		}
		actual := ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		obj = map[string]interface{}{
			"title": map[string]interface{}{
				"$ne": "null",
			},
		}

		expected = map[string]string{
			"title": ".",
			"$ne":   ".title",
			"null":  ".title.$ne",
		}
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		obj = map[string]interface{}{
			"age":        "whaat1",
			"user_input": []string{"whaat", "dangerous"},
		}

		expected = map[string]string{
			"user_input":      ".",
			"age":             ".",
			"whaat1":          ".age",
			"whaat":           ".user_input.[0]",
			"dangerous":       ".user_input.[1]",
			"whaat,dangerous": ".user_input",
		}
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

	})

	t.Run("it can extract cookie objects", func(t *testing.T) {
		obj := map[string]interface{}{
			"session":  "ABC",
			"session2": "DEF",
		}

		expected := map[string]string{
			"session2": ".",
			"session":  ".",
			"ABC":      ".session",
			"DEF":      ".session2",
		}
		actual := ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		obj = map[string]interface{}{
			"session":  "ABC",
			"session2": 1234,
		}

		expected = map[string]string{
			"session2": ".",
			"session":  ".",
			"ABC":      ".session",
		}
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}
	})

	t.Run("it can extract header objects", func(t *testing.T) {
		obj := map[string]interface{}{
			"Content-Type": "application/json",
		}

		expected := map[string]string{
			"Content-Type":     ".",
			"application/json": ".Content-Type",
		}
		actual := ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		obj = map[string]interface{}{
			"Content-Type": 54321,
		}
		expected = map[string]string{
			"Content-Type": ".",
		}
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		obj = map[string]interface{}{
			"Content-Type": "application/json",
			"ExtraHeader":  "value",
		}
		expected = map[string]string{
			"Content-Type":     ".",
			"application/json": ".Content-Type",
			"ExtraHeader":      ".",
			"value":            ".ExtraHeader",
		}
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}
	})

	t.Run("it can extract body objects", func(t *testing.T) {
		obj := map[string]interface{}{
			"nested": map[string]interface{}{
				"nested": map[string]interface{}{
					"$ne": nil,
				},
			},
		}

		expected := map[string]string{
			"nested": ".nested",
			"$ne":    ".nested.nested",
		}
		actual := ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		obj = map[string]interface{}{
			"age": map[string]interface{}{
				"$gt": "21",
				"$lt": "100",
			},
		}

		expected = map[string]string{
			"age": ".",
			"$lt": ".age",
			"$gt": ".age",
			"21":  ".age.$gt",
			"100": ".age.$lt",
		}
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}
	})

	t.Run("it decodes JWTs", func(t *testing.T) {
		obj := map[string]interface{}{
			"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwidXNlcm5hbWUiOnsiJG5lIjpudWxsfSwiaWF0IjoxNTE2MjM5MDIyfQ._jhGJw9WzB6gHKPSozTFHDo9NOHs3CNOlvJ8rWy6VrQ",
		}

		expected := map[string]string{
			"token":      ".",
			"iat":        ".token<jwt>",
			"username":   ".token<jwt>",
			"sub":        ".token<jwt>",
			"1234567890": ".token<jwt>.sub",
			"1516239022": ".token<jwt>.iat",
			"$ne":        ".token<jwt>.username",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwidXNlcm5hbWUiOnsiJG5lIjpudWxsfSwiaWF0IjoxNTE2MjM5MDIyfQ._jhGJw9WzB6gHKPSozTFHDo9NOHs3CNOlvJ8rWy6VrQ": ".token",
		}

		actual := ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}
	})

	t.Run("it also adds the JWT itself as string", func(t *testing.T) {
		obj := map[string]interface{}{
			"header": "/;ping%20localhost;.e30=.",
		}

		expected := map[string]string{
			"header":                    ".",
			"/;ping%20localhost;.e30=.": ".header",
		}

		actual := ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}
	})

	t.Run("it concatenates array values", func(t *testing.T) {
		obj := map[string]interface{}{
			"arr": []interface{}{"1", "2", "3"},
		}

		expected := map[string]string{
			"arr":   ".",
			"1,2,3": ".arr",
			"1":     ".arr.[0]",
			"2":     ".arr.[1]",
			"3":     ".arr.[2]",
		}
		actual := ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		obj = map[string]interface{}{
			"arr": []interface{}{"1", 2, true, nil, nil, map[string]interface{}{"test": "test"}},
		}

		expected = map[string]string{
			"arr":  ".",
			"1":    ".arr.[0]",
			"test": ".arr.[5].test",
			"1,<int Value>,<bool Value>,<invalid Value>,<invalid Value>,<map[string]interface {} Value>": ".arr",
		}
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		obj = map[string]interface{}{
			"arr": []interface{}{"1", 2, true, nil, nil, map[string]interface{}{"test": []string{"test123", "test345"}}},
		}

		expected = map[string]string{
			"arr":             ".",
			"1":               ".arr.[0]",
			"test":            ".arr.[5]",
			"test123":         ".arr.[5].test.[0]",
			"test345":         ".arr.[5].test.[1]",
			"test123,test345": ".arr.[5].test",
			"1,<int Value>,<bool Value>,<invalid Value>,<invalid Value>,<map[string]interface {} Value>": ".arr",
		}
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}
	})

	t.Run("it respects depth parameter", func(t *testing.T) {
		// Test with depth 0 - should work normally
		obj := map[string]interface{}{
			"key": "value",
		}
		expected := map[string]string{
			"key":   ".",
			"value": ".key",
		}
		actual := ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		// Test with depth 1 - should work normally
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 1)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		// Test with depth at maxDepth (1024) - key will be extracted but value won't
		// because processing the value would require depth 1025
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 1024)
		expected = map[string]string{
			"key": ".",
		}
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		// Test with depth exceeding maxDepth (1025) - should return empty
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 1025)
		expected = map[string]string{}
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected empty map when depth exceeds maxDepth, got %v", actual)
		}

		// Test with very large depth - should return empty
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 10000)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected empty map when depth is very large, got %v", actual)
		}
	})

	t.Run("it stops recursion at maxDepth", func(t *testing.T) {
		// Create a deeply nested structure that would exceed maxDepth
		// Build a nested map that goes 1025 levels deep
		deepObj := map[string]interface{}{}
		current := deepObj
		for i := 0; i < 1025; i++ {
			next := map[string]interface{}{}
			current["nested"] = next
			current = next
		}
		current["value"] = "should_not_be_extracted"

		// When starting at depth 0, it should stop before extracting the deep value
		actual := ExtractStringsFromUserInput(deepObj, []PathPart{}, 0)

		// The value at depth 1025+ should not be extracted
		if _, exists := actual["should_not_be_extracted"]; exists {
			t.Error("Expected deep value to not be extracted when exceeding maxDepth")
		}

		// But keys at shallower depths should still be extracted
		if _, exists := actual["nested"]; !exists {
			t.Error("Expected nested keys at shallow depths to still be extracted")
		}
	})

	t.Run("it tracks depth correctly through nested structures", func(t *testing.T) {
		// Test that depth increments correctly through nested maps
		obj := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": "value",
				},
			},
		}

		expected := map[string]string{
			"level1": ".",
			"level2": ".level1",
			"level3": ".level1.level2",
			"value":  ".level1.level2.level3",
		}
		actual := ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		// Test that depth increments correctly through arrays
		obj = map[string]interface{}{
			"arr": []interface{}{
				[]interface{}{
					[]interface{}{"deep_value"},
				},
			},
		}

		expected = map[string]string{
			"arr":                    ".",
			"deep_value":             ".arr.[0].[0]",
			"<[]interface {} Value>": ".arr",
		}
		actual = ExtractStringsFromUserInput(obj, []PathPart{}, 0)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}
	})

}

func TestExtractResourceOrOriginal(t *testing.T) {
	t.Run("php://filter/convert.base64-encode/resource=/etc/passwd", func(t *testing.T) {
		if ExtractResourceOrOriginal("php://filter/convert.base64-encode/resource=/etc/passwd") != "/etc/passwd" {
			t.Error("expected /etc/passwd")
		}
	})
	t.Run("php://filter/convert.base64-encode/resource=../../../../file", func(t *testing.T) {
		if ExtractResourceOrOriginal("php://filter/convert.base64-encode/resource=../../../../file") != "../../../../file" {
			t.Error("expected ../../../../file")
		}
	})
	t.Run("php://filter/resource=php://filter/resource=../../../../file", func(t *testing.T) {
		if ExtractResourceOrOriginal("php://filter/resource=php://filter/resource=../../../../file") != "../../../../file" {
			t.Error("expected ../../../../file")
		}
	})
	t.Run("file.txt", func(t *testing.T) {
		if ExtractResourceOrOriginal("file.txt") != "file.txt" {
			t.Error("expected file.txt")
		}
	})
	t.Run("test.txt/resource=../../../../file", func(t *testing.T) {
		if ExtractResourceOrOriginal("test.txt/resource=../../../../file") != "test.txt/resource=../../../../file" {
			t.Error("expected test.txt/resource=../../../../file")
		}
	})
	t.Run("php://filter/", func(t *testing.T) {
		if ExtractResourceOrOriginal("php://filter/") != "php://filter/" {
			t.Error("expected php://filter/")
		}
	})
	t.Run("php://filter/resource=", func(t *testing.T) {
		if ExtractResourceOrOriginal("php://filter/resource=") != "" {
			t.Error("expected empty")
		}
	})

	t.Run("Case insensitive", func(t *testing.T) {
		if ExtractResourceOrOriginal("php://FiltEr/convert.base64-encode/resource=/etc/passwd") != "/etc/passwd" {
			t.Error("expected /etc/passwd")
		}
	})

}
