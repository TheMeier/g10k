package internal

// GitModule contains information about a Git Puppet module
type GitModule struct {
	PrivateKey        string
	Git               string
	Branch            string
	Tag               string
	Commit            string
	Ref               string
	Tree              string
	Link              bool
	IgnoreUnreachable bool
	Fallback          []string
	InstallPath       string
	Local             bool
	ModuleDir         string
	UseSSHAgent       bool
}
