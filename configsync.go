package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"github.com/ecnepsnai/configsync/git"
	"github.com/ecnepsnai/logtic"
)

func startSync(config *configType) {
	if config.Verbose {
		logtic.Log.Level = logtic.LevelDebug
	} else {
		logtic.Log.Level = logtic.LevelWarn
	}
	logtic.Open()

	if config.Git.Author == "" {
		config.Git.Author = "configsync <configsync@" + getHostname() + ">"
	}

	log.Debug("git author: %s", config.Git.Author)

	workDir := config.Workdir
	makeDirectoryIfNotExists(workDir)

	log.Debug("working in directory: %s", workDir)

	git, err := git.New(config.Git.Path, config.Workdir)
	if err != nil {
		log.Fatal("error opening git instance: %s", err.Error())
	}
	if err := git.InitIfNeeded(); err != nil {
		log.Fatal("error initalizing git repo: %s", err.Error())
	}
	if err := git.Checkout(getHostname()); err != nil {
		log.Fatal("error checking out git branch: %s", err.Error())
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
			log.Info("Previously synced file '%s' no longer exists, removing", file.Path)
			git.Remove(syncPath)
		}
	}

	metadata.Files = []fileType{}

	filesToBackup := []string{}
	for _, pattern := range config.Files {
		if fileExists(pattern) {
			filesToBackup = append(filesToBackup, pattern)
			continue
		}

		files, err := filepath.Glob(pattern)
		if err != nil {
			log.Error("No files matched glob '%s'", pattern)
			continue
		}
		log.Info("Expanding glob '%s' to -> %v", pattern, files)
		filesToBackup = append(filesToBackup, files...)
	}

	for _, filePath := range filesToBackup {
		log.Info("Syncing file '%s'", filePath)
		var destHash uint64 = 0
		syncAtomicPath := path.Join(workDir, filePath+"_")
		syncPath := path.Join(workDir, filePath)
		if fileExists(syncPath) {
			destHash = hashFile(syncPath)
		}
		sourceHash := hashFile(filePath)

		if sourceHash == destHash {
			log.Info("No changes to already synced file '%s'", syncPath)
			metadata.Files = append(metadata.Files, fileType{
				Path: filePath,
				Hash: destHash,
			})
			continue
		}

		syncDir := pathWithoutFile(syncPath)
		makeDirectoryIfNotExists(syncDir)

		info, err := os.Stat(filePath)
		if err != nil {
			log.Error("Error stat-ing file: %s", err.Error())
			continue
		}

		source, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
		if err != nil {
			log.Error("Error opening source file: %s", err.Error())
			continue
		}
		dest, err := os.OpenFile(syncAtomicPath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Error("Error opening destination file: %s", err.Error())
			continue
		}

		wrote, err := io.CopyBuffer(dest, source, nil)
		dest.Close()
		if err != nil {
			log.Error("Error copying source file: %s", err.Error())
			continue
		}
		if wrote != info.Size() {
			log.Error("Did not copy entire source file")
			continue
		}

		if err := os.Rename(syncAtomicPath, syncPath); err != nil {
			log.Error("Error writing replacement file '%s': %s", syncPath, err.Error())
			continue
		}

		destHash = hashFile(syncPath)
		if sourceHash != destHash {
			log.Error("Source and destination hash do not match. %d != %d", sourceHash, destHash)
			continue
		}

		var UID int
		var GID int
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			UID = int(stat.Uid)
			GID = int(stat.Gid)
		}

		metadata.Files = append(metadata.Files, fileType{
			Path: filePath,
			Hash: destHash,
			Info: fileInfoType{
				Mode: uint32(info.Mode()),
				UID:  UID,
				GID:  GID,
			},
		})
		log.Info("Successfully synced file '%s'", filePath)
	}

	for _, command := range config.Commands {
		log.Info("Running command '%s' -> '%s'", command.CommandLine, command.Filepath)
		syncAtomicPath := path.Join(workDir, command.Filepath+"_")
		syncPath := path.Join(workDir, command.Filepath)
		syncDir := pathWithoutFile(syncPath)
		makeDirectoryIfNotExists(syncDir)

		cmd := exec.Command("/bin/bash", "-c", command.CommandLine)
		var buf bytes.Buffer
		cmd.Stdout = &buf
		if err := cmd.Run(); err != nil {
			log.Error("Error running command '%s': %s", command.CommandLine, err.Error())
			continue
		}

		dest, err := os.OpenFile(syncAtomicPath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Error("Error opening destination file: %s", err.Error())
			continue
		}

		_, err = io.CopyBuffer(dest, &buf, nil)
		dest.Close()
		if err != nil {
			os.Remove(syncAtomicPath)
			continue
		}

		if err := os.Rename(syncAtomicPath, syncPath); err != nil {
			log.Error("Error writing replacement file '%s': %s", syncPath, err.Error())
			continue
		}

		destHash := hashFile(syncPath)
		metadata.Files = append(metadata.Files, fileType{
			Path:        command.Filepath,
			Hash:        destHash,
			FromCommand: true,
		})
		log.Info("Successfully synced file '%s'", command.Filepath)
	}

	saveMetadata(metadataPath, metadata)

	if git.HasChanges() {
		git.Add(workDir)
		git.Commit("Automatic config sync", config.Git.Author)
		if config.Git.RemoteEnabled {
			git.Push(config.Git.RemoteName, getHostname())
		}
	}

	log.Info("Finished")
}
