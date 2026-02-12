package shell_injection

import (
	"main/helpers"
	"main/utils"
	"regexp"
	"sort"
	"strings"
)

var dangerousChars = []string{
	"#",
	"!",
	`"`,
	"$",
	"&",
	"'",
	"(",
	")",
	"*",
	";",
	"<",
	"=",
	">",
	"?",
	"[",
	"\\",
	"]",
	"^",
	"`",
	"{",
	"|",
	"}",
	" ",
	"\n",
	"\t",
	"~",
	"\r",
	"\f",
}
var commands = []string{
	"sleep",
	"shutdown",
	"reboot",
	"poweroff",
	"halt",
	"ifconfig",
	"chmod",
	"chown",
	"ping",
	"ssh",
	"scp",
	"curl",
	"wget",
	"telnet",
	"kill",
	"killall",
	"rm",
	"mv",
	"cp",
	"touch",
	"echo",
	"cat",
	"head",
	"tail",
	"grep",
	"find",
	"awk",
	"sed",
	"sort",
	"uniq",
	"wc",
	"ls",
	"env",
	"ps",
	"who",
	"whoami",
	"id",
	"w",
	"df",
	"du",
	"pwd",
	"uname",
	"hostname",
	"netstat",
	"passwd",
	"arch",
	"printenv",

	// Colon is a null command
	// it might occur in URLs that are passed as arguments to a binary
	// we should flag if it's surrounded by separators
	":",
}

var pathPrefixes = []string{
	"/bin/",
	"/sbin/",
	"/usr/bin/",
	"/usr/sbin/",
	"/usr/local/bin/",
	"/usr/local/sbin/",
}

var separators = []string{" ", "\t", "\n", ";", "&", "|", "(", ")", "<", ">", "\r", "\f"}

// "killall" should be matched before "kill"
func byLength(a, b string) bool {
	return len(b)-len(a) < 0
}

// Build the commands regex
func buildCommandsRegex() *regexp.Regexp {
	// Escape each path prefix and join them
	prefixes := make([]string, len(pathPrefixes))
	for i, prefix := range pathPrefixes {
		prefixes[i] = helpers.EscapeStringRegexp(prefix)
	}

	// Sort commands by length (longer first)
	commandsSorted := make([]string, len(commands))
	copy(commandsSorted, commands)
	sort.Slice(commandsSorted, func(i, j int) bool {
		return byLength(commandsSorted[i], commandsSorted[j])
	})

	// Create the regex pattern
	pattern := `(?i)([/.]*(` + strings.Join(prefixes, "|") + `)?(` + strings.Join(commandsSorted, "|") + `))`
	return regexp.MustCompile(pattern)
}

var commandsRegex = buildCommandsRegex()

type Match struct {
	value string
	start int
}

func matchAll(str string, regex *regexp.Regexp) []Match {
	matches := regex.FindAllStringSubmatch(str, -1)
	matches_index := regex.FindAllStringSubmatchIndex(str, -1)
	result := make([]Match, len(matches))
	for i, match := range matches {
		result[i] = Match{value: match[0], start: matches_index[i][0]}
	}
	return result
}

func containsShellSyntax(command, userInput string) bool {
	// Check if the user input is all whitespace
	if strings.TrimSpace(userInput) == "" {
		return false
	}

	// Check for dangerous characters
	for _, char := range dangerousChars {
		if strings.Contains(userInput, char) {
			return true
		}
	}

	// The command is the same as the user input
	// Rare case, but it's possible
	// e.g. command is `shutdown` and user input is `shutdown`
	// (`shutdown -h now` will be caught by the dangerous chars as it contains a space)
	if command == userInput {

		match := commandsRegex.FindStringIndex(command)

		// Check if the entire command matches a known dangerous command
		if match != nil && match[0] == 0 && len(command) == match[1] {
			return true
		}
		return false
	}

	// Check if the command contains a commonly used command
	for _, match := range matchAll(command, commandsRegex) {
		// We found a command like `rm` or `/sbin/shutdown` in the command
		// Check if the command is the same as the user input
		// If it's not the same, continue searching

		if userInput != match.value {
			continue
		}

		// Otherwise, we'll check if the command is surrounded by separators
		// These separators are used to separate commands and arguments
		// e.g. `rm<space>-rf`
		// e.g. `ls<newline>whoami`
		// e.g. `echo<tab>hello`
		charBefore := ""
		charAfter := ""
		if match.start > 0 {
			charBefore = string(command[match.start-1])
		}
		if match.start+len(match.value) < len(command) {
			charAfter = string(command[match.start+len(match.value)])
		}

		// e.g. <separator>command<separator>
		if utils.ArrayContains(separators, charBefore) && utils.ArrayContains(separators, charAfter) {
			return true
		}

		// e.g. <separator>command
		if utils.ArrayContains(separators, charBefore) && charAfter == "" {
			return true
		}

		// e.g. command<separator>
		if charBefore == "" && utils.ArrayContains(separators, charAfter) {
			return true
		}
	}

	return false
}
