package handler

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()
	tmpRoot := t.TempDir()
	project1Name, project2Name := "project1", "project2"
	projectDir1 := filepath.Join(tmpRoot, project1Name)
	projectDir2 := filepath.Join(tmpRoot, project2Name)
	projects := []string{projectDir1, projectDir2}
	if err := os.Mkdir(projectDir1, 0755); err != nil {
		t.Errorf("failed to make project1 directory")
	}
	if err := os.Mkdir(projectDir2, 0755); err != nil {
		t.Errorf("failed to make project2 directory")
	}

	configContent := "default=" + tmpRoot

	tmpConfigFile := filepath.Join(tmpRoot, ".tmux-sessionizer")
	if err := os.WriteFile(tmpConfigFile, []byte(configContent), 0600); err != nil {
		t.Errorf("failed to write config file: %v", err)
	}

	sh := &SessionHandler{}
	cfg, err := sh.parseConfig(tmpConfigFile)
	if err != nil {
		t.Errorf("failed to parse config file: %v", err)
	}

	if len(cfg.projects) != len(projects) {
		t.Errorf("expected %v projects, but got %v", len(projects), len(cfg.projects))
	}
	wantNames := map[string]bool{project1Name: true, project2Name: true}
	for _, p := range cfg.projects {
		if !wantNames[p.name] {
			t.Errorf("unexpected project: %+v", p)
		}
	}
}

func TestExpandPath(t *testing.T) {
	t.Parallel()
	type args struct {
		relative string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "home directory case",
			args: args{
				relative: "~/hoge/fuga",
			},
			/* I don't care about windows, sorry.*/
			want:    os.Getenv("HOME") + "/hoge/fuga",
			wantErr: nil,
		},
		{
			name: "complex relative path case",
			args: args{
				relative: "~/puga/hoge/../../fuga",
			},
			want:    os.Getenv("HOME") + "/fuga",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sh := SessionHandler{}
			got, err := sh.expandPath(tt.args.relative)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("failed to expand path: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("failed to expandPath. expected: %v, but got %v", tt.want, got)
			}
		})
	}
}

func TestReplaceHomeDir(t *testing.T) {
	t.Parallel()
	homeDir, _ := os.UserHomeDir()
	type args struct {
		fullpath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "Normal test case1",
			args: args{
				fullpath: filepath.Join(homeDir, "a/b/c"),
			},
			want:    "~/a/b/c",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sh := SessionHandler{}
			got, err := sh.replaceHomeDir(tt.args.fullpath)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("failed to replace home directory: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("failed to replace home directory, expected: %v, but got %v", tt.want, got)
			}
		})
	}
}
