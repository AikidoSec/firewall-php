package helpers

import (
	"bytes"
	"encoding/json"
)

func ParseJSON(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber() // unmarshal a number into an interface{} as a [Number] instead of as a float64
	return dec.Decode(v)
}
