package webscanner

import (
	"main/vulnerabilities/web-scanner/paths"
	"slices"
	"strings"
)

var fileExtensions = []string{"env", "bak", "sql", "sqlite", "sqlite3", "db", "old", "save", "orig", "sqlitedb", "sqlite3db"}

func isWebScanPath(path string) bool {
	normalized := strings.ToLower(path)
	segments := strings.Split(normalized, "/")
	filename := segments[len(segments)-1]
	if filename != "" {
		if _, ok := paths.FileNames[filename]; ok {
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
		if _, ok := paths.DirectoryNames[dir]; ok {
			return true
		}
	}

	return false
}
