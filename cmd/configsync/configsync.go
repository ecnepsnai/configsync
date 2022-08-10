package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/ecnepsnai/configsync"
	"github.com/ecnepsnai/logtic"
	"github.com/pelletier/go-toml"
)

var log = logtic.Log.Connect("configsync")

func printHelpAndExit() {
	fmt.Fprintf(os.Stderr, "Usage %s [Override config path]\n", os.Args[0])
	os.Exit(1)
}

func main() {
	args := os.Args
	if len(args) >= 2 && (args[1] == "-h" || args[1] == "--help") {
		printHelpAndExit()
	}

	configPath := "configsync.conf"
	if len(args) == 2 {
		configPath = os.Args[1]
	}

	f, err := os.OpenFile(configPath, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	config := configSyncOptionsType{}
	if err := toml.NewDecoder(f).Decode(&config); err != nil {
		panic(err)
	}

	if config.Workdir == "" {
		fmt.Fprintf(os.Stderr, "Workdir is required\n")
		os.Exit(1)
	}

	if len(config.commands()) == 0 && len(config.filePatterns()) == 0 {
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

	if config.Verbose {
		logtic.Log.Level = logtic.LevelDebug
	} else {
		logtic.Log.Level = logtic.LevelWarn
	}
	logtic.Log.Open()

	configsync.Start(config.Workdir, config.filePatterns(), config.commands(), config.Git)
}

func findGitBin() string {
	envPath, present := os.LookupEnv("PATH")
	if !present {
		return ""
	}
	for _, p := range strings.Split(envPath, ":") {
		gitPath := path.Join(p, "git")
		info, _ := os.Stat(gitPath)
		if info != nil {
			return gitPath
		}
	}

	return ""
}

type configSyncOptionsType struct {
	ConfInclude string                    `toml:"conf_include"`
	Workdir     string                    `toml:"workdir"`
	Git         configsync.GitOptionsType `toml:"git"`
	Verbose     bool                      `toml:"verbose"`
}

func (c configSyncOptionsType) includeFilesWithExtension(ext string) []string {
	incFiles := []string{}
	files, _ := os.ReadDir(c.ConfInclude)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ext) {
			incFiles = append(incFiles, file.Name())
		}
	}

	return incFiles
}

func (c configSyncOptionsType) filePatterns() []string {
	patterns := []string{}

	for _, includeFile := range c.includeFilesWithExtension(".files") {
		f, err := os.OpenFile(path.Join(c.ConfInclude, includeFile), os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Error("Error opening file %s: %s", includeFile, err.Error())
			continue
		}
		defer f.Close()
		data, _ := io.ReadAll(f)

		for _, line := range strings.Split(string(data), "\n") {
			if line == "" {
				continue
			}
			if line[0] == '#' {
				continue
			}
			patterns = append(patterns, line)
		}
	}

	return patterns
}

func (c configSyncOptionsType) commands() []configsync.CommandType {
	commands := []configsync.CommandType{}

	for _, includeFile := range c.includeFilesWithExtension(".cmd") {
		f, err := os.OpenFile(path.Join(c.ConfInclude, includeFile), os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Error("Error opening file %s: %s", includeFile, err.Error())
			continue
		}
		defer f.Close()
		command := configsync.CommandType{}
		if err := toml.NewDecoder(f).Decode(&command); err != nil {
			log.Error("Error decoding command file %s: %s", includeFile, err.Error())
			continue
		}
		if command.CommandLine == "" {
			log.Error("Invalid command file %s: Empty or missing command_line property", includeFile)
			continue
		}
		if command.Filepath == "" {
			log.Error("Invalid command file %s: Empty or missing file_path property", includeFile)
			continue
		}
		commands = append(commands, command)
	}

	return commands
}
