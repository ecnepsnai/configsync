// Command configsync is a command line tool to sync system configuration files and command outputs in a git repo
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ecnepsnai/logtic"
)

var log = logtic.Connect("configsync")

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage %s <config JSON path>\n", os.Args[0])
		os.Exit(1)
	}

	f, err := os.OpenFile(args[1], os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	config := configType{}
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		panic(err)
	}

	if config.Workdir == "" {
		fmt.Fprintf(os.Stderr, "Workdir is required\n")
		os.Exit(1)
	}

	if len(config.Commands) == 0 && len(config.Files) == 0 {
		fmt.Fprintf(os.Stderr, "At least one file or command is required\n")
		os.Exit(1)
	}

	if config.Git.RemoteEnabled && config.Git.RemoteName == "" {
		fmt.Fprintf(os.Stderr, "Remote name is required if git remote is enabled\n")
		os.Exit(1)
	}

	if config.Git.Path == "" {
		gitPath := findGitBin()
		if gitPath == "" {
			fmt.Fprintf(os.Stderr, "Git binary not specified and not found anywhere on $PATH\n")
			os.Exit(1)
		}
		config.Git.Path = gitPath
	}

	startSync(&config)
}

func findGitBin() string {
	envPath, present := os.LookupEnv("PATH")
	if !present {
		return ""
	}
	for _, p := range strings.Split(envPath, ":") {
		gitPath := path.Join(p, "git")
		if fileExists(gitPath) {
			return gitPath
		}
	}

	return ""
}
