package aikido_types

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSessionID(t *testing.T) {
	before := time.Now().UnixMilli()
	id := newSessionID()
	after := time.Now().UnixMilli()

	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, id)

	ms, err := strconv.ParseInt(id[0:8]+id[9:13], 16, 64)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, ms, before)
	assert.LessOrEqual(t, ms, after)

	assert.NotEqual(t, id, newSessionID())
}
