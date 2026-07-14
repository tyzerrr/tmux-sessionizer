package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/TlexCypher/my-tmux-sessionizer/handler"
	iohelper "github.com/TlexCypher/my-tmux-sessionizer/internal/io"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/session"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/tmux"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/types"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/validate"
	"github.com/urfave/cli/v3"
)

const (
	ExitCodeOK int = iota
	ExitCodeError
)

const (
	configFile = "~/.tmux-sessionizer"
)

const (
	CommandName  = "tmux-sessionizer"
	CommandUsage = "tmux session manager"
)

var (
	ErrNoSuchCmd = errors.New("no such command")
)

func Core() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cmd := newCmd()

	err := cmd.Run(ctx, os.Args)
	if err != nil {
		fmt.Println(err)
		return ExitCodeError
	}

	return ExitCodeOK
}

func newCmd() *cli.Command {
	return &cli.Command{
		Name:   CommandName,
		Usage:  CommandUsage,
		Action: run,
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	filer := iohelper.NewFiler()
	configFile, err := filer.ExpandTildeAsHomeDir(configFile)
	if err != nil {
		return err
	}
	configFileAbs, err := filepath.Abs(configFile)
	if err != nil {
		return err
	}

	return runWithHandler(ctx, cmd, filer, configFileAbs)
}

func runWithHandler(
	ctx context.Context,
	cmd *cli.Command,
	filer *iohelper.Filer,
	configFileAbs string,
) error {
	args := cmd.Args().Slice()
	// initialization does not need config file validation
	ph := handler.NewProjectHandler(configFileAbs)
	if len(args) == 1 && args[0] == "init" {
		return ph.Init(ctx, configFileAbs)
	}
	config, err := readConfig(ctx, filer, configFileAbs)
	if err != nil {
		return fmt.Errorf("failed to read config:%w", err)
	}
	// register does not require to gather tmux sessions
	if len(args) == 2 && args[0] == "register" {
		return registerProject(ctx, ph, filer, config, args[1])
	}

	sh := buildSessionHandler(ctx, config)
	if len(args) == 1 && args[0] == "list" {
		return sh.GrabExistingSession(ctx)
	} else if len(args) == 1 && args[0] == "delete" {
		return sh.DeleteSessions(ctx)
	} else if len(args) > 0 {
		return ErrNoSuchCmd
	} else {
		return sh.NewSession(ctx)
	}
}

func buildSessionHandler(ctx context.Context, config *iohelper.Config) handler.ISessionHandler {
	tmux := tmux.NewTmux()
	sessions, err := tmux.GatherExistingSessions(ctx)
	if err != nil {
		// tmux may simply not be running yet; start from an empty session map.
		sessions = make(map[types.String]*session.Session, 0)
	}

	sessionNameTransformer := session.NewTransformer().WithRule(
		session.NewTransformRule(
			func(in string) string { return strings.ReplaceAll(in, ".", "_") },
			func(in string) string { return strings.ReplaceAll(in, "_", ".") },
		),
		session.NewTransformRule(
			func(in string) string { return strings.ReplaceAll(in, ":", ";") },
			func(in string) string { return strings.ReplaceAll(in, ";", ":") },
		),
	)
	sm := session.NewSessionManager(sessions, sessionNameTransformer)
	return handler.NewSessionHandler(config, sm, tmux)
}

func readConfig(_ context.Context, filer *iohelper.Filer, configFileAbs string) (*iohelper.Config, error) {
	if err := validate.ValidateConfig(configFileAbs); err != nil {
		return nil, fmt.Errorf("failed to validate config file:%w", err)
	}

	configParser := iohelper.NewConfigParser()
	config, err := configParser.ReadConfig(filer, configFileAbs)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func registerProject(
	ctx context.Context,
	ph *handler.ProjectHandler,
	filer *iohelper.Filer,
	config *iohelper.Config,
	rawPath string,
) error {
	registerPath, err := filer.ExpandTildeAsHomeDir(rawPath)
	if err != nil {
		return fmt.Errorf("failed to register %s as a tmux-sessionizer project:%w", rawPath, err)
	}
	// Tilde expansion alone keeps plain relative paths relative;
	// Register requires an absolute path so it never depends on a later CWD.
	registerAbs, err := filepath.Abs(registerPath)
	if err != nil {
		return fmt.Errorf("failed to convert %s to absolute path:%w", registerPath, err)
	}

	// A comma is the config entry separator and cannot be escaped, so a
	// path containing one would split into bogus entries on the next read.
	if strings.Contains(registerAbs, ",") {
		return fmt.Errorf("project path %s must not contain ',', the config file separator", registerAbs)
	}

	register := types.NewString(registerAbs)
	for _, project := range config.Projects {
		if register.Value() == project.Value() {
			return errors.New("tmux-sessionizer does not allow duplicated project registration")
		}
	}
	return ph.Register(ctx, registerAbs)
}
