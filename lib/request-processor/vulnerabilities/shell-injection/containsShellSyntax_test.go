package shell_injection

import (
	"testing"
)

func TestContainsShellSyntax(t *testing.T) {
	// Helper function for checking containsShellSyntax
	containsSyntax := func(str string, expected bool) {
		if result := containsShellSyntax(str, str); result != expected {
			t.Errorf("Expected %v for input '%s', but got %v", expected, str, result)
		}
	}

	t.Run("it detects shell syntax", func(t *testing.T) {
		containsSyntax("", false)
		containsSyntax("hello", false)
		containsSyntax("\n", false)
		containsSyntax("\n\n", false)

		containsSyntax("$(command)", true)
		containsSyntax("$(command arg arg)", true)
		containsSyntax("`command`", true)
		containsSyntax("\narg", true)
		containsSyntax("\targ", true)
		containsSyntax("\narg\n", true)
		containsSyntax("arg\n", true)
		containsSyntax("arg\narg", true)
		containsSyntax("rm -rf", true)
		containsSyntax("/bin/rm -rf", true)
		containsSyntax("/bin/rm", true)
		containsSyntax("/sbin/sleep", true)
		containsSyntax("/usr/bin/kill", true)
		containsSyntax("/usr/bin/killall", true)
		containsSyntax("/usr/bin/env", true)
		containsSyntax("/bin/ps", true)
		containsSyntax("/usr/bin/W", true)
	})

	t.Run("it detects commands surrounded by separators", func(t *testing.T) {
		expected := true
		result := containsShellSyntax(`find /path/to/search -type f -name "pattern" -exec rm {} \\;`, "rm")
		if result != expected {
			t.Errorf("Expected %v for 'rm' in command, but got %v", expected, result)
		}
	})

	t.Run("it detects commands with separator before", func(t *testing.T) {
		expected := true
		result := containsShellSyntax(`find /path/to/search -type f -name "pattern" | xargs rm`, "rm")
		if result != expected {
			t.Errorf("Expected %v for 'rm' in command, but got %v", expected, result)
		}
	})

	t.Run("it detects commands with separator after", func(t *testing.T) {
		expected := true
		result := containsShellSyntax("rm arg", "rm")
		if result != expected {
			t.Errorf("Expected %v for 'rm' in command, but got %v", expected, result)
		}
	})

	t.Run("it checks if the same command occurs in the user input", func(t *testing.T) {
		expected := false
		result := containsShellSyntax("find cp", "rm")
		if result != expected {
			t.Errorf("Expected %v for 'rm' in command, but got %v", expected, result)
		}
	})

	t.Run("it treats colon as a command", func(t *testing.T) {
		expected := true
		result := containsShellSyntax(":|echo", ":|")
		if result != expected {
			t.Errorf("Expected %v for ':|' in command, but got %v", expected, result)
		}

		expected = false
		result = containsShellSyntax("https://www.google.com", "https://www.google.com")
		if result != expected {
			t.Errorf("Expected %v for 'https://www.google.com' in command, but got %v", expected, result)
		}
	})

	t.Run("it detects newline as separator", func(t *testing.T) {
		if !containsShellSyntax("ls\nrm", "rm") {
			t.Errorf("Expected true for newline separator")
		}
		if !containsShellSyntax("echo test\nrm -rf /", "rm") {
			t.Errorf("Expected true for newline separator in command")
		}
		if !containsShellSyntax("rm\nls", "rm") {
			t.Errorf("Expected true for newline separator after command")
		}
	})

	t.Run("it detects tab as separator", func(t *testing.T) {
		if !containsShellSyntax("ls\trm", "rm") {
			t.Errorf("Expected true for tab separator")
		}
		if !containsShellSyntax("echo test\trm -rf /", "rm") {
			t.Errorf("Expected true for tab separator in command")
		}
		if !containsShellSyntax("rm\tls", "rm") {
			t.Errorf("Expected true for tab separator after command")
		}
	})

	t.Run("it detects carriage return as separator", func(t *testing.T) {
		if !containsShellSyntax("ls\rrm", "rm") {
			t.Errorf("Expected true for carriage return separator")
		}
		if !containsShellSyntax("echo test\rrm -rf /", "rm") {
			t.Errorf("Expected true for carriage return separator in command")
		}
		if !containsShellSyntax("rm\rls", "rm") {
			t.Errorf("Expected true for carriage return separator after command")
		}
	})

	t.Run("it detects form feed as separator", func(t *testing.T) {
		if !containsShellSyntax("ls\frm", "rm") {
			t.Errorf("Expected true for form feed separator")
		}
		if !containsShellSyntax("echo test\frm -rf /", "rm") {
			t.Errorf("Expected true for form feed separator in command")
		}
		if !containsShellSyntax("rm\fls", "rm") {
			t.Errorf("Expected true for form feed separator after command")
		}
	})

	t.Run("it flags input as shell injection", func(t *testing.T) {
		expected := true
		result := containsShellSyntax("command -disable-update-check -target https://examplx.com|curl+https://cde-123.abc.domain.com+%23 -json-export /tmp/5891/8526757.json -tags microsoft,windows,exchange,iis,gitlab,oracle,cisco,joomla -stats -stats-interval 3 -retries 3 -no-stdin", "https://examplx.com|curl+https://cde-123.abc.domain.com+%23")
		if result != expected {
			t.Errorf("Expected %v for shell injection detection, but got %v", expected, result)
		}
	})
}
