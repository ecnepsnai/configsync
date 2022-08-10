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
		log.PError("Error opening metadata", map[string]interface{}{
			"file_path": metaPath,
			"error":     err.Error(),
		})
		return &metadataType{}
	}
	defer f.Close()
	metadata := metadataType{}
	if err := json.NewDecoder(f).Decode(&metadata); err != nil {
		log.PError("Error decoding metadata", map[string]interface{}{
			"file_path": metaPath,
			"error":     err.Error(),
		})
		return &metadataType{}
	}

	log.Debug("Metadata loaded: %+v", metadata)
	return &metadata
}

func saveMetadata(metaPath string, metadata *metadataType) {
	syncPath := metaPath + ".atomic"
	f, err := os.OpenFile(syncPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.PFatal("Error opening atomic path for metadata", map[string]interface{}{
			"atomic_path": syncPath,
			"file_path":   metaPath,
			"error":       err.Error(),
		})
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(metadata); err != nil {
		f.Close()
		log.PFatal("Error encoding metadata JSON", map[string]interface{}{
			"atomic_path": syncPath,
			"file_path":   metaPath,
			"error":       err.Error(),
		})
	}
	f.Close()

	if err := os.Rename(syncPath, metaPath); err != nil {
		log.PFatal("Error encoding metadata JSON", map[string]interface{}{
			"atomic_path": syncPath,
			"file_path":   metaPath,
			"error":       err.Error(),
		})
	}
	log.Debug("Synced metadata")
}
