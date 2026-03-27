package shell_injection

import (
	"testing"
)

func TestIsSafelyEncapsulated(t *testing.T) {
	t.Run("safe between single quotes", func(t *testing.T) {
		if got := isSafelyEncapsulated("echo '$USER'", "$USER"); got != true {
			t.Errorf("isSafelyEncapsulated('echo '$USER'', '$USER') = %v; want true", got)
		}
		if got := isSafelyEncapsulated("echo '`$USER'", "`USER"); got != true {
			t.Errorf("isSafelyEncapsulated('echo '`$USER'', '`USER') = %v; want true", got)
		}
	})

	t.Run("single quote in single quotes", func(t *testing.T) {
		if got := isSafelyEncapsulated("echo ''USER'", "'USER"); got != false {
			t.Errorf("isSafelyEncapsulated('echo ''USER'', ''USER') = %v; want false", got)
		}
	})

	t.Run("dangerous chars between double quotes", func(t *testing.T) {
		if got := isSafelyEncapsulated(`echo "=USER"`, "=USER"); got != true {
			t.Errorf("isSafelyEncapsulated('echo \"=USER\"', '=USER') = %v; want true", got)
		}
		if got := isSafelyEncapsulated(`echo "$USER"`, "$USER"); got != false {
			t.Errorf("isSafelyEncapsulated('echo \"$USER\"', '$USER') = %v; want false", got)
		}
		if got := isSafelyEncapsulated(`echo "!USER"`, "!USER"); got != false {
			t.Errorf("isSafelyEncapsulated('echo \"!USER\"', '!USER') = %v; want false", got)
		}
		if got := isSafelyEncapsulated("echo \"`USER\"", "`USER"); got != false {
			t.Errorf("isSafelyEncapsulated('echo \"\\`USER\"', '`USER') = %v; want false", got)
		}
		if got := isSafelyEncapsulated(`echo "\\USER"`, "\\USER"); got != false {
			t.Errorf("isSafelyEncapsulated('echo \"\\\\USER\"', '\\USER') = %v; want false", got)
		}
	})

	t.Run("same user input multiple times", func(t *testing.T) {
		if got := isSafelyEncapsulated(`echo '$USER' '$USER'`, "$USER"); got != true {
			t.Errorf("isSafelyEncapsulated('echo '$USER' '$USER'', '$USER') = %v; want true", got)
		}
		if got := isSafelyEncapsulated(`echo "$USER" '$USER'`, "$USER"); got != false {
			t.Errorf("isSafelyEncapsulated('echo \"$USER\" '$USER'', '$USER') = %v; want false", got)
		}
		if got := isSafelyEncapsulated(`echo "$USER" "$USER"`, "$USER"); got != false {
			t.Errorf("isSafelyEncapsulated('echo \"$USER\" \"$USER\"', '$USER') = %v; want false", got)
		}
	})

	t.Run("the first and last quote doesn't match", func(t *testing.T) {
		if got := isSafelyEncapsulated(`echo '$USER"`, "$USER"); got != false {
			t.Errorf("isSafelyEncapsulated('echo '$USER\"', '$USER') = %v; want false", got)
		}
		if got := isSafelyEncapsulated(`echo "$USER'`, "$USER"); got != false {
			t.Errorf("isSafelyEncapsulated('echo \"$USER'\", '$USER') = %v; want false", got)
		}
	})

	t.Run("the first or last character is not an escape char", func(t *testing.T) {
		if got := isSafelyEncapsulated(`echo $USER'`, "$USER"); got != false {
			t.Errorf("isSafelyEncapsulated('echo $USER'', '$USER') = %v; want false", got)
		}
		if got := isSafelyEncapsulated(`echo $USER"`, "$USER"); got != false {
			t.Errorf("isSafelyEncapsulated('echo $USER\"', '$USER') = %v; want false", got)
		}
	})

	t.Run("user input does not occur in the command", func(t *testing.T) {
		if got := isSafelyEncapsulated(`echo 'USER'`, "$USER"); got != true {
			t.Errorf("isSafelyEncapsulated('echo 'USER'', '$USER') = %v; want true", got)
		}
		if got := isSafelyEncapsulated(`echo "USER"`, "$USER"); got != true {
			t.Errorf("isSafelyEncapsulated('echo \"USER\"', '$USER') = %v; want true", got)
		}
	})

	t.Run("user input is substring of single-quoted content", func(t *testing.T) {
		if got := isSafelyEncapsulated("ls '/var/a b/c'", "a b"); got != true {
			t.Errorf("isSafelyEncapsulated(\"ls '/var/a b/c'\", \"a b\") = %v; want true", got)
		}
		if got := isSafelyEncapsulated("grep 'hello world' file.txt", "hello world"); got != true {
			t.Errorf("isSafelyEncapsulated(\"grep 'hello world' file.txt\", \"hello world\") = %v; want true", got)
		}
		if got := isSafelyEncapsulated("echo 'prefix foo suffix'", "foo"); got != true {
			t.Errorf("isSafelyEncapsulated(\"echo 'prefix foo suffix'\", \"foo\") = %v; want true", got)
		}
	})

	t.Run("user input is substring of double-quoted content without dangerous chars", func(t *testing.T) {
		if got := isSafelyEncapsulated(`ls "/var/a b/c"`, "a b"); got != true {
			t.Errorf(`isSafelyEncapsulated("ls \"/var/a b/c\"", "a b") = %v; want true`, got)
		}
		if got := isSafelyEncapsulated(`grep "hello world" file.txt`, "hello world"); got != true {
			t.Errorf(`isSafelyEncapsulated("grep \"hello world\" file.txt", "hello world") = %v; want true`, got)
		}
	})

	t.Run("user input is substring of double-quoted content with dangerous chars", func(t *testing.T) {
		if got := isSafelyEncapsulated(`echo "/tmp/$USER/dir"`, "$USER"); got != false {
			t.Errorf(`isSafelyEncapsulated("echo \"/tmp/$USER/dir\"", "$USER") = %v; want false`, got)
		}
		if got := isSafelyEncapsulated("echo \"/tmp/`whoami`/dir\"", "`whoami`"); got != false {
			t.Errorf("isSafelyEncapsulated with backticks substring in double quotes = %v; want false", got)
		}
	})

	t.Run("user input spans across quote boundary", func(t *testing.T) {
		if got := isSafelyEncapsulated("echo 'hello' world", "lo' wo"); got != false {
			t.Errorf("isSafelyEncapsulated spanning quote boundary = %v; want false", got)
		}
		if got := isSafelyEncapsulated(`echo "hello" world`, `lo" wo`); got != false {
			t.Errorf("isSafelyEncapsulated spanning double-quote boundary = %v; want false", got)
		}
	})

	t.Run("user input is the entire quoted content", func(t *testing.T) {
		if got := isSafelyEncapsulated("echo 'hello'", "hello"); got != true {
			t.Errorf("isSafelyEncapsulated entire single-quoted content = %v; want true", got)
		}
		if got := isSafelyEncapsulated(`echo "hello"`, "hello"); got != true {
			t.Errorf("isSafelyEncapsulated entire double-quoted content = %v; want true", got)
		}
	})

	t.Run("unterminated quotes", func(t *testing.T) {
		if got := isSafelyEncapsulated("echo 'hello", "hello"); got != false {
			t.Errorf("isSafelyEncapsulated unterminated single quote = %v; want false", got)
		}
		if got := isSafelyEncapsulated(`echo "hello`, "hello"); got != false {
			t.Errorf("isSafelyEncapsulated unterminated double quote = %v; want false", got)
		}
	})

	t.Run("single quotes inside double quotes are literal", func(t *testing.T) {
		if got := isSafelyEncapsulated(`echo "it's fine"`, "it's fine"); got != true {
			t.Errorf(`isSafelyEncapsulated("echo \"it's fine\"", "it's fine") = %v; want true`, got)
		}
	})

	t.Run("double quotes inside single quotes are literal", func(t *testing.T) {
		if got := isSafelyEncapsulated(`echo 'say "hi"'`, `say "hi"`); got != true {
			t.Errorf(`isSafelyEncapsulated single-quoted with inner double quotes = %v; want true`, got)
		}
	})

	t.Run("escaped double quote inside double quotes", func(t *testing.T) {
		if got := isSafelyEncapsulated(`echo "hello \"world\""`, "hello"); got != true {
			t.Errorf("isSafelyEncapsulated with escaped quotes = %v; want true", got)
		}
	})

	t.Run("user input appears both inside and outside quotes", func(t *testing.T) {
		if got := isSafelyEncapsulated("echo 'foo' foo", "foo"); got != false {
			t.Errorf("isSafelyEncapsulated with one safe and one unsafe occurrence = %v; want false", got)
		}
	})

	t.Run("multiple safe occurrences in different quote types", func(t *testing.T) {
		if got := isSafelyEncapsulated(`echo 'foo' "foo"`, "foo"); got != true {
			t.Errorf("isSafelyEncapsulated with safe occurrences in both quote types = %v; want true", got)
		}
	})

}
