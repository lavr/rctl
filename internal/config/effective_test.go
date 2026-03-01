package config

import (
	"testing"
)

func TestComputeEffective_EnvOverride(t *testing.T) {
	defaults := DomainConfig{
		Env: map[string]string{"A": "1", "B": "2"},
	}
	clientDefaults := DomainConfig{
		Env: map[string]string{"B": "override", "C": "3"},
	}
	domain := DomainConfig{
		Env: map[string]string{"A": "final"},
	}

	eff := ComputeEffective(defaults, clientDefaults, domain)

	if eff.Env["A"] != "final" {
		t.Errorf("Env[A] = %q, want %q", eff.Env["A"], "final")
	}
	if eff.Env["B"] != "override" {
		t.Errorf("Env[B] = %q, want %q", eff.Env["B"], "override")
	}
	if eff.Env["C"] != "3" {
		t.Errorf("Env[C] = %q, want %q", eff.Env["C"], "3")
	}
}

func TestComputeEffective_ListConcat(t *testing.T) {
	defaults := DomainConfig{
		DefaultArgs:  []string{"--verbose"},
		CommandsPath: []string{"/usr/share/rctl"},
	}
	clientDefaults := DomainConfig{
		DefaultArgs:  []string{"--color"},
		CommandsPath: []string{"/opt/rctl"},
	}
	domain := DomainConfig{
		DefaultArgs: []string{"--debug"},
	}

	eff := ComputeEffective(defaults, clientDefaults, domain)

	wantArgs := []string{"--verbose", "--color", "--debug"}
	if len(eff.DefaultArgs) != len(wantArgs) {
		t.Fatalf("DefaultArgs = %v, want %v", eff.DefaultArgs, wantArgs)
	}
	for i, v := range wantArgs {
		if eff.DefaultArgs[i] != v {
			t.Errorf("DefaultArgs[%d] = %q, want %q", i, eff.DefaultArgs[i], v)
		}
	}

	wantPaths := []string{"/usr/share/rctl", "/opt/rctl"}
	if len(eff.CommandsPath) != len(wantPaths) {
		t.Fatalf("CommandsPath = %v, want %v", eff.CommandsPath, wantPaths)
	}
}

func TestComputeEffective_DirOverride(t *testing.T) {
	defaults := DomainConfig{Dir: "/default"}
	clientDefaults := DomainConfig{}
	domain := DomainConfig{Dir: "/domain"}

	eff := ComputeEffective(defaults, clientDefaults, domain)

	if eff.Dir != "/domain" {
		t.Errorf("Dir = %q, want %q", eff.Dir, "/domain")
	}
}

func TestComputeEffective_EmptyLayers(t *testing.T) {
	eff := ComputeEffective(DomainConfig{}, DomainConfig{}, DomainConfig{})

	if eff.Dir != "" {
		t.Errorf("Dir = %q, want empty", eff.Dir)
	}
	if len(eff.Env) != 0 {
		t.Errorf("Env = %v, want empty", eff.Env)
	}
	if len(eff.DefaultArgs) != 0 {
		t.Errorf("DefaultArgs = %v, want empty", eff.DefaultArgs)
	}
}

func TestComputeEffective_ProfilesMerge(t *testing.T) {
	defaults := DomainConfig{
		Profiles: map[string]Profile{
			"up":   {Cmd: "default-up"},
			"down": {Cmd: "default-down"},
		},
	}
	clientDefaults := DomainConfig{}
	domain := DomainConfig{
		Profiles: map[string]Profile{
			"up": {Cmd: "domain-up", Args: []string{"--fast"}},
		},
	}

	eff := ComputeEffective(defaults, clientDefaults, domain)

	if eff.Profiles["up"].Cmd != "domain-up" {
		t.Errorf("profile up cmd = %q, want %q", eff.Profiles["up"].Cmd, "domain-up")
	}
	if eff.Profiles["down"].Cmd != "default-down" {
		t.Errorf("profile down cmd = %q, want %q", eff.Profiles["down"].Cmd, "default-down")
	}
}

func TestComputeEffective_TaskRunnerOverride(t *testing.T) {
	defaults := DomainConfig{TaskRunner: "make"}
	clientDefaults := DomainConfig{}
	domain := DomainConfig{TaskRunner: "just"}

	eff := ComputeEffective(defaults, clientDefaults, domain)

	if eff.TaskRunner != "just" {
		t.Errorf("TaskRunner = %q, want %q", eff.TaskRunner, "just")
	}
}

func TestComputeEffective_TaskRunnerInherited(t *testing.T) {
	defaults := DomainConfig{TaskRunner: "auto"}
	clientDefaults := DomainConfig{}
	domain := DomainConfig{}

	eff := ComputeEffective(defaults, clientDefaults, domain)

	if eff.TaskRunner != "auto" {
		t.Errorf("TaskRunner = %q, want %q", eff.TaskRunner, "auto")
	}
}

func TestComputeEffective_VarsMerge(t *testing.T) {
	defaults := DomainConfig{
		Vars: map[string]string{"region": "us", "env": "prod"},
	}
	clientDefaults := DomainConfig{
		Vars: map[string]string{"region": "eu"},
	}
	domain := DomainConfig{}

	eff := ComputeEffective(defaults, clientDefaults, domain)

	if eff.Vars["region"] != "eu" {
		t.Errorf("Vars[region] = %q, want %q", eff.Vars["region"], "eu")
	}
	if eff.Vars["env"] != "prod" {
		t.Errorf("Vars[env] = %q, want %q", eff.Vars["env"], "prod")
	}
}
