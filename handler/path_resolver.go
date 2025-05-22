package handler

import (
	"os"
	"path/filepath"
	"strings"
)

type PathResolver struct{}

func NewPathResolver() *PathResolver {
	return &PathResolver{}
}

func (p *PathResolver) ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
	}
	return path, nil
}

func (p *PathResolver) ReplaceHomeDir(fullpath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return strings.Replace(fullpath, home, "~", 1), nil
}
