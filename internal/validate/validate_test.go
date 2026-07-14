package validate

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

func TestValidateConfig_ValidPrefix_ReturnsNil(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfig(t, io.ConfigPrefix+"/home/user/project")

	if err := ValidateConfig(configFileAbs); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidateConfig_PrefixOnly_ReturnsNil(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfig(t, io.ConfigPrefix)

	if err := ValidateConfig(configFileAbs); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidateConfig_MissingFile_ReturnsError(t *testing.T) {
	t.Parallel()

	configFileAbs := filepath.Join(t.TempDir(), ".tmux-sessionizer")

	if err := ValidateConfig(configFileAbs); err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestValidateConfig_EmptyFile_ReturnsError(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfig(t, "")

	if err := ValidateConfig(configFileAbs); err == nil {
		t.Error("expected error for empty file, got nil")
	}
}

func TestValidateConfig_WrongPrefix_ReturnsError(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfig(t, "not a tmux-sessionizer config\n")

	if err := ValidateConfig(configFileAbs); err == nil {
		t.Error("expected error for wrong prefix, got nil")
	}
}
