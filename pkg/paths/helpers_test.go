package paths

import (
	"os"
)

func homeForTest() (string, error) { return os.UserHomeDir() }

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}
