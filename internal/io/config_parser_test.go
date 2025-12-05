package io

import (
	"errors"
	"os"
	"testing"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/types"
	"github.com/google/go-cmp/cmp"
)

type testContainer struct {
	dirs []types.String
}

func NewTestContainer(dirs []types.String) *testContainer {
	return &testContainer{
		dirs: dirs,
	}
}

func setupHomeEnv() {
	os.Setenv("HOME", "/tmp/tmuxsessionizer")
}

func (s *testContainer) setup() {
	setupHomeEnv()

	for _, p := range s.dirs {
		_ = os.MkdirAll(p.Value(), 0755)
	}
}

func (s *testContainer) teardown() {
	for _, p := range s.dirs {
		_ = os.RemoveAll(p.Value())
	}
}

func Test_ConfigParser_parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		projectList []string
		want        *Config
		wantErr     error
	}{
		{
			name: "including tilde(~) expansion and '\n'",
			projectList: []string{
				"~/projects/",
			},
			want: &Config{
				Projects: []types.String{
					types.NewString("/tmp/tmuxsessionizer/projects\n"),
					types.NewString("/tmp/tmuxsessionizer/projects/project1\n"),
					types.NewString("/tmp/tmuxsessionizer/projects/project2\n"),
					types.NewString("/tmp/tmuxsessionizer/projects/project3\n"),
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		c := NewTestContainer(tt.want.Projects)
		c.setup()
		t.Cleanup(c.teardown)

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cp := NewConfigParser()
			got, err := cp.parse(tt.projectList)

			if diff := cmp.Diff(tt.want.Projects, got.Projects, []cmp.Option{
				cmp.Comparer(func(a, b types.String) bool {
					return a.Value() == b.Value()
				}),
			}...); diff != "" {
				t.Errorf("ConfigParser.parse() mismatch (-want +got):\n%s", diff)
			}

			if !errors.Is(tt.wantErr, err) {
				t.Errorf("ConfigParser.parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
