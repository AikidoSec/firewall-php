package log

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A symlink planted under the predictable log file name must not be followed.
func TestOpenLogFileRefusesSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "protected")
	if err := os.WriteFile(target, []byte("ORIGINAL\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "aikido-agent.log")
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}

	if f, err := openLogFile(path); err == nil {
		f.Close()
		t.Fatal("opened a log file through a planted symlink")
	}
	content, _ := os.ReadFile(target)
	if string(content) != "ORIGINAL\n" {
		t.Fatalf("protected file was modified: %q", content)
	}
}

// A dangling symlink (target does not exist yet) must not be created through either.
func TestOpenLogFileRefusesDanglingSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "does-not-exist")
	path := filepath.Join(dir, "aikido-agent.log")
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}

	if f, err := openLogFile(path); err == nil {
		f.Close()
		t.Fatal("opened a log file through a dangling symlink")
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatal("symlink target was created through the dangling link")
	}
}

// A pre-existing regular file under the log name is never reused.
func TestOpenLogFileRefusesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "aikido-agent.log")
	if err := os.WriteFile(path, []byte("planted\n"), 0666); err != nil {
		t.Fatal(err)
	}

	if f, err := openLogFile(path); err == nil {
		f.Close()
		t.Fatal("reused a pre-existing log file")
	}
	content, _ := os.ReadFile(path)
	if string(content) != "planted\n" {
		t.Fatalf("pre-existing file was modified: %q", content)
	}
}

func TestOpenLogFileCreatesNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aikido-agent.log")

	f, err := openLogFile(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0022 != 0 {
		t.Fatalf("log file is group/world-writable: %o", info.Mode().Perm())
	}
}

// Caller-controlled strings must not be able to forge log lines.
func TestFormatEscapesNewlines(t *testing.T) {
	f := &AikidoFormatter{}
	out := f.Format(DebugLevel, "Received domain: \n/tmp/evil.so\r\n:80")

	if strings.Count(out, "\n") != 1 || !strings.HasSuffix(out, "\n") {
		t.Fatalf("log line contains injected newlines: %q", out)
	}
	if strings.Contains(out, "\r") {
		t.Fatalf("log line contains a carriage return: %q", out)
	}
	if !strings.Contains(out, `\n/tmp/evil.so\r\n:80`) {
		t.Fatalf("escaped payload missing from log line: %q", out)
	}
}

// Existing callers end messages with "\n"; that must not show up as a literal `\n`.
func TestFormatTrimsTrailingNewline(t *testing.T) {
	f := &AikidoFormatter{}
	out := f.Format(InfoLevel, "Run directory created successfully.\n")

	if !strings.HasSuffix(out, "Run directory created successfully.\n") {
		t.Fatalf("trailing newline was not trimmed: %q", out)
	}
}
