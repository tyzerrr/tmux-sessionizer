package io

import (
	"errors"
	"os"
	"testing"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/types"
	"github.com/google/go-cmp/cmp"
)

func setupHomeEnv() {
	os.Setenv("HOME", "/tmp/tmuxsessionizer")
}

func Test_ConfigParser_parse(t *testing.T) {
	setupHomeEnv()
	t.Parallel()

	tests := []struct {
		name        string
		projectList []string
		want        *Config
		wantErr     error
	}{
		{
			name: "including tilde(~) expansion and trailing slash",
			projectList: []string{
				"~/projects/",
			},
			want: &Config{
				Projects: []types.String{
					types.NewString("/tmp/tmuxsessionizer/projects"),
				},
			},
			wantErr: nil,
		},
		{
			name: "multiple entries, surrounding blanks and empty entries are normalized",
			projectList: []string{
				" ~/personal ",
				"",
				"~/projects",
			},
			want: &Config{
				Projects: []types.String{
					types.NewString("/tmp/tmuxsessionizer/personal"),
					types.NewString("/tmp/tmuxsessionizer/projects"),
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
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
