package handler

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
)

type SessionHandler struct {
	config                   *Config
	pathResolver             *PathResolver
	projectNameFullPathMap   map[string]string
	projectExpressionNameMap map[string]string
}

func NewSessionHandler(config *Config, pathResolver *PathResolver,
	projectNameFullPathMap, projectExpressionNameMap map[string]string) ISessionHandler {
	return &SessionHandler{
		config:                   config,
		pathResolver:             pathResolver,
		projectNameFullPathMap:   projectNameFullPathMap,
		projectExpressionNameMap: projectExpressionNameMap,
	}
}

func (sh *SessionHandler) NewSession(ctx context.Context) error {
	/*build fzf input and build hashmap to retrieve filepath from entry's name.*/
	var input bytes.Buffer
	for _, project := range sh.config.Projects {
		replaced, err := sh.pathResolver.ExpandPath(project.filepath)
		if err != nil {
			return errors.New("failed to replace home directory to ~")
		}
		input.WriteString(replaced + "\n")
	}
	fzfOut, err := sh.toFzf(input)
	if err != nil {
		return fmt.Errorf("failed to create new project from projects: %w", err)
	}

	sessionName := sh.projectExpressionNameMap[strings.Trim(fzfOut.String(), "\n")]
	if sh.isInSession() {
		return sh.switchToNewClient(sessionName, sh.projectNameFullPathMap[sessionName])
	}
	if sh.sessionExists(sessionName) {
		return sh.attach(sessionName)
	}
	if err := sh.newTmuxCmd(
		"tmux", "new-session", "-s",
		sessionName, "-c", sh.projectNameFullPathMap[sessionName]).Run(); err != nil {
		return fmt.Errorf("failed to create new session %w: ", err)
	}
	return nil
}

func (sh *SessionHandler) GrabExistingSession(ctx context.Context) error {
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

func (sh *SessionHandler) CreateNewProjectSession(ctx context.Context) error {
	parentDirSet, err := sh.buildParentDirSet()
	if err != nil {
		return fmt.Errorf("failed to build parent directory set: %w", err)
	}
	parentDirs := make([]string, 0)
	for key := range parentDirSet {
		parentDirs = append(parentDirs, key)
	}

	for {
		parentDir, newProjectName, err := sh.getNewProjectInfo(parentDirs)
		if err != nil {
			return fmt.Errorf("faield to get new project information interactively: %w", err)
		}
		newProjectPath, err := sh.pathResolver.ExpandPath(filepath.Join(parentDir, fmt.Sprintf("%v/", newProjectName)))
		if err != nil {
			return fmt.Errorf("failed to expand path: %w", err)
		}
		_, err = os.Stat(newProjectPath)
		if err == nil {
			fmt.Printf("%v is already exist\n", newProjectPath)
			continue
		}
		if os.IsNotExist(err) {
			if err := exec.Command("mkdir", "-p", newProjectPath).Run(); err != nil {
				return fmt.Errorf("failed to create a new project: %w", err)
			}
			if sh.isInSession() {
				return sh.switchToNewClient(newProjectName, newProjectPath)
			}
			if err := sh.newTmuxCmd(
				"tmux", "new-session", "-s",
				newProjectName, "-c", newProjectPath).Run(); err != nil {
				return fmt.Errorf("failed to create new session %w: ", err)
			}
			break
		} else {
			return fmt.Errorf("failed to create a new project: %w", err)
		}
	}
	return nil
}

func (sh *SessionHandler) DeleteProjectSession(ctx context.Context) error {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		return fmt.Errorf("failed to execute tmux list-sessions cmd: %w", err)
	}
	sessions := strings.Split(strings.TrimSpace(string(out)), "\n")
	deleteSession, err := sh.getSessionInfoToDelete(sessions)
	if err != nil {
		return fmt.Errorf("failed to get a session name to delete: %w", err)
	}
	if err = sh.newTmuxCmd("tmux", "kill-session", "-t", deleteSession).Run(); err != nil {
		return fmt.Errorf("failed to kill session: %w", err)
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

func (sh *SessionHandler) buildParentDirSet() (map[string]struct{}, error) {
	var configFiles = []string{"./.tmux-sessionizer", "~/.tmux-sessionizer"}
	parentDirSet := make(map[string]struct{}, 0)
	for _, configFile := range configFiles {
		path, err := sh.pathResolver.ExpandPath(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to expand path: %w", err)
		}
		file, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to open config file: %w", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "default=") {
				raw := strings.TrimPrefix(line, "default=")
				projects := strings.Split(raw, ",")
				for _, pr := range projects {
					trimmed := strings.TrimSpace(pr)
					parentDirSet[trimmed] = struct{}{}
				}
			}
		}
	}
	return parentDirSet, nil
}

func (sh *SessionHandler) getNewProjectInfo(parentDirs []string) (string, string, error) {
	chooseParentDirPrompt := promptui.Select{
		Label: "Which project do you choose to create a new project (session) ?",
		Items: parentDirs,
	}
	_, parentDir, err := chooseParentDirPrompt.Run()
	if err != nil {
		return "", "", fmt.Errorf("failed to choose a parent directory: %w", err)
	}

	newProjectNamePrompt := promptui.Prompt{
		Label: "Enter a new project name",
	}
	newProjectName, err := newProjectNamePrompt.Run()
	if err != nil {
		return "", "", fmt.Errorf("failed to get a new project name: %w", err)
	}
	return parentDir, newProjectName, nil
}

func (sh *SessionHandler) getSessionInfoToDelete(existingProjects []string) (string, error) {
	deleteSessionPrompt := promptui.Select{
		Label: "Which tmux session do you want to delete?",
		Items: existingProjects,
	}
	_, deleteSessionName, err := deleteSessionPrompt.Run()
	if err != nil {
		return "", err
	}
	return deleteSessionName, nil
}

func (sh *SessionHandler) isInSession() bool {
	return len(os.Getenv("TMUX")) > 0
}
