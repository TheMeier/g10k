package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/xorpaul/g10k/internal"
	"github.com/xorpaul/g10k/pkg/puppetfile"
)

func equalPuppetfile(a, b puppetfile.Puppetfile) bool {
	if &a == &b {
		return true
	}
	if a.ForgeBaseURL != b.ForgeBaseURL ||
		a.ForgeCacheTTL != b.ForgeCacheTTL ||
		a.PrivateKey != b.PrivateKey ||
		a.ControlRepoBranch != b.ControlRepoBranch ||
		a.Source != b.Source {
		Debugf("forgeBaseURL, forgeCacheTTL, privateKey, controlRepoBranch or source isn't equal!")
		return false
	}

	if len(a.GitModules) != len(b.GitModules) ||
		len(a.ForgeModules) != len(b.ForgeModules) {
		Debugf("size of gitModules or forgeModules isn't equal!")
		return false
	}

	for gitModuleName, gm := range a.GitModules {
		if _, ok := b.GitModules[gitModuleName]; !ok {
			Debugf("git module " + gitModuleName + " missing!")
			return false
		}
		if !equalGitModule(gm, b.GitModules[gitModuleName]) {
			Debugf("git module " + gitModuleName + " isn't equal!")
			return false
		}
	}

	for forgeModuleName, fm := range a.ForgeModules {
		if _, ok := b.ForgeModules[forgeModuleName]; !ok {
			Debugf("forge module " + forgeModuleName + " missing!")
			return false
		}
		//fmt.Println("checking Forge module: ", forgeModuleName, fm)
		if !equalForgeModule(fm, b.ForgeModules[forgeModuleName]) {
			Debugf("forge module " + forgeModuleName + " isn't equal!")
			return false
		}
	}

	return true
}

func equalForgeResult(a, b ForgeResult) bool {
	if &a == &b {
		return true
	}
	if a.needToGet != b.needToGet || a.versionNumber != b.versionNumber ||
		a.fileSize != b.fileSize {
		return false
	}
	return true
}

func equalForgeModule(a, b internal.ForgeModule) bool {
	if &a == &b {
		return true
	}
	if a.Author != b.Author || a.Name != b.Name ||
		a.Version != b.Version ||
		a.Md5sum != b.Md5sum ||
		a.Sha256sum != b.Sha256sum ||
		a.FileSize != b.FileSize ||
		a.BaseURL != b.BaseURL ||
		a.CacheTTL != b.CacheTTL {
		return false
	}
	return true
}

func equalGitModule(a, b internal.GitModule) bool {
	if &a == &b {
		return true
	}
	if a.Git != b.Git ||
		a.PrivateKey != b.PrivateKey ||
		a.Branch != b.Branch ||
		a.Tag != b.Tag ||
		a.Commit != b.Commit ||
		a.Ref != b.Ref ||
		a.Link != b.Link ||
		a.IgnoreUnreachable != b.IgnoreUnreachable ||
		a.InstallPath != b.InstallPath ||
		a.Local != b.Local ||
		a.UseSSHAgent != b.UseSSHAgent {
		return false
	}
	if len(a.Fallback) != len(b.Fallback) {
		return false
	}
	for i, v := range a.Fallback {
		if b.Fallback[i] != v {
			return false
		}
	}
	return true
}

func checkExitCodeAndOutputOfReadPuppetfileSubprocess(t *testing.T, forceForgeVersions bool, expectedExitCode int, expectedOutput string) {
	pc, _, _, _ := runtime.Caller(1)
	testFunctionName := strings.Split(runtime.FuncForPC(pc).Name(), ".")[len(strings.Split(runtime.FuncForPC(pc).Name(), "."))-1]
	if os.Getenv("TEST_FOR_CRASH_"+testFunctionName) == "1" {
		readPuppetfile("tests/"+testFunctionName, "", "test", "test", forceForgeVersions, false)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run="+testFunctionName+"$")
	cmd.Env = append(os.Environ(), "TEST_FOR_CRASH_"+testFunctionName+"=1")
	out, err := cmd.CombinedOutput()
	if debug {
		fmt.Print(string(out))
	}

	exitCode := 0
	if msg, ok := err.(*exec.ExitError); ok { // there is error code
		exitCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}

	if expectedExitCode != exitCode {
		t.Errorf("readPuppetfile() terminated with %v, but we expected exit status %v", exitCode, expectedExitCode)
	}
	if !strings.Contains(string(out), expectedOutput) {
		t.Errorf("readPuppetfile() terminated with the correct exit code, but the expected output was missing. out: %s", string(out))
	}
}

func TestPreparePuppetfile(t *testing.T) {
	expected := regexp.MustCompile("(moduledir 'external_modules'\nmod 'puppetlabs/ntp')")
	got := preparePuppetfile("tests/TestPreparePuppetfile")

	if !expected.MatchString(got) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestCommentPuppetfile(t *testing.T) {
	expected := regexp.MustCompile(`mod 'sensu',\s*:git => 'https://github.com/sensu/sensu-puppet.git',\s*:commit => '8f4fc5780071c4895dec559eafc6030511b0caaa'`)
	got := preparePuppetfile("tests/TestCommentPuppetfile")

	if !expected.MatchString(got) {
		spew.Dump(expected)
		spew.Dump(got)
		t.Error("Expected", expected, "got", got)
	}
}

func TestReadPuppetfile(t *testing.T) {
	funcName := strings.Split(funcName(), ".")[len(strings.Split(funcName(), "."))-1]
	got := readPuppetfile("tests/"+funcName, "", "test", "test", false, false)

	fallbackMapExample := make([]string, 1)
	fallbackMapExample[0] = "master"

	fallbackMapExampleFull := make([]string, 3)
	fallbackMapExampleFull[0] = "b"
	fallbackMapExampleFull[1] = "a"
	fallbackMapExampleFull[2] = "r"

	fallbackMapAnother := make([]string, 4)
	fallbackMapAnother[0] = "dev"
	fallbackMapAnother[1] = "qa"
	fallbackMapAnother[2] = "prelive"
	fallbackMapAnother[3] = "live"

	gm := make(map[string]internal.GitModule)
	gm["sensu"] = internal.GitModule{Git: "https://github.com/sensu/sensu-puppet.git",
		Commit: "8f4fc5780071c4895dec559eafc6030511b0caaa", IgnoreUnreachable: false}
	gm["example_module"] = internal.GitModule{Git: "git@somehost.com/foo/example-module.git",
		Link: true, IgnoreUnreachable: false, Fallback: fallbackMapExample}
	gm["another_module"] = internal.GitModule{Git: "git@somehost.com/foo/another-module.git",
		Link: true, IgnoreUnreachable: false, Fallback: fallbackMapAnother}
	gm["example_module_full"] = internal.GitModule{Git: "git@somehost.com/foo/example-module.git",
		Branch: "foo", IgnoreUnreachable: true, Fallback: fallbackMapExampleFull}

	fm := make(map[string]internal.ForgeModule)
	fm["apt"] = internal.ForgeModule{Version: "2.3.0", Author: "puppetlabs", Name: "apt"}
	fm["ntp"] = internal.ForgeModule{Version: "present", Author: "puppetlabs", Name: "ntp"}
	fm["stdlib"] = internal.ForgeModule{Version: "latest", Author: "puppetlabs", Name: "stdlib"}

	expected := puppetfile.Puppetfile{GitModules: gm, ForgeModules: fm, Source: "test", ForgeCacheTTL: time.Duration(50 * time.Minute), ForgeBaseURL: "foobar"}

	if !equalPuppetfile(got, expected) {
		spew.Dump(expected)
		spew.Dump(got)
		t.Error("Expected Puppetfile:", expected, ", but got Puppetfile:", got)
	}
}

func TestFallbackPuppetfile(t *testing.T) {
	fallbackMapExample := make([]string, 1)
	fallbackMapExample[0] = "master"

	fallbackMapAnother := make([]string, 4)
	fallbackMapAnother[0] = "dev"
	fallbackMapAnother[1] = "qa"
	fallbackMapAnother[2] = "prelive"
	fallbackMapAnother[3] = "live"

	gm := make(map[string]internal.GitModule)
	gm["example_module"] = internal.GitModule{Git: "git@somehost.com/foo/example-module.git",
		Link: true, IgnoreUnreachable: false, Fallback: fallbackMapExample}
	gm["another_module"] = internal.GitModule{Git: "git@somehost.com/foo/another-module.git",
		Branch: "master", IgnoreUnreachable: false, Fallback: fallbackMapAnother}

	expected := puppetfile.Puppetfile{GitModules: gm, Source: "test"}
	got := readPuppetfile("tests/TestFallbackPuppetfile", "", "test", "test", false, false)

	if !equalGitModule(got.GitModules["example_module"], expected.GitModules["example_module"]) {
		t.Error("Expected gitModules:", expected.GitModules["example_module"], ", but got gitModules:", got.GitModules["example_module"])
	}

	if !equalGitModule(got.GitModules["another_module"], expected.GitModules["another_module"]) {
		t.Error("Expected gitModules:", expected.GitModules["another_module"], ", but got gitModules:", got.GitModules["another_module"])
	}
}

func TestForgeCacheTTLPuppetfile(t *testing.T) {
	expected := regexp.MustCompile("(moduledir 'external_modules'\nforge.cacheTtl 50m\n)")
	got := preparePuppetfile("tests/TestForgeCacheTTLPuppetfile")

	if !expected.MatchString(got) {
		t.Error("Expected", expected, "got", got)
	}

	expectedPuppetfile := puppetfile.Puppetfile{ForgeCacheTTL: 50 * time.Minute}
	gotPuppetfile := readPuppetfile("tests/TestForgeCacheTTLPuppetfile", "", "test", "test", false, false)

	if gotPuppetfile.ForgeCacheTTL != expectedPuppetfile.ForgeCacheTTL {
		t.Error("Expected for forgeCacheTTL", expectedPuppetfile.ForgeCacheTTL, "got", gotPuppetfile.ForgeCacheTTL)
	}

}

func TestForceForgeVersionsPuppetfile(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, true, 1, "")
}

func TestForceForgeVersionsPuppetfileCorrect(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, true, 0, "")
}

func TestReadPuppetfileDuplicateGitAttribute(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileTrailingComma(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileInvalidForgeModuleName(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileDuplicateForgeModule(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileMissingGitAttribute(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileTooManyGitAttributes(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileConflictingGitAttributesTag(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileConflictingGitAttributesBranch(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileConflictingGitAttributesCommit(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileConflictingGitAttributesRef(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileIgnoreUnreachable(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileForgeCacheTTL(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "Error: Can not convert value 300x of parameter forge.cacheTtl 300x to a golang Duration. Valid time units are 300ms, 1.5h or 2h45m. In tests/TestReadPuppetfileForgeCacheTTL line: forge.cacheTtl 300x")
}

func TestReadPuppetfileLink(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "Error: Found conflicting git attributes :branch, :link, in tests/TestReadPuppetfileLink for module example_module line: mod 'example_module',:git => 'git@somehost.com/foo/example-module.git',:branch => 'foo',:link => true")
}

func TestReadPuppetfileDuplicateForgeGitModule(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "Error: Git Puppet module with same name found in tests/TestReadPuppetfileDuplicateForgeGitModule for module bar line: mod 'bar',:git => 'https://github.com/foo/bar.git'")
}

func TestReadPuppetfileChecksumAttribute(t *testing.T) {
	funcName := strings.Split(funcName(), ".")[len(strings.Split(funcName(), "."))-1]
	got := readPuppetfile("tests/"+funcName, "", "test", "test", false, false)

	fm := make(map[string]internal.ForgeModule)
	fm["ntp"] = internal.ForgeModule{Version: "6.0.0", Author: "puppetlabs", Name: "ntp", Sha256sum: "a988a172a3edde6ac2a26d0e893faa88d37bc47465afc50d55225a036906c944"}
	fm["stdlib"] = internal.ForgeModule{Version: "2.3.0", Author: "puppetlabs", Name: "stdlib", Sha256sum: "433c69fb99a46185e81619fadb70e0961bce2f4e952294a16e61364210d1519d"}
	fm["apt"] = internal.ForgeModule{Version: "2.3.0", Author: "puppetlabs", Name: "apt", Sha256sum: "a09290c207bbfed7f42dd0356ff4dee16e138c7f9758d2134a21aeb66e14072f"}
	fm["concat"] = internal.ForgeModule{Version: "2.2.0", Author: "puppetlabs", Name: "concat", Sha256sum: "ec0407abab71f57e106ade6ed394410d08eec29bdad4c285580e7b56514c5194"}

	expected := puppetfile.Puppetfile{ForgeModules: fm, Source: "test"}

	if !equalPuppetfile(got, expected) {
		spew.Dump(expected)
		spew.Dump(got)
		t.Error("Expected Puppetfile:", expected, ", but got Puppetfile:", got)
	}
}

func TestReadPuppetfileForgeSlashNotation(t *testing.T) {
	funcName := strings.Split(funcName(), ".")[len(strings.Split(funcName(), "."))-1]

	got := readPuppetfile("tests/"+funcName, "", "test", "test", false, false)
	fm := make(map[string]internal.ForgeModule)
	fm["filebeat"] = internal.ForgeModule{Version: "0.10.4", Author: "pcfens", Name: "filebeat"}
	expected := puppetfile.Puppetfile{ForgeModules: fm, Source: "test"}
	if !equalPuppetfile(got, expected) {
		spew.Dump(expected)
		spew.Dump(got)
		t.Error("Expected Puppetfile:", expected, ", but got Puppetfile:", got)
	}

}

func TestReadPuppetfileForgeDash(t *testing.T) {
	funcName := strings.Split(funcName(), ".")[len(strings.Split(funcName(), "."))-1]
	got := readPuppetfile("tests/"+funcName, "", "test", "test", false, false)

	fm := make(map[string]internal.ForgeModule)
	fm["php"] = internal.ForgeModule{Version: "4.0.0-beta1", Author: "mayflower", Name: "php"}

	expected := puppetfile.Puppetfile{ForgeModules: fm, Source: "test"}

	if !equalPuppetfile(got, expected) {
		spew.Dump(expected)
		spew.Dump(got)
		t.Errorf("Expected Puppetfile: %+v, but got Puppetfile: %+v", expected, got)
	}
}

func TestReadPuppetfileInstallPath(t *testing.T) {
	quiet = true
	funcName := strings.Split(funcName(), ".")[len(strings.Split(funcName(), "."))-1]
	got := readPuppetfile("tests/"+funcName, "", "test", "test", false, false)

	gm := make(map[string]internal.GitModule)
	gm["sensu"] = internal.GitModule{Git: "https://github.com/sensu/sensu-puppet.git", Commit: "8f4fc5780071c4895dec559eafc6030511b0caaa", InstallPath: "external"}

	expected := puppetfile.Puppetfile{GitModules: gm, Source: "test"}
	//fmt.Println(got)

	if !equalPuppetfile(got, expected) {
		spew.Dump(expected)
		spew.Dump(got)
		t.Errorf("Expected Puppetfile: %+v, but got Puppetfile: %+v", expected, got)
	}
}

func TestReadPuppetfileLocalModule(t *testing.T) {
	quiet = true
	funcName := strings.Split(funcName(), ".")[len(strings.Split(funcName(), "."))-1]
	got := readPuppetfile("tests/"+funcName, "", "test", "test", false, false)

	gm := make(map[string]internal.GitModule)
	gm["localstuff"] = internal.GitModule{Local: true}
	gm["localstuff2"] = internal.GitModule{Local: true}
	gm["localstuff3"] = internal.GitModule{Local: false}
	gm["external"] = internal.GitModule{Local: true, InstallPath: "modules"}

	expected := puppetfile.Puppetfile{Source: "test", GitModules: gm}
	//fmt.Println(got)

	if !equalPuppetfile(got, expected) {
		spew.Dump(expected)
		spew.Dump(got)
		t.Errorf("Expected Puppetfile: %+v, but got Puppetfile: %+v", expected, got)
	}
}

func TestReadPuppetfileMissingTrailingComma(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileMissingTrailingComma2(t *testing.T) {
	checkExitCodeAndOutputOfReadPuppetfileSubprocess(t, false, 1, "")
}

func TestReadPuppetfileForgeNotationGitModule(t *testing.T) {
	quiet = true
	funcName := strings.Split(funcName(), ".")[len(strings.Split(funcName(), "."))-1]
	got := readPuppetfile("tests/"+funcName, "", "test", "test", false, false)

	gm := make(map[string]internal.GitModule)
	gm["elasticsearch"] = internal.GitModule{Git: "https://github.com/elastic/puppet-elasticsearch.git", Branch: "5.x"}

	expected := puppetfile.Puppetfile{Source: "test", GitModules: gm}
	//fmt.Println(got)

	if !equalPuppetfile(got, expected) {
		spew.Dump(expected)
		spew.Dump(got)
		t.Errorf("Expected Puppetfile: %+v, but got Puppetfile: %+v", expected, got)
	}
}

func TestReadPuppetfileGitSlashNotation(t *testing.T) {
	quiet = true
	funcName := strings.Split(funcName(), ".")[len(strings.Split(funcName(), "."))-1]
	got := readPuppetfile("tests/"+funcName, "", "test", "test", false, false)

	fm := make(map[string]internal.ForgeModule)
	fm["stdlib"] = internal.ForgeModule{Version: "present", Author: "puppetlabs", Name: "stdlib"}
	fm["apache"] = internal.ForgeModule{Version: "present", Author: "puppetlabs", Name: "apache"}
	fm["apt"] = internal.ForgeModule{Version: "latest", Author: "puppetlabs", Name: "apt"}
	fm["rsync"] = internal.ForgeModule{Version: "latest", Author: "puppetlabs", Name: "rsync"}

	gm := make(map[string]internal.GitModule)
	gm["puppetboard"] = internal.GitModule{Git: "https://github.com/nibalizer/puppet-module-puppetboard.git", Ref: "2.7.1"}
	gm["elasticsearch"] = internal.GitModule{Git: "https://github.com/alexharv074/puppet-elasticsearch.git", Ref: "alex_master"}

	expected := puppetfile.Puppetfile{Source: "test", GitModules: gm, ForgeModules: fm}
	//fmt.Println(got)

	if !equalPuppetfile(got, expected) {
		spew.Dump(expected)
		spew.Dump(got)
		t.Errorf("Expected Puppetfile: %+v, but got Puppetfile: %+v", expected, got)
	}
}

func TestReadPuppetfileGitDashNotation(t *testing.T) {
	quiet = true
	funcName := strings.Split(funcName(), ".")[len(strings.Split(funcName(), "."))-1]
	got := readPuppetfile("tests/"+funcName, "", "test", "test", false, false)

	fm := make(map[string]internal.ForgeModule)
	fm["stdlib"] = internal.ForgeModule{Version: "present", Author: "puppetlabs", Name: "stdlib"}
	fm["apache"] = internal.ForgeModule{Version: "present", Author: "puppetlabs", Name: "apache"}
	fm["apt"] = internal.ForgeModule{Version: "latest", Author: "puppetlabs", Name: "apt"}
	fm["rsync"] = internal.ForgeModule{Version: "latest", Author: "puppetlabs", Name: "rsync"}

	gm := make(map[string]internal.GitModule)
	gm["puppetboard"] = internal.GitModule{Git: "https://github.com/nibalizer/puppet-module-puppetboard.git", Ref: "2.7.1"}
	gm["elasticsearch"] = internal.GitModule{Git: "https://github.com/alexharv074/puppet-elasticsearch.git", Ref: "alex_master"}

	expected := puppetfile.Puppetfile{Source: "test", GitModules: gm, ForgeModules: fm}
	//fmt.Println(got)

	if !equalPuppetfile(got, expected) {
		spew.Dump(expected)
		spew.Dump(got)
		t.Errorf("Expected Puppetfile: %+v, but got Puppetfile: %+v", expected, got)
	}
}

func TestReadPuppetfileGitDashNSlashNotation(t *testing.T) {
	quiet = true
	funcName := strings.Split(funcName(), ".")[len(strings.Split(funcName(), "."))-1]
	got := readPuppetfile("tests/"+funcName, "", "test", "test", false, false)

	fm := make(map[string]internal.ForgeModule)
	fm["stdlib"] = internal.ForgeModule{Version: "present", Author: "puppetlabs", Name: "stdlib"}
	fm["apache"] = internal.ForgeModule{Version: "present", Author: "puppetlabs", Name: "apache"}
	fm["apt"] = internal.ForgeModule{Version: "latest", Author: "puppetlabs", Name: "apt"}
	fm["rsync"] = internal.ForgeModule{Version: "latest", Author: "puppetlabs", Name: "rsync"}

	gm := make(map[string]internal.GitModule)
	gm["puppetboard"] = internal.GitModule{Git: "https://github.com/nibalizer/puppet-module-puppetboard.git", Ref: "2.7.1"}
	gm["elasticsearch"] = internal.GitModule{Git: "https://github.com/alexharv074/puppet-elasticsearch.git", Ref: "alex_master"}

	expected := puppetfile.Puppetfile{Source: "test", GitModules: gm, ForgeModules: fm}
	//fmt.Println(got)

	if !equalPuppetfile(got, expected) {
		spew.Dump(expected)
		spew.Dump(got)
		t.Errorf("Expected Puppetfile: %+v, but got Puppetfile: %+v", expected, got)
	}
}

func TestReadPuppetfileSSHKeyAlreadyLoaded(t *testing.T) {
	quiet = true
	funcName := strings.Split(funcName(), ".")[len(strings.Split(funcName(), "."))-1]
	got := readPuppetfile("tests/"+funcName, "", "test", "test", false, false)

	fm := make(map[string]internal.ForgeModule)
	gm := make(map[string]internal.GitModule)
	gm["example_module"] = internal.GitModule{Git: "git@somehost.com/foo/example-module.git", Branch: "foo", UseSSHAgent: true}

	expected := puppetfile.Puppetfile{Source: "test", GitModules: gm, ForgeModules: fm}
	//fmt.Println(got)

	if !equalPuppetfile(got, expected) {
		fmt.Println("Expected:")
		spew.Dump(expected)
		fmt.Println("Got:")
		spew.Dump(got)
		t.Errorf("Expected Puppetfile: %+v, but got Puppetfile: %+v", expected, got)
	}
}
