package session

import "github.com/TlexCypher/my-tmux-sessionizer/internal/types"

type Session struct {
	Name        types.String
	ProjectPath types.String
}

func NewSession(name types.String, projectPath types.String) *Session {
	return &Session{
		Name:        name,
		ProjectPath: projectPath,
	}
}

func (s *Session) Project() types.String {
	return s.ProjectPath
}
