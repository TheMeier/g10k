package puppetfile

import (
	"time"

	"github.com/xorpaul/g10k/internal"
)

// Puppetfile contains the key value pairs from the Puppetfile
type Puppetfile struct {
	ForgeBaseURL      string
	ForgeCacheTTL     time.Duration
	ForgeModules      map[string]internal.ForgeModule
	GitModules        map[string]internal.GitModule
	PrivateKey        string
	Source            string
	SourceBranch      string
	WorkDir           string
	GitDir            string
	GitURL            string
	ModuleDirs        []string
	ControlRepoBranch string
}
