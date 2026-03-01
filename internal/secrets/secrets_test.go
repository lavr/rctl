package secrets

import (
	"os"
	"testing"
)

func TestEnvProvider_Get(t *testing.T) {
	os.Setenv("TEST_SECRET_123", "my-secret-value")
	defer os.Unsetenv("TEST_SECRET_123")

	p := &EnvProvider{}
	val, err := p.Get("TEST_SECRET_123")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "my-secret-value" {
		t.Errorf("val = %q, want %q", val, "my-secret-value")
	}
}

func TestEnvProvider_NotSet(t *testing.T) {
	os.Unsetenv("TEST_MISSING_SECRET_XYZ")

	p := &EnvProvider{}
	_, err := p.Get("TEST_MISSING_SECRET_XYZ")
	if err == nil {
		t.Error("expected error for missing env var")
	}
}

func TestRegistry_Caching(t *testing.T) {
	registry := NewRegistry()

	callCount := 0
	registry.Register(&mockProvider{
		name: "mock",
		getFunc: func(path string) (string, error) {
			callCount++
			return "cached-value", nil
		},
	})

	// First call
	v1, err := registry.Get("mock", "key1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if v1 != "cached-value" {
		t.Errorf("v1 = %q, want cached-value", v1)
	}

	// Second call should use cache
	v2, err := registry.Get("mock", "key1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if v2 != "cached-value" {
		t.Errorf("v2 = %q, want cached-value", v2)
	}

	if callCount != 1 {
		t.Errorf("provider called %d times, want 1 (caching)", callCount)
	}
}

func TestRegistry_UnknownProvider(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("nonexistent", "key")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestRegistry_SecretFunc(t *testing.T) {
	os.Setenv("TEST_FUNC_SECRET", "func-value")
	defer os.Unsetenv("TEST_FUNC_SECRET")

	registry := NewRegistry()
	registry.Register(&EnvProvider{})

	fn := registry.SecretFunc()
	val, err := fn("env", "TEST_FUNC_SECRET")
	if err != nil {
		t.Fatalf("SecretFunc: %v", err)
	}
	if val != "func-value" {
		t.Errorf("val = %q, want func-value", val)
	}
}

func TestCommandProvider_Echo(t *testing.T) {
	p := &CommandProvider{
		ProviderName: "command:test",
		CmdTemplate:  "echo {{ .Path }}",
	}

	val, err := p.Get("hello-secret")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "hello-secret" {
		t.Errorf("val = %q, want hello-secret", val)
	}
}

type mockProvider struct {
	name    string
	getFunc func(string) (string, error)
}

func (p *mockProvider) Name() string                 { return p.name }
func (p *mockProvider) Get(path string) (string, error) { return p.getFunc(path) }
