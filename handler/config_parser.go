package handler

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Projects []project
}

func newConfig() *Config {
	return &Config{
		Projects: []project{},
	}
}

type project struct {
	name     string
	filepath string
}

type ConfigParser struct {
	PathResolver *PathResolver
}

func NewConfigParser(pr *PathResolver) *ConfigParser {
	return &ConfigParser{
		PathResolver: pr,
	}
}

func (c *ConfigParser) ReadConfig() (*Config, error) {
	var configFiles = []string{"./.tmux-sessionizer", "~/.tmux-sessionizer"}
	for _, cf := range configFiles {
		config, err := c.ParseConfig(cf)
		if err == nil {
			return config, nil
		} else {
			return nil, err
		}
	}
	return nil, errors.New("failed to read both config files, ${pwd}/.tmux-sessionizer, ${HOME}/.tmux-sessionizer")
}

func (c *ConfigParser) ParseConfig(configFile string) (*Config, error) {
	path, err := c.PathResolver.ExpandPath(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path: %w", err)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	config := newConfig()

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "default=") {
			raw := strings.TrimPrefix(line, "default=")
			projects := strings.Split(raw, ",")
			for _, pr := range projects {
				trimmed := strings.TrimSpace(pr)
				if trimmed != "" {
					path, err := c.PathResolver.ExpandPath(trimmed)
					if err != nil {
						return nil, fmt.Errorf("failed to expand path: %w", err)
					}
					dirs, err := os.ReadDir(path)
					if err != nil {
						return nil, fmt.Errorf("failed to get all directories: %w", err)
					}
					for _, dir := range dirs {
						if dir.IsDir() {
							config.Projects = append(config.Projects,
								project{name: dir.Name(), filepath: filepath.Join(path, dir.Name())},
							)
						}
					}
				}
			}
		}
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("failed to read .tmux-sessionizer: %w", err)
	}
	return config, nil
}
