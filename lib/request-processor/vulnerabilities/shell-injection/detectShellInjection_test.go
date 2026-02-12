package shell_injection

import (
	"testing"
)

func isShellInjection(t *testing.T, command, userInput string) {
	if !detectShellInjection(command, userInput) {
		t.Errorf("Expected shell injection for command: '%s' with user input: '%s'", command, userInput)
	}
}

func isNotShellInjection(t *testing.T, command, userInput string) {
	if detectShellInjection(command, userInput) {
		t.Errorf("Did not expect shell injection for command: '%s' with user input: '%s'", command, userInput)
	}
}

func TestDetectShellInjection(t *testing.T) {
	t.Run("single characters are ignored", func(t *testing.T) {
		isNotShellInjection(t, "ls `", "`")
		isNotShellInjection(t, "ls *", "*")
		isNotShellInjection(t, "ls a", "a")
	})

	t.Run("no shell injection when there is no user input", func(t *testing.T) {
		isNotShellInjection(t, "ls", "")
		isNotShellInjection(t, "ls", " ")
		isNotShellInjection(t, "ls", "  ")
		isNotShellInjection(t, "ls", "   ")
	})

	t.Run("no shell injection if user input does not occur in the command", func(t *testing.T) {
		isNotShellInjection(t, "ls", "$(echo)")
	})

	t.Run("user input longer than the command", func(t *testing.T) {
		isNotShellInjection(t, "`ls`", "`ls` `ls`")
	})

	t.Run("detects $(command)", func(t *testing.T) {
		isShellInjection(t, "ls $(echo)", "$(echo)")
		isShellInjection(t, `ls "$(echo)"`, "$(echo)")
		isShellInjection(t, `echo $(echo "Inner: $(echo \"This is nested\")")`, `$(echo "Inner: $(echo \"This is nested\")")`)
		isNotShellInjection(t, "ls '$(echo)'", "$(echo)")
		isNotShellInjection(t, `ls '$(echo "Inner: $(echo \"This is nested\")")'`, `$(echo "Inner: $(echo \"This is nested\")")`)
	})

	t.Run("detects `command`", func(t *testing.T) {
		isShellInjection(t, "echo `echo`", "`echo`")
	})

	t.Run("checks unsafely quoted", func(t *testing.T) {
		isShellInjection(t, "ls '$(echo)", "$(echo)")
	})

	t.Run("single quote between single quotes", func(t *testing.T) {
		isShellInjection(t, "ls ''single quote''", "'single quote'")
	})

	t.Run("ignores escaped backticks", func(t *testing.T) {
		domain := "www.example`whoami`.com"
		isNotShellInjection(t, "--domain www.example\\`whoami\\`.com", domain)
	})

	t.Run("does not allow special chars inside double quotes", func(t *testing.T) {
		isShellInjection(t, `ls "whatever$"`, "whatever$")
		isShellInjection(t, `ls "whatever!"`, "whatever!")
		isShellInjection(t, "ls \"whatever`\"", "whatever`")
	})

	t.Run("does not allow semi", func(t *testing.T) {
		isShellInjection(t, `ls whatever;`, "whatever;")
		isNotShellInjection(t, `ls "whatever;"`, "whatever;")
		isNotShellInjection(t, `ls 'whatever;'`, "whatever;")
	})

	t.Run("rm rf executed by using semicolon", func(t *testing.T) {
		isShellInjection(t, `ls; rm -rf`, "; rm -rf")
	})

	t.Run("rm rf is flagged as shell injection", func(t *testing.T) {
		isShellInjection(t, `rm -rf`, "rm -rf")
	})

	t.Run("detects shell injection with chained commands using &&", func(t *testing.T) {
		isShellInjection(t, "ls && rm -rf /", "&& rm -rf /")
	})

	t.Run("detects shell injection with OR logic using ||", func(t *testing.T) {
		isShellInjection(t, "ls || echo 'malicious code'", "|| echo 'malicious code'")
	})

	t.Run("detects redirection attempts", func(t *testing.T) {
		isShellInjection(t, "ls > /dev/null", "> /dev/null")
		isShellInjection(t, "cat file.txt > /etc/passwd", "> /etc/passwd")
	})

	t.Run("detects append redirection attempts", func(t *testing.T) {
		isShellInjection(t, "echo 'data' >> /etc/passwd", ">> /etc/passwd")
	})

	t.Run("detects pipe character as potential shell injection", func(t *testing.T) {
		isShellInjection(t, "cat file.txt | grep 'password'", "| grep 'password'")
	})

	t.Run("allows safe use of pipe character within quotes", func(t *testing.T) {
		isNotShellInjection(t, "echo '|'", "|")
	})

	t.Run("detects nested command substitution", func(t *testing.T) {
		isShellInjection(t, `echo $(cat $(ls))`, `$(cat $(ls))`)
	})

	t.Run("allows safe commands within single quotes", func(t *testing.T) {
		isNotShellInjection(t, "echo 'safe command'", "safe command")
	})

	t.Run("detects unsafe use of variables", func(t *testing.T) {
		isShellInjection(t, "echo $USER", "$USER")
		isShellInjection(t, "echo ${USER}", "${USER}")
		isShellInjection(t, `echo "${USER}"`, "${USER}")
	})

	t.Run("allows safe use of variables within quotes", func(t *testing.T) {
		isNotShellInjection(t, "echo '$USER'", "$USER")
	})

	t.Run("detects subshell execution within backticks inside double quotes", func(t *testing.T) {
		isShellInjection(t, "ls \"$(echo `whoami`)\"", "`whoami`")
	})

	t.Run("carriage return in user input is flagged", func(t *testing.T) {
		isShellInjection(t, "ls \rrm", "\rrm")
		isShellInjection(t, "ls \rrm -rf", "\rrm -rf")
	})

	t.Run("form feed in user input is flagged", func(t *testing.T) {
		isShellInjection(t, "ls \frm", "\frm")
		isShellInjection(t, "ls \frm -rf", "\frm -rf")
	})

	t.Run("carriage return in user input is flagged when user input is command", func(t *testing.T) {
		isShellInjection(t, "sleep\r10", "sleep\r10")
		isShellInjection(t, "shutdown\r-h\rnow", "shutdown\r-h\rnow")
	})

	t.Run("form feed in user input is flagged when user input is command", func(t *testing.T) {
		isShellInjection(t, "sleep\f10", "sleep\f10")
		isShellInjection(t, "shutdown\f-h\fnow", "shutdown\f-h\fnow")
	})

	t.Run("carriage return as separator between commands", func(t *testing.T) {
		isShellInjection(t, "ls\rrm", "rm")
		isShellInjection(t, "echo test\rrm -rf /", "rm")
		isShellInjection(t, "rm\rls", "rm")
	})

	t.Run("form feed as separator between commands", func(t *testing.T) {
		isShellInjection(t, "ls\frm", "rm")
		isShellInjection(t, "echo test\frm -rf /", "rm")
		isShellInjection(t, "rm\fls", "rm")
	})
}
