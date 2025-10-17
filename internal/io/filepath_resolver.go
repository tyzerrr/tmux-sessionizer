package io

import (
	"os"
	"path/filepath"
	"strings"
)

type FilePathResolver struct{}

func NewFilePathResolver() *FilePathResolver {
	return &FilePathResolver{}
}

func (fp *FilePathResolver) ExpandTildeAsHomeDir(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	trimmed := strings.TrimPrefix(path, "~")
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHome, trimmed), nil
}
