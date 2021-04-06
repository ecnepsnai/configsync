package configsync_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/ecnepsnai/configsync"
	"github.com/ecnepsnai/logtic"
)

var tmpDir string
var verbose bool

func isTestVerbose() bool {
	for _, arg := range os.Args {
		if arg == "-test.v=true" {
			return true
		}
	}

	return false
}

func testSetup() {
	tmp, err := os.MkdirTemp("", "configsync")
	if err != nil {
		panic(err)
	}
	tmpDir = tmp

	if verbose {
		logtic.Log.FilePath = "/dev/null"
		logtic.Log.Level = logtic.LevelDebug
		if err := logtic.Open(); err != nil {
			panic(err)
		}
	}
}

func testTeardown() {
	os.RemoveAll(tmpDir)
	logtic.Close()
}

func TestMain(m *testing.M) {
	verbose = isTestVerbose()
	testSetup()
	retCode := m.Run()
	testTeardown()
	os.Exit(retCode)
}

func randomString(length uint16) string {
	randB := make([]byte, length)
	rand.Read(randB)
	return hex.EncodeToString(randB)
}

func touchFile(filePath string) error {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	if _, err := f.Write([]byte(randomString(6))); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

func TestConfigsyncGlob(t *testing.T) {
	t.Parallel()

	workDir, err := os.MkdirTemp("", "configsync")
	if err != nil {
		panic(err)
	}
	tmp, err := os.MkdirTemp("", "configsync")
	if err != nil {
		panic(err)
	}

	i := 0
	for i < 10 {
		touchFile(path.Join(tmp, fmt.Sprintf("%d.txt", i)))
		i++
	}

	config := configsync.ConfigType{
		Workdir: workDir,
		Files: []string{
			path.Join(tmp, "*.txt"),
		},
		Git: configsync.GitConfigType{
			Path: "/usr/bin/git",
		},
	}
	configsync.Start(&config)

	i = 5
	for i < 10 {
		os.Remove(path.Join(tmp, fmt.Sprintf("%d.txt", i)))
		touchFile(path.Join(tmp, fmt.Sprintf("%d.txt", i)))
		i++
	}
	os.Remove(path.Join(tmp, fmt.Sprintf("%d.txt", 0)))
	touchFile(path.Join(tmp, fmt.Sprintf("%d.txt", 11)))
	config.Files = config.Files[1:]

	configsync.Start(&config)

	os.RemoveAll(workDir)
	os.RemoveAll(tmp)
}

func TestConfigsyncFile(t *testing.T) {
	t.Parallel()

	workDir, err := os.MkdirTemp("", "configsync")
	if err != nil {
		panic(err)
	}
	tmp, err := os.MkdirTemp("", "configsync")
	if err != nil {
		panic(err)
	}

	for _, idx := range []string{"1", "2", "3"} {
		os.Mkdir(path.Join(tmp, idx), os.ModePerm)
		touchFile(path.Join(tmp, idx, "foo.txt"))
	}

	config := configsync.ConfigType{
		Workdir: workDir,
		Files: []string{
			path.Join(tmp, "/1/foo.txt"),
			path.Join(tmp, "/2/foo.txt"),
			path.Join(tmp, "/3/foo.txt"),
		},
		Git: configsync.GitConfigType{
			Path: "/usr/bin/git",
		},
	}
	configsync.Start(&config)

	os.Remove(path.Join(tmp, "1", "foo.txt"))
	touchFile(path.Join(tmp, "1", "foo.txt"))
	config.Files = config.Files[:2]

	configsync.Start(&config)

	os.RemoveAll(workDir)
	os.RemoveAll(tmp)
}

func TestConfigsyncCommand(t *testing.T) {
	t.Parallel()

	workDir, err := os.MkdirTemp("", "configsync")
	if err != nil {
		panic(err)
	}
	tmp, err := os.MkdirTemp("", "configsync")
	if err != nil {
		panic(err)
	}

	config := configsync.ConfigType{
		Workdir: workDir,
		Commands: []configsync.CommandType{
			{
				CommandLine: "openssl rand -hex 10",
				Filepath:    "/rand",
			},
			{
				CommandLine: "hostname",
				Filepath:    "/hostname",
			},
			{
				CommandLine: "date",
				Filepath:    "/date",
			},
		},
		Git: configsync.GitConfigType{
			Path: "/usr/bin/git",
		},
	}
	configsync.Start(&config)
	config.Commands = config.Commands[:2]

	configsync.Start(&config)

	os.RemoveAll(workDir)
	os.RemoveAll(tmp)
}

func TestConfigsyncInvalidGlob(t *testing.T) {
	t.Parallel()

	workDir, err := os.MkdirTemp("", "configsync")
	if err != nil {
		panic(err)
	}

	config := configsync.ConfigType{
		Workdir: workDir,
		Files: []string{
			`\`,
			"/does/not/**/map/to/anything",
		},
		Git: configsync.GitConfigType{
			Path: "/usr/bin/git",
		},
	}
	configsync.Start(&config)
	os.RemoveAll(workDir)
}

func TestConfigsyncInvalidCommand(t *testing.T) {
	t.Parallel()

	workDir, err := os.MkdirTemp("", "configsync")
	if err != nil {
		panic(err)
	}

	config := configsync.ConfigType{
		Workdir: workDir,
		Commands: []configsync.CommandType{
			{
				CommandLine: "doesnotexist anymore",
				Filepath:    "/blah",
			},
		},
		Git: configsync.GitConfigType{
			Path: "/usr/bin/git",
		},
	}
	configsync.Start(&config)
	os.RemoveAll(workDir)
}
