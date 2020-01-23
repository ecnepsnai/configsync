package configsync

import (
	"encoding/json"
	"os"
)

// Metadata describes a metadata object
type Metadata struct {
	Files []File
}

// File describes metadata about a file
type File struct {
	Path        string
	Hash        uint64
	FromCommand bool
}

func tryLoadMeta(metaPath string) *Metadata {
	if !fileExists(metaPath) {
		return &Metadata{}
	}

	f, err := os.OpenFile(metaPath, os.O_RDONLY, 0644)
	if err != nil {
		return &Metadata{}
	}
	defer f.Close()
	metadata := Metadata{}
	if err := json.NewDecoder(f).Decode(&metadata); err != nil {
		return &Metadata{}
	}

	return &metadata
}

func saveMetadata(metaPath string, metadata *Metadata) {
	os.Remove(metaPath)
	f, err := os.OpenFile(metaPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(metadata); err != nil {
		panic(err)
	}
}
