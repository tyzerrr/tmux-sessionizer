package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/TlexCypher/my-tmux-sessionizer/handler"
	"github.com/urfave/cli/v3"
)

const (
	ExitCodeOK        = 0
	ExitCodeError int = iota
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
		Name:   "tmux-sessionizer",
		Usage:  "tmux session manager",
		Action: run,
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	sh := handler.NewSessionHandler()
	if err := sh.Start(); err != nil {
		return fmt.Errorf("failed to run tmux-sessionizer command: %w", err)
	}
	return nil
}
