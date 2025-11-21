package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

func findFileInPath(file string) (string, error) {
	pathEnv := os.Getenv("PATH")
	dirs := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			var pathErr *fs.PathError
			if errors.As(err, &pathErr) {
				continue
			}
			return "", err
		}

		// 遍历所有条目
		for _, entry := range entries {
			if entry.Name() == file {
				absPath := filepath.Join(dir, file)
				hasPerm, err := hasExecutePermission(absPath)
				if err != nil {
					return "", err
				}
				if hasPerm {
					return absPath, nil
				}
			}
		}
	}
	return "", nil
}

func hasExecutePermission(filePath string) (bool, error) {
	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	// Directories are not executable files (Unix: execute = enter directory)
	if fileInfo.IsDir() {
		return false, fmt.Errorf("path is a directory, not a file")
	}

	switch runtime.GOOS {
	case "windows":
		return isWindowsExecutable(filePath)
	case "linux", "darwin", "freebsd", "openbsd", "netbsd":
		return isUnixExecutable(fileInfo)
	default:
		return false, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func isUnixExecutable(fileInfo os.FileInfo) (bool, error) {
	perm := fileInfo.Mode().Perm()
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("failed to get system stat")
	}

	currentUID := os.Getuid()
	currentGID := os.Getgid()

	// Check owner, group, or others execute permission
	if stat.Uid == uint32(currentUID) && (perm&0100) != 0 { // Owner
		return true, nil
	} else if stat.Gid == uint32(currentGID) && (perm&0010) != 0 { // Group
		return true, nil
	} else if (perm & 0001) != 0 { // Others
		return true, nil
	}

	return false, nil
}

func isWindowsExecutable(filePath string) (bool, error) {
	// Windows recognizes executable extensions
	execExts := map[string]bool{
		".exe": true,
		".bat": true,
		".cmd": true,
		".com": true,
		".ps1": true,
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if !execExts[ext] {
		return false, nil
	}

	// Optional: Verify access rights via Windows API (advanced)
	// Using syscall to check if the current user can execute
	// (Simplified here; use golang.org/x/sys/windows for full ACL checks)
	return true, nil
}
