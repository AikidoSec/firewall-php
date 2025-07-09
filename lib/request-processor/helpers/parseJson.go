package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func ParseJSON(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber() // unmarshal a number into an interface{} as a [Number] instead of as a float64
	// decode first value
	if err := dec.Decode(v); err != nil {
		return err
	}

	// check for trailing garbage
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("unexpected extra JSON values")
		}
		return err
	}
	return nil
}
