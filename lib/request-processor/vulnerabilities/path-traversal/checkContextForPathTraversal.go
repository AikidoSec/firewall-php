package path_traversal

import (
	. "main/aikido_types"
	"main/context"
	"main/utils"
)

func CheckContextForPathTraversal(fileAccessed *FileAccessed) *utils.InterceptorResult {
	for _, source := range context.SOURCES {
		mapss := source.CacheGet()

		for str, path := range mapss {
			inputString := utils.SanitizePath(str)
			if detectPathTraversal(fileAccessed.Filename, inputString, true) {
				return &utils.InterceptorResult{
					Operation:     fileAccessed.Operation,
					Kind:          utils.Path_traversal,
					Source:        source.Name,
					PathToPayload: path,
					Metadata: map[string]string{
						"filename": fileAccessed.Filename,
					},
					Payload: str,
				}
			}
		}
	}
	return nil
}
