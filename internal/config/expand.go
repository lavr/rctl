package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandTilde replaces a leading "~/" with the user's home directory.
func ExpandTilde(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// expandPathFields expands ~ in all path-like fields of a DomainConfig.
func expandPathFields(dc *DomainConfig) {
	dc.Dir = ExpandTilde(dc.Dir)
	for i, p := range dc.PathAdd {
		dc.PathAdd[i] = ExpandTilde(p)
	}
	for i, p := range dc.CommandsPath {
		dc.CommandsPath[i] = ExpandTilde(p)
	}
}

// expandRootPaths expands ~ in all path-like fields throughout the Root config.
func expandRootPaths(root *Root) {
	expandPathFields(&root.Defaults)
	if root.Defaults.Policy != nil && root.Defaults.Policy.Audit != nil {
		root.Defaults.Policy.Audit.File = ExpandTilde(root.Defaults.Policy.Audit.File)
	}
	for _, c := range root.Clients {
		expandPathFields(&c.Defaults)
		for name, d := range c.Domains {
			expandPathFields(&d.DomainConfig)
			c.Domains[name] = d
		}
	}
	for i, inc := range root.Includes {
		root.Includes[i] = ExpandTilde(inc)
	}
}
