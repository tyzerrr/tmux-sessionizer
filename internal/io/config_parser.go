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
	// Registered holds the normalized root directories listed in the config
	// file, while Projects holds their immediate subdirectories. Duplicate
	// registration must be checked against Registered, not Projects.
	Registered []types.String
	Projects   []types.String
}

func newConfig() *Config {
	return &Config{
		Registered: []types.String{},
		Projects:   []types.String{},
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

		// Record the entry even when its directory is gone: it still lives in
		// the config file, so re-registering it would duplicate the line.
		config.Registered = append(config.Registered, types.NewString(absPath))

		if err := filer.Exists(absPath); err != nil {
			// NOTE: registered directory might be deleted, we need to skip in this case.
			continue
		}

		entries, err := os.ReadDir(absPath)
		if err != nil {
			return nil, err
		}

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(absPath, e.Name())
			if err := filer.Exists(path); err != nil {
				return nil, err
			}
			c.createProjects(config, path)
		}

	}

	return config, nil
}

func (c *ConfigParser) createProjects(config *Config, path string) {
	config.Projects = append(config.Projects, types.NewString(path))
}
