package configsync_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/ecnepsnai/configsync"
	"github.com/ecnepsnai/logtic"
)

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
	if verbose {
		logtic.Log.Level = logtic.LevelDebug
		if err := logtic.Log.Open(); err != nil {
			panic(err)
		}
	}
}

func testTeardown() {
	logtic.Log.Close()
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

	workDir := t.TempDir()
	tmp := t.TempDir()

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
}

func TestConfigGlobNest(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	tmp := t.TempDir()
	if err := os.Mkdir(path.Join(tmp, "files"), 0700); err != nil {
		panic(err)
	}
	if err := os.Mkdir(path.Join(tmp, "files", "more_files"), 0700); err != nil {
		panic(err)
	}

	i := 0
	for i < 10 {
		if i%2 == 0 {
			touchFile(path.Join(path.Join(tmp, "files", "more_files"), fmt.Sprintf("%d.txt", i)))
		} else {
			touchFile(path.Join(path.Join(tmp, "files"), fmt.Sprintf("%d.txt", i)))
		}
		i++
	}

	files := []string{
		path.Join(tmp, "*"),
	}
	commands := []configsync.CommandType{}

	configsync.Start(workDir, files, commands, gitOptions)

	i = 5
	for i < 10 {
		if i%2 == 0 {
			os.Remove(path.Join(tmp, "files", "more_files", fmt.Sprintf("%d.txt", i)))
			touchFile(path.Join(tmp, "files", "more_files", fmt.Sprintf("%d.txt", i)))
		} else {
			os.Remove(path.Join(tmp, "files", fmt.Sprintf("%d.txt", i)))
			touchFile(path.Join(tmp, "files", fmt.Sprintf("%d.txt", i)))
		}
		i++
	}
	os.Remove(path.Join(tmp, "files", fmt.Sprintf("%d.txt", 0)))
	touchFile(path.Join(tmp, "files", fmt.Sprintf("%d.txt", 11)))
	files = files[1:]

	configsync.Start(workDir, files, commands, gitOptions)
}

func TestConfigsyncFile(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	tmp := t.TempDir()

	for _, idx := range []string{"1", "2", "3"} {
		os.Mkdir(path.Join(tmp, idx), os.ModePerm)
		touchFile(path.Join(tmp, idx, "foo.txt"))
	}

	files := []string{
		path.Join(tmp, "/1/foo.txt"),
		path.Join(tmp, "/2/foo.txt"),
		path.Join(tmp, "/3/foo.txt"),
		path.Join(tmp, "/4/foo.txt"),
	}
	commands := []configsync.CommandType{}
	configsync.Start(workDir, files, commands, gitOptions)

	os.Remove(path.Join(tmp, "1", "foo.txt"))
	touchFile(path.Join(tmp, "1", "foo.txt"))
	files = files[:2]

	configsync.Start(workDir, files, commands, gitOptions)
}

func TestConfigsyncCommand(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()

	files := []string{}
	commands := []configsync.CommandType{
		{
			ExePath:   "/usr/bin/openssl",
			Arguments: []string{"rand", "-hex", "10"},
			FilePath:  "/rand",
		},
		{
			ExePath:  "/usr/bin/hostname",
			FilePath: "/hostname",
		},
		{
			ExePath:  "/usr/bin/date",
			FilePath: "/date",
		},
	}
	configsync.Start(workDir, files, commands, gitOptions)

	commands = commands[:2]

	configsync.Start(workDir, files, commands, gitOptions)
}

func TestConfigsyncInvalidGlob(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()

	files := []string{
		`\`,
		"/does/not/**/map/to/anything",
	}
	commands := []configsync.CommandType{}
	configsync.Start(workDir, files, commands, gitOptions)
}

func TestConfigsyncInvalidCommand(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()

	files := []string{}
	commands := []configsync.CommandType{
		{
			ExePath:  "doesnotexist anymore",
			FilePath: "/blah",
		},
	}
	configsync.Start(workDir, files, commands, gitOptions)
}

func TestConfigsyncUnsuccessfulCommand(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()

	files := []string{}
	commands := []configsync.CommandType{
		{
			ExePath:   "/bin/bash",
			Arguments: []string{"-c", "exit 1"},
			FilePath:  "/test",
		},
	}
	configsync.Start(workDir, files, commands, gitOptions)
}

func TestConfigsyncExistingGitWorkdir(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()

	c := exec.Command("git", "init")
	c.Dir = workDir
	c.CombinedOutput()

	files := []string{}
	commands := []configsync.CommandType{
		{
			ExePath:   "/usr/bin/openssl",
			Arguments: []string{"rand", "-hex", "10"},
			FilePath:  "/rand",
		},
	}
	configsync.Start(workDir, files, commands, gitOptions)
}
