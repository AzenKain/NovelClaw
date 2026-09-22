//go:build windows

package main

func checkAndElevate() {
	// On Windows, UAC and privileges can be managed via manifest or are not automatically elevated through osascript/pkexec equivalent.
}

func ensureDataDir() {
	// On Windows, data dir is "." (current directory), no setup needed.
}
