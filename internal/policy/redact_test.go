package policy

import (
	"testing"
)

func TestRedactEnv_GlobPatterns(t *testing.T) {
	env := map[string]string{
		"PATH":          "/usr/bin",
		"AWS_TOKEN":     "secret123",
		"DB_PASSWORD":   "pass123",
		"MY_SECRET_KEY": "key123",
		"NORMAL_VAR":    "value",
	}

	patterns := []string{"*_TOKEN", "*PASSWORD*", "*SECRET*"}
	result := RedactEnv(env, patterns)

	if result["PATH"] != "/usr/bin" {
		t.Errorf("PATH should not be redacted, got %q", result["PATH"])
	}
	if result["AWS_TOKEN"] != redactedValue {
		t.Errorf("AWS_TOKEN should be redacted, got %q", result["AWS_TOKEN"])
	}
	if result["DB_PASSWORD"] != redactedValue {
		t.Errorf("DB_PASSWORD should be redacted, got %q", result["DB_PASSWORD"])
	}
	if result["MY_SECRET_KEY"] != redactedValue {
		t.Errorf("MY_SECRET_KEY should be redacted, got %q", result["MY_SECRET_KEY"])
	}
	if result["NORMAL_VAR"] != "value" {
		t.Errorf("NORMAL_VAR should not be redacted, got %q", result["NORMAL_VAR"])
	}
}

func TestRedactEnv_NoPatterns(t *testing.T) {
	env := map[string]string{"TOKEN": "secret"}
	result := RedactEnv(env, nil)
	if result["TOKEN"] != "secret" {
		t.Errorf("TOKEN should not be redacted when no patterns, got %q", result["TOKEN"])
	}
}

func TestRedactArgv_FlagValue(t *testing.T) {
	argv := []string{"cmd", "--token", "secret123", "--verbose", "--password", "pass456"}
	result := RedactArgv(argv)

	if result[0] != "cmd" {
		t.Errorf("result[0] = %q, want cmd", result[0])
	}
	if result[1] != "--token" {
		t.Errorf("result[1] = %q, want --token", result[1])
	}
	if result[2] != redactedValue {
		t.Errorf("result[2] should be redacted, got %q", result[2])
	}
	if result[3] != "--verbose" {
		t.Errorf("result[3] = %q, want --verbose", result[3])
	}
	if result[5] != redactedValue {
		t.Errorf("result[5] should be redacted, got %q", result[5])
	}
}

func TestRedactArgv_FlagEqualValue(t *testing.T) {
	argv := []string{"cmd", "--token=secret123", "--verbose"}
	result := RedactArgv(argv)

	if result[1] != "--token="+redactedValue {
		t.Errorf("result[1] = %q, want --token=%s", result[1], redactedValue)
	}
	if result[2] != "--verbose" {
		t.Errorf("result[2] = %q, want --verbose", result[2])
	}
}

func TestRedactArgv_Empty(t *testing.T) {
	result := RedactArgv(nil)
	if result != nil {
		t.Errorf("result = %v, want nil", result)
	}
}

func TestRedactArgv_NoSensitive(t *testing.T) {
	argv := []string{"echo", "hello", "world"}
	result := RedactArgv(argv)

	for i, v := range argv {
		if result[i] != v {
			t.Errorf("result[%d] = %q, want %q", i, result[i], v)
		}
	}
}
