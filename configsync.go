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

const fileSourceCommand = "cmd"

type fileToBackupT struct {
	FilePath string
	Source   string
}

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

	filesToRemove := []string{}
	for _, file := range metadata.Files {
		syncPath := path.Join(workDir, file.Path)
		shouldRemove := false
		if file.Source == fileSourceCommand {
			if !commandFileMap[file.Path] {
				log.Warn("Will remove command output '%s' ('%s') because it was removed from the config", file.Path, syncPath)
				shouldRemove = true
			}
		} else {
			if !fileMap[file.Source] {
				log.Warn("Will remove file '%s' ('%s') because it was removed from the config", file.Path, syncPath)
				shouldRemove = true
			}
			if !fileExists(file.Path) {
				log.Warn("Will remove file '%s' ('%s') because the source no longer exists", file.Path, syncPath)
				shouldRemove = true
			}
		}
		if shouldRemove {
			filesToRemove = append(filesToRemove, syncPath)
		}
	}
	if len(filesToRemove) > 0 {
		git.Remove(filesToRemove...)
	}

	metadata.Files = []fileType{}

	filesToBackup := []fileToBackupT{}
	for _, pattern := range config.Files {
		if fileExists(pattern) {
			filesToBackup = append(filesToBackup, fileToBackupT{
				FilePath: pattern,
				Source:   pattern,
			})
			continue
		}

		files, err := filepath.Glob(pattern)
		if err != nil {
			log.Error("No files matched glob '%s'", pattern)
			continue
		}
		log.Info("Expanding glob '%s' to -> %v", pattern, files)
		for _, file := range files {
			filesToBackup = append(filesToBackup, fileToBackupT{
				FilePath: file,
				Source:   pattern,
			})
		}
	}

	for _, fileToBackup := range filesToBackup {
		log.Info("Syncing file '%s'", fileToBackup.FilePath)
		var destHash uint64 = 0
		syncAtomicPath := path.Join(workDir, fileToBackup.FilePath+"_")
		syncPath := path.Join(workDir, fileToBackup.FilePath)
		if fileExists(syncPath) {
			destHash = hashFile(syncPath)
		}
		sourceHash := hashFile(fileToBackup.FilePath)

		info, err := os.Stat(fileToBackup.FilePath)
		if err != nil {
			log.Error("Error stat-ing file: %s", err.Error())
			continue
		}

		var UID int
		var GID int
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			UID = int(stat.Uid)
			GID = int(stat.Gid)
		}

		if sourceHash == destHash {
			log.Info("No changes to already synced file '%s'", syncPath)
			metadata.Files = append(metadata.Files, fileType{
				Path: fileToBackup.FilePath,
				Hash: destHash,
				Info: fileInfoType{
					Mode: uint32(info.Mode()),
					UID:  UID,
					GID:  GID,
				},
				Source: fileToBackup.Source,
			})
			continue
		}

		syncDir := pathWithoutFile(syncPath)
		makeDirectoryIfNotExists(syncDir)

		source, err := os.OpenFile(fileToBackup.FilePath, os.O_RDONLY, 0644)
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

		metadata.Files = append(metadata.Files, fileType{
			Path: fileToBackup.FilePath,
			Hash: destHash,
			Info: fileInfoType{
				Mode: uint32(info.Mode()),
				UID:  UID,
				GID:  GID,
			},
			Source: fileToBackup.Source,
		})
		log.Info("Successfully synced file '%s'", fileToBackup.FilePath)
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
			Path:   command.Filepath,
			Hash:   destHash,
			Source: fileSourceCommand,
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
