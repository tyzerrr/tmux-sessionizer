package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/TlexCypher/my-tmux-sessionizer/handler"
	iohelper "github.com/TlexCypher/my-tmux-sessionizer/internal/io"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/session"
	"github.com/TlexCypher/my-tmux-sessionizer/internal/tmux"
	"github.com/urfave/cli/v3"
)

const (
	ExitCodeOK        = 0
	ExitCodeError int = iota
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
	if err := cmd.Run(ctx, os.Args); err != nil {
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
	filepathResolver := iohelper.NewFilePathResolver()
	configFile, err := filepathResolver.ExpandTildeAsHomeDir(configFile)
	if err != nil {
		return err
	}
	configParser := iohelper.NewConfigParser()
	configFileAbs, err := filepath.Abs(configFile)
	if err != nil {
		return err
	}

	config, err := configParser.ReadConfig(configFileAbs)
	if err != nil {
		return err
	}
	tmux := tmux.NewTmux()
	sessions, err := tmux.GatherExistingSessions()
	if err != nil {
		fmt.Println("Warning: could not gather existing tmux sessions:", err)
		return err
	}
	sm := session.NewSessionManager(sessions)
	sh := handler.NewSessionHandler(config, sm, tmux)
	return runWithHandler(sh, ctx, cmd)
}

func runWithHandler(h handler.ISessionHandler, ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) > 0 && args[0] == "list" {
		return h.GrabExistingSession(ctx)
	} else if len(args) > 0 {
		return ErrNoSuchCmd
	} else {
		return h.NewSession(ctx)
	}
}
