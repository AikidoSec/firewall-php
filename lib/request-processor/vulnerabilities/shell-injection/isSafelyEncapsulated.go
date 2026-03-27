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
// escape mechanism.
func parseQuoteRegions(command string) []quoteRegion {
	var regions []quoteRegion
	i := 0
	for i < len(command) {
		ch := command[i]
		if ch == '\'' || ch == '"' {
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
					for _, dc := range dangerousCharsInsideDoubleQuotes {
						if strings.Contains(userInput, dc) {
							hasDangerous = true
							break
						}
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
