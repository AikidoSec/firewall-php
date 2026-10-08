package shell_injection

import (
	"strings"
)

var dangerousCharsInsideDoubleQuotes = []string{"$", "`", "\\", "!"}

type quoteRegion struct {
	start     int
	end       int
	quoteChar byte
}

// parseQuoteRegions walks the command and returns all properly closed
// single-quote and double-quote regions. Inside double quotes, backslash
// escapes are respected (per POSIX/bash rules); single quotes have no
// escape mechanism. This function now properly tracks escape state to
// avoid treating escaped quotes as region delimiters.
func parseQuoteRegions(command string) []quoteRegion {
	var regions []quoteRegion
	i := 0
	for i < len(command) {
		ch := command[i]
		// Check if this character is escaped (preceded by an odd number of backslashes)
		numBackslashes := 0
		j := i - 1
		for j >= 0 && command[j] == '\\' {
			numBackslashes++
			j--
		}
		isEscaped := numBackslashes%2 == 1

		if !isEscaped && (ch == '\'' || ch == '"') {
			start := i
			i++
			for i < len(command) && command[i] != ch {
				if ch == '"' && command[i] == '\\' {
					i++
				}
				i++
			}
			if i < len(command) {
				regions = append(regions, quoteRegion{start: start, end: i, quoteChar: ch})
			}
			i++
		} else {
			i++
		}
	}
	return regions
}

// containsCommandSubstitution checks if the input contains command substitution
// patterns like $(command) or `command` that would be executed inside double quotes
func containsCommandSubstitution(input string) bool {
	// Check for $(...)
	if strings.Contains(input, "$(") {
		return true
	}
	// Check for `...`
	if strings.Contains(input, "`") {
		return true
	}
	return false
}

// isInsideCommandSubstitution checks if a position range in the command
// falls within a command substitution context $(...) or `...`
func isInsideCommandSubstitution(command string, start, end int) bool {
	// Check for $(...) command substitution
	depth := 0
	for i := start - 1; i >= 0; i-- {
		if command[i] == ')' {
			depth++
		} else if command[i] == '(' {
			if depth > 0 {
				depth--
			} else if i > 0 && command[i-1] == '$' {
				// Found opening $( before our position
				// Now check if there's a closing ) after our position
				for j := end + 1; j < len(command); j++ {
					if command[j] == ')' {
						return true
					} else if command[j] == '(' && j > 0 && command[j-1] == '$' {
						// Found another $( before closing, continue
						break
					}
				}
			}
		}
	}

	// Check for `...` command substitution (backticks)
	// Count backticks before the start position
	backticksBefore := 0
	for i := start - 1; i >= 0; i-- {
		if command[i] == '`' {
			backticksBefore++
		}
	}
	// If odd number of backticks before, we're inside a backtick substitution
	// Check if there's a closing backtick after
	if backticksBefore%2 == 1 {
		for j := end + 1; j < len(command); j++ {
			if command[j] == '`' {
				return true
			}
		}
	}

	return false
}

func isSafelyEncapsulated(command, userInput string) bool {
	regions := parseQuoteRegions(command)

	idx := 0
	for {
		pos := strings.Index(command[idx:], userInput)
		if pos == -1 {
			break
		}
		absStart := idx + pos
		absEnd := absStart + len(userInput) - 1

		inSafeQuote := false
		for _, region := range regions {
			if absStart > region.start && absEnd < region.end {
				if region.quoteChar == '\'' {
					inSafeQuote = true
					break
				}
				if region.quoteChar == '"' {
					hasDangerous := false
					// Check for dangerous characters
					for _, dc := range dangerousCharsInsideDoubleQuotes {
						if strings.Contains(userInput, dc) {
							hasDangerous = true
							break
						}
					}
					// Also check for command substitution patterns in the input
					if !hasDangerous && containsCommandSubstitution(userInput) {
						hasDangerous = true
					}
					// Check if the input position is inside a command substitution
					if !hasDangerous && isInsideCommandSubstitution(command, absStart, absEnd) {
						hasDangerous = true
					}
					if !hasDangerous {
						inSafeQuote = true
						break
					}
				}
			}
		}

		if !inSafeQuote {
			return false
		}

		idx = absStart + 1
	}

	return true
}
