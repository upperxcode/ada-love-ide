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
	if len(cfg.AllowedCapabilities) != 2 || cfg.AllowedCapabilities[0] != CapRead || cfg.AllowedCapabilities[1] != CapSearch {
		t.Errorf("AllowedCapabilities = %v, want [read search]", cfg.AllowedCapabilities)
	}
	if cfg.NeedsPermission {
		t.Error("NeedsPermission = true, want false")
	}
	if !cfg.CanOverrideOnce {
		t.Error("CanOverrideOnce = false, want true")
	}
	if !cfg.AllowSessionGrant {
		t.Error("AllowSessionGrant = false, want true")
	}
	if cfg.RequiresExplicitActivation {
		t.Error("RequiresExplicitActivation = true, want false")
	}
}

func TestGetModeConfig_Plan(t *testing.T) {
	cfg := GetModeConfig(ModePlan)
	if cfg.Mode != ModePlan {
		t.Errorf("Mode = %v, want %v", cfg.Mode, ModePlan)
	}
	if len(cfg.AllowedCapabilities) != 2 || cfg.AllowedCapabilities[0] != CapRead || cfg.AllowedCapabilities[1] != CapSearch {
		t.Errorf("AllowedCapabilities = %v, want [read search]", cfg.AllowedCapabilities)
	}
	if cfg.NeedsPermission {
		t.Error("NeedsPermission = true, want false")
	}
	if !cfg.CanOverrideOnce {
		t.Error("CanOverrideOnce = false, want true")
	}
	if cfg.AllowSessionGrant {
		t.Error("AllowSessionGrant = true, want false")
	}
	if cfg.RequiresExplicitActivation {
		t.Error("RequiresExplicitActivation = true, want false")
	}
	if cfg.DefaultTTL != TTLTask {
		t.Errorf("DefaultTTL = %v, want %v", cfg.DefaultTTL, TTLTask)
	}
}

func TestGetModeConfig_Edit(t *testing.T) {
	cfg := GetModeConfig(ModeEdit)
	if cfg.Mode != ModeEdit {
		t.Errorf("Mode = %v, want %v", cfg.Mode, ModeEdit)
	}
	expected := []ToolCapability{CapRead, CapSearch, CapWrite, CapPlan}
	if len(cfg.AllowedCapabilities) != len(expected) {
		t.Errorf("AllowedCapabilities len = %d, want %d", len(cfg.AllowedCapabilities), len(expected))
	}
	for i, v := range expected {
		if i >= len(cfg.AllowedCapabilities) || cfg.AllowedCapabilities[i] != v {
			t.Errorf("AllowedCapabilities[%d] = %v, want %v", i, cfg.AllowedCapabilities, expected)
			break
		}
	}
	if !cfg.NeedsPermission {
		t.Error("NeedsPermission = false, want true")
	}
	if !cfg.CanOverrideOnce {
		t.Error("CanOverrideOnce = false, want true")
	}
	if !cfg.AllowSessionGrant {
		t.Error("AllowSessionGrant = false, want true")
	}
	if cfg.RequiresExplicitActivation {
		t.Error("RequiresExplicitActivation = true, want false")
	}
	if cfg.DefaultTTL != TTLAction {
		t.Errorf("DefaultTTL = %v, want %v", cfg.DefaultTTL, TTLAction)
	}
}

func TestGetModeConfig_Exec(t *testing.T) {
	cfg := GetModeConfig(ModeExec)
	if cfg.Mode != ModeExec {
		t.Errorf("Mode = %v, want %v", cfg.Mode, ModeExec)
	}
	expected := []ToolCapability{CapRead, CapSearch, CapWrite, CapExec, CapPlan}
	if len(cfg.AllowedCapabilities) != len(expected) {
		t.Errorf("AllowedCapabilities len = %d, want %d", len(cfg.AllowedCapabilities), len(expected))
	}
	if !cfg.NeedsPermission {
		t.Error("NeedsPermission = false, want true")
	}
	if cfg.AllowSessionGrant {
		t.Error("AllowSessionGrant = true, want false")
	}
	if len(cfg.DeniedCommands) == 0 {
		t.Error("DeniedCommands should not be empty for EXEC mode")
	}
	if len(cfg.DeniedActions) == 0 {
		t.Error("DeniedActions should not be empty for EXEC mode")
	}
}

func TestGetModeConfig_Full(t *testing.T) {
	cfg := GetModeConfig(ModeFull)
	if cfg.Mode != ModeFull {
		t.Errorf("Mode = %v, want %v", cfg.Mode, ModeFull)
	}
	expected := []ToolCapability{CapRead, CapSearch, CapWrite, CapExec, CapPlan}
	if len(cfg.AllowedCapabilities) != len(expected) {
		t.Errorf("AllowedCapabilities len = %d, want %d", len(cfg.AllowedCapabilities), len(expected))
	}
	for i, v := range expected {
		if i >= len(cfg.AllowedCapabilities) || cfg.AllowedCapabilities[i] != v {
			t.Errorf("AllowedCapabilities[%d] = %v, want %v", i, cfg.AllowedCapabilities, expected)
			break
		}
	}
	if cfg.NeedsPermission {
		t.Error("NeedsPermission = true, want false")
	}
	if cfg.CanOverrideOnce {
		t.Error("CanOverrideOnce = true, want false")
	}
	if !cfg.RequiresExplicitActivation {
		t.Error("RequiresExplicitActivation = false, want true")
	}
	if cfg.DefaultTTL != TTLTemporary {
		t.Errorf("DefaultTTL = %v, want %v", cfg.DefaultTTL, TTLTemporary)
	}
}

func TestGetModeConfig_Admin(t *testing.T) {
	cfg := GetModeConfig(ModeAdmin)
	if cfg.Mode != ModeAdmin {
		t.Errorf("Mode = %v, want %v", cfg.Mode, ModeAdmin)
	}
	expected := []ToolCapability{CapRead, CapSearch, CapWrite, CapExec, CapPlan, CapAdmin, CapConfig}
	if len(cfg.AllowedCapabilities) != len(expected) {
		t.Errorf("AllowedCapabilities len = %d, want %d", len(cfg.AllowedCapabilities), len(expected))
	}
	if cfg.NeedsPermission {
		t.Error("NeedsPermission = true, want false")
	}
	if !cfg.RequiresExplicitActivation {
		t.Error("RequiresExplicitActivation = false, want true")
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
		{ModeExec, true},
		{ModeAdmin, true},
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
	// Cada modo deve ter um system prompt default
	for _, m := range []ChatMode{ModeAsk, ModeEdit, ModePlan, ModeFull, ModeExec, ModeAdmin} {
		p := GetSystemPrompt(m, "")
		if p == "" {
			t.Errorf("GetSystemPrompt(%v) returned empty", m)
		}
	}
}

func TestAllowedChunkTypes_Ask(t *testing.T) {
	types := AllowedChunkTypes(ModeAsk)
	for _, ct := range types {
		if ct == stream.ChunkExec || ct == stream.ChunkDiff {
			t.Errorf("AllowedChunkTypes(ASK) contains %v, should not", ct)
		}
	}
}

func TestAllowedChunkTypes_Plan_NoExec(t *testing.T) {
	types := AllowedChunkTypes(ModePlan)
	for _, ct := range types {
		if ct == stream.ChunkExec || ct == stream.ChunkDiff {
			t.Errorf("AllowedChunkTypes(PLAN) contains %v, should not", ct)
		}
	}
}

func TestAllowedChunkTypes_Exec_ContainsExec(t *testing.T) {
	types := AllowedChunkTypes(ModeExec)
	found := false
	for _, ct := range types {
		if ct == stream.ChunkExec {
			found = true
			break
		}
	}
	if !found {
		t.Error("AllowedChunkTypes(EXEC) should contain ChunkExec")
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

func TestAllowedChunkTypes_Admin_NoExec(t *testing.T) {
	types := AllowedChunkTypes(ModeAdmin)
	for _, ct := range types {
		if ct == stream.ChunkExec || ct == stream.ChunkDiff || ct == stream.ChunkPlan {
			t.Errorf("AllowedChunkTypes(ADMIN) contains %v, should not", ct)
		}
	}
}

func TestGetCapabilityForTool(t *testing.T) {
	cases := []struct {
		tool string
		want ToolCapability
	}{
		{"read", CapRead},
		{"read_file", CapRead},
		{"cat", CapRead},
		{"search", CapSearch},
		{"grep", CapSearch},
		{"explore", CapSearch},
		{"write", CapWrite},
		{"write_file", CapWrite},
		{"edit", CapWrite},
		{"exec", CapExec},
		{"execute", CapExec},
		{"run", CapExec},
		{"shell", CapExec},
		{"terminal", CapExec},
		{"plan", CapPlan},
		{"planning", CapPlan},
		{"admin", CapAdmin},
		{"config", CapConfig},
		{"unknown_tool", CapExec}, // fallback
	}
	for _, c := range cases {
		got := GetCapabilityForTool(c.tool)
		if got != c.want {
			t.Errorf("GetCapabilityForTool(%q) = %v, want %v", c.tool, got, c.want)
		}
	}
}

func TestClassifyAction_Read(t *testing.T) {
	action, risk := ClassifyAction("read", `{"path":"/tmp/file"}`, ModeAsk)
	if action != ActionRead {
		t.Errorf("action = %v, want %v", action, ActionRead)
	}
	if risk != RiskNone {
		t.Errorf("risk = %v, want %v", risk, RiskNone)
	}
}

func TestClassifyAction_Write_Env(t *testing.T) {
	action, risk := ClassifyAction("write", `{"file_path":".env"}`, ModeEdit)
	if action != ActionWriteEnv {
		t.Errorf("action = %v, want %v", action, ActionWriteEnv)
	}
	if risk != RiskCritical {
		t.Errorf("risk = %v, want %v", risk, RiskCritical)
	}
}

func TestClassifyAction_Write_Outside(t *testing.T) {
	action, risk := ClassifyAction("write", `{"file_path":"/etc/hosts"}`, ModeEdit)
	if action != ActionWriteOutside {
		t.Errorf("action = %v, want %v", action, ActionWriteOutside)
	}
	if risk != RiskCritical {
		t.Errorf("risk = %v, want %v", risk, RiskCritical)
	}
}

func TestClassifyAction_Write_Project(t *testing.T) {
	action, risk := ClassifyAction("write", `{"file_path":"src/main.go"}`, ModeEdit)
	if action != ActionWriteProject {
		t.Errorf("action = %v, want %v", action, ActionWriteProject)
	}
	if risk != RiskMedium {
		t.Errorf("risk = %v, want %v", risk, RiskMedium)
	}
}

func TestClassifyAction_Exec(t *testing.T) {
	action, risk := ClassifyAction("exec", `{"command":"ls -la"}`, ModeEdit)
	if action != ActionExec {
		t.Errorf("action = %v, want %v", action, ActionExec)
	}
	if risk != RiskHigh {
		t.Errorf("risk = %v, want %v", risk, RiskHigh)
	}
}

func TestClassifyAction_Exec_HighRisk(t *testing.T) {
	action, risk := ClassifyAction("exec", `{"command":"rm -rf /tmp"}`, ModeEdit)
	if action != ActionExecHighRisk {
		t.Errorf("action = %v, want %v", action, ActionExecHighRisk)
	}
	if risk != RiskCritical {
		t.Errorf("risk = %v, want %v", risk, RiskCritical)
	}
}

func TestClassifyAction_Exec_GitPushForce(t *testing.T) {
	action, risk := ClassifyAction("exec", `{"command":"git push --force origin main"}`, ModeFull)
	if action != ActionExecHighRisk {
		t.Errorf("action = %v, want %v", action, ActionExecHighRisk)
	}
	if risk != RiskHigh {
		t.Errorf("risk = %v, want %v", risk, RiskHigh)
	}
}

func TestClassifyAction_Search(t *testing.T) {
	action, risk := ClassifyAction("search", `{"query":"TODO"}`, ModeAsk)
	if action != ActionSearch {
		t.Errorf("action = %v, want %v", action, ActionSearch)
	}
	if risk != RiskNone {
		t.Errorf("risk = %v, want %v", risk, RiskNone)
	}
}

func TestRiskLevel_String(t *testing.T) {
	cases := []struct {
		r    RiskLevel
		want string
	}{
		{RiskNone, "none"},
		{RiskLow, "low"},
		{RiskMedium, "medium"},
		{RiskHigh, "high"},
		{RiskCritical, "critical"},
		{RiskLevel(99), "unknown"},
	}
	for _, c := range cases {
		got := c.r.String()
		if got != c.want {
			t.Errorf("RiskLevel(%d).String() = %q, want %q", c.r, got, c.want)
		}
	}
}

func TestIsHighRiskCommand(t *testing.T) {
	cases := []struct {
		cmd  string
		want bool
	}{
		{"rm -rf /", true},
		{"rm -rf --no-preserve-root /", true},
		{"git push --force", true},
		{"sudo apt install", true},
		{"ls -la", false},
		{"go test ./...", false},
		{"git status", false},
		{"cat file.txt", false},
		{"docker rm -f container", true},
		{"chmod 777 /tmp", true},
	}
	for _, c := range cases {
		got := isHighRiskCommand(c.cmd)
		if got != c.want {
			t.Errorf("isHighRiskCommand(%q) = %v, want %v", c.cmd, got, c.want)
		}
	}
}

func TestIsEnvFile(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{".env", true},
		{".env.production", true},
		{"src/main.go", false},
		{".ssh/id_rsa", true},
		{"credentials.json", true},
		{"README.md", false},
		{".gitconfig", true},
	}
	for _, c := range cases {
		got := isEnvFile(c.path)
		if got != c.want {
			t.Errorf("isEnvFile(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestGetRisk(t *testing.T) {
	// ActionRead no ASK deve ser RiskNone
	risk := GetRisk(ActionRead, ModeAsk)
	if risk != RiskNone {
		t.Errorf("GetRisk(Read, ASK) = %v, want %v", risk, RiskNone)
	}
	// ActionExec no ASK deve ser RiskCritical
	risk = GetRisk(ActionExec, ModeAsk)
	if risk != RiskCritical {
		t.Errorf("GetRisk(Exec, ASK) = %v, want %v", risk, RiskCritical)
	}
	// ActionExec no FULL deve ser RiskLow
	risk = GetRisk(ActionExec, ModeFull)
	if risk != RiskLow {
		t.Errorf("GetRisk(Exec, FULL) = %v, want %v", risk, RiskLow)
	}
	// Action desconhecida: default RiskMedium
	risk = GetRisk("unknown_action", ModeAsk)
	if risk != RiskMedium {
		t.Errorf("GetRisk(unknown, ASK) = %v, want %v", risk, RiskMedium)
	}
}
