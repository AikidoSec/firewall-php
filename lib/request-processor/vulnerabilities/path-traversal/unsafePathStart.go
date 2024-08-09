package path_traversal

import (
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

var dangerousPathStarts = append(linuxRootFolders, "c:/", "c:\\")

func startsWithUnsafePath(filePath string, userInput string) bool {
	lowerCasePath := strings.ToLower(filePath)
	lowerCaseUserInput := strings.ToLower(userInput)

	for _, dangerousStart := range dangerousPathStarts {
		if strings.HasPrefix(lowerCasePath, dangerousStart) && strings.HasPrefix(lowerCasePath, lowerCaseUserInput) {
			return true
		}
	}
	return false
}
