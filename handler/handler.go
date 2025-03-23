package handler

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ISessionHandler interface {
	NewSession() error
	GrabExistingSession() error
}

type SessionHandler struct{}

type config struct {
	projects []project
}

type project struct {
	name     string
	filepath string
}

func NewSessionHandler() ISessionHandler {
	return &SessionHandler{}
}

func newConfig() *config {
	return &config{
		projects: []project{},
	}
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
	path, err := sh.expandPath(configFile)
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

func (sh *SessionHandler) NewSession() error {
	config := sh.readConfig()
	if config == nil {
		return errors.New("failed to create new project from projects")
	}
	/*build fzf input and build hashmap to retrieve filepath from entry's name.*/
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
	sessionName := strings.Trim(fzfOut.String(), "\n")
	if sh.isInSession() {
		return sh.switchToNewClient(sessionName, projectHashMap[sessionName])
	}
	if sh.sessionExists(sessionName) {
		return sh.attach(sessionName)
	}
	if err := sh.newTmuxCmd(
		"tmux", "new-session", "-s",
		sessionName, "-c", projectHashMap[sessionName]).Run(); err != nil {
		return fmt.Errorf("failed to create new session %w: ", err)
	}
	return nil
}

func (sh *SessionHandler) sessionExists(sessionName string) bool {
	cmd := sh.newTmuxCmd("tmux", "has-session", "-t", sessionName)
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

func (sh *SessionHandler) attach(sessionName string) error {
	/*
	   When tmux try to attach, real tty is necessary.
	   But as default, exec.Command provides virtual in-memory pipe.
	   So tmux throws the error.
	*/
	cmd := sh.newTmuxCmd("tmux", "attach", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute tmux attach -t: %w", err)
	}
	return nil
}

func (sh *SessionHandler) switchClient(sessionName string) error {
	cmd := sh.newTmuxCmd("tmux", "switch-client", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute tmux switch-client -t: %w", err)
	}
	return nil
}

func (sh *SessionHandler) switchToNewClient(sessionName, path string) error {
	if sh.sessionExists(sessionName) {
		return sh.switchClient(sessionName)
	}
	cmd := sh.newTmuxCmd("tmux", "new-session", "-ds", sessionName, "-c", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute tmux new-sessio -ds %v -c %v: %w", sessionName, path, err)
	}
	return sh.switchClient(sessionName)
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

func (sh *SessionHandler) GrabExistingSession() error {
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
	if sh.isInSession() {
		return sh.switchClient(selected)
	}
	if err := sh.attach(selected); err != nil {
		return fmt.Errorf("failed to attach an existing session: %w", err)
	}
	return nil
}

func (sh *SessionHandler) isInSession() bool {
	return len(os.Getenv("TMUX")) > 0
}
