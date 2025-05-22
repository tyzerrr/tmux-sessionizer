package handler

import (
	"bytes"
	"errors"
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

func (p *PathResolver) BuildProjectInfo(config *Config) (map[string]string, map[string]string, error) {
	var input bytes.Buffer
	projectNameFullPathMap, projectExpressionNameMap := make(map[string]string, 0), make(map[string]string, 0)
	for _, project := range config.Projects {
		replaced, err := p.ExpandPath(project.filepath)
		if err != nil {
			return nil, nil, errors.New("failed to replace home directory to ~")
		}
		input.WriteString(replaced + "\n")
		projectNameFullPathMap[project.name] = project.filepath
		projectExpressionNameMap[replaced] = project.name
	}
	return projectNameFullPathMap, projectExpressionNameMap, nil
}
