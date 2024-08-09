package path_traversal

import "strings"

var dangerousPathParts = []string{"../", "..\\"}

func containsUnsafePathParts(filePath string) bool {
	for _, dangerousPart := range dangerousPathParts {
		if strings.Contains(filePath, dangerousPart) {
			return true
		}
	}
	return false
}
