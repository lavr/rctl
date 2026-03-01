package policy

import (
	"path/filepath"
	"strings"
)

const redactedValue = "***REDACTED***"

// RedactEnv redacts environment variable values whose keys match any of the given patterns.
// Patterns support glob matching (e.g., "*_TOKEN", "*SECRET*", "*PASSWORD*").
func RedactEnv(env map[string]string, patterns []string) map[string]string {
	if len(patterns) == 0 {
		return env
	}

	result := make(map[string]string, len(env))
	for k, v := range env {
		if matchesAny(k, patterns) {
			result[k] = redactedValue
		} else {
			result[k] = v
		}
	}
	return result
}

// RedactArgv redacts values that follow sensitive flag names.
// Sensitive flags: --token, --password, --secret, --key, --api-key (and single-dash variants).
func RedactArgv(argv []string) []string {
	if len(argv) == 0 {
		return argv
	}

	sensitiveFlags := map[string]bool{
		"--token":    true,
		"--password": true,
		"--secret":   true,
		"--key":      true,
		"--api-key":  true,
		"-token":     true,
		"-password":  true,
		"-secret":    true,
		"-key":       true,
	}

	result := make([]string, len(argv))
	copy(result, argv)

	for i := 0; i < len(result)-1; i++ {
		flag := strings.ToLower(result[i])
		if sensitiveFlags[flag] {
			result[i+1] = redactedValue
		}
		// Also handle --flag=value form
		if eqIdx := strings.Index(result[i], "="); eqIdx > 0 {
			flagPart := strings.ToLower(result[i][:eqIdx])
			if sensitiveFlags[flagPart] {
				result[i] = result[i][:eqIdx+1] + redactedValue
			}
		}
	}

	return result
}

func matchesAny(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}
