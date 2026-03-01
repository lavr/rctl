package runner

import (
	"strings"
	"testing"
)

func TestBuildEnv_Override(t *testing.T) {
	current := []string{"A=1", "B=2", "C=3"}
	overrides := map[string]string{"B": "new", "D": "4"}

	result := BuildEnv(current, overrides)

	env := envMap(result)
	if env["A"] != "1" {
		t.Errorf("A = %q, want %q", env["A"], "1")
	}
	if env["B"] != "new" {
		t.Errorf("B = %q, want %q", env["B"], "new")
	}
	if env["C"] != "3" {
		t.Errorf("C = %q, want %q", env["C"], "3")
	}
	if env["D"] != "4" {
		t.Errorf("D = %q, want %q", env["D"], "4")
	}
}

func TestBuildEnv_Empty(t *testing.T) {
	current := []string{"A=1"}
	result := BuildEnv(current, nil)

	if len(result) != 1 || result[0] != "A=1" {
		t.Errorf("result = %v, want [A=1]", result)
	}
}

func TestPrependPath(t *testing.T) {
	result := PrependPath([]string{"/custom/bin", "/other/bin"})
	if !strings.HasPrefix(result, "/custom/bin:") {
		t.Errorf("PATH should start with /custom/bin:, got %q", result)
	}
	if !strings.Contains(result, "/other/bin:") {
		t.Errorf("PATH should contain /other/bin:, got %q", result)
	}
}

func TestPrependPath_Empty(t *testing.T) {
	result := PrependPath(nil)
	if result == "" {
		t.Error("PrependPath(nil) should return current PATH, got empty")
	}
}

func envMap(env []string) map[string]string {
	m := make(map[string]string)
	for _, e := range env {
		if k, v, ok := strings.Cut(e, "="); ok {
			m[k] = v
		}
	}
	return m
}
