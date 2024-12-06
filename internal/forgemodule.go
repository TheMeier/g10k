package internal

import "time"

// ForgeModule contains information (Version, Name, Author, md5 checksum, file size of the tar.gz archive, Forge BaseURL if custom) about a Puppetlabs Forge module
type ForgeModule struct {
	Version      string
	Name         string
	Author       string
	Md5sum       string
	FileSize     int64
	BaseURL      string
	CacheTTL     time.Duration
	Sha256sum    string
	ModuleDir    string
	SourceBranch string
}
