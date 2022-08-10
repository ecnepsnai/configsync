package configsync

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cespare/xxhash/v2"
)

func directoryExists(dirPath string) bool {
	stat, err := os.Stat(dirPath)
	return err == nil && stat.IsDir()
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func makeDirectoryIfNotExists(dirPath string) error {
	if directoryExists(dirPath) {
		return nil
	}

	return os.MkdirAll(dirPath, 0755)
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}

func hashFile(filePath string) (uint64, error) {
	w := xxhash.New()
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		log.PError("Error opening file for hasing", map[string]interface{}{
			"path":  filePath,
			"error": err.Error(),
		})
		return 0, err
	}
	defer f.Close()
	if _, err := io.CopyBuffer(w, f, nil); err != nil {
		log.PError("Error hasing file", map[string]interface{}{
			"path":  filePath,
			"error": err.Error(),
		})
		return 0, err
	}
	return w.Sum64(), nil
}

func pathWithoutFile(filePath string) string {
	components := strings.Split(filePath, "/")
	return strings.Join(components[0:len(components)-1], "/")
}

func listAllFilesInDirectory(dir string) ([]string, error) {
	paths := []string{}

	err := filepath.WalkDir(dir, func(pathName string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			paths = append(paths, pathName)
		}
		return nil
	})
	return paths, err
}
