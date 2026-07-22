package chat

import (
	"testing"

	stream "github.com/upperxcode/ada-stream"
)

func TestGetModeConfig_Ask(t *testing.T) {
	cfg := GetModeConfig(ModeAsk)
	if cfg.Mode != ModeAsk {
		t.Errorf("Mode = %v, want %v", cfg.Mode, ModeAsk)
	}
	if len(cfg.AllowedTools) != 2 || cfg.AllowedTools[0] != "read" || cfg.AllowedTools[1] != "search" {
		t.Errorf("AllowedTools = %v, want [read search]", cfg.AllowedTools)
	}
	if cfg.NeedsPermission {
		t.Error("NeedsPermission = true, want false")
	}
	if !cfg.CanOverrideOnce {
		t.Error("CanOverrideOnce = false, want true")
	}
}

func TestGetModeConfig_Plan(t *testing.T) {
	cfg := GetModeConfig(ModePlan)
	if cfg.Mode != ModePlan {
		t.Errorf("Mode = %v, want %v", cfg.Mode, ModePlan)
	}
	if len(cfg.AllowedTools) != 2 || cfg.AllowedTools[0] != "read" || cfg.AllowedTools[1] != "search" {
		t.Errorf("AllowedTools = %v, want [read search]", cfg.AllowedTools)
	}
	if cfg.NeedsPermission {
		t.Error("NeedsPermission = true, want false")
	}
	if !cfg.CanOverrideOnce {
		t.Error("CanOverrideOnce = false, want true")
	}
}

func TestGetModeConfig_Edit(t *testing.T) {
	cfg := GetModeConfig(ModeEdit)
	if cfg.Mode != ModeEdit {
		t.Errorf("Mode = %v, want %v", cfg.Mode, ModeEdit)
	}
	if len(cfg.AllowedTools) != 3 || cfg.AllowedTools[0] != "read" || cfg.AllowedTools[1] != "search" || cfg.AllowedTools[2] != "write" {
		t.Errorf("AllowedTools = %v, want [read search write]", cfg.AllowedTools)
	}
	if !cfg.NeedsPermission {
		t.Error("NeedsPermission = false, want true")
	}
	if !cfg.CanOverrideOnce {
		t.Error("CanOverrideOnce = false, want true")
	}
}

func TestGetModeConfig_Full(t *testing.T) {
	cfg := GetModeConfig(ModeFull)
	if cfg.Mode != ModeFull {
		t.Errorf("Mode = %v, want %v", cfg.Mode, ModeFull)
	}
	expected := []string{"read", "search", "write", "exec", "plan"}
	if len(cfg.AllowedTools) != len(expected) {
		t.Errorf("AllowedTools len = %d, want %d", len(cfg.AllowedTools), len(expected))
	}
	for i, v := range expected {
		if i >= len(cfg.AllowedTools) || cfg.AllowedTools[i] != v {
			t.Errorf("AllowedTools[%d] = %v, want %v", i, cfg.AllowedTools, expected)
			break
		}
	}
	if cfg.NeedsPermission {
		t.Error("NeedsPermission = true, want false")
	}
	if cfg.CanOverrideOnce {
		t.Error("CanOverrideOnce = true, want false")
	}
}

func TestGetModeConfig_Unknown_FallbackToAsk(t *testing.T) {
	cfg := GetModeConfig("INVALID")
	if cfg.Mode != ModeAsk {
		t.Errorf("Mode = %v, want %v (default fallback)", cfg.Mode, ModeAsk)
	}
}

func TestIsValid(t *testing.T) {
	cases := []struct {
		mode ChatMode
		want bool
	}{
		{ModeAsk, true},
		{ModeEdit, true},
		{ModePlan, true},
		{ModeFull, true},
		{"INVALID", false},
		{"", false},
	}
	for _, c := range cases {
		got := c.mode.IsValid()
		if got != c.want {
			t.Errorf("IsValid(%v) = %v, want %v", c.mode, got, c.want)
		}
	}
}

func TestGetSystemPrompt(t *testing.T) {
	base := "custom prompt"
	got := GetSystemPrompt(ModeAsk, base)
	if got != base {
		t.Errorf("GetSystemPrompt with base = %q, want %q", got, base)
	}
	got2 := GetSystemPrompt(ModeAsk, "")
	if got2 == "" {
		t.Error("GetSystemPrompt with empty base returned empty")
	}
}

func TestAllowedChunkTypes_Plan_NoExec(t *testing.T) {
	types := AllowedChunkTypes(ModePlan)
	for _, ct := range types {
		if ct == stream.ChunkExec {
			t.Error("AllowedChunkTypes(PLAN) contains ChunkExec, should not")
		}
	}
}

func TestAllowedChunkTypes_Full_ContainsExec(t *testing.T) {
	types := AllowedChunkTypes(ModeFull)
	found := false
	for _, ct := range types {
		if ct == stream.ChunkExec {
			found = true
			break
		}
	}
	if !found {
		t.Error("AllowedChunkTypes(FULL) should contain ChunkExec")
	}
}
