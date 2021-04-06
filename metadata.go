package configsync

import (
	"encoding/json"
	"os"
)

type metadataType struct {
	Files []fileType
}

type fileType struct {
	Path   string
	Hash   uint64
	Info   fileInfoType
	Source string
}

type fileInfoType struct {
	Mode uint32
	UID  int
	GID  int
}

func tryLoadMeta(metaPath string) *metadataType {
	log.Debug("Trying to read metadata...")

	if !fileExists(metaPath) {
		log.Debug("Metadata does not exist")
		return &metadataType{}
	}

	f, err := os.OpenFile(metaPath, os.O_RDONLY, 0644)
	if err != nil {
		log.Error("Error opening metadata: file_path='%s' error='%s'", metaPath, err.Error())
		return &metadataType{}
	}
	defer f.Close()
	metadata := metadataType{}
	if err := json.NewDecoder(f).Decode(&metadata); err != nil {
		log.Error("Error decoding metadata: file_path='%s' error='%s'", metaPath, err.Error())
		return &metadataType{}
	}

	log.Debug("Metadata loaded: %+v", metadata)
	return &metadata
}

func saveMetadata(metaPath string, metadata *metadataType) {
	syncPath := metaPath + ".atomic"
	f, err := os.OpenFile(syncPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("Error opening atomic path for metadata: atomic_path='%s' file_path='%s' error='%s'", syncPath, metaPath, err.Error())
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(metadata); err != nil {
		f.Close()
		log.Fatal("Error encoding metadata JSON: atomic_path='%s' file_path='%s' error='%s'", syncPath, metaPath, err.Error())
	}
	f.Close()

	if err := os.Rename(syncPath, metaPath); err != nil {
		log.Fatal("Error overwriting metadata: atomic_path='%s' file_path='%s' error='%s'", syncPath, metaPath, err.Error())
	}
	log.Debug("Synced metadata")
}
