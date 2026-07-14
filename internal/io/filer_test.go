package io

import (
	"errors"
	"path/filepath"
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

			fl := NewFiler()

			got, err := fl.ExpandTildeAsHomeDir(tt.path)
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

func TestFiler_Exists_ReturnsNilForExistingDirectory(t *testing.T) {
	t.Parallel()

	fl := NewFiler()

	if err := fl.Exists(t.TempDir()); err != nil {
		t.Errorf("expected nil for existing directory, got %v", err)
	}
}

func TestFiler_Exists_ReturnsErrorForMissingPath(t *testing.T) {
	t.Parallel()

	fl := NewFiler()

	if err := fl.Exists(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("expected error for missing path, got nil")
	}
}
