package handler

import (
	"fmt"
	"testing"

	testutils "github.com/TlexCypher/my-tmux-sessionizer/test_utils"
	"github.com/google/go-cmp/cmp"
)

func TestExpandPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		description string
		path        string
		want        string
		wantErr     error
	}{

		{
			description: "Home directory expansion (Normal case)",
			path:        "~/hoge/fuga",
			want:        fmt.Sprintf("%v/hoge/fuga", testutils.GetUserHomeDir()),
			wantErr:     nil,
		},
		{
			description: "No home directory",
			path:        "/home/hoge",
			want:        "/home/hoge",
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			p := NewPathResolver()
			got, err := p.ExpandPath(tt.path)
			if err != tt.wantErr {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("PathResolver.ExpandPath() result diff (-expect, +got)\n%s", diff)
			}
		})
	}
}

type want struct {
	projectNameFullPathMap   map[string]string
	projectExpressionNameMap map[string]string
}

type args struct {
	config *Config
}

func TestBuildProjectInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description string
		args        args
		want        want
		wantErr     error
	}{
		{
			description: "Normal case",
			args: args{
				config: &Config{
					Projects: []project{
						{
							name:     "tmux-sessionizer",
							filepath: "~/tmux-sessionizer",
						},
						{
							name:     "nvim",
							filepath: "~/.config/nvim",
						},
					},
				},
			},
			want: want{
				projectNameFullPathMap: map[string]string{
					"tmux-sessionizer": "~/tmux-sessionizer",
					"nvim":             "~/.config/nvim",
				},
				projectExpressionNameMap: map[string]string{
					fmt.Sprintf("%v/tmux-sessionizer", testutils.GetUserHomeDir()): "tmux-sessionizer",
					fmt.Sprintf("%v/.config/nvim", testutils.GetUserHomeDir()):     "nvim",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			p := NewPathResolver()
			got1, got2, err := p.BuildProjectInfo(tt.args.config)
			if tt.wantErr != err {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want.projectNameFullPathMap, got1); diff != "" {
				t.Errorf("PathResolver.BuildProjectInfo() result diff (-expect, +got)\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.projectExpressionNameMap, got2); diff != "" {
				t.Errorf("PathResolver.BuildProjectInfo() result diff (-expect, +got)\n%s", diff)
			}
		})
	}
}
