package handler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/io"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	configFileAbs := filepath.Join(t.TempDir(), ".tmux-sessionizer")
	if err := os.WriteFile(configFileAbs, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return configFileAbs
}

func readConfig(t *testing.T, configFileAbs string) string {
	t.Helper()

	b, err := os.ReadFile(configFileAbs)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestProjectHandler_Init_CreatesConfigWithPrefix(t *testing.T) {
	t.Parallel()

	configFileAbs := filepath.Join(t.TempDir(), ".tmux-sessionizer")
	ph := NewProjectHandler(configFileAbs)

	if err := ph.Init(t.Context(), configFileAbs); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := readConfig(t, configFileAbs); got != io.ConfigPrefix {
		t.Errorf("expected %q, got %q", io.ConfigPrefix, got)
	}
}

func TestProjectHandler_Init_KeepsExistingConfigIntact(t *testing.T) {
	t.Parallel()

	content := io.ConfigPrefix + "/home/user/project"
	configFileAbs := writeConfig(t, content)
	ph := NewProjectHandler(configFileAbs)

	if err := ph.Init(t.Context(), configFileAbs); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := readConfig(t, configFileAbs); got != content {
		t.Errorf("expected existing config to be kept as %q, got %q", content, got)
	}
}

func TestProjectHandler_Register_AppendsFirstProjectWithoutComma(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfig(t, io.ConfigPrefix)
	project := t.TempDir()
	ph := NewProjectHandler(configFileAbs)

	if err := ph.Register(t.Context(), project); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := io.ConfigPrefix + project
	if got := readConfig(t, configFileAbs); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestProjectHandler_Register_SeparatesProjectsWithComma(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfig(t, io.ConfigPrefix)
	first, second := t.TempDir(), t.TempDir()
	ph := NewProjectHandler(configFileAbs)

	if err := ph.Register(t.Context(), first); err != nil {
		t.Fatalf("expected no error on first register, got %v", err)
	}
	if err := ph.Register(t.Context(), second); err != nil {
		t.Fatalf("expected no error on second register, got %v", err)
	}

	want := io.ConfigPrefix + first + "," + second
	if got := readConfig(t, configFileAbs); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestProjectHandler_Register_TruncatesTrailingNewlineBeforeAppend(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfig(t, io.ConfigPrefix+"\n")
	project := t.TempDir()
	ph := NewProjectHandler(configFileAbs)

	if err := ph.Register(t.Context(), project); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := io.ConfigPrefix + project
	if got := readConfig(t, configFileAbs); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestProjectHandler_Register_RejectsRegularFile(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfig(t, io.ConfigPrefix)
	file := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	ph := NewProjectHandler(configFileAbs)

	if err := ph.Register(t.Context(), file); err == nil {
		t.Error("expected error when registering a regular file, got nil")
	}

	if got := readConfig(t, configFileAbs); got != io.ConfigPrefix {
		t.Errorf("expected config to stay %q, got %q", io.ConfigPrefix, got)
	}
}
