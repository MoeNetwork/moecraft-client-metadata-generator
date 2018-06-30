package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type FileEntry struct {
	Path string `json:"path"`
	MD5  string `json:"md5"`
}

type DirEntry struct {
	Path  string       `json:"path"`
	Files []*FileEntry `json:"files"`
}

var metadata struct {
	SyncedDirs  []*DirEntry  `json:"synced_dirs"`
	SyncedFiles []*FileEntry `json:"synced_files"`
}

var config struct {
	SyncedDirs  []string `json:"synced_dirs"`
	SyncedFiles []string `json:"synced_files"`
}

var lock sync.Mutex

// Limit concurrency to 5
var sem = make(chan bool, 5)

// After the first run, I realized that md5 is fast enough, where parallel computing is not necessary
func parallel(f func()) {
	sem <- true
	go func() {
		f()
		<-sem
	}()
}

func wait() {
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}

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
	dir := DirEntry{
		Path: dirPath,
	}
	metadata.SyncedDirs = append(metadata.SyncedDirs, &dir)

	filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		bullshit(err)

		parallel(func() {
			file := FileEntry{
				Path: filePath,
				MD5:  hashFile(filePath),
			}
			lock.Lock()
			dir.Files = append(dir.Files, &file)
			lock.Unlock()
		})

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
		parallel(func() {
			file := FileEntry{
				Path: filePath,
				MD5:  hashFile(filePath),
			}
			lock.Lock()
			metadata.SyncedFiles = append(metadata.SyncedFiles, &file)
			lock.Unlock()
		})
	}

	wait()

	data, err = json.Marshal(metadata)
	bullshit(err)

	err = ioutil.WriteFile("metadata.json", data, 0755)
	bullshit(err)
}
