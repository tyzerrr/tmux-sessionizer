package handler

import "context"

type IConfigParser interface {
	ReadConfig() *Config
	ParseConfig(configFile string) (*Config, error)
}

type IPathResolver interface {
	ExpandPath(path string) (string, error)
	ReplaceHomeDir(fullpath string) (string, error)
}

type ISessionHandler interface {
	NewSession(ctx context.Context) error
	GrabExistingSession(ctx context.Context) error
	CreateNewProjectSession(ctx context.Context) error
	DeleteProjectSession(ctx context.Context) error
}
