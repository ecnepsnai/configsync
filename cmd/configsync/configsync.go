package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
	if len(args) >= 2 {
		if args[1] == "-h" || args[1] == "--help" {
			printHelpAndExit()
		}
		if args[1] == "-v" || args[1] == "--version" {
			fmt.Printf("configsync v%s built on %s\n", Version, BuildDate)
			os.Exit(0)
		}
	}

	configPath := "configsync.conf"
	if len(args) == 2 {
		configPath = os.Args[1]
	}

	f, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Config file not found at path '%s'\n", configPath)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Unable to read config file at '%s': %s\n", configPath, err.Error())
		os.Exit(1)
	}
	defer f.Close()
	config := configSyncOptionsType{}
	if err := toml.NewDecoder(f).Decode(&config); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read config file at '%s': %s\n", configPath, err.Error())
		os.Exit(1)
	}
	config.ConfigFilePath = configPath

	if config.Workdir == "" {
		fmt.Fprintf(os.Stderr, "Invalid configuration: Workdir is required\n")
		os.Exit(1)
	}

	if len(config.commands()) == 0 && len(config.filePatterns()) == 0 {
		fmt.Fprintf(os.Stderr, "Invalid configuration: At least one file or command is required\n")
		os.Exit(1)
	}

	if config.Git.RemoteEnabled && config.Git.RemoteName == "" {
		fmt.Fprintf(os.Stderr, "Invalid configuration: Remote name is required if git remote is enabled\n")
		os.Exit(1)
	}

	if config.Git.Path == "" {
		gitPath, err := exec.LookPath("git")
		if err != nil {
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

type configSyncOptionsType struct {
	ConfInclude string                    `toml:"conf_include"`
	Workdir     string                    `toml:"workdir"`
	Git         configsync.GitOptionsType `toml:"git"`
	Verbose     bool                      `toml:"verbose"`

	// Populated at runtime with the absolute path to the original config file
	ConfigFilePath string `toml:"-"`
}

func (c configSyncOptionsType) includeDir() string {
	if filepath.IsAbs(c.ConfInclude) {
		return c.ConfInclude
	}
	return path.Join(filepath.Dir(c.ConfigFilePath), c.ConfInclude)
}

func (c configSyncOptionsType) includeFilesWithExtension(ext string) []string {
	incFiles := []string{}
	files, _ := os.ReadDir(c.includeDir())
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
		f, err := os.OpenFile(path.Join(c.includeDir(), includeFile), os.O_RDONLY, os.ModePerm)
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
		f, err := os.OpenFile(path.Join(c.includeDir(), includeFile), os.O_RDONLY, os.ModePerm)
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
		if command.ExePath == "" {
			log.Error("Invalid command file %s: Empty or missing exe_path property", includeFile)
			continue
		}
		if command.FilePath == "" {
			log.Error("Invalid command file %s: Empty or missing file_path property", includeFile)
			continue
		}
		commands = append(commands, command)
	}

	return commands
}
