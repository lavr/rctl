package config

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// MergeNodes merges overlay into base using the following rules:
// - Maps: key-wise merge, overlay keys override base keys
// - Lists: concatenation by default, unless the key has "@override" suffix
// - Scalars: overlay replaces base
//
// The @override mechanism: if overlay contains a key like "default_args@override",
// it replaces the base "default_args" entirely instead of concatenating.
func MergeNodes(base, overlay *yaml.Node) *yaml.Node {
	if base == nil {
		return overlay
	}
	if overlay == nil {
		return base
	}

	// Resolve aliases
	base = resolveAlias(base)
	overlay = resolveAlias(overlay)

	switch {
	case base.Kind == yaml.MappingNode && overlay.Kind == yaml.MappingNode:
		return mergeMappings(base, overlay)
	case base.Kind == yaml.SequenceNode && overlay.Kind == yaml.SequenceNode:
		return concatSequences(base, overlay)
	default:
		// Scalar or different kinds: overlay wins
		return overlay
	}
}

func mergeMappings(base, overlay *yaml.Node) *yaml.Node {
	result := &yaml.Node{
		Kind:    yaml.MappingNode,
		Tag:     base.Tag,
		Style:   base.Style,
		Content: make([]*yaml.Node, 0, len(base.Content)+len(overlay.Content)),
	}

	// Collect override keys from overlay (keys ending with @override)
	overrideKeys := make(map[string]bool)
	for i := 0; i < len(overlay.Content)-1; i += 2 {
		key := overlay.Content[i].Value
		if strings.HasSuffix(key, "@override") {
			baseKey := strings.TrimSuffix(key, "@override")
			overrideKeys[baseKey] = true
		}
	}

	// Index overlay keys for quick lookup
	overlayIndex := make(map[string]int) // key -> index of value in overlay.Content
	for i := 0; i < len(overlay.Content)-1; i += 2 {
		key := overlay.Content[i].Value
		overlayIndex[key] = i + 1
	}

	// Track which overlay keys have been used
	used := make(map[string]bool)

	// Process base keys
	for i := 0; i < len(base.Content)-1; i += 2 {
		keyNode := base.Content[i]
		baseVal := base.Content[i+1]
		key := keyNode.Value

		// Check if overlay has this key with @override
		if overrideKeys[key] {
			overrideKey := key + "@override"
			if idx, ok := overlayIndex[overrideKey]; ok {
				result.Content = append(result.Content, keyNode, overlay.Content[idx])
				used[overrideKey] = true
				used[key] = true // mark base key as handled
				continue
			}
		}

		// Check if overlay has this key (normal merge)
		if idx, ok := overlayIndex[key]; ok {
			merged := MergeNodes(baseVal, overlay.Content[idx])
			result.Content = append(result.Content, keyNode, merged)
			used[key] = true
		} else {
			result.Content = append(result.Content, keyNode, baseVal)
		}
	}

	// Add overlay-only keys (not yet used, and skip @override keys themselves)
	for i := 0; i < len(overlay.Content)-1; i += 2 {
		key := overlay.Content[i].Value
		if !used[key] && !strings.HasSuffix(key, "@override") {
			result.Content = append(result.Content, overlay.Content[i], overlay.Content[i+1])
		}
	}

	return result
}

func concatSequences(base, overlay *yaml.Node) *yaml.Node {
	result := &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     base.Tag,
		Style:   base.Style,
		Content: make([]*yaml.Node, 0, len(base.Content)+len(overlay.Content)),
	}
	result.Content = append(result.Content, base.Content...)
	result.Content = append(result.Content, overlay.Content...)
	return result
}

func resolveAlias(n *yaml.Node) *yaml.Node {
	if n.Kind == yaml.AliasNode {
		return n.Alias
	}
	return n
}

// MergeMultiple merges multiple yaml.Node documents in order.
func MergeMultiple(nodes []*yaml.Node) *yaml.Node {
	if len(nodes) == 0 {
		return nil
	}
	result := nodes[0]
	for _, n := range nodes[1:] {
		result = MergeNodes(result, n)
	}
	return result
}
