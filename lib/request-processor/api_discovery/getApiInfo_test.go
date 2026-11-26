package api_discovery

import (
	"main/aikido_types"
	"main/context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsObject(t *testing.T) {
	t.Run("Test isObject", func(t *testing.T) {
		assert.Equal(t, false, isObject(""))
		assert.Equal(t, false, isObject([]string{"1"}))
		assert.Equal(t, false, isObject(nil))

		assert.Equal(t, true, isObject(map[string]any{"1": "2"}))
		assert.Equal(t, true, isObject(map[string]string{"1": "2"}))
		assert.Equal(t, true, isObject(map[string]int{"1": 500}))
		assert.Equal(t, true, isObject(map[string][]string{"1": {"2"}}))
		assert.Equal(t, true, isObject(map[string][]any{"1": {"2"}}))
	})
}

func TestRequestSizeForSchema(t *testing.T) {
	atLimit := strconv.Itoa(maxBodyBytesForSchema)
	overLimit := strconv.Itoa(maxBodyBytesForSchema + 1)
	for _, tc := range []struct {
		name, contentType, length string
		wantSpec, wantBody        bool
	}{
		{"small body", "application/json", "14", true, true},
		{"form at limit", "application/x-www-form-urlencoded", atLimit, true, true},
		{"file over limit", "multipart/form-data", overLimit, false, false},
		{"oversized length", "application/json", "18446744073709551616", false, false},
		{"chunked body", "application/json", "", true, false},
		{"form without length", "multipart/form-data", "", true, false},
		{"invalid length", "application/json", "unknown", true, false},
		{"zero length", "", "0", true, false},
		{"query only", "", "", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			headers := map[string]interface{}{"authorization": "Bearer test"}
			if tc.length != "" {
				headers["content_length"] = tc.length
			}
			if tc.contentType != "" {
				headers["content_type"] = tc.contentType
			}
			inst := context.LoadForUnitTests(map[string]string{
				"headers": context.GetJsonString(headers),
				"body":    `{"last":"two"}`,
				"query":   "page=one",
			})
			t.Cleanup(context.UnloadForUnitTests)
			server := &aikido_types.ServerData{}
			server.AikidoConfig.CollectApiSchema = true

			spec := GetApiInfo(inst, server)
			assert.Equal(t, tc.wantBody, context.GetContext(inst).BodyRaw != nil)
			if tc.wantSpec {
				require.NotNil(t, spec)
				assert.Equal(t, tc.wantBody, spec.Body != nil)
				assert.NotNil(t, spec.GetQuery().GetProperties()["page"])
				assert.NotEmpty(t, spec.Auth)
			} else {
				assert.Nil(t, spec)
			}
			assert.Equal(t, ".last", context.GetBodyParsedFlattened(inst)["two"])
		})
	}
}
