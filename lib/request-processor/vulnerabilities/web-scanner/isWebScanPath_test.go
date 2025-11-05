package webscanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWebScanPath(t *testing.T) {
	t.Run("Test isWebScanPath", func(t *testing.T) {
		var tests = []struct {
			path     string
			expected bool
		}{
			{path: "/.env", expected: true},
			{path: "/test/.env", expected: true},
			{path: "/test/.env.bak", expected: true},
			{path: "/.git/config", expected: true},
			{path: "/.aws/config", expected: true},
			{path: "/some/path/.git/test", expected: true},
			{path: "/some/path/.gitlab-ci.yml", expected: true},
			{path: "/some/path/.github/workflows/test.yml", expected: true},
			{path: "/.travis.yml", expected: true},
			{path: "/../example/", expected: true},
			{path: "/./test", expected: true},
			{path: "/Cargo.lock", expected: true},
			{path: "/System32/test", expected: true},
		}

		for _, test := range tests {
			t.Run(test.path, func(t *testing.T) {
				result := isWebScanPath(test.path)
				assert.Equal(t, test.expected, result)
			})
		}
	})

	t.Run("Test is not a web scan path", func(t *testing.T) {
		var tests = []struct {
			path     string
			expected bool
		}{

			{path: "/test/file.txt", expected: false},
			{path: "/some/route/to/file.txt", expected: false},
			{path: "/some/route/to/file.json", expected: false},
			{path: "/en", expected: false},
			{path: "/", expected: false},
			{path: "/test/route", expected: false},
			{path: "/static/file.css", expected: false},
			{path: "/static/file.a461f56e.js", expected: false},
		}

		for _, test := range tests {
			t.Run(test.path, func(t *testing.T) {
				result := isWebScanPath(test.path)
				assert.Equal(t, test.expected, result)
			})
		}
	})
}
