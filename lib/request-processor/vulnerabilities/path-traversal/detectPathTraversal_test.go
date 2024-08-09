package path_traversal

import (
	"testing"
)

func TestDetectPathTraversal(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		userInput  string
		checkStart bool
		expected   bool
	}{
		{"empty user input", "test.txt", "", true, false},
		{"empty file input", "", "test", true, false},
		{"empty user input and file input", "", "", true, false},
		{"user input is a single character", "test.txt", "t", true, false},
		{"file input is a single character", "t", "test", true, false},
		{"same as user input", "text.txt", "text.txt", true, false},
		{"with directory before", "directory/text.txt", "text.txt", true, false},
		{"with both directory before", "directory/text.txt", "directory/text.txt", true, false},
		{"user input and file input are single characters", "t", "t", true, false},
		{"it flags ../", "../test.txt", "../", true, true},
		{"it flags ..\\", "..\\test.txt", "..\\", true, true},
		{"it flags ../../", "../../test.txt", "../../", true, true},
		{"it flags ..\\..\\", "..\\..\\test.txt", "..\\..\\", true, true},
		{"it flags ../../../../", "../../../../test.txt", "../../../../", true, true},
		{"it flags ..\\..\\..\\", "..\\..\\..\\test.txt", "..\\..\\..\\", true, true},
		{"user input is longer than file path", "../file.txt", "../../file.txt", true, false},
		{"absolute linux path", "/etc/passwd", "/etc/passwd", true, true},
		{"linux user directory", "/home/user/file.txt", "/home/user/", true, true},
		{"windows drive letter", "C:\\file.txt", "C:\\", true, true},
		{"no path traversal", "/appdata/storage/file.txt", "/storage/file.txt", true, false},
		{"does not flag test", "/app/test.txt", "test", true, false},
		{"does not flag example/test.txt", "/app/data/example/test.txt", "example/test.txt", true, false},
		{"does not absolute path with different folder", "/etc/app/config", "/etc/hack/config", true, false},
		{"does not absolute path inside another folder", "/etc/app/data/etc/config", "/etc/config", true, false},
		{"disable checkPathStart", "/etc/passwd", "/etc/passwd", false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectPathTraversal(tc.filePath, tc.userInput, tc.checkStart)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
