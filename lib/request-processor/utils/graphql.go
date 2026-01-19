package utils

import (
	"encoding/json"
	"main/helpers"
	"main/log"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// IsGraphQLOverHTTP checks if the current request is a GraphQL over HTTP request
// Similar to firewall-node/library/sources/graphql/isGraphQLOverHTTP.ts
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
			looksLikeGraphQLQuery(getQueryString(body))
	}

	if method == "GET" {
		queryStr := getQueryStringFromQueryParams(query)
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

// getQueryString extracts the query string from the body
func getQueryString(body map[string]interface{}) string {
	if body == nil {
		return ""
	}
	queryField, exists := body["query"]
	if !exists {
		return ""
	}
	queryStr, ok := queryField.(string)
	if !ok {
		return ""
	}
	return queryStr
}

// getQueryStringFromQueryParams extracts the query string from query parameters
func getQueryStringFromQueryParams(query map[string]interface{}) string {
	if query == nil {
		return ""
	}
	queryField, exists := query["query"]
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
// Similar to firewall-node/library/sources/graphql/extractInputsFromDocument.ts
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
		queryString = getQueryString(body)
		if varsField, exists := body["variables"]; exists {
			if varsMap, ok := varsField.(map[string]interface{}); ok {
				variables = varsMap
			}
		}
	} else if method == "GET" && query != nil {
		queryString = getQueryStringFromQueryParams(query)
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
			result[input] = ".graphql.query"
		}
	}

	// Extract string values from variables
	if variables != nil {
		varInputs := helpers.ExtractStringsFromUserInput(variables, []helpers.PathPart{{Type: "object", Key: "graphql.variables"}}, 0)
		for k, v := range varInputs {
			result[k] = ".graphql.variables" + v
		}
	}

	return result
}

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
	visitNode(doc, &inputs, 0)

	return inputs
}

const maxGraphQLRecursionDepth = 100

// visitNode recursively visits AST nodes and extracts string values
// depth tracking prevents stack overflow from malicious deeply nested queries
func visitNode(node interface{}, inputs *[]string, depth int) {
	if node == nil {
		return
	}

	// Prevent excessive recursion
	if depth > maxGraphQLRecursionDepth {
		log.Warnf("GraphQL query exceeds maximum recursion depth of %d", maxGraphQLRecursionDepth)
		return
	}

	nextDepth := depth + 1

	switch n := node.(type) {
	case *ast.Document:
		for _, def := range n.Definitions {
			visitNode(def, inputs, nextDepth)
		}
	case *ast.OperationDefinition:
		if n.SelectionSet != nil {
			visitNode(n.SelectionSet, inputs, nextDepth)
		}
		for _, varDef := range n.VariableDefinitions {
			visitNode(varDef, inputs, nextDepth)
		}
	case *ast.VariableDefinition:
		if n.DefaultValue != nil {
			visitNode(n.DefaultValue, inputs, nextDepth)
		}
	case *ast.SelectionSet:
		for _, sel := range n.Selections {
			visitNode(sel, inputs, nextDepth)
		}
	case *ast.Field:
		for _, arg := range n.Arguments {
			visitNode(arg, inputs, nextDepth)
		}
		if n.SelectionSet != nil {
			visitNode(n.SelectionSet, inputs, nextDepth)
		}
	case *ast.Argument:
		visitNode(n.Value, inputs, nextDepth)
	case *ast.StringValue:
		*inputs = append(*inputs, n.Value)
	case *ast.IntValue:
		// Skip int values
	case *ast.FloatValue:
		// Skip float values
	case *ast.BooleanValue:
		// Skip boolean values
	case *ast.EnumValue:
		// Skip enum values
	case *ast.ListValue:
		for _, val := range n.Values {
			visitNode(val, inputs, nextDepth)
		}
	case *ast.ObjectValue:
		for _, field := range n.Fields {
			visitNode(field, inputs, nextDepth)
		}
	case *ast.ObjectField:
		visitNode(n.Value, inputs, nextDepth)
	case *ast.Variable:
		// Skip variables (they are handled separately)
	case *ast.FragmentDefinition:
		if n.SelectionSet != nil {
			visitNode(n.SelectionSet, inputs, nextDepth)
		}
	case *ast.InlineFragment:
		if n.SelectionSet != nil {
			visitNode(n.SelectionSet, inputs, nextDepth)
		}
	case *ast.FragmentSpread:
		// Skip fragment spreads
	}
}

// ExtractTopLevelFields extracts the top-level fields from a GraphQL document
// Returns the operation type (query/mutation) and field names
// Similar to firewall-node/library/sources/graphql/extractTopLevelFieldsFromDocument.ts
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
