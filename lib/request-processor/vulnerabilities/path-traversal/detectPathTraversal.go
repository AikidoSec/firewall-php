package path_traversal

import "strings"

func extractResourceOrOriginal(filePath string) string {
	if strings.HasPrefix(filePath, "php://filter/") {
		parts := strings.Split(filePath, "/resource=")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return filePath
}

func detectPathTraversal(filePath string, userInput string, checkPathStart bool) bool {

	if len(userInput) <= 1 {
		// We ignore single characters since they don't pose a big threat.
		return false
	}

	if len(userInput) > len(filePath) {
		// We ignore cases where the user input is longer than the file path.
		// Because the user input can't be part of the file path.
		return false
	}

	if !strings.Contains(filePath, userInput) {
		// We ignore cases where the user input is not part of the file path.
		return false
	}

	filePath = extractResourceOrOriginal(filePath)
	userInput = extractResourceOrOriginal(userInput)

	if containsUnsafePathParts(filePath) && containsUnsafePathParts(userInput) {
		return true
	}

	if checkPathStart {
		return startsWithUnsafePath(filePath, userInput)
	}

	return false
}
