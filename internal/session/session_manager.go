package session

import (
	"fmt"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/types"
)

var (
	ErrSessionNotFound = fmt.Errorf("session not found")
)

type SessionManager struct {
	sessions map[types.String]*Session
}

func NewSessionManager(sessions map[types.String]*Session) *SessionManager {
	return &SessionManager{
		sessions: sessions,
	}
}

func (sm *SessionManager) CreateSession(rawName string, rawPath string) *Session {
	name, projectPath := types.NewString(rawName), types.NewString(rawPath)
	if _, exists := sm.sessions[projectPath]; !exists {
		sm.sessions[projectPath] = NewSession(name, projectPath)
	}
	return sm.sessions[projectPath]
}

func (sm *SessionManager) ListSessions() []*Session {
	var sessions []*Session
	for _, v := range sm.sessions {
		sessions = append(sessions, v)
	}
	return sessions
}

func (sm *SessionManager) GetSession(rawPath string) (*Session, error) {
	projectPath := types.NewString(rawPath)
	if session, exists := sm.sessions[projectPath]; exists {
		return session, nil
	}
	return nil, ErrSessionNotFound
}
