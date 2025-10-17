package session

import "github.com/google/uuid"

type SessionID struct {
	uuid.UUID
}

func NewSessionID() SessionID {
	return SessionID{UUID: uuid.New()}
}

func (s *SessionID) String() string {
	return s.UUID.String()
}

func (s *SessionID) Same(other SessionID) bool {
	return s.UUID.String() == other.UUID.String()
}
