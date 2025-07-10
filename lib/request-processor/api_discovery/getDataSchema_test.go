package api_discovery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"main/ipc/protos"
	"reflect"
	"strings"
	"testing"
)

// Helper function for comparing two DataSchema structs
func compareSchemas(t *testing.T, got, expected *protos.DataSchema) {
	gotJson, _ := json.Marshal(got)
	expectedJson, _ := json.Marshal(expected)

	if string(gotJson) != string(expectedJson) {
		t.Errorf("Got %s, expected %s", string(gotJson), string(expectedJson))
	}
}

func TestGetDataSchema(t *testing.T) {

	t.Run("it works", func(t *testing.T) {
		compareSchemas(t, GetDataSchema("test", 0), &protos.DataSchema{
			Type: []string{"string"},
		})

		compareSchemas(t, GetDataSchema([]string{"test"}, 0), &protos.DataSchema{
			Type: []string{"array"},
			Items: &protos.DataSchema{
				Type: []string{"string"},
			},
		})

		compareSchemas(t, GetDataSchema(map[string]interface{}{"test": "abc"}, 0), &protos.DataSchema{
			Type: []string{"object"},
			Properties: map[string]*protos.DataSchema{
				"test": {Type: []string{"string"}},
			},
		})

		compareSchemas(t, GetDataSchema(map[string]interface{}{"test": 123, "arr": []int{1, 2, 3}}, 0), &protos.DataSchema{
			Type: []string{"object"},
			Properties: map[string]*protos.DataSchema{
				"test": {Type: []string{"number"}},
				"arr": {
					Type: []string{"array"},
					Items: &protos.DataSchema{
						Type: []string{"number"},
					},
				},
			},
		})

		compareSchemas(t, GetDataSchema(map[string]interface{}{
			"test": 123,
			"arr":  []interface{}{map[string]interface{}{"sub": true}},
			"x":    nil,
		}, 0), &protos.DataSchema{
			Type: []string{"object"},
			Properties: map[string]*protos.DataSchema{
				"test": {Type: []string{"number"}},
				"arr": {
					Type: []string{"array"},
					Items: &protos.DataSchema{
						Type: []string{"object"},
						Properties: map[string]*protos.DataSchema{
							"sub": {Type: []string{"boolean"}},
						},
					},
				},
				"x": {Type: []string{"null"}},
			},
		})

		compareSchemas(t, GetDataSchema(map[string]interface{}{
			"test": map[string]interface{}{
				"x": map[string]interface{}{
					"y": map[string]interface{}{
						"z": 123,
					},
				},
			},
			"arr": []interface{}{},
		}, 0), &protos.DataSchema{
			Type: []string{"object"},
			Properties: map[string]*protos.DataSchema{
				"test": {
					Type: []string{"object"},
					Properties: map[string]*protos.DataSchema{
						"x": {
							Type: []string{"object"},
							Properties: map[string]*protos.DataSchema{
								"y": {
									Type: []string{"object"},
									Properties: map[string]*protos.DataSchema{
										"z": {Type: []string{"number"}},
									},
								},
							},
						},
					},
				},
				"arr": {
					Type:  []string{"array"},
					Items: nil,
				},
			},
		})
	})

	t.Run("test max depth", func(t *testing.T) {
		var generateTestObjectWithDepth func(depth int) interface{}

		generateTestObjectWithDepth = func(depth int) interface{} {
			if depth == 0 {
				return "testValue"
			}
			return map[string]interface{}{
				"prop": generateTestObjectWithDepth(depth - 1),
			}
		}

		obj := generateTestObjectWithDepth(10)
		schema := GetDataSchema(obj, 0)
		schemaJson, _ := json.Marshal(schema)
		if !json.Valid([]byte(schemaJson)) || !strings.Contains(string(schemaJson), `"type":["string"]`) {
			t.Errorf("Expected schema to contain 'type: string'! Got %s", string(schemaJson))
		}

		obj2 := generateTestObjectWithDepth(21)
		schema2 := GetDataSchema(obj2, 0)
		schema2Json, _ := json.Marshal(schema2)
		schema2JsonStr := string(schema2Json)
		if strings.Contains(schema2JsonStr, `"type":["string"]`) {
			t.Errorf("Expected schema to not contain 'type: string' for depth beyond limit! Got %s", schema2JsonStr)
		}
	})

	t.Run("test max properties", func(t *testing.T) {
		generateObjectWithProperties := func(count int) map[string]interface{} {
			obj := make(map[string]interface{})
			for i := 0; i < count; i++ {
				obj[fmt.Sprintf("prop%d", i)] = i
			}
			return obj
		}

		obj := generateObjectWithProperties(80)
		schema := GetDataSchema(obj, 0)
		if len(schema.Properties) != 80 {
			t.Errorf("Expected 80 properties, got %d", len(schema.Properties))
		}

		obj2 := generateObjectWithProperties(120)
		schema2 := GetDataSchema(obj2, 0)
		if len(schema2.Properties) != 100 {
			t.Errorf("Expected 100 properties, got %d", len(schema2.Properties))
		}
	})

	t.Run("test json.Number is treated as number", func(t *testing.T) {
		var data interface{}
		rowData := []byte(`{"age": 30, "name": "John", "isAdmin": true, "array": [1, 2, 3], "object": {"name": "John", "age": 30}}`)
		dec := json.NewDecoder(bytes.NewReader(rowData))
		dec.UseNumber()
		dec.Decode(&data)
		schema := GetDataSchema(data, 0)
		expected := &protos.DataSchema{
			Type: []string{"object"},
			Properties: map[string]*protos.DataSchema{
				"age":     {Type: []string{"number"}},
				"name":    {Type: []string{"string"}},
				"isAdmin": {Type: []string{"boolean"}},
				"array":   {Type: []string{"array"}, Items: &protos.DataSchema{Type: []string{"number"}}},
				"object": {Type: []string{"object"}, Properties: map[string]*protos.DataSchema{
					"name": {Type: []string{"string"}},
					"age":  {Type: []string{"number"}},
				}},
			},
		}

		if !reflect.DeepEqual(schema, expected) {
			t.Errorf("expected %v, got %v", expected, schema)
		}
	})

	t.Run("test max property key length", func(t *testing.T) {
		key := strings.Repeat("a", 101)
		shorterKey := strings.Repeat("b", 99)
		value := "test"
		obj := map[string]interface{}{
			key:        value,
			"test":     []int{1, 2, 3},
			shorterKey: value,
		}
		schema := GetDataSchema(obj, 0)
		compareSchemas(t, schema, &protos.DataSchema{
			Type: []string{"object"},
			Properties: map[string]*protos.DataSchema{
				shorterKey: {Type: []string{"string"}},
				"test": {
					Type: []string{"array"},
					Items: &protos.DataSchema{
						Type: []string{"number"},
					},
				},
			},
		})
	})
}
