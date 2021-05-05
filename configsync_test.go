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
var gitOptions = configsync.GitOptionsType{
	Path: "/usr/bin/git",
}

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

	files := []string{
		path.Join(tmp, "*.txt"),
	}
	commands := []configsync.CommandType{}

	configsync.Start(workDir, files, commands, gitOptions)

	i = 5
	for i < 10 {
		os.Remove(path.Join(tmp, fmt.Sprintf("%d.txt", i)))
		touchFile(path.Join(tmp, fmt.Sprintf("%d.txt", i)))
		i++
	}
	os.Remove(path.Join(tmp, fmt.Sprintf("%d.txt", 0)))
	touchFile(path.Join(tmp, fmt.Sprintf("%d.txt", 11)))
	files = files[1:]

	configsync.Start(workDir, files, commands, gitOptions)

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

	files := []string{
		path.Join(tmp, "/1/foo.txt"),
		path.Join(tmp, "/2/foo.txt"),
		path.Join(tmp, "/3/foo.txt"),
	}
	commands := []configsync.CommandType{}
	configsync.Start(workDir, files, commands, gitOptions)

	os.Remove(path.Join(tmp, "1", "foo.txt"))
	touchFile(path.Join(tmp, "1", "foo.txt"))
	files = files[:2]

	configsync.Start(workDir, files, commands, gitOptions)

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

	files := []string{}
	commands := []configsync.CommandType{
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
	}
	configsync.Start(workDir, files, commands, gitOptions)

	commands = commands[:2]

	configsync.Start(workDir, files, commands, gitOptions)

	os.RemoveAll(workDir)
	os.RemoveAll(tmp)
}

func TestConfigsyncInvalidGlob(t *testing.T) {
	t.Parallel()

	workDir, err := os.MkdirTemp("", "configsync")
	if err != nil {
		panic(err)
	}

	files := []string{
		`\`,
		"/does/not/**/map/to/anything",
	}
	commands := []configsync.CommandType{}
	configsync.Start(workDir, files, commands, gitOptions)

	os.RemoveAll(workDir)
}

func TestConfigsyncInvalidCommand(t *testing.T) {
	t.Parallel()

	workDir, err := os.MkdirTemp("", "configsync")
	if err != nil {
		panic(err)
	}

	files := []string{}
	commands := []configsync.CommandType{
		{
			CommandLine: "doesnotexist anymore",
			Filepath:    "/blah",
		},
	}
	configsync.Start(workDir, files, commands, gitOptions)

	os.RemoveAll(workDir)
}
