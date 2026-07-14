package validate

import (
	"errors"
	"fmt"
	"os"

	"github.com/TlexCypher/my-tmux-sessionizer/internal/io"
)

func ValidateConfig(configFileAbs string) error {
	f, err := os.Open(configFileAbs)
	if err != nil {
		// Without a readable file the content checks below are meaningless.
		return fmt.Errorf("config file could not be opened:%w", err)
	}
	defer f.Close()

	var errs []error

	buf := make([]byte, len(io.ConfigPrefix))
	if _, err := f.ReadAt(buf, 0); err != nil {
		errs = append(errs, fmt.Errorf("failed to get first %d bytes:%w", len(io.ConfigPrefix), err))
	}
	if string(buf) != io.ConfigPrefix {
		errs = append(errs, fmt.Errorf("config prefix must be %s, you need to initialize config file with 'tmux-sessionizer init'", io.ConfigPrefix))
	}

	return errors.Join(errs...)
}
