package attackwavedetection

import (
	"main/attack-wave-detection/paths"
	"slices"
	"strings"
)

var fileExtensions = []string{"env", "bak", "sql", "sqlite", "sqlite3", "db", "old", "save", "orig", "sqlitedb", "sqlite3db"}

func isWebScanPath(path string) bool {
	normalized := strings.ToLower(path)
	segments := strings.Split(normalized, "/")
	filename := segments[len(segments)-1]
	if filename != "" {
		if slices.Contains(paths.FileNames, filename) {
			return true
		}

		if strings.Contains(filename, ".") {
			// last one
			parts := strings.Split(filename, ".")
			ext := parts[len(parts)-1]
			if ext != "" && slices.Contains(fileExtensions, ext) {
				return true
			}
		}
	}

	// Check all directory names
	for _, dir := range segments {
		if slices.Contains(paths.DirectoryNames, dir) {
			return true
		}
	}

	return false
}
