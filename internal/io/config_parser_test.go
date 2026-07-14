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

	// parse skips entries whose directory does not exist, so create every
	// directory the cases expect and guarantee the "ghost" one is absent —
	// otherwise the results would depend on leftovers in /tmp.
	for _, dir := range []string{"/tmp/tmuxsessionizer/projects", "/tmp/tmuxsessionizer/personal"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.RemoveAll("/tmp/tmuxsessionizer/ghost"); err != nil {
		t.Fatal(err)
	}

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
		{
			name: "entries whose directory does not exist are skipped",
			projectList: []string{
				"~/ghost",
				"~/projects",
			},
			want: &Config{
				Projects: []types.String{
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
			got, err := cp.parse(tt.projectList, NewFiler())

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
