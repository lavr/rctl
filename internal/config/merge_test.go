package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func parseNode(t *testing.T, s string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(s), &doc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0]
	}
	t.Fatal("empty document")
	return nil
}

func decodeMap(t *testing.T, n *yaml.Node) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := n.Decode(&m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return m
}

func TestMergeNodes_MapOverride(t *testing.T) {
	base := parseNode(t, `
env:
  A: "1"
  B: "2"
`)
	overlay := parseNode(t, `
env:
  B: "override"
  C: "3"
`)

	result := MergeNodes(base, overlay)
	m := decodeMap(t, result)

	env := m["env"].(map[string]interface{})
	if env["A"] != "1" {
		t.Errorf("A = %v, want 1", env["A"])
	}
	if env["B"] != "override" {
		t.Errorf("B = %v, want override", env["B"])
	}
	if env["C"] != "3" {
		t.Errorf("C = %v, want 3", env["C"])
	}
}

func TestMergeNodes_ListConcat(t *testing.T) {
	base := parseNode(t, `
default_args:
  - "--verbose"
`)
	overlay := parseNode(t, `
default_args:
  - "--debug"
`)

	result := MergeNodes(base, overlay)
	m := decodeMap(t, result)

	args := m["default_args"].([]interface{})
	if len(args) != 2 {
		t.Fatalf("args len = %d, want 2", len(args))
	}
	if args[0] != "--verbose" || args[1] != "--debug" {
		t.Errorf("args = %v, want [--verbose --debug]", args)
	}
}

func TestMergeNodes_ListOverride(t *testing.T) {
	base := parseNode(t, `
default_args:
  - "--verbose"
  - "--old"
`)
	overlay := parseNode(t, `
default_args@override:
  - "--new-only"
`)

	result := MergeNodes(base, overlay)
	m := decodeMap(t, result)

	args := m["default_args"].([]interface{})
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
	if args[0] != "--new-only" {
		t.Errorf("args[0] = %v, want --new-only", args[0])
	}
}

func TestMergeNodes_DeepMerge(t *testing.T) {
	base := parseNode(t, `
clients:
  acme:
    defaults:
      env:
        A: "1"
`)
	overlay := parseNode(t, `
clients:
  acme:
    defaults:
      env:
        B: "2"
`)

	result := MergeNodes(base, overlay)
	m := decodeMap(t, result)

	clients := m["clients"].(map[string]interface{})
	acme := clients["acme"].(map[string]interface{})
	defaults := acme["defaults"].(map[string]interface{})
	env := defaults["env"].(map[string]interface{})

	if env["A"] != "1" {
		t.Errorf("A = %v, want 1", env["A"])
	}
	if env["B"] != "2" {
		t.Errorf("B = %v, want 2", env["B"])
	}
}

func TestMergeNodes_MultiLayer(t *testing.T) {
	n1 := parseNode(t, `env: { A: "1" }`)
	n2 := parseNode(t, `env: { B: "2" }`)
	n3 := parseNode(t, `env: { A: "3", C: "3" }`)

	result := MergeMultiple([]*yaml.Node{n1, n2, n3})
	m := decodeMap(t, result)

	env := m["env"].(map[string]interface{})
	if env["A"] != "3" {
		t.Errorf("A = %v, want 3", env["A"])
	}
	if env["B"] != "2" {
		t.Errorf("B = %v, want 2", env["B"])
	}
	if env["C"] != "3" {
		t.Errorf("C = %v, want 3", env["C"])
	}
}

func TestMergeNodes_NilBase(t *testing.T) {
	overlay := parseNode(t, `key: value`)
	result := MergeNodes(nil, overlay)
	m := decodeMap(t, result)
	if m["key"] != "value" {
		t.Errorf("key = %v, want value", m["key"])
	}
}

func TestMergeNodes_NilOverlay(t *testing.T) {
	base := parseNode(t, `key: value`)
	result := MergeNodes(base, nil)
	m := decodeMap(t, result)
	if m["key"] != "value" {
		t.Errorf("key = %v, want value", m["key"])
	}
}
