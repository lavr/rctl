package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// loadIncludeNodes loads all include files matching the given patterns and
// returns their parsed yaml.Node documents.
func loadIncludeNodes(patterns []string) ([]*yaml.Node, error) {
	var nodes []*yaml.Node

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid include glob %q: %w", pattern, err)
		}
		sort.Strings(matches)

		for _, path := range matches {
			node, err := loadYAMLNode(path)
			if err != nil {
				return nil, fmt.Errorf("loading include %s: %w", path, err)
			}
			if node != nil {
				nodes = append(nodes, node)
			}
		}
	}

	return nodes, nil
}

// loadYAMLNode reads a file and returns its root yaml.Node.
func loadYAMLNode(path string) (*yaml.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	// yaml.Unmarshal wraps in a document node
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0], nil
	}
	return nil, nil
}
