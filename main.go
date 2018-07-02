package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type FileEntry struct {
	Path string `json:"path"`
	MD5  string `json:"md5"`
}

type DirEntry struct {
	Path  string       `json:"path"`
	Files []*FileEntry `json:"files"`
}

type Metadata struct {
	UpdatedAt    int64        `json:"updated_at"`
	SyncedDirs   []*DirEntry  `json:"synced_dirs"`
	SyncedFiles  []*FileEntry `json:"synced_files"`
	DefaultFiles []*FileEntry `json:"default_files"`
}

type Config struct {
	SyncedDirs   []string `json:"synced_dirs"`
	SyncedFiles  []string `json:"synced_files"`
	DefaultFiles []string `json:"default_files"`
}

var metadata Metadata
var config Config

func bullshit(err error) {
	// Fuck the shitty golang error handling
	if err != nil {
		panic(err)
	}
}

func hashFile(path string) string {
	f, err := os.Open(path)
	bullshit(err)
	defer f.Close()

	hash := md5.New()
	_, err = io.Copy(hash, f)
	bullshit(err)

	sum := hash.Sum(nil)[:16]
	return hex.EncodeToString(sum)
}

func scanDir(dirPath string) {
	dir := &DirEntry{
		Path: dirPath,
	}
	metadata.SyncedDirs = append(metadata.SyncedDirs, dir)

	filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		bullshit(err)

		if info.Name() == ".DS_Store" {
			return nil
		}

		file := &FileEntry{
			Path: filePath,
			MD5:  hashFile(filePath),
		}
		dir.Files = append(dir.Files, file)

		return nil
	})
}

func main() {
	data, err := ioutil.ReadFile("metadata_generator.json")
	bullshit(err)

	err = json.Unmarshal(data, &config)
	bullshit(err)

	for _, dirPath := range config.SyncedDirs {
		scanDir(dirPath)
	}

	for _, filePath := range config.SyncedFiles {
		file := &FileEntry{
			Path: filePath,
			MD5:  hashFile(filePath),
		}
		metadata.SyncedFiles = append(metadata.SyncedFiles, file)
	}

	for _, filePath := range config.DefaultFiles {
		file := &FileEntry{
			Path: filePath,
			MD5:  hashFile(filePath),
		}
		metadata.DefaultFiles = append(metadata.DefaultFiles, file)
	}

	metadata.UpdatedAt = time.Now().Unix()

	data, err = json.Marshal(metadata)
	bullshit(err)

	err = ioutil.WriteFile("metadata.json", data, 0644)
	bullshit(err)
}
