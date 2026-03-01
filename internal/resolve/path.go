package resolve

import (
	"os"
	"path/filepath"
	"runtime"
)

// findExecutable searches for an executable file named name in the given directory.
// Returns the absolute path if found and executable, empty string otherwise.
func findExecutable(dir, name string) string {
	p := filepath.Join(dir, name)
	if isExecutable(p) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return p
		}
		return abs
	}
	return ""
}

// isExecutable checks if the given path exists and is executable.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0111 != 0
}
