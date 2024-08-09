package path_traversal

import (
	"main/helpers"
	"main/utils"
)

var SOURCES = []string{"body", "routeParams", "query", "params", "headers", "cookies"}

type Context map[string]interface{}

func CheckContextForPathTraversal(filename string, operartion string, context Context, checkPathStart bool) *utils.InterceptorResult {
	for _, source := range SOURCES {
		if context[source] == nil {
			continue
		}

		mapss := helpers.ExtractStringsFromUserInput(context[source], []helpers.PathPart{})
		for str, path := range mapss {
			if detectPathTraversal(filename, str, checkPathStart) {
				return &utils.InterceptorResult{
					Operation:     operartion,
					Kind:          utils.Path_traversal,
					Source:        source,
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
