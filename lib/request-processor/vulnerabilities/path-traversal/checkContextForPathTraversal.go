package path_traversal

import (
	"main/context"
	"main/utils"
	"strings"
)

func CheckContextForPathTraversal(filename string, operation string, checkPathStart bool) *utils.InterceptorResult {
	for _, source := range context.SOURCES {
		mapss := source.CacheGet()
		sanitizedPath := SanitizePath(filename)

		for str, path := range mapss {
			inputString := SanitizePath(str)
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

func SanitizePath(path string) string {
	// If path starts with file:// -> remove it (case insensitive)
	if len(path) > 7 && strings.HasPrefix(strings.ToLower(path), "file://") {
		path = path[7:]
	}
	return path
}
