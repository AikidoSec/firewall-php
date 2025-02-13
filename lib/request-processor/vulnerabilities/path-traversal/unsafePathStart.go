package path_traversal

import (
	"path/filepath"
	"strings"
)

var linuxRootFolders = []string{
	"/bin/",
	"/boot/",
	"/dev/",
	"/etc/",
	"/home/",
	"/init/",
	"/lib/",
	"/media/",
	"/mnt/",
	"/opt/",
	"/proc/",
	"/root/",
	"/run/",
	"/sbin/",
	"/srv/",
	"/sys/",
	"/tmp/",
	"/usr/",
	"/var/",
}

var dangerousPathStarts = linuxRootFolders

// To enable when doing a Windows build (filepath.IsAbs does not work on Linux when a Windows path is checked)
// var dangerousPathStarts = append(linuxRootFolders, "c:/", "c:\\")

func normalizePath(p string) (string, error) {
	p, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return strings.ToLower(p), nil
}

func startsWithUnsafePath(accessFilePath string, userInput string) bool {
	if !filepath.IsAbs(accessFilePath) || !filepath.IsAbs(userInput) {
		return false
	}

	var err error
	normalizedAccessFilePath, err := normalizePath(accessFilePath)
	if err != nil {
		return false
	}
	normalizedUserInput, err := normalizePath(userInput)
	if err != nil {
		return false
	}

	for _, dangerousStart := range dangerousPathStarts {
		if strings.HasPrefix(normalizedAccessFilePath, dangerousStart) && strings.HasPrefix(normalizedAccessFilePath, normalizedUserInput) {
			if userInput == dangerousStart || userInput == dangerousStart[:len(dangerousStart)-1] {
				// If the user input is the same as the dangerous start, we don't want to flag it to prevent false positives
				// e.g. if user input is /etc/ and the path is /etc/passwd, we don't want to flag it, as long as the
				// user input does not contain a subdirectory or filename
				return false
			}
			return true
		}
	}
	return false
}
