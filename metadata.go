package main

import (
	"encoding/json"
	"os"
)

type metadataType struct {
	Files []fileType
}

type fileType struct {
	Path        string
	Hash        uint64
	Info        fileInfoType
	FromCommand bool
}

type fileInfoType struct {
	Mode uint32
	UID  int
	GID  int
}

func tryLoadMeta(metaPath string) *metadataType {
	if !fileExists(metaPath) {
		return &metadataType{}
	}

	f, err := os.OpenFile(metaPath, os.O_RDONLY, 0644)
	if err != nil {
		return &metadataType{}
	}
	defer f.Close()
	metadata := metadataType{}
	if err := json.NewDecoder(f).Decode(&metadata); err != nil {
		return &metadataType{}
	}

	return &metadata
}

func saveMetadata(metaPath string, metadata *metadataType) {
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
