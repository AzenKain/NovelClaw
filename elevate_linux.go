//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func checkAndElevate() {
	dataDir := GetDataDir()

	if dataDir == "" {
		return
	}

	absDir, err := filepath.Abs(dataDir)
	if err != nil {
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	absHome, err := filepath.Abs(homeDir)
	if err != nil {
		return
	}

	if !strings.HasPrefix(absDir, absHome+string(os.PathSeparator)) {
		fmt.Printf("Refusing to repair permissions outside user home: %s\n", absDir)
		return
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return
	}

	testFile := filepath.Join(absDir, ".permissions_test")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err == nil {
		_ = os.Remove(testFile)
		return
	}

	out, err := exec.Command("id", "-un").Output()
	if err != nil {
		return
	}

	currentUser := strings.TrimSpace(string(out))

	cmdChown := exec.Command("pkexec", "chown", "-R", currentUser, absDir)
	if err := cmdChown.Run(); err != nil {
		fmt.Printf("Failed to repair owner permissions on Linux: %v\n", err)
		return
	}

	cmdChmod := exec.Command("pkexec", "chmod", "-R", "u+rwX", absDir)
	if err := cmdChmod.Run(); err != nil {
		fmt.Printf("Failed to repair mode permissions on Linux: %v\n", err)
	}
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
