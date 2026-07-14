package io

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Filer struct {
}

func NewFiler() *Filer {
	return &Filer{}
}

func (fl *Filer) ExpandTildeAsHomeDir(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	trimmed := strings.TrimPrefix(path, "~")

	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userHome, trimmed), nil
}

// Exists reports why the path cannot be used: missing or unreadable.
// Stat is enough here; opening the file would consume a descriptor for nothing.
func (fl *Filer) Exists(path string) error {
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%s does not exist:%w", path, err)
	} else if errors.Is(err, fs.ErrPermission) {
		return fmt.Errorf("you don't have enough permission to %s:%w", path, err)
	} else if err != nil {
		return err
	}
	return nil
}
