package runner

import (
	"os"
	"strings"
)

// BuildEnv creates an environment slice by taking the current process env
// and applying overrides. Keys in overrides replace existing values or are appended.
func BuildEnv(current []string, overrides map[string]string) []string {
	if len(overrides) == 0 {
		return current
	}

	// Index existing env by key
	seen := make(map[string]int, len(current))
	result := make([]string, len(current))
	copy(result, current)

	for i, entry := range result {
		if k, _, ok := strings.Cut(entry, "="); ok {
			seen[k] = i
		}
	}

	for k, v := range overrides {
		entry := k + "=" + v
		if idx, ok := seen[k]; ok {
			result[idx] = entry
		} else {
			result = append(result, entry)
			seen[k] = len(result) - 1
		}
	}

	return result
}

// PrependPath returns a new PATH value with dirs prepended to the current PATH.
func PrependPath(dirs []string) string {
	if len(dirs) == 0 {
		return os.Getenv("PATH")
	}
	current := os.Getenv("PATH")
	parts := append(dirs, current)
	return strings.Join(parts, string(os.PathListSeparator))
}
