package utils

import (
	"encoding/json"
	"main/helpers"
	"main/log"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/graphql-go/graphql/language/visitor"
)

// IsGraphQLOverHTTP checks if the current request is a GraphQL over HTTP request
func IsGraphQLOverHTTP(
	method string,
	url string,
	contentType string,
	body map[string]interface{},
	query map[string]interface{},
) bool {
	if method == "POST" {
		return isGraphQLRoute(url) &&
			isJSONContentType(contentType) &&
			hasGraphQLQuery(body) &&
			looksLikeGraphQLQuery(extractQueryString(body))
	}

	if method == "GET" {
		queryStr := extractQueryString(query)
		return isGraphQLRoute(url) &&
			queryStr != "" &&
			looksLikeGraphQLQuery(queryStr)
	}

	return false
}

// isGraphQLRoute checks if the URL path contains graphql
// Matches common patterns like /graphql, /api/graphql, /graphql/api, etc.
func isGraphQLRoute(url string) bool {
	if url == "" {
		return false
	}
	urlLower := strings.ToLower(url)
	return strings.Contains(urlLower, "/graphql") || strings.Contains(urlLower, "graphql/")
}

// isJSONContentType checks if the content type is JSON
func isJSONContentType(contentType string) bool {
	if contentType == "" {
		return false
	}
	contentTypeLower := strings.ToLower(contentType)
	return strings.Contains(contentTypeLower, "application/json") ||
		strings.Contains(contentTypeLower, "application/graphql")
}

// hasGraphQLQuery checks if body has a query field that is a string
func hasGraphQLQuery(body map[string]interface{}) bool {
	if body == nil {
		return false
	}
	queryField, exists := body["query"]
	if !exists {
		return false
	}
	_, ok := queryField.(string)
	return ok
}

// extractQueryString extracts the query string from a map (body or query params)
func extractQueryString(data map[string]interface{}) string {
	if data == nil {
		return ""
	}
	queryField, exists := data["query"]
	if !exists {
		return ""
	}
	queryStr, ok := queryField.(string)
	if !ok {
		return ""
	}
	return queryStr
}

// looksLikeGraphQLQuery checks if the query string looks like a GraphQL query
// Every GraphQL query should have at least curly braces
func looksLikeGraphQLQuery(query string) bool {
	return strings.Contains(query, "{") && strings.Contains(query, "}")
}

// ExtractInputsFromGraphQL extracts user inputs from a GraphQL request
// This includes:
// - String values from the GraphQL document AST
// - Variable values (strings only)
func ExtractInputsFromGraphQL(
	body map[string]interface{},
	query map[string]interface{},
	method string,
) map[string]string {
	result := make(map[string]string)

	var queryString string
	var variables map[string]interface{}

	// Extract query and variables based on method
	if method == "POST" && body != nil {
		queryString = extractQueryString(body)
		// We don't extract variables from body, because they are already in sources (body.variables)
	} else if method == "GET" && query != nil {
		queryString = extractQueryString(query)
		if varsField, exists := query["variables"]; exists {
			// Variables in GET requests might be JSON-encoded strings
			if varsStr, ok := varsField.(string); ok {
				var varsMap map[string]interface{}
				if err := json.Unmarshal([]byte(varsStr), &varsMap); err == nil {
					variables = varsMap
				}
			} else if varsMap, ok := varsField.(map[string]interface{}); ok {
				variables = varsMap
			}
		}
	}

	// Parse GraphQL document and extract string values
	if queryString != "" {
		inputs := extractStringValuesFromDocument(queryString)
		for _, input := range inputs {
			result[input] = ".query"
		}
	}

	// Extract string values from variables
	if variables != nil {
		varInputs := helpers.ExtractStringsFromUserInput(variables, []helpers.PathPart{{Type: "object", Key: "graphql.variables"}}, 0)
		for k, v := range varInputs {
			result[k] = ".variables" + v
		}
	}

	return result
}

const maxGraphQLRecursionDepth = 200

// extractStringValuesFromDocument parses a GraphQL document and extracts all string values
// This is similar to the Node.js implementation using visit()
func extractStringValuesFromDocument(queryString string) []string {
	var inputs []string

	// Parse the GraphQL document
	src := source.NewSource(&source.Source{
		Body: []byte(queryString),
		Name: "GraphQL request",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		log.Warnf("Failed to parse GraphQL document: %v", err)
		return inputs
	}

	// Recursively visit all nodes in the AST and extract string values
	// Start with depth 0
	// Walk AST and collect string values
	// Walk AST with depth tracking
	depth := 0
	visitor.Visit(doc, &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			depth++
			if depth > maxGraphQLRecursionDepth {
				return visitor.ActionSkip, nil
			}

			if node, ok := p.Node.(*ast.StringValue); ok {
				inputs = append(inputs, node.Value)
			}

			return visitor.ActionNoChange, nil
		},
		Leave: func(p visitor.VisitFuncParams) (string, interface{}) {
			depth--
			return visitor.ActionNoChange, nil
		},
	}, nil)

	return inputs
}

// ExtractTopLevelFields extracts the top-level fields from a GraphQL document
// Returns the operation type (query/mutation) and field names
func ExtractTopLevelFields(queryString string, operationName string) (operationType string, fields []string) {
	if queryString == "" {
		return "", nil
	}

	// Parse the GraphQL document
	src := source.NewSource(&source.Source{
		Body: []byte(queryString),
		Name: "GraphQL request",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		log.Warnf("Failed to parse GraphQL document: %v", err)
		return "", nil
	}

	// Find the operation definition
	var operation *ast.OperationDefinition
	for _, def := range doc.Definitions {
		if opDef, ok := def.(*ast.OperationDefinition); ok {
			// If no operation name is specified and there's only one operation, use it
			if operationName == "" && len(doc.Definitions) == 1 {
				operation = opDef
				break
			}
			// If operation name is specified, find the matching operation
			if operationName != "" && opDef.Name != nil && opDef.Name.Value == operationName {
				operation = opDef
				break
			}
			// If no operation name and multiple operations, use the first one (not ideal but matches Node.js behavior)
			if operation == nil {
				operation = opDef
			}
		}
	}

	if operation == nil {
		return "", nil
	}

	// Extract operation type
	operationType = operation.Operation

	// Extract top-level field names
	if operation.SelectionSet != nil {
		for _, selection := range operation.SelectionSet.Selections {
			if field, ok := selection.(*ast.Field); ok {
				if field.Name != nil {
					fields = append(fields, field.Name.Value)
				}
			}
		}
	}

	return operationType, fields
}
