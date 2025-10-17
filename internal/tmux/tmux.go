package tmux

import (
	"fmt"
	"os"
	"strings"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/command"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/session"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/types"
)

var (
	tmux = "TMUX"
)

type Tmux struct{}

func NewTmux() *Tmux {
	return &Tmux{}
}

func (t *Tmux) GatherExistingSessions() (map[types.String]*session.Session, error) {
	tmuxCmd := command.NewTmuxCommand("list-sessions", "-F", "#{session_name}:#{session_path}")
	if err := tmuxCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to gather existing tmux sessions with `tmux list-sessions -F '#{session_name}:#{session_path}'`: %w", err)
	}
	existingSessions := make(map[types.String]*session.Session, 0)
	listSessions := types.NewString(tmuxCmd.OutBuf().String())
	itr := strings.SplitSeq(listSessions.Value(), "\n")
	for line := range itr {
		parts := strings.SplitN(line, ":", 2)
		sessionName, projectPath := types.NewString(parts[0]), types.NewString(parts[1])
		existingSessions[projectPath] = session.NewSession(sessionName, projectPath)
	}
	return existingSessions, nil
}

func (t *Tmux) CreateAndAttach(session *session.Session) error {
	tmuxCmd := command.NewTmuxCommand("new-session", "-s", session.Name.Value(), "-c", session.ProjectPath.Value())
	return tmuxCmd.Run()
}

func (t *Tmux) Attach(session *session.Session) error {
	tmuxCmd := command.NewTmuxCommand("attach", "-t", session.Name.Value())
	return tmuxCmd.Run()
}

func (t *Tmux) SwitchClient(switchTo *session.Session) error {
	tmuxCmd := command.NewTmuxCommand("switch-client", "-t", switchTo.Name.Value())
	return tmuxCmd.Run()
}

func (t *Tmux) SwitchToNewClient(switchTo *session.Session) error {
	tmuxCmd := command.NewTmuxCommand("new-session", "-ds", switchTo.Name.Value(), "-c", switchTo.ProjectPath.Value())
	if err := tmuxCmd.Run(); err != nil {
		return err
	}
	return t.SwitchClient(switchTo)
}

func (t *Tmux) Delete(session *session.Session) error {
	tmuxCmd := command.NewTmuxCommand("kill-session", "-t", session.Name.Value())
	return tmuxCmd.Run()
}

func (t *Tmux) IsInSession() bool {
	return len(os.Getenv(tmux)) > 0
}

func (t *Tmux) HasSession(session *session.Session) bool {
	if session == nil {
		return false
	}
	tmuxCmd := command.NewTmuxCommand("has-session", "-t", session.Name.Value())
	return tmuxCmd.Run() == nil
}
