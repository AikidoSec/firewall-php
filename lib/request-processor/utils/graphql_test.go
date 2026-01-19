package utils

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsGraphQLOverHTTP_POST(t *testing.T) {
	// Test POST request with valid GraphQL query
	body := map[string]interface{}{
		"query": "{ user(id: \"123\") { id name } }",
	}
	result := IsGraphQLOverHTTP("POST", "/graphql", "application/json", body, nil)
	assert.True(t, result, "Should detect valid GraphQL POST request")

	// Test with application/graphql content type
	result = IsGraphQLOverHTTP("POST", "/graphql", "application/graphql", body, nil)
	assert.True(t, result, "Should detect GraphQL with application/graphql content type")

	// Test without JSON content type
	result = IsGraphQLOverHTTP("POST", "/graphql", "text/plain", body, nil)
	assert.False(t, result, "Should not detect GraphQL without JSON content type")

	// Test without /graphql route
	result = IsGraphQLOverHTTP("POST", "/api/data", "application/json", body, nil)
	assert.False(t, result, "Should not detect GraphQL without /graphql route")

	// Test without query field
	bodyWithoutQuery := map[string]interface{}{
		"data": "something",
	}
	result = IsGraphQLOverHTTP("POST", "/graphql", "application/json", bodyWithoutQuery, nil)
	assert.False(t, result, "Should not detect GraphQL without query field")

	// Test with non-string query
	bodyWithNonStringQuery := map[string]interface{}{
		"query": 123,
	}
	result = IsGraphQLOverHTTP("POST", "/graphql", "application/json", bodyWithNonStringQuery, nil)
	assert.False(t, result, "Should not detect GraphQL with non-string query")

	// Test with query that doesn't look like GraphQL
	bodyWithBadQuery := map[string]interface{}{
		"query": "SELECT * FROM users",
	}
	result = IsGraphQLOverHTTP("POST", "/graphql", "application/json", bodyWithBadQuery, nil)
	assert.False(t, result, "Should not detect GraphQL with SQL query")
}

func TestIsGraphQLOverHTTP_GET(t *testing.T) {
	// Test GET request with valid GraphQL query
	query := map[string]interface{}{
		"query": "{ user(id: \"123\") { id name } }",
	}
	result := IsGraphQLOverHTTP("GET", "/graphql", "", nil, query)
	assert.True(t, result, "Should detect valid GraphQL GET request")

	// Test without /graphql route
	result = IsGraphQLOverHTTP("GET", "/api/data", "", nil, query)
	assert.False(t, result, "Should not detect GraphQL without /graphql route")

	// Test without query parameter
	emptyQuery := map[string]interface{}{}
	result = IsGraphQLOverHTTP("GET", "/graphql", "", nil, emptyQuery)
	assert.False(t, result, "Should not detect GraphQL without query parameter")
}

func TestLooksLikeGraphQLQuery(t *testing.T) {
	assert.True(t, looksLikeGraphQLQuery("{ user { id } }"))
	assert.True(t, looksLikeGraphQLQuery("query { user { id } }"))
	assert.True(t, looksLikeGraphQLQuery("mutation { createUser { id } }"))
	assert.False(t, looksLikeGraphQLQuery("SELECT * FROM users"))
	assert.False(t, looksLikeGraphQLQuery("plain text"))
	assert.False(t, looksLikeGraphQLQuery(""))
}

func TestExtractInputsFromGraphQL_POST(t *testing.T) {
	// Test extracting inputs from query document
	body := map[string]interface{}{
		"query": `query { user(id: "123", name: "John") { id name } }`,
	}
	result := ExtractInputsFromGraphQL(body, nil, "POST")

	// Should extract string literals from query
	assert.Contains(t, result, "123")
	assert.Contains(t, result, "John")

	// Test with variables
	bodyWithVariables := map[string]interface{}{
		"query": `query GetUser($id: ID!) { user(id: $id) { id name } }`,
		"variables": map[string]interface{}{
			"id":    "456",
			"name":  "Jane",
			"age":   30, // Non-string should be handled
			"email": "jane@example.com",
		},
	}
	result = ExtractInputsFromGraphQL(bodyWithVariables, nil, "POST")

	// Should extract string variables
	assert.Contains(t, result, "456")
	assert.Contains(t, result, "Jane")
	assert.Contains(t, result, "jane@example.com")

	//  "query": "mutation { uploadFile(url: \"http://localhost/secrets\") { success } }",
	bodyWithMutation := map[string]interface{}{
		"query": "mutation { uploadFile(url: \"http://localhost/secrets\") { success } }",
	}
	result = ExtractInputsFromGraphQL(bodyWithMutation, nil, "POST")
	assert.Contains(t, result, "http://localhost/secrets")

	// Test with variables
	bodyWithVariablesMutation := map[string]interface{}{
		"query": "mutation { uploadFile(url: \"http://localhost/secrets\") { success } }",
		"variables": map[string]interface{}{
			"url": "http://localhost/secrets",
		},
	}
	result = ExtractInputsFromGraphQL(bodyWithVariablesMutation, nil, "POST")
	assert.Contains(t, result, "http://localhost/secrets")
}

func TestExtractInputsFromGraphQL_GET(t *testing.T) {
	// Test GET request with query in query parameters
	query := map[string]interface{}{
		"query": `{ user(id: "789") { id name } }`,
	}
	result := ExtractInputsFromGraphQL(nil, query, "GET")

	assert.Contains(t, result, "789")

	// Test with JSON-encoded variables
	queryWithVariables := map[string]interface{}{
		"query":     `query GetUser($id: ID!) { user(id: $id) { id } }`,
		"variables": `{"id": "999", "name": "Test"}`,
	}
	result = ExtractInputsFromGraphQL(nil, queryWithVariables, "GET")

	assert.Contains(t, result, "999")
	assert.Contains(t, result, "Test")
}

func TestExtractStringValuesFromDocument(t *testing.T) {
	// Test simple query
	query := `{ user(id: "123") { id name } }`
	inputs := extractStringValuesFromDocument(query)
	assert.Equal(t, 1, len(inputs))
	assert.Contains(t, inputs, "123")

	// Test query with multiple string values
	query = `{ user(id: "123", email: "test@example.com") { id name address(city: "NYC") } }`
	inputs = extractStringValuesFromDocument(query)
	assert.Equal(t, 3, len(inputs))
	assert.Contains(t, inputs, "123")
	assert.Contains(t, inputs, "test@example.com")
	assert.Contains(t, inputs, "NYC")

	// Test mutation
	mutation := `mutation { createUser(name: "John", email: "john@example.com") { id } }`
	inputs = extractStringValuesFromDocument(mutation)
	assert.Equal(t, 2, len(inputs))
	assert.Contains(t, inputs, "John")
	assert.Contains(t, inputs, "john@example.com")

	// Test with invalid query (should not crash)
	invalidQuery := `this is not valid GraphQL`
	inputs = extractStringValuesFromDocument(invalidQuery)
	assert.Equal(t, 0, len(inputs))
}

func TestExtractTopLevelFields(t *testing.T) {
	// Test query operation
	query := `query { user { id } posts { title } }`
	opType, fields := ExtractTopLevelFields(query, "")
	assert.Equal(t, "query", opType)
	assert.Equal(t, 2, len(fields))
	assert.Contains(t, fields, "user")
	assert.Contains(t, fields, "posts")

	// Test mutation operation
	mutation := `mutation { createUser { id } deletePost { success } }`
	opType, fields = ExtractTopLevelFields(mutation, "")
	assert.Equal(t, "mutation", opType)
	assert.Equal(t, 2, len(fields))
	assert.Contains(t, fields, "createUser")
	assert.Contains(t, fields, "deletePost")

	// Test with operation name
	queryWithName := `query GetUser { user { id } }`
	opType, fields = ExtractTopLevelFields(queryWithName, "GetUser")
	assert.Equal(t, "query", opType)
	assert.Equal(t, 1, len(fields))
	assert.Contains(t, fields, "user")

	// Test with invalid query
	invalidQuery := `this is not valid`
	opType, fields = ExtractTopLevelFields(invalidQuery, "")
	assert.Equal(t, "", opType)
	assert.Nil(t, fields)

	// Test with empty query
	opType, fields = ExtractTopLevelFields("", "")
	assert.Equal(t, "", opType)
	assert.Nil(t, fields)
}

func TestIsGraphQLRoute(t *testing.T) {
	// Standard patterns
	assert.True(t, isGraphQLRoute("/graphql"))
	assert.True(t, isGraphQLRoute("/api/graphql"))
	assert.True(t, isGraphQLRoute("/v1/graphql"))

	// GraphQL in the middle of path
	assert.True(t, isGraphQLRoute("/graphql/api"))
	assert.True(t, isGraphQLRoute("/index.php?p=admin/actions/graphql/api"))

	// Case insensitive
	assert.True(t, isGraphQLRoute("/GraphQL"))
	assert.True(t, isGraphQLRoute("/api/GRAPHQL"))

	// Should NOT match
	assert.False(t, isGraphQLRoute("/api/users"))
	assert.False(t, isGraphQLRoute(""))
}

func TestIsJSONContentType(t *testing.T) {
	assert.True(t, isJSONContentType("application/json"))
	assert.True(t, isJSONContentType("application/json; charset=utf-8"))
	assert.True(t, isJSONContentType("Application/JSON"))
	assert.True(t, isJSONContentType("application/graphql"))
	assert.False(t, isJSONContentType("text/plain"))
	assert.False(t, isJSONContentType("application/xml"))
	assert.False(t, isJSONContentType(""))
}

func TestExtractStringValuesFromDocument_DeeplyNested(t *testing.T) {
	// Test with a deeply nested query to ensure recursion limit works
	// Without a limit, extremely nested queries could cause stack overflow
	query := `{ user(id: "level0") {`

	// Add many nested levels (well beyond what's reasonable)
	for i := 1; i <= 150; i++ {
		query += ` friends { user(id: "level` + strconv.Itoa(i) + `") {`
	}

	// Close with a field selection
	query += ` id`

	// Close all braces
	for i := 0; i <= 150; i++ {
		query += ` }}`
	}

	// Should not crash - the recursion limit protects against stack overflow
	inputs := extractStringValuesFromDocument(query)

	// Should extract values from at least the first levels before hitting the limit
	assert.NotEmpty(t, inputs, "Should extract values from nested query")
	assert.Contains(t, inputs, "level0", "Should extract value from first level")
	assert.Contains(t, inputs, "level10", "Should extract value from early levels")

	// Should NOT extract all 150 levels - the recursion limit should stop it
	assert.Less(t, len(inputs), 150, "Recursion limit should prevent extracting all levels")

	// But should extract a reasonable number before hitting the limit
	assert.Greater(t, len(inputs), 10, "Should extract values before hitting limit")
}
