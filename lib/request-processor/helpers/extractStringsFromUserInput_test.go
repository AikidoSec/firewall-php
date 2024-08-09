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
		actual := ExtractStringsFromUserInput(obj, pathToPayload)
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
		actual := ExtractStringsFromUserInput(obj, []PathPart{})
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
		actual = ExtractStringsFromUserInput(obj, []PathPart{})
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
		actual = ExtractStringsFromUserInput(obj, []PathPart{})
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
		actual := ExtractStringsFromUserInput(obj, []PathPart{})
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
		actual = ExtractStringsFromUserInput(obj, []PathPart{})
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
		actual := ExtractStringsFromUserInput(obj, []PathPart{})
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}

		obj = map[string]interface{}{
			"Content-Type": 54321,
		}
		expected = map[string]string{
			"Content-Type": ".",
		}
		actual = ExtractStringsFromUserInput(obj, []PathPart{})
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
		actual = ExtractStringsFromUserInput(obj, []PathPart{})
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
		actual := ExtractStringsFromUserInput(obj, []PathPart{})
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
		actual = ExtractStringsFromUserInput(obj, []PathPart{})
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
			"$ne":        ".token<jwt>.username",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwidXNlcm5hbWUiOnsiJG5lIjpudWxsfSwiaWF0IjoxNTE2MjM5MDIyfQ._jhGJw9WzB6gHKPSozTFHDo9NOHs3CNOlvJ8rWy6VrQ": ".token",
		}

		actual := ExtractStringsFromUserInput(obj, []PathPart{})
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

		actual := ExtractStringsFromUserInput(obj, []PathPart{})
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
		actual := ExtractStringsFromUserInput(obj, []PathPart{})
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
		actual = ExtractStringsFromUserInput(obj, []PathPart{})
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
		actual = ExtractStringsFromUserInput(obj, []PathPart{})
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, got %v", expected, actual)
		}
	})

}
