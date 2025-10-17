package tmux

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/command"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/session"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/types"
)

const (
	tmux = "TMUX"
)

type Tmux struct{}

func NewTmux() *Tmux {
	return &Tmux{}
}

func (t *Tmux) GatherExistingSessions(ctx context.Context) (map[types.String]*session.Session, error) {
	tmuxCmd := command.NewTmuxCommand(ctx, "list-sessions", "-F", "#{session_name}:#{session_path}")

	err := tmuxCmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to gather existing tmux sessions with `tmux list-sessions -F '#{session_name}:#{session_path}'`: %w", err)
	}

	existingSessions := make(map[types.String]*session.Session, 0)
	listSessions := types.NewString(tmuxCmd.OutBuf().String())
	splitCnt := 2

	itr := strings.SplitSeq(listSessions.Value(), "\n")
	for line := range itr {
		parts := strings.SplitN(line, ":", splitCnt)
		sessionName, projectPath := types.NewString(parts[0]), types.NewString(parts[1])
		existingSessions[projectPath] = session.NewSession(sessionName, projectPath)
	}

	return existingSessions, nil
}

func (t *Tmux) CreateAndAttach(ctx context.Context, session *session.Session) error {
	tmuxCmd := command.NewTmuxCommand(ctx, "new-session", "-s", session.Name.Value(), "-c", session.ProjectPath.Value())
	return tmuxCmd.Run()
}

func (t *Tmux) Attach(ctx context.Context, session *session.Session) error {
	tmuxCmd := command.NewTmuxCommand(ctx, "attach", "-t", session.Name.Value())
	return tmuxCmd.Run()
}

func (t *Tmux) SwitchClient(ctx context.Context, switchTo *session.Session) error {
	tmuxCmd := command.NewTmuxCommand(ctx, "switch-client", "-t", switchTo.Name.Value())
	return tmuxCmd.Run()
}

func (t *Tmux) SwitchToNewClient(ctx context.Context, switchTo *session.Session) error {
	tmuxCmd := command.NewTmuxCommand(ctx, "new-session", "-ds", switchTo.Name.Value(), "-c", switchTo.ProjectPath.Value())

	err := tmuxCmd.Run()
	if err != nil {
		return err
	}

	return t.SwitchClient(ctx, switchTo)
}

func (t *Tmux) Delete(ctx context.Context, session *session.Session) error {
	tmuxCmd := command.NewTmuxCommand(ctx, "kill-session", "-t", session.Name.Value())
	return tmuxCmd.Run()
}

func (t *Tmux) IsInSession() bool {
	return len(os.Getenv(tmux)) > 0
}

func (t *Tmux) HasSession(ctx context.Context, session *session.Session) bool {
	if session == nil {
		return false
	}

	tmuxCmd := command.NewTmuxCommand(ctx, "has-session", "-t", session.Name.Value())

	return tmuxCmd.Run() == nil
}
