package testutils

import (
	"os"

	"github.com/samber/lo"
)

func GetUserHomeDir() string {
	return lo.Must(os.UserHomeDir())
}
