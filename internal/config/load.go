package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "~/.config/rctl/config.yaml"

// DefaultConfigPath returns the default config file path with ~ expanded.
func DefaultConfigPath() string {
	return ExpandTilde(defaultConfigPath)
}

// Load reads and parses a YAML config file using the node-based merge pipeline:
// 1. Load base config as yaml.Node
// 2. Extract and expand includes, load include nodes
// 3. Merge: base -> includes
// 4. Decode merged node into Root struct
// 5. Validate version and expand paths
func Load(path string) (*Root, error) {
	// Step 1: Load base config as node
	baseNode, err := loadYAMLNode(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	if baseNode == nil {
		return nil, fmt.Errorf("empty config file: %s", path)
	}

	// Step 2: Pre-parse to extract includes (need to know them before full merge)
	var preRoot struct {
		Version  int      `yaml:"version"`
		Includes []string `yaml:"includes,omitempty"`
	}
	data, _ := os.ReadFile(path)
	if err := yaml.Unmarshal(data, &preRoot); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	if preRoot.Version != 1 {
		return nil, fmt.Errorf("unsupported config version %d in %s (expected 1)", preRoot.Version, path)
	}

	// Expand includes
	var expandedIncludes []string
	for _, inc := range preRoot.Includes {
		expandedIncludes = append(expandedIncludes, ExpandTilde(inc))
	}

	// Load include nodes
	includeNodes, err := loadIncludeNodes(expandedIncludes)
	if err != nil {
		return nil, err
	}

	// Step 3: Merge all nodes
	allNodes := make([]*yaml.Node, 0, 1+len(includeNodes))
	allNodes = append(allNodes, baseNode)
	allNodes = append(allNodes, includeNodes...)

	merged := MergeMultiple(allNodes)

	// Step 4: Decode merged node into Root
	var root Root
	if err := merged.Decode(&root); err != nil {
		return nil, fmt.Errorf("decoding merged config: %w", err)
	}

	// Step 5: Validate and expand
	if root.Version != 1 {
		return nil, fmt.Errorf("unsupported config version %d (expected 1)", root.Version)
	}
	expandRootPaths(&root)

	return &root, nil
}

// LoadSimple reads a single YAML config without includes.
// Used for testing and simple scenarios.
func LoadSimple(path string) (*Root, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var root Root
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	if root.Version != 1 {
		return nil, fmt.Errorf("unsupported config version %d in %s (expected 1)", root.Version, path)
	}
	expandRootPaths(&root)
	return &root, nil
}
