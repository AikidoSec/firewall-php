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

// to lowercase all directory names
var DirectoryNames = func() []string {
	lowercaseDirectoryNames := make([]string, len(directoryNamesList))
	for i, directoryName := range directoryNamesList {
		lowercaseDirectoryNames[i] = strings.ToLower(directoryName)
	}
	return lowercaseDirectoryNames
}()
