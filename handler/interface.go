package handler

type IConfigParser interface {
	ReadConfig() *Config
	ParseConfig(configFile string) (*Config, error)
}

type IPathResolver interface {
	ExpandPath(path string) (string, error)
	ReplaceHomeDir(fullpath string) (string, error)
}
