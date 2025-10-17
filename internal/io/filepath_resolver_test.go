package io

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFilePathResolver_ExpandTildeAsHomeDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr error
	}{
		{
			name:    "expand tilde to home directory",
			path:    "~/projects/myproject",
			want:    "/tmp/tmuxsessionizer/projects/myproject",
			wantErr: nil,
		},
		{
			name:    "no tilde, return as is",
			path:    "/var/www/html",
			want:    "/var/www/html",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		setupHomeEnv()
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fp := NewFilePathResolver()

			got, err := fp.ExpandTildeAsHomeDir(tt.path)
			if !errors.Is(tt.wantErr, err) {
				t.Errorf("ExpandTildeAsHomeDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ExpandTildeAsHomeDir() got = %v, want %v", got, tt.want)
			}
		})
	}
}
