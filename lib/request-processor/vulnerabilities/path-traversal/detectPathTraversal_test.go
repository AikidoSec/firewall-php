package path_traversal

import (
	"testing"
)

func TestDetectPathTraversal(t *testing.T) {
	t.Run("empty user input", func(t *testing.T) {
		if detectPathTraversal("test.txt", "", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("empty file input", func(t *testing.T) {
		if detectPathTraversal("", "test", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("empty user input and file input", func(t *testing.T) {
		if detectPathTraversal("", "", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("user input is a single character", func(t *testing.T) {
		if detectPathTraversal("test.txt", "t", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("file input is a single character", func(t *testing.T) {
		if detectPathTraversal("t", "test", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("same as user input", func(t *testing.T) {
		if detectPathTraversal("text.txt", "text.txt", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("with directory before", func(t *testing.T) {
		if detectPathTraversal("directory/text.txt", "text.txt", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("with both directory before", func(t *testing.T) {
		if detectPathTraversal("directory/text.txt", "directory/text.txt", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("user input and file input are single characters", func(t *testing.T) {
		if detectPathTraversal("t", "t", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("it flags ../", func(t *testing.T) {
		if detectPathTraversal("../test.txt", "../", true) != true {
			t.Error("expected true")
		}
	})

	t.Run("it flags ..\\", func(t *testing.T) {
		if detectPathTraversal("..\\test.txt", "..\\", true) != true {
			t.Error("expected true")
		}
	})

	t.Run("it flags ../../", func(t *testing.T) {
		if detectPathTraversal("../../test.txt", "../../", true) != true {
			t.Error("expected true")
		}
	})

	t.Run("it flags ..\\..\\", func(t *testing.T) {
		if detectPathTraversal("..\\..\\test.txt", "..\\..\\", true) != true {
			t.Error("expected true")
		}
	})

	t.Run("it flags ../../../../", func(t *testing.T) {
		if detectPathTraversal("../../../../test.txt", "../../../../", true) != true {
			t.Error("expected true")
		}
	})

	t.Run("it flags ..\\..\\..\\", func(t *testing.T) {
		if detectPathTraversal("..\\..\\..\\test.txt", "..\\..\\..\\", true) != true {
			t.Error("expected true")
		}
	})

	t.Run("it flags ./../", func(t *testing.T) {
		if detectPathTraversal("./../test.txt", "./../", true) != true {
			t.Error("expected true")
		}
	})

	t.Run("user input is longer than file path", func(t *testing.T) {
		if detectPathTraversal("../file.txt", "../../file.txt", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("absolute linux path", func(t *testing.T) {
		if detectPathTraversal("/etc/passwd", "/etc/passwd", true) != true {
			t.Error("expected true")
		}
	})

	t.Run("linux user directory", func(t *testing.T) {
		if detectPathTraversal("/home/user/file.txt", "/home/user/", true) != true {
			t.Error("expected true")
		}
	})

	t.Run("possible bypass", func(t *testing.T) {
		if detectPathTraversal("/./etc/passwd", "/./etc/passwd", true) != true {
			t.Error("expected true")
		}
	})

	t.Run("another bypass", func(t *testing.T) {
		if detectPathTraversal("/./././root/test.txt", "/./././root/test.txt", true) != true {
			t.Error("expected true")
		}
		if detectPathTraversal("/./././root/test.txt", "/./././root", true) != true {
			t.Error("expected true")
		}
	})

	t.Run("no path traversal", func(t *testing.T) {
		if detectPathTraversal("/appdata/storage/file.txt", "/storage/file.txt", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("does not flag test", func(t *testing.T) {
		if detectPathTraversal("/app/test.txt", "test", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("does not flag example/test.txt", func(t *testing.T) {
		if detectPathTraversal("/app/data/example/test.txt", "example/test.txt", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("does not absolute path with different folder", func(t *testing.T) {
		if detectPathTraversal("/etc/app/config", "/etc/hack/config", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("does not absolute path inside another folder", func(t *testing.T) {
		if detectPathTraversal("/etc/app/data/etc/config", "/etc/config", true) != false {
			t.Error("expected false")
		}
	})

	t.Run("disable checkPathStart", func(t *testing.T) {
		if detectPathTraversal("/etc/passwd", "/etc/passwd", false) != false {
			t.Error("expected false")
		}
	})

	// To enable when doing a Windows build (filepath.IsAbs does not work on linux when a windows path is checked)
	//t.Run("windows drive letter", func(t *testing.T) {
	//	if detectPathTraversal("C:\\file.txt", "C:\\", true) != true {
	//		t.Error("expected true")
	//	}
	//})

	t.Run("does not detect if user input path contains no filename or subfolder", func(t *testing.T) {
		testCases := []struct {
			inputPath string
			userPath  string
			expected  bool
		}{
			{"/etc/app/test.txt", "/etc/", false},
			{"/etc/app/", "/etc/", false},
			{"/etc/app/", "/etc", false},
			{"/etc/", "/etc/", false},
			{"/etc", "/etc", false},
			{"/var/a", "/var/", false},
			{"/var/a", "/var/b", false},
			{"/var/a", "/var/b/test.txt", false},
		}
		for _, tc := range testCases {
			if detectPathTraversal(tc.inputPath, tc.userPath, true) != tc.expected {
				t.Errorf("expected %v for input %q and user path %q", tc.expected, tc.inputPath, tc.userPath)
			}
		}
	})

	t.Run("it does detect if user input path contains a filename or subfolder", func(t *testing.T) {
		testCases := []struct {
			inputPath string
			userPath  string
			expected  bool
		}{
			{"/etc/app/file.txt", "/etc/app", true},
			{"/etc/app/file.txt", "/etc/app/file.txt", true},
			{"/var/backups/file.txt", "/var/backups", true},
			{"/var/backups/file.txt", "/var/backups/file.txt", true},
			{"/var/a", "/var/a", true},
			{"/var/a/b", "/var/a", true},
			{"/var/a/b/test.txt", "/var/a", true},
		}
		for _, tc := range testCases {
			if detectPathTraversal(tc.inputPath, tc.userPath, true) != tc.expected {
				t.Errorf("expected %v for input %q and user path %q", tc.expected, tc.inputPath, tc.userPath)
			}
		}
	})
}
