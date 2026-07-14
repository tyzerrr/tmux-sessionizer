package io

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/types"
)

const (
	ConfigPrefix = "default="
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

func (c *ConfigParser) ReadConfig(filer *Filer, configFileAbs string) (*Config, error) {
	f, err := os.Open(configFileAbs)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	projectList := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, ConfigPrefix) {
			continue
		}

		seq := strings.Split(strings.TrimPrefix(line, ConfigPrefix), ",")
		projectList = append(projectList, seq...)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return c.parse(projectList, filer)
}

func (c *ConfigParser) parse(projectList []string, filer *Filer) (*Config, error) {
	config := newConfig()

	for _, p := range projectList {
		tp := strings.TrimSpace(p)
		if len(tp) == 0 {
			continue
		}

		tp, err := filer.ExpandTildeAsHomeDir(tp)
		if err != nil {
			return nil, err
		}

		absPath, err := filepath.Abs(tp)
		if err != nil {
			return nil, err
		}

		if err := filer.Exists(absPath); err != nil {
			// NOTE: registered directory might be deleted, we need to skip in this case.
			continue
		}
		c.createProjects(config, absPath)
	}

	return config, nil
}

func (c *ConfigParser) createProjects(config *Config, path string) {
	config.Projects = append(config.Projects, types.NewString(path))
}
