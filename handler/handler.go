package handler

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SessionHandler struct {
	config config
}

type config struct {
	projects []project
}

type project struct {
	name     string
	filepath string
}

func NewSessionHandler() *SessionHandler {
	return &SessionHandler{}
}

func newConfig() *config {
	return &config{
		projects: []project{},
	}
}

func (sh *SessionHandler) Start() error {
	if err := sh.newSessionFromProjects(); err != nil {
		return err
	}
	return nil
}

func (sh *SessionHandler) newTmuxCmd(name string, args ...string) *exec.Cmd {
	tmuxCmd := exec.Command(name, args...)
	tmuxCmd.Stdin = os.Stdin
	tmuxCmd.Stdout = os.Stdout
	tmuxCmd.Stderr = os.Stderr
	return tmuxCmd
}

func (sh *SessionHandler) readConfig() *config {
	configFiles := []string{"./.tmux-sessionizer", "~/.tmux-sessionizer"}
	for _, cf := range configFiles {
		config, err := sh.parseConfig(cf)
		if err == nil {
			return config
		}
	}
	return nil
}

func (sh *SessionHandler) parseConfig(configFile string) (*config, error) {
	file, err := os.Open(configFile)
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
					path, err := sh.expandPath(trimmed)
					if err != nil {
						return nil, fmt.Errorf("failed to expand path: %w", err)
					}
					dirs, err := os.ReadDir(path)
					if err != nil {
						return nil, fmt.Errorf("failed to get all directories: %w", err)
					}
					for _, dir := range dirs {
						if dir.IsDir() {
							config.projects = append(config.projects,
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

func (sh *SessionHandler) expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
	}
	return path, nil
}

func (sh *SessionHandler) newSessionFromProjects() error {
	config := sh.readConfig()
	if config == nil {
		return fmt.Errorf("failed to create new project from projects")
	}

	/*build fzf input*/
	var input bytes.Buffer
	projectHashMap := make(map[string]string, 0)
	for _, project := range config.projects {
		input.WriteString(project.name + "\n")
		projectHashMap[project.name] = project.filepath
	}
	fzfOut, err := sh.toFzf(input)
	if err != nil {
		return fmt.Errorf("failed to create new project from projects: %w", err)
	}
	if err := sh.newTmuxCmd("tmux", "new-session", "-s", strings.Trim(fzfOut.String(), "\n"), "-c", projectHashMap[strings.Trim(fzfOut.String(), "\n")]).Run(); err != nil {
		return fmt.Errorf("failed to create new session %w: ", err)
	}
	return nil
}

func (sh *SessionHandler) toFzf(input bytes.Buffer) (bytes.Buffer, error) {
	fzfCmd := exec.Command("fzf")
	fzfCmd.Stdin = &input
	var fzfOut bytes.Buffer
	fzfCmd.Stdout = &fzfOut
	if err := fzfCmd.Run(); err != nil {
		return bytes.Buffer{}, err
	}
	return fzfOut, nil
}

func (sh *SessionHandler) switchTo() error {
	tmuxCmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	var tmuxOut bytes.Buffer
	tmuxCmd.Stdout = &tmuxOut
	if err := tmuxCmd.Run(); err != nil {
		return fmt.Errorf("failed to execute tmux list-sessions command: %w", err)
	}
	fzfOut, err := sh.toFzf(tmuxOut)
	if err != nil {
		return fmt.Errorf("failed to execute fzf command: %w", err)
	}
	selected := strings.TrimSpace(fzfOut.String())
	/*
	   When tmux try to attach, real tty is necessary.
	   But as default, exec.Command provides virtual in-memory pipe.
	   So tmux throws the error.
	*/
	cmd := sh.newTmuxCmd("tmux", "attach", "-t", selected)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute tmux attach -t: %w", err)
	}
	return nil
}
