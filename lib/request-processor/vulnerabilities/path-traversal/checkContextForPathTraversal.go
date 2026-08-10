package path_traversal

import (
	"main/context"
	"main/helpers"
	"main/instance"
	"main/utils"
	"path/filepath"
	"strings"
)

func CheckContextForPathTraversal(instance *instance.RequestProcessorInstance, filename string, operation string, checkPathStart bool) *utils.InterceptorResult {
	trimmedFilename := helpers.TrimInvisible(filename)
	sanitizedPath := SanitizePath(trimmedFilename)

	for _, source := range context.SOURCES {
		mapss := context.GetPathTraversalCandidates(instance, source.Name, source.CacheGet(instance), isPathTraversalCandidate)

		for str, path := range mapss {
			trimmedInputString := helpers.TrimInvisible(str)
			inputString := SanitizePath(trimmedInputString)
			if detectPathTraversal(sanitizedPath, inputString, checkPathStart) {
				return &utils.InterceptorResult{
					Operation:     operation,
					Kind:          utils.Path_traversal,
					Source:        source.Name,
					PathToPayload: path,
					Metadata: map[string]string{
						"filename": filename,
					},
					Payload: str,
				}
			}
		}

	}
	return nil
}

func isPathTraversalCandidate(input string) bool {
	sanitizedInput := SanitizePath(helpers.TrimInvisible(input))
	userInput := helpers.ExtractResourceOrOriginal(sanitizedInput)
	return len(sanitizedInput) > 1 && (containsUnsafePathParts(userInput) || filepath.IsAbs(userInput))
}

func SanitizePath(path string) string {
	if len(path) > 7 && strings.ToLower(path[:7]) == "file://" {
		path = path[7:]
	}
	return path
}
