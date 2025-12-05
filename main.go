package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/TlexCypher/my-tmux-sessionizer/cmd"
)

var version = "v0.0.0"

func getVersion() string {
	if version != "" {
		return version
	}

	i, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	return i.Main.Version
}

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Print the version and exit")
	flag.Parse()

	if showVersion {
		fmt.Println(getVersion())
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	os.Exit(cmd.Core())
}
