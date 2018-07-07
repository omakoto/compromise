package comptest

import (
	"github.com/ungerik/go-dry"
	"io/ioutil"
	"path/filepath"
)

type FileEntry struct {
	Name  string
	IsDir bool
}

var injectedFiles []string

func IsEmptyDir(path string) bool {
	files, _ := ioutil.ReadDir(path + "/")
	return len(files) == 0
}

func ListFiles(path string) ([]FileEntry, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	ret := make([]FileEntry, 0, len(files))
	for _, file := range files {
		baseName := file.Name()
		relPath := filepath.Join(path, baseName)
		isDir := dry.FileIsDir(relPath)

		ret = append(ret, FileEntry{baseName, isDir})
	}
	return ret, nil
}
