package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	iohelper "github.com/TlexCypher/my-tmux-sessionizer/internal/io"
	"github.com/urfave/cli/v3"
)

// runTestCmd drives runWithHandler through a real cli.Command with a
// temporary config file, so tests exercise the actual dispatch logic
// without touching the user's config.
func runTestCmd(t *testing.T, configFileAbs string, args ...string) error {
	t.Helper()

	cmd := &cli.Command{
		Name:  "mock tmux-sessionizer",
		Usage: "mock tmux session manager",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runWithHandler(ctx, cmd, iohelper.NewFiler(), configFileAbs)
		},
	}

	return cmd.Run(t.Context(), append([]string{"tmux-sessionizer"}, args...))
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	configFileAbs := filepath.Join(t.TempDir(), ".tmux-sessionizer")
	if err := os.WriteFile(configFileAbs, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return configFileAbs
}

func readConfigFile(t *testing.T, configFileAbs string) string {
	t.Helper()

	b, err := os.ReadFile(configFileAbs)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRunWithHandler_InvalidCommand_ReturnsErrNoSuchCmd(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfigFile(t, iohelper.ConfigPrefix)

	err := runTestCmd(t, configFileAbs, "invalid")

	if !errors.Is(err, ErrNoSuchCmd) {
		t.Errorf("expected ErrNoSuchCmd, got %v", err)
	}
}

func TestRunWithHandler_Init_CreatesConfigFileWithPrefix(t *testing.T) {
	t.Parallel()

	configFileAbs := filepath.Join(t.TempDir(), ".tmux-sessionizer")

	if err := runTestCmd(t, configFileAbs, "init"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := readConfigFile(t, configFileAbs); got != iohelper.ConfigPrefix {
		t.Errorf("expected %q, got %q", iohelper.ConfigPrefix, got)
	}
}

func TestRunWithHandler_Init_KeepsExistingConfigIntact(t *testing.T) {
	t.Parallel()

	content := iohelper.ConfigPrefix + "/home/user/project"
	configFileAbs := writeConfigFile(t, content)

	if err := runTestCmd(t, configFileAbs, "init"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := readConfigFile(t, configFileAbs); got != content {
		t.Errorf("expected existing config to be kept as %q, got %q", content, got)
	}
}

func TestRunWithHandler_Register_AppendsFirstProjectWithoutComma(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfigFile(t, iohelper.ConfigPrefix)
	project := t.TempDir()

	if err := runTestCmd(t, configFileAbs, "register", project); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := iohelper.ConfigPrefix + project
	if got := readConfigFile(t, configFileAbs); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestRunWithHandler_Register_SeparatesSecondProjectWithComma(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfigFile(t, iohelper.ConfigPrefix)
	first, second := t.TempDir(), t.TempDir()

	if err := runTestCmd(t, configFileAbs, "register", first); err != nil {
		t.Fatalf("expected no error on first register, got %v", err)
	}
	if err := runTestCmd(t, configFileAbs, "register", second); err != nil {
		t.Fatalf("expected no error on second register, got %v", err)
	}

	want := iohelper.ConfigPrefix + first + "," + second
	if got := readConfigFile(t, configFileAbs); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

//nolint:paralleltest // t.Chdir is incompatible with t.Parallel.
func TestRunWithHandler_Register_ResolvesRelativePathToAbsolute(t *testing.T) {
	configFileAbs := writeConfigFile(t, iohelper.ConfigPrefix)
	parent := t.TempDir()
	if err := os.Mkdir(filepath.Join(parent, "project"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(parent)

	if err := runTestCmd(t, configFileAbs, "register", "project"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Getwd resolves symlinks (e.g. /var -> /private/var on macOS),
	// so derive the expected path from it rather than from parent.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	want := iohelper.ConfigPrefix + filepath.Join(wd, "project")
	if got := readConfigFile(t, configFileAbs); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestRunWithHandler_Register_RejectsPathWithComma(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfigFile(t, iohelper.ConfigPrefix)
	project := filepath.Join(t.TempDir(), "my,dir")
	if err := os.Mkdir(project, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := runTestCmd(t, configFileAbs, "register", project); err == nil {
		t.Error("expected error for path containing a comma, got nil")
	}

	if got := readConfigFile(t, configFileAbs); got != iohelper.ConfigPrefix {
		t.Errorf("expected config to stay %q, got %q", iohelper.ConfigPrefix, got)
	}
}

func TestRunWithHandler_Register_RejectsDuplicatedProject(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfigFile(t, iohelper.ConfigPrefix)
	project := t.TempDir()

	if err := runTestCmd(t, configFileAbs, "register", project); err != nil {
		t.Fatalf("expected no error on first register, got %v", err)
	}
	if err := runTestCmd(t, configFileAbs, "register", project); err == nil {
		t.Error("expected error on duplicated register, got nil")
	}

	want := iohelper.ConfigPrefix + project
	if got := readConfigFile(t, configFileAbs); got != want {
		t.Errorf("expected config to stay %q, got %q", want, got)
	}
}

func TestRunWithHandler_UninitializedConfig_FailsValidation(t *testing.T) {
	t.Parallel()

	configFileAbs := writeConfigFile(t, "not a tmux-sessionizer config\n")

	// "invalid" never dispatches to fzf/tmux, so reaching ErrNoSuchCmd
	// would mean the broken config slipped through validation.
	err := runTestCmd(t, configFileAbs, "invalid")

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if errors.Is(err, ErrNoSuchCmd) {
		t.Errorf("expected validation error before dispatch, got ErrNoSuchCmd")
	}
}
