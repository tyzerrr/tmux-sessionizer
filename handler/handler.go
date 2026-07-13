package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/command"
	iohelper "github.com/TlexCypher/my-tmux-sessionizer/internal/io"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/session"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/tmux"
	"golang.org/x/sync/errgroup"
)

type ISessionHandler interface {
	NewSession(ctx context.Context) error
	GrabExistingSession(ctx context.Context) error
	DeleteSession(ctx context.Context) error
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

	rawPath := fzfCmd.OutBuf().String()
	session, err := sh.manager.GetSession(rawPath)
	// NOTE: if session is not found, create a new one.
	if err != nil {
		session := sh.manager.CreateSession(
			rawPath,
			rawPath,
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
		fzfCmd.InBuf().WriteString(session.ProjectPath.Value() + "\n")
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

func (sh *SessionHandler) DeleteSession(ctx context.Context) error {
	sessions := sh.manager.ListSessions()
	fzfCmd := command.NewFzfCommand(ctx, "-m")

	for _, session := range sessions {
		fzfCmd.InBuf().WriteString(session.ProjectPath.Value() + "\n")
	}

	if err := fzfCmd.Run(); err != nil {
		return err
	}

	deletes := fzfCmd.OutBuf().String()
	// split all with "\n"
	ds := strings.Split(strings.TrimSuffix(deletes, "\n"), "\n")
	filtered := sh.manager.FilterSessions(ds)
	if err := sh.manager.DeleteSessions(ds); err != nil {
		return err
	}
	eg := new(errgroup.Group)
	for _, f := range filtered {
		eg.Go(func() error {
			if err := sh.tmux.Delete(ctx, f); err != nil {
				return err
			}
			return nil
		})
	}
	return eg.Wait()
}
