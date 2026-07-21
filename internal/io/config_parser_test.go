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

	// parse returns the immediate subdirectories of each registered entry,
	// so rebuild the fixture tree from scratch — otherwise the results
	// would depend on leftovers in /tmp.
	if err := os.RemoveAll("/tmp/tmuxsessionizer"); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{
		"/tmp/tmuxsessionizer/projects/app",
		"/tmp/tmuxsessionizer/projects/tool",
		"/tmp/tmuxsessionizer/personal/blog",
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// plain files are not session candidates and must be ignored
	//nolint:gosec // the fixture must live under the fixed fake HOME shared with filer_test, not t.TempDir.
	if err := os.WriteFile("/tmp/tmuxsessionizer/projects/README.md", []byte("readme"), 0o600); err != nil {
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
				Registered: []types.String{
					types.NewString("/tmp/tmuxsessionizer/projects"),
				},
				Projects: []types.String{
					types.NewString("/tmp/tmuxsessionizer/projects/app"),
					types.NewString("/tmp/tmuxsessionizer/projects/tool"),
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
				Registered: []types.String{
					types.NewString("/tmp/tmuxsessionizer/personal"),
					types.NewString("/tmp/tmuxsessionizer/projects"),
				},
				Projects: []types.String{
					types.NewString("/tmp/tmuxsessionizer/personal/blog"),
					types.NewString("/tmp/tmuxsessionizer/projects/app"),
					types.NewString("/tmp/tmuxsessionizer/projects/tool"),
				},
			},
			wantErr: nil,
		},
		{
			name: "entries whose directory does not exist are kept as registered but yield no projects",
			projectList: []string{
				"~/ghost",
				"~/projects",
			},
			want: &Config{
				Registered: []types.String{
					types.NewString("/tmp/tmuxsessionizer/ghost"),
					types.NewString("/tmp/tmuxsessionizer/projects"),
				},
				Projects: []types.String{
					types.NewString("/tmp/tmuxsessionizer/projects/app"),
					types.NewString("/tmp/tmuxsessionizer/projects/tool"),
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

			opts := []cmp.Option{
				cmp.Comparer(func(a, b types.String) bool {
					return a.Value() == b.Value()
				}),
			}
			if diff := cmp.Diff(tt.want.Registered, got.Registered, opts...); diff != "" {
				t.Errorf("ConfigParser.parse() Registered mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.Projects, got.Projects, opts...); diff != "" {
				t.Errorf("ConfigParser.parse() Projects mismatch (-want +got):\n%s", diff)
			}

			if !errors.Is(tt.wantErr, err) {
				t.Errorf("ConfigParser.parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
