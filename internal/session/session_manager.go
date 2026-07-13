package session

import (
	"errors"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/types"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type SessionManager struct {
	sessions               map[types.String]*Session
	sessionNameTransformer *Transformer
}

func NewSessionManager(sessions map[types.String]*Session, transformer *Transformer) *SessionManager {
	return &SessionManager{
		sessions:               sessions,
		sessionNameTransformer: transformer,
	}
}

func (sm *SessionManager) CreateSession(rawName string, rawPath string) *Session {
	transformed := sm.sessionNameTransformer.Transform(rawPath)

	sessionName, projectPath := types.NewString(transformed), types.NewString(rawPath)
	if _, exists := sm.sessions[projectPath]; !exists {
		sm.sessions[projectPath] = NewSession(sessionName, projectPath)
	}

	return sm.sessions[projectPath]
}

func (sm *SessionManager) ListSessions() (sessions []*Session) {
	for _, v := range sm.sessions {
		sessions = append(sessions, v)
	}

	return sessions
}

func (sm *SessionManager) FilterSessions(rawPaths []string) (sessions []*Session) {
	paths := make(map[string]struct{})
	for _, rp := range rawPaths {
		paths[rp] = struct{}{}
	}

	for _, v := range sm.sessions {
		if _, exists := paths[v.ProjectPath.Value()]; exists {
			sessions = append(sessions, v)
		}
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

// DeleteSessions removes every session identified by rawPaths. Deletion runs
// sequentially because the underlying map is not safe for concurrent writes.
func (sm *SessionManager) DeleteSessions(rawPaths []string) error {
	for _, path := range rawPaths {
		if err := sm.deleteSession(path); err != nil {
			return err
		}
	}

	return nil
}

func (sm *SessionManager) deleteSession(rawPath string) error {
	projectPath := types.NewString(rawPath)
	if _, exists := sm.sessions[projectPath]; !exists {
		return ErrSessionNotFound
	}

	delete(sm.sessions, projectPath)

	return nil
}
