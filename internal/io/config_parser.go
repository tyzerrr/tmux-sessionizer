package io

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/types"
)

const (
	configPrefix = "default="
)

type Config struct {
	Projects []types.String
}

func newConfig() *Config {
	return &Config{
		Projects: []types.String{},
	}
}

type ConfigParser struct{}

func NewConfigParser() *ConfigParser {
	return &ConfigParser{}
}

func (c *ConfigParser) ReadConfig(configFile string) (*Config, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	projectList := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, configPrefix) {
			continue
		}

		seq := strings.Split(strings.TrimPrefix(line, configPrefix), ",")
		projectList = append(projectList, seq...)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return c.parse(projectList)
}

func (c *ConfigParser) parse(projectList []string) (*Config, error) {
	config, filepathResolver := newConfig(), NewFilePathResolver()

	for _, p := range projectList {
		tp := strings.TrimSpace(p)
		if len(tp) == 0 {
			continue
		}

		tp, err := filepathResolver.ExpandTildeAsHomeDir(tp)
		if err != nil {
			return nil, err
		}

		absPath, err := filepath.Abs(tp)
		if err != nil {
			return nil, err
		}

		files, err := os.ReadDir(absPath)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if !f.IsDir() {
				continue
			}

			config.Projects = append(config.Projects, types.NewString(filepath.Join(absPath, f.Name())))
		}
	}

	return config, nil
}
