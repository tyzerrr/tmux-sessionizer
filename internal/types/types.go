package types

import "strings"

type String struct {
	value string
}

func NewString(value string) String {
	return String{value: strings.TrimSpace(value)}
}

func (s *String) Value() string {
	return s.value
}
