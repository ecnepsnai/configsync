package configsync

import (
	"os"
	"path"
	"testing"
)

func TestPathWithoutFile(t *testing.T) {
	filePath := "/etc/passwd"
	expected := "/etc"
	result := pathWithoutFile(filePath)
	if result != expected {
		t.Errorf("Unexpected result. Expected '%s' got '%s'", expected, result)
	}
}

func TestListAllFilesInDirectory(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(path.Join(dir, "1", "2", "3", "4"), 0755)
	os.WriteFile(path.Join(dir, "file.txt"), []byte("hello"), 0644)
	os.WriteFile(path.Join(dir, "1", "file.txt"), []byte("hello"), 0644)
	os.WriteFile(path.Join(dir, "1", "2", "file.txt"), []byte("hello"), 0644)
	os.WriteFile(path.Join(dir, "1", "2", "3", "file.txt"), []byte("hello"), 0644)
	os.WriteFile(path.Join(dir, "1", "2", "3", "4", "file.txt"), []byte("hello"), 0644)
	os.WriteFile(path.Join(dir, "1", "2", "3", "4", "5", "file.txt"), []byte("hello"), 0644)

	files, err := listAllFilesInDirectory(dir)
	if err != nil {
		panic(err)
	}

	if len(files) != 5 {
		t.Errorf("Incorrect number of files returned. Expected 5 got %d", len(files))
	}
}
