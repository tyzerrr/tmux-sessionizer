package main

import (
	"log/slog"
	"os"

	"github.com/TlexCypher/my-tmux-sessionizer/cmd"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	os.Exit(cmd.Core())
}
