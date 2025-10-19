package session

import (
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/types"
	"github.com/google/go-cmp/cmp"
)

func TestSessionManager_CreateSession(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sessionName string
		path        string
		want        *Session
	}{
		{
			name:        "Create new session even for including newline characters",
			sessionName: "/path/to/project1\n",
			path:        "/path/to/project1\n",
			want: &Session{
				Name:        types.NewString("/path/to/project1"),
				ProjectPath: types.NewString("/path/to/project1"),
			},
		},
		{
			name:        "Create new session with dot in name",
			sessionName: "/path/to/.config/tmux-sessionizer\n",
			path:        "/path/to/.config/tmux-sessionizer\n",
			want: &Session{
				Name:        types.NewString("/path/to/_config/tmux-sessionizer"),
				ProjectPath: types.NewString("/path/to/.config/tmux-sessionizer"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sessionNameTransformer := NewTransformer().WithRule(
				NewTransformRule(
					func(in string) string { return strings.ReplaceAll(in, ".", "_") },
					func(in string) string { return strings.ReplaceAll(in, "_", ".") },
				),
				NewTransformRule(
					func(in string) string { return strings.ReplaceAll(in, ":", ";") },
					func(in string) string { return strings.ReplaceAll(in, ";", ":") },
				),
			)

			sm := NewSessionManager(make(map[types.String]*Session), sessionNameTransformer)

			got := sm.CreateSession(tt.sessionName, tt.path)
			if diff := cmp.Diff(tt.want, got, []cmp.Option{
				cmp.Comparer(func(a, b Session) bool {
					return a.Name.Value() == b.Name.Value() && a.ProjectPath.Value() == b.ProjectPath.Value()
				}),
			}...); diff != "" {
				t.Errorf("CreateSession() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSesssionManager_GetSession(t *testing.T) {
	t.Parallel()

	session1 := &Session{Name: types.NewString("project1"), ProjectPath: types.NewString("/path/to/project1")}
	sessionNameTransformer := NewTransformer().WithRule(
		NewTransformRule(
			func(in string) string { return strings.ReplaceAll(in, ".", "_") },
			func(in string) string { return strings.ReplaceAll(in, "_", ".") },
		),
		NewTransformRule(
			func(in string) string { return strings.ReplaceAll(in, ":", ";") },
			func(in string) string { return strings.ReplaceAll(in, ";", ":") },
		),
	)
	sm := NewSessionManager(map[types.String]*Session{
		types.NewString("/path/to/project1"): session1,
	}, sessionNameTransformer)

	tests := []struct {
		name    string
		path    string
		want    *Session
		wantErr error
	}{
		{
			name:    "Get existing session",
			path:    "/path/to/project1\n",
			want:    session1,
			wantErr: nil,
		},
		{
			name:    "Get non-existing session",
			path:    "/path/to/nonexistent\n",
			want:    nil,
			wantErr: ErrSessionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := sm.GetSession(tt.path)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got, []cmp.Option{
				cmp.Comparer(func(a, b *Session) bool {
					if a == nil || b == nil {
						return a == b
					}
					return a.Name.Value() == b.Name.Value() && a.ProjectPath.Value() == b.ProjectPath.Value()
				}),
			}...); diff != "" {
				t.Errorf("GetSession() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSessionManager_ListSessions(t *testing.T) {
	t.Parallel()

	session1 := &Session{Name: types.NewString("project1"), ProjectPath: types.NewString("/path/to/project1")}
	session2 := &Session{Name: types.NewString("project2"), ProjectPath: types.NewString("/path/to/project2")}
	session3 := &Session{Name: types.NewString("project3"), ProjectPath: types.NewString("/path/to/project3")}

	sessionNameTransformer := NewTransformer().WithRule(
		NewTransformRule(
			func(in string) string { return strings.ReplaceAll(in, ".", "_") },
			func(in string) string { return strings.ReplaceAll(in, "_", ".") },
		),
		NewTransformRule(
			func(in string) string { return strings.ReplaceAll(in, ":", ";") },
			func(in string) string { return strings.ReplaceAll(in, ";", ":") },
		),
	)
	sm := NewSessionManager(map[types.String]*Session{
		types.NewString("/path/to/project1"): session1,
		types.NewString("/path/to/project2"): session2,
		types.NewString("/path/to/project3"): session3,
	}, sessionNameTransformer)

	tests := []struct {
		name string
		want []*Session
	}{
		{
			name: "List all sessions",
			want: []*Session{session1, session2, session3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := sm.ListSessions()

			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].Name.Value() < tt.want[j].Name.Value()
			})
			sort.Slice(got, func(i, j int) bool {
				return got[i].Name.Value() < got[j].Name.Value()
			})

			if diff := cmp.Diff(tt.want, got, []cmp.Option{
				cmp.Comparer(func(a, b *Session) bool {
					return a.Name.Value() == b.Name.Value() && a.ProjectPath.Value() == b.ProjectPath.Value()
				}),
			}...); diff != "" {
				t.Errorf("ListSessions() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
