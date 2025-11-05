package paths

import "strings"

var directoryNamesList = []string{
	".",
	"..",
	".anydesk",
	".aptitude",
	".aws",
	".azure",
	".cache",
	".circleci",
	".config",
	".dbus",
	".docker",
	".drush",
	".gem",
	".git",
	".github",
	".gnupg",
	".gsutil",
	".hg",
	".idea",
	".java",
	".kube",
	".lftp",
	".minikube",
	".npm",
	".nvm",
	".pki",
	".snap",
	".ssh",
	".subversion",
	".svn",
	".tconn",
	".thunderbird",
	".tor",
	".vagrant.d",
	".vidalia",
	".vim",
	".vmware",
	".vscode",
	"apache",
	"apache2",
	"grub",
	"System32",
	"tmp",
	"xampp",
	"cgi-bin",
	"%systemroot%",
}

// to lowercase all directory names and return a map (map lookup is faster than list lookup)
var DirectoryNames = func() map[string]struct{} {
	lowercaseDirectoryNames := make(map[string]struct{})
	for _, directoryName := range directoryNamesList {
		lowercaseDirectoryNames[strings.ToLower(directoryName)] = struct{}{}
	}
	return lowercaseDirectoryNames
}()
