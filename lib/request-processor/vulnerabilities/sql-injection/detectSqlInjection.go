package sql_injection

import (
	zen_internals "main/vulnerabilities/zen-internals"
	"regexp"
	"strings"
)

var isAlphanumeric = regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString

func shouldReturnEarly(query, userInput string) bool {
	// User input too small or larger than query
	if len(userInput) <= 1 || len(query) < len(userInput) {
		return true
	}

	// Lowercase versions of query and user input
	queryLowercase := strings.ToLower(query)
	userInputLowercase := strings.ToLower(userInput)

	// User input not in query
	if !strings.Contains(queryLowercase, userInputLowercase) {
		return true
	}

	// User input is alphanumerical (with underscores allowed)
	if isAlphanumeric(userInputLowercase) {
		return true
	}

	// Check if user input is a valid comma-separated list of numbers
	cleanedInputForList := strings.ReplaceAll(strings.ReplaceAll(userInputLowercase, " ", ""), ",", "")
	match, _ := regexp.MatchString(`^\d+$`, cleanedInputForList)
	return match
}

func detectSQLInjection(query string, userInput string, dialect int) int {
	if shouldReturnEarly(query, userInput) {
		return 0
	}

	// Executing our final check with zen_internals
	return zen_internals.DetectSQLInjection(query, userInput, dialect)
}
