package session

import (
	"errors"
	"sort"
	"testing"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/types"
	"github.com/google/go-cmp/cmp"
)

func TestSessionManager_CreateSession(t *testing.T) {
	t.Parallel()

	sm := NewSessionManager(make(map[types.String]*Session))

	tests := []struct {
		name        string
		sessionName string
		path        string
		want        *Session
	}{
		{
			name:        "Create new session even for including newline characters",
			sessionName: "project1\n",
			path:        "/path/to/project1\n",
			want: &Session{
				Name:        types.NewString("project1"),
				ProjectPath: types.NewString("/path/to/project1"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

	sm := NewSessionManager(map[types.String]*Session{
		types.NewString("/path/to/project1"): session1,
	})

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

	sm := NewSessionManager(map[types.String]*Session{
		types.NewString("/path/to/project1"): session1,
		types.NewString("/path/to/project2"): session2,
		types.NewString("/path/to/project3"): session3,
	})

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

			// 順序に依存しない比較のために、両方のスライスをソート
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
