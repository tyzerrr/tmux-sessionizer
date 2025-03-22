package handler

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type SessionHandler struct {
}

func NewSessionHandler() *SessionHandler {
	return &SessionHandler{}
}

func (sh *SessionHandler) Start() error {
	if err := sh.switchTo(); err != nil {
		return err
	}
	return nil
}

func (sh *SessionHandler) switchTo() error {
	tmuxCmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	var tmuxOut bytes.Buffer
	tmuxCmd.Stdout = &tmuxOut
	if err := tmuxCmd.Run(); err != nil {
		return fmt.Errorf("failed to execute tmux list-sessions command: %w", err)
	}

	fzfCmd := exec.Command("fzf")
	fzfCmd.Stdin = &tmuxOut
	var fzfOut bytes.Buffer
	fzfCmd.Stdout = &fzfOut
	if err := fzfCmd.Run(); err != nil {
		return fmt.Errorf("failed to execute fzf command: %w", err)
	}

	selected := strings.TrimSpace(fzfOut.String())
	slog.Debug("switchTo", slog.String("selected session", selected))

	/*
	   When tmux try to attach, real tty is necessary.
	   But as default, exec.Command provides virtual in-memory pipe.
	   So tmux throws the error.
	*/
	cmd := exec.Command("tmux", "attach", "-t", selected)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute tmux attach -t: %w", err)
	}
	return nil
}
