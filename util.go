package configsync

import (
	"io"
	"os"
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

func makeDirectoryIfNotExists(dirPath string) {
	if directoryExists(dirPath) {
		return
	}

	os.MkdirAll(dirPath, 0755)
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("error getting hostname: %s", err)
	}
	return hostname
}

func hashFile(filePath string) uint64 {
	w := xxhash.New()
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal("error opening file for hasing: %s", err)
	}
	defer f.Close()
	if _, err := io.CopyBuffer(w, f, nil); err != nil {
		log.Fatal("error hashing file: %s", err)
	}
	return w.Sum64()
}

func pathWithoutFile(filePath string) string {
	components := strings.Split(filePath, "/")
	return strings.Join(components[0:len(components)-1], "/")
}
