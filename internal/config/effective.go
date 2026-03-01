package config

// ComputeEffective merges three layers: global defaults, client defaults, and domain config.
// Maps: keys from later layers override earlier.
// Lists: later layers are appended.
// Scalars: later non-zero values override.
func ComputeEffective(defaults, clientDefaults DomainConfig, domain DomainConfig) *EffectiveConfig {
	eff := &EffectiveConfig{
		Env:      make(map[string]string),
		Profiles: make(map[string]Profile),
		Vars:     make(map[string]string),
	}

	// Apply layers in order: defaults -> clientDefaults -> domain
	layers := []DomainConfig{defaults, clientDefaults, domain}
	for _, layer := range layers {
		// Dir: last non-empty wins
		if layer.Dir != "" {
			eff.Dir = layer.Dir
		}

		// Env: map merge (later overrides)
		for k, v := range layer.Env {
			eff.Env[k] = v
		}

		// Lists: concat
		eff.DefaultArgs = append(eff.DefaultArgs, layer.DefaultArgs...)
		eff.PathAdd = append(eff.PathAdd, layer.PathAdd...)
		eff.CommandsPath = append(eff.CommandsPath, layer.CommandsPath...)

		// Profiles: map merge (later overrides)
		for k, v := range layer.Profiles {
			eff.Profiles[k] = v
		}

		// Vars: map merge (later overrides)
		for k, v := range layer.Vars {
			eff.Vars[k] = v
		}

		// TaskRunner: last non-empty wins
		if layer.TaskRunner != "" {
			eff.TaskRunner = layer.TaskRunner
		}
	}

	return eff
}
