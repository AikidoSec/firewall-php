package api_discovery

import (
	"encoding/json"
	"main/ipc/protos"
	"reflect"
)

const maxDepth = 20
const maxProperties = 100
const maxPropertyKeyLength = 100

// GetDataSchema returns the schema of the given data as a DataSchema
func GetDataSchema(data interface{}, depth int) *protos.DataSchema {
	// If the data is not an object (or an array), return the type
	if data == nil {
		return &protos.DataSchema{Type: []string{"null"}}
	}
	//interface {}(encoding/json.Number)
	if _, ok := data.(json.Number); ok {
		return &protos.DataSchema{Type: []string{"number"}}
	}

	dataType := reflect.TypeOf(data)

	switch dataType.Kind() {
	case reflect.Slice, reflect.Array:
		// If the data is an array/slice, return an array schema
		v := reflect.ValueOf(data)
		if v.Len() > 0 {
			return &protos.DataSchema{
				Type:  []string{"array"},
				Items: GetDataSchema(v.Index(0).Interface(), depth+1),
			}
		} else {
			return &protos.DataSchema{Type: []string{"array"}}
		}
	case reflect.Map:
		// Create an object schema with properties
		schema := protos.DataSchema{
			Type:       []string{"object"},
			Properties: make(map[string]*protos.DataSchema),
		}

		// Traverse properties if within depth
		if depth < maxDepth {
			keys := reflect.ValueOf(data).MapKeys()
			for i, key := range keys {
				if i >= maxProperties {
					break
				}
				if len(key.String()) > maxPropertyKeyLength {
					continue // skip key if it's too long
				}
				value := reflect.ValueOf(data).MapIndex(key).Interface()
				schema.Properties[key.String()] = GetDataSchema(value, depth+1)
			}
		}

		return &schema

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return &protos.DataSchema{Type: []string{"number"}}

	case reflect.Float32, reflect.Float64:
		return &protos.DataSchema{Type: []string{"number"}}

	case reflect.Bool:
		return &protos.DataSchema{Type: []string{"boolean"}}

	default:
		// If the data is not an object or array, return its type
		return &protos.DataSchema{Type: []string{dataType.Kind().String()}}
	}
}
