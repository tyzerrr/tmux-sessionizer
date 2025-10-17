package session

import "github.com/TlexCypher/my-tmux-sessionizer/internal/types"

type Session struct {
	Id          SessionID
	Name        types.String
	ProjectPath types.String
}

func NewSession(name types.String, projectPath types.String) *Session {
	return &Session{
		Id:          NewSessionID(),
		Name:        name,
		ProjectPath: projectPath,
	}
}

func (s *Session) Project() types.String {
	return s.ProjectPath
}

func (s *Session) String() string {
	return s.Id.String()
}

func (s *Session) Same(other *Session) bool {
	return s.Id.Same(other.Id)
}
