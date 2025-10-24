package webscanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWebScanner(t *testing.T) {
	t.Run("Test is a web scanner", func(t *testing.T) {
		assert.Equal(t, true, IsWebScanner("GET", "/wp-config.php", nil))
		assert.Equal(t, true, IsWebScanner("GET", "/.env", nil))
		assert.Equal(t, true, IsWebScanner("GET", "/test/.env.bak", nil))

		assert.Equal(t, true, IsWebScanner("GET", "/.git/config", nil))
		assert.Equal(t, true, IsWebScanner("GET", "/.aws/config", nil))
		assert.Equal(t, true, IsWebScanner("GET", "/../secret", nil))
		assert.Equal(t, true, IsWebScanner("BADMETHOD", "/", nil))
		assert.Equal(t, true, IsWebScanner("GET", "/", map[string]interface{}{"test": "SELECT * FROM admin"}))
		assert.Equal(t, true, IsWebScanner("GET", "/", map[string]interface{}{"test": "../etc/passwd"}))

	})

	t.Run("Test is not a web scanner", func(t *testing.T) {
		assert.Equal(t, false, IsWebScanner("POST", "graphql", nil))
		assert.Equal(t, false, IsWebScanner("GET", "/api/v1/users", nil))
		assert.Equal(t, false, IsWebScanner("GET", "/public/index.html", nil))
		assert.Equal(t, false, IsWebScanner("GET", "/static/js/app.js", nil))
		assert.Equal(t, false, IsWebScanner("GET", "/uploads/image.png", nil))
		assert.Equal(t, false, IsWebScanner("GET", "/", map[string]interface{}{"test": "1'"}))
		assert.Equal(t, false, IsWebScanner("GET", "/", map[string]interface{}{"test": "abcd"}))
	})
}
