package configsync

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ecnepsnai/configsync/git"
	"github.com/ecnepsnai/logtic"
)

var log = logtic.Log.Connect("configsync")

const fileSourceCommand = "cmd"

type fileToBackupT struct {
	FilePath string
	Source   string
}

// Start beging the sync process
func Start(workDir string, filePatterns []string, commands []CommandType, gitOptions GitOptionsType) {
	start := time.Now()

	if gitOptions.Author == "" {
		gitOptions.Author = "configsync <configsync@" + getHostname() + ">"
	}
	if gitOptions.BranchName == "" {
		gitOptions.BranchName = getHostname()
	}

	log.Debug("Work directory: %s", workDir)
	log.Debug("File patterns: %v", filePatterns)
	log.Debug("Commands: %+v", commands)
	log.Debug("Git options: %+v", gitOptions)

	if err := makeDirectoryIfNotExists(workDir); err != nil {
		log.PFatal("Error making work directory", map[string]interface{}{
			"path":  workDir,
			"error": err.Error(),
		})
	}

	git, err := git.New(gitOptions.Path, workDir)
	if err != nil {
		log.Fatal("error opening git instance: %s", err.Error())
	}
	if err := git.InitIfNeeded(); err != nil {
		log.Fatal("error initalizing git repo: %s", err.Error())
	}
	if git.HasChanges() {
		log.Warn("working directory is dirty (has unstaged or untracked files)!")
	}
	if err := git.Checkout(gitOptions.BranchName); err != nil {
		log.Fatal("error checking out git branch: %s", err.Error())
	}
	if gitOptions.RemoteEnabled {
		git.Pull()
	}

	metadataPath := path.Join(workDir, "configsync_meta.json")
	metadata := tryLoadMeta(metadataPath)

	commandFileMap := map[string]bool{}
	for _, command := range commands {
		commandFileMap[command.FilePath] = true
	}
	fileMap := map[string]bool{}
	for _, pattern := range filePatterns {
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
	for _, pattern := range filePatterns {
		if fileExists(pattern) {
			filesToBackup = append(filesToBackup, fileToBackupT{
				FilePath: pattern,
				Source:   pattern,
			})
			continue
		}

		paths, err := filepath.Glob(pattern)
		if err != nil {
			log.Error("Invalid glob pattern '%s'", pattern)
			continue
		}
		if len(paths) == 0 {
			log.Warn("No files matched glob '%s'", pattern)
			continue
		}
		log.Info("Expanding glob '%s' to -> %v", pattern, paths)
		for _, globPath := range paths {
			info, err := os.Stat(globPath)
			if err != nil {
				log.PError("Error querying path from glob", map[string]interface{}{
					"path":  globPath,
					"error": err.Error(),
					"glob":  pattern,
				})
				continue
			}
			if info.IsDir() {
				files, err := listAllFilesInDirectory(globPath)
				if err != nil {
					log.PError("Error listing files in directory", map[string]interface{}{
						"path":  globPath,
						"error": err.Error(),
					})
					continue
				}
				log.Info("Expanding directory '%s' to -> %v", globPath, files)
				for _, file := range files {
					filesToBackup = append(filesToBackup, fileToBackupT{
						FilePath: file,
						Source:   pattern,
					})
				}
			} else {
				filesToBackup = append(filesToBackup, fileToBackupT{
					FilePath: globPath,
					Source:   pattern,
				})
			}
		}
	}

	for _, fileToBackup := range filesToBackup {
		log.Info("Syncing file '%s'", fileToBackup.FilePath)
		var destHash uint64 = 0
		syncAtomicPath := path.Join(workDir, fileToBackup.FilePath+"_")
		syncPath := path.Join(workDir, fileToBackup.FilePath)
		if fileExists(syncPath) {
			destHash, err = hashFile(syncPath)
			if err != nil {
				continue
			}
		}
		sourceHash, err := hashFile(fileToBackup.FilePath)
		if err != nil {
			continue
		}

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
		if err := makeDirectoryIfNotExists(syncDir); err != nil {
			log.PError("Error making sync directory", map[string]interface{}{
				"path":  syncDir,
				"error": err.Error(),
			})
			continue
		}

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

		destHash, err = hashFile(syncPath)
		if err != nil {
			continue
		}
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

	for _, command := range commands {
		log.Info("Running command '%s %s' -> '%s'", command.ExePath, command.Arguments, command.FilePath)
		syncAtomicPath := path.Join(workDir, command.FilePath+"_")
		syncPath := path.Join(workDir, command.FilePath)
		syncDir := pathWithoutFile(syncPath)
		if err := makeDirectoryIfNotExists(syncDir); err != nil {
			log.PError("Error making sync directory", map[string]interface{}{
				"path":  syncDir,
				"error": err.Error(),
			})
			continue
		}

		cmd := exec.Command(command.ExePath, command.Arguments...)
		if command.WorkDir != "" {
			cmd.Dir = command.WorkDir
			log.Debug("Setting command workdir: %s", command.WorkDir)
		}
		if len(command.Env) > 0 {
			cmd.Env = command.Env
			log.Debug("Setting command environment variables: %s", command.Env)
		}
		if command.User > 0 && command.Group > 0 {
			cmd.SysProcAttr.Credential = &syscall.Credential{
				Uid: command.User,
				Gid: command.Group,
			}
			log.Debug("Setting command UID and GID: %d, %d", command.User, command.Group)
		}
		var buf bytes.Buffer
		cmd.Stdout = &buf
		if err := cmd.Run(); err != nil {
			log.Error("Error running command '%s %s': %s", command.ExePath, command.Arguments, err.Error())
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

		destHash, err := hashFile(syncPath)
		if err != nil {
			continue
		}

		file := fileType{
			Path:   command.FilePath,
			Hash:   destHash,
			Source: fileSourceCommand,
			Info: fileInfoType{
				Mode: uint32(os.ModePerm),
			},
		}
		if command.User > 0 && command.Group > 0 {
			file.Info.UID = int(command.User)
			file.Info.GID = int(command.Group)
		}

		metadata.Files = append(metadata.Files, file)
		log.Info("Successfully synced file '%s'", command.FilePath)
	}

	saveMetadata(metadataPath, metadata)

	if git.HasChanges() {
		git.Add(workDir)
		git.Commit("Automatic config sync", gitOptions.Author)
		if gitOptions.RemoteEnabled {
			git.Push(gitOptions.RemoteName, gitOptions.BranchName)
		}
	}

	finished := time.Since(start)
	log.Info("Finished in %s", finished)
}
