package context

import "testing"

func TestBodyInputsIncludeQueryParameterFromNestedURL(t *testing.T) {
	// The HTTP body contains encodedPath as one JSON string. WordPress decodes
	// author_exclude into decodedParameterValue before interpolating it into SQL.
	const decodedParameterValue = "SELECT IF((1=1),SLEEP(0.15),0)"
	const encodedPath = "/wp/v2/categories?author_exclude=SELECT%20IF((1%3D1)%2CSLEEP(0.15)%2C0)"

	body := GetJsonString(map[string]interface{}{
		"requests": []interface{}{
			map[string]interface{}{"method": "POST", "path": "http://:"},
			map[string]interface{}{
				"method": "POST",
				"path":   "/wp/v2/posts",
				"body": map[string]interface{}{
					"requests": []interface{}{
						map[string]interface{}{"method": "GET", "path": "http://:"},
						map[string]interface{}{"method": "GET", "path": encodedPath},
						map[string]interface{}{"method": "GET", "path": "/wp/v2/posts"},
					},
				},
			},
			map[string]interface{}{"method": "POST", "path": "/batch/v1"},
		},
	})

	instance := LoadForUnitTests(map[string]string{"body": body})
	t.Cleanup(UnloadForUnitTests)

	inputs := GetBodyParsedFlattened(instance)
	if _, found := inputs[encodedPath]; !found {
		t.Fatalf("expected raw nested URL %q in body inputs, got %#v", encodedPath, inputs)
	}

	_, found := inputs[decodedParameterValue]
	if !found {
		t.Fatalf("expected decoded nested URL query parameter %q in body inputs, got %#v", decodedParameterValue, inputs)
	}
}
