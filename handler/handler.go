package handler

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/command"
	iohelper "github.com/TlexCypher/my-tmux-sessionizer/internal/io"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/session"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/tmux"
)

type ISessionHandler interface {
	NewSession(ctx context.Context) error
	GrabExistingSession(ctx context.Context) error
}

type SessionHandler struct {
	config  *iohelper.Config
	manager *session.SessionManager
	tmux    *tmux.Tmux
}

func NewSessionHandler(config *iohelper.Config, manager *session.SessionManager, tmux *tmux.Tmux) ISessionHandler {
	return &SessionHandler{
		config:  config,
		manager: manager,
		tmux:    tmux,
	}
}

func (sh *SessionHandler) NewSession(ctx context.Context) error {
	/*build fzf input and build hashmap to retrieve filepath from entry's name.*/
	fzfCmd := command.NewFzfCommand(ctx)
	for _, project := range sh.config.Projects {
		fzfCmd.InBuf().WriteString(project.Value() + "\n")
	}

	if err := fzfCmd.Run(); err != nil {
		return fmt.Errorf("failed to grab project path with fzf: %w", err)
	}

	projectPath := fzfCmd.OutBuf().String()
	session, err := sh.manager.GetSession(projectPath)
	// NOTE: if session is not found, create a new one.
	if err != nil {
		session := sh.manager.CreateSession(
			filepath.Base(projectPath),
			projectPath,
		)
		if sh.tmux.IsInSession() {
			return sh.tmux.SwitchToNewClient(ctx, session)
		} else {
			return sh.tmux.CreateAndAttach(ctx, session)
		}
	}

	if sh.tmux.IsInSession() {
		return sh.tmux.SwitchClient(ctx, session)
	} else {
		return sh.tmux.Attach(ctx, session)
	}
}

func (sh *SessionHandler) GrabExistingSession(ctx context.Context) error {
	sessions := sh.manager.ListSessions()

	fzfCmd := command.NewFzfCommand(ctx)
	for _, session := range sessions {
		fzfCmd.InBuf().WriteString(session.Name.Value() + "\n")
	}

	if err := fzfCmd.Run(); err != nil {
		return err
	}

	grabbed := fzfCmd.OutBuf().String()

	session, err := sh.manager.GetSession(grabbed)
	if err != nil {
		return err
	}

	if sh.tmux.IsInSession() {
		return sh.tmux.SwitchClient(ctx, session)
	}

	return sh.tmux.Attach(ctx, session)
}
