package configsync

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/ecnepsnai/configsync/git"
)

// Sync sync all config files
func Sync(config *Config) {
	if config.Git.Author == "" {
		config.Git.Author = "configsync <configsync@" + getHostname() + ">"
	}

	workDir := config.Workdir
	makeDirectoryIfNotExists(workDir)

	git, err := git.New(config.Git.Path, config.Workdir)
	if err != nil {
		panic(err)
	}
	if err := git.InitIfNeeded(); err != nil {
		panic(err)
	}
	if err := git.Checkout(getHostname()); err != nil {
		panic(err)
	}
	if config.Git.RemoteEnabled {
		git.Pull()
	}

	metadataPath := path.Join(workDir, "configsync_meta.json")
	metadata := tryLoadMeta(metadataPath)

	commandFileMap := map[string]bool{}
	for _, command := range config.Commands {
		commandFileMap[command.Filepath] = true
	}
	fileMap := map[string]bool{}
	for _, pattern := range config.Files {
		fileMap[pattern] = true
	}

	for _, file := range metadata.Files {
		syncPath := path.Join(workDir, file.Path)
		shouldRemove := false
		if file.FromCommand {
			if !commandFileMap[file.Path] {
				shouldRemove = true
			}
		} else {
			if !fileMap[file.Path] {
				shouldRemove = true
			}
			if !fileExists(file.Path) {
				shouldRemove = true
			}
		}
		if shouldRemove {
			stdout("Previously synced file '%s' no longer exists, removing", file.Path)
			git.Remove(syncPath)
		}
	}

	metadata.Files = []File{}

	filesToBackup := []string{}
	for _, pattern := range config.Files {
		if fileExists(pattern) {
			filesToBackup = append(filesToBackup, pattern)
			continue
		}

		files, err := filepath.Glob(pattern)
		if err != nil {
			stderr("No files matched glob '%s'", pattern)
			continue
		}
		stdout("Expanding glob '%s' to -> %v", pattern, files)
		filesToBackup = append(filesToBackup, files...)
	}

	for _, filePath := range filesToBackup {
		stdout("Syncing file '%s'", filePath)
		var destHash uint64 = 0
		syncPath := path.Join(workDir, filePath)
		if fileExists(syncPath) {
			destHash = hashFile(syncPath)
		}
		sourceHash := hashFile(filePath)

		if sourceHash == destHash {
			stdout("No changes to already synced file '%s'", syncPath)
			metadata.Files = append(metadata.Files, File{
				Path: filePath,
				Hash: destHash,
			})
			continue
		}

		syncDir := pathWithoutFile(syncPath)
		makeDirectoryIfNotExists(syncDir)

		info, err := os.Stat(filePath)
		if err != nil {
			// Error
			stderr("Error stat-ing file: %s", err.Error())
			continue
		}

		source, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
		if err != nil {
			// Error
			stderr("Error opening source file: %s", err.Error())
			continue
		}
		dest, err := os.OpenFile(syncPath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			// Error
			stderr("Error opening destination file: %s", err.Error())
			continue
		}

		wrote, err := io.CopyBuffer(dest, source, nil)
		if err != nil {
			// Error
			stderr("Error copying source file: %s", err.Error())
			continue
		}
		if wrote != info.Size() {
			// Error
			stderr("Did not copy entire source file")
			continue
		}

		destHash = hashFile(syncPath)
		if sourceHash != destHash {
			// Error
			stderr("Source and destination hash do not match. %d != %d", sourceHash, destHash)
			continue
		}

		metadata.Files = append(metadata.Files, File{
			Path: filePath,
			Hash: destHash,
		})
		stdout("Successfully synced file '%s'", filePath)
	}

	for _, command := range config.Commands {
		stdout("Running command '%s' -> '%s'", command.CommandLine, command.Filepath)
		syncPath := path.Join(workDir, command.Filepath)
		syncDir := pathWithoutFile(syncPath)
		makeDirectoryIfNotExists(syncDir)

		cmd := exec.Command("/bin/bash", "-c", command.CommandLine)
		var buf bytes.Buffer
		cmd.Stdout = &buf
		if err := cmd.Run(); err != nil {
			stderr("Error running command '%s': %s", command.CommandLine, err.Error())
			continue
		}

		dest, err := os.OpenFile(syncPath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			// Error
			stderr("Error opening destination file: %s", err.Error())
			continue
		}

		_, err = io.CopyBuffer(dest, &buf, nil)
		if err != nil {
			// Error
			continue
		}

		destHash := hashFile(syncPath)
		metadata.Files = append(metadata.Files, File{
			Path:        command.Filepath,
			Hash:        destHash,
			FromCommand: true,
		})
		stdout("Successfully synced file '%s'", command.Filepath)
	}

	saveMetadata(metadataPath, metadata)

	if git.HasChanges() {
		git.Add(workDir)
		git.Commit("Automatic config sync", config.Git.Author)
		if config.Git.RemoteEnabled {
			git.Push(config.Git.RemoteName, getHostname())
		}
	}
}
