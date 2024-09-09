package path_traversal

import (
	"main/utils"
)

func CheckContextForPathTraversal(filename string, operation string, checkPathStart bool) *utils.InterceptorResult {
	for _, source := range utils.SOURCES {
		mapss := source.CacheGet()

		for str, path := range mapss {
			if detectPathTraversal(filename, str, checkPathStart) {
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
