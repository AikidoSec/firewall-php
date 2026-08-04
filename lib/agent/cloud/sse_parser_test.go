package cloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func feedLines(p *sseParser, lines ...string) (sseEvent, bool) {
	var event sseEvent
	var ok bool
	for _, line := range lines {
		event, ok = p.feedLine(line)
	}
	return event, ok
}

func TestSSEParser(t *testing.T) {
	t.Run("it does not dispatch on non-blank lines", func(t *testing.T) {
		p := &sseParser{}

		_, ok := p.feedLine("event: config-updated")
		assert.False(t, ok)

		_, ok = p.feedLine(`data: {"configUpdatedAt":100}`)
		assert.False(t, ok)
	})

	t.Run("it dispatches an event with name and data on a blank line", func(t *testing.T) {
		p := &sseParser{}
		event, ok := feedLines(p,
			"event: config-updated",
			`data: {"configUpdatedAt":100}`,
			"",
		)

		require.True(t, ok)
		assert.Equal(t, "config-updated", event.name)
		assert.Equal(t, `{"configUpdatedAt":100}`, event.data)
	})

	t.Run("it joins multiple data lines with newlines", func(t *testing.T) {
		p := &sseParser{}
		event, ok := feedLines(p,
			"event: config-updated",
			"data: line one",
			"data: line two",
			"",
		)

		require.True(t, ok)
		assert.Equal(t, "line one\nline two", event.data)
	})

	t.Run("it accepts fields without a space after the colon", func(t *testing.T) {
		p := &sseParser{}
		event, ok := feedLines(p, "event:config-updated", "data:{}", "")

		require.True(t, ok)
		assert.Equal(t, "config-updated", event.name)
		assert.Equal(t, "{}", event.data)
	})

	t.Run("it ignores comments used as keep-alives", func(t *testing.T) {
		p := &sseParser{}
		event, ok := feedLines(p, ": keep-alive", "data: 1", "")

		require.True(t, ok)
		assert.Equal(t, "", event.name)
		assert.Equal(t, "1", event.data)
	})

	t.Run("it ignores retry and unknown fields", func(t *testing.T) {
		p := &sseParser{}
		event, ok := feedLines(p, "retry: 1000", "id: 42", "data: 1", "")

		require.True(t, ok)
		assert.Equal(t, "1", event.data)
	})

	t.Run("it resets state after dispatching", func(t *testing.T) {
		p := &sseParser{}
		feedLines(p, "event: config-updated", `data: {"configUpdatedAt":100}`, "")

		event, ok := feedLines(p, "")

		require.True(t, ok)
		assert.Equal(t, "", event.name)
		assert.Equal(t, "", event.data)
	})

	t.Run("it parses multiple events in sequence", func(t *testing.T) {
		p := &sseParser{}

		first, ok := feedLines(p, "event: a", "data: 1", "")
		require.True(t, ok)
		assert.Equal(t, "a", first.name)
		assert.Equal(t, "1", first.data)

		second, ok := feedLines(p, "event: b", "data: 2", "")
		require.True(t, ok)
		assert.Equal(t, "b", second.name)
		assert.Equal(t, "2", second.data)
	})
}
