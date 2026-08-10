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
		mapss := getPathTraversalCandidates(instance, source.Name, source.CacheGet(instance))

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

func getPathTraversalCandidates(instance *instance.RequestProcessorInstance, sourceName string, full map[string]string) map[string]string {
	candidates := context.GetPathTraversalCandidatesCache(instance)
	if cached, ok := candidates[sourceName]; ok {
		return cached
	}

	filtered := make(map[string]string)
	for str, path := range full {
		if isPathTraversalCandidate(str) {
			filtered[str] = path
		}
	}
	candidates[sourceName] = filtered
	return filtered
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
