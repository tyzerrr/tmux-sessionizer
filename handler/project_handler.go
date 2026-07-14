package handler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/io"
)

const configFilePermission = 0o644

type ProjectHandler struct {
	configFile string
}

func NewProjectHandler(configFile string) *ProjectHandler {
	return &ProjectHandler{
		configFile: configFile,
	}
}

func (ph *ProjectHandler) Init(ctx context.Context, configFileAbs string) error {
	// Check if configFile does not exist.
	if _, err := os.Stat(configFileAbs); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to get status of config file:%w", err)
	}

	// If configfile does not exist, create it.
	f, err := os.Create(configFileAbs)
	if err != nil {
		return fmt.Errorf("failed to create config file of tmux-sessionizer:%w", err)
	}
	defer f.Close()

	// On init phase, we only add prefix(default=) to config file.
	if _, err = f.WriteString(io.ConfigPrefix); err != nil {
		return fmt.Errorf("failed to add config prefix to newly created config file:%w", err)
	}
	return nil
}

func (ph *ProjectHandler) Register(ctx context.Context, projectPathAbs string) error {
	// NOTE: tmux-sessionizer only allows to handle directory as a project, so at first we need to check the status.
	fi, err := os.Stat(projectPathAbs)
	if err != nil {
		return fmt.Errorf("failed to get status of file:%w", err)
	}
	if !fi.IsDir() {
		return errors.New("tmux-sessionizer does not allow to register file as a project")
	}

	// read config file
	f, err := os.OpenFile(ph.configFile, os.O_APPEND|os.O_RDWR, configFilePermission)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// The config file may end with \n, so we need to truncate new line.
	if err := ph.truncate(ctx, f); err != nil {
		return fmt.Errorf("failed to truncate last /\n:%w", err)
	}

	// Append new directory to config file. To improve performance, using strings.Builder
	if err := ph.appendProject(ctx, f, projectPathAbs); err != nil {
		return fmt.Errorf("failed to append project to configFile:%w", err)
	}
	return nil
}

func (ph *ProjectHandler) truncate(_ context.Context, f *os.File) error {
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get status of file: %w", err)
	}
	if size := info.Size(); size > 0 {
		last := make([]byte, 1)
		if _, err := f.ReadAt(last, size-1); err != nil {
			return fmt.Errorf("failed to read last byte:%w", err)
		}
		if last[0] == '\n' {
			if err := f.Truncate(size - 1); err != nil {
				return fmt.Errorf("failed to truncate last trailing new line:%w", err)
			}
		}
	}
	return nil
}

func (ph *ProjectHandler) appendProject(_ context.Context, f *os.File, projectPath string) error {
	// Stat the handle, not the path: the size must reflect the file this
	// handle just truncated, even if the path was swapped out meanwhile.
	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get status of project config file:%w", err)
	}

	// If any projects has not been registered to config file, we don't need to add ",".
	var builder strings.Builder
	if fi.Size() != int64(len(io.ConfigPrefix)) {
		if _, err := builder.WriteString(","); err != nil {
			return fmt.Errorf("failed to write ,:%w", err)
		}
	}

	_, err = builder.WriteString(projectPath)
	if err != nil {
		return fmt.Errorf("failed to write %s:%w", projectPath, err)
	}
	if _, err = f.WriteString(builder.String()); err != nil {
		return fmt.Errorf("failed to register new project to config file:%w", err)
	}
	return nil
}
