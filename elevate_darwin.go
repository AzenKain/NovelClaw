//go:build darwin

package main

import (
	"os"
	"path/filepath"
)

func checkAndElevate() {
	dataDir := GetDataDir()
	if dataDir == "" || dataDir == "." {
		return
	}

	absDir, err := filepath.Abs(dataDir)
	if err != nil {
		return
	}

	_ = os.MkdirAll(absDir, 0o755)
}

func ensureDataDir() {
	dataDir := GetDataDir()
	if dataDir == "" || dataDir == "." {
		return
	}

	subdirs := []string{
		"",
		"skills",
		"souls",
		"data",
	}

	for _, sub := range subdirs {
		dir := filepath.Join(dataDir, sub)
		_ = os.MkdirAll(dir, 0o755)
	}
}
