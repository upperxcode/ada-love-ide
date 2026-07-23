package chat

import (
	"fmt"
	"testing"
)

type testEmitter2 struct {
	events []string
}

func (e *testEmitter2) Emit(event string, data ...any) {
	e.events = append(e.events, event)
	fmt.Printf("[testEmitter] %s\n", event)
}

func TestModePowerLevel(t *testing.T) {
	cases := []struct {
		mode ChatMode
		want int
	}{
		{ModeAdmin, 100},
		{ModeFull, 80},
		{ModeExec, 60},
		{ModeEdit, 40},
		{ModePlan, 20},
		{ModeAsk, 10},
		{"UNKNOWN", 0},
	}
	for _, c := range cases {
		got := ModePowerLevel(c.mode)
		if got != c.want {
			t.Errorf("ModePowerLevel(%v) = %d, want %d", c.mode, got, c.want)
		}
	}
}

func TestIsUpgrade(t *testing.T) {
	cases := []struct {
		oldMode ChatMode
		newMode ChatMode
		want    bool
	}{
		{ModeAsk, ModeFull, true},      // upgrade
		{ModeEdit, ModeFull, true},      // upgrade
		{ModeFull, ModeFull, true},      // same
		{ModeFull, ModeEdit, false},     // downgrade
		{ModeFull, ModeAsk, false},      // downgrade
		{ModeExec, ModeEdit, false},     // downgrade
		{ModeEdit, ModeExec, true},      // upgrade
		{ModeAdmin, ModeFull, false},    // downgrade
		{ModeFull, ModeAdmin, true},     // upgrade
	}
	for _, c := range cases {
		got := IsUpgrade(c.oldMode, c.newMode)
		if got != c.want {
			t.Errorf("IsUpgrade(%v, %v) = %v, want %v", c.oldMode, c.newMode, got, c.want)
		}
	}
}

func TestDowngradeClearsSessionGrants(t *testing.T) {
	ps := newTestStore()
	emitter := &testEmitter2{}

	// Concede grant no modo FULL
	ps.SetCurrentMode("sess-1", ModeFull, emitter)
	ps.GrantAllowOnce("sess-1", "exec", ModeFull)

	// Verifica que o grant existe
	grants := ps.ListGrants("sess-1")
	if len(grants) != 1 {
		t.Fatalf("expected 1 grant, got %d: %+v", len(grants), grants)
	}

	// Downgrade para EDIT
	cleared := ps.SetCurrentMode("sess-1", ModeEdit, emitter)
	if !cleared {
		t.Error("SetCurrentMode(FULL→EDIT) should clear grants")
	}

	// Verifica que os grants foram limpos
	grants2 := ps.ListGrants("sess-1")
	if len(grants2) != 0 {
		t.Errorf("grants should be empty after downgrade, got %+v", grants2)
	}
}

func TestUpgradeKeepsGrants(t *testing.T) {
	ps := newTestStore()
	emitter := &testEmitter2{}

	// Concede grant no modo EDIT
	ps.SetCurrentMode("sess-1", ModeEdit, emitter)
	ps.GrantAllowOnce("sess-1", "write_file", ModeEdit)

	// Upgrade para FULL
	cleared := ps.SetCurrentMode("sess-1", ModeFull, emitter)
	if cleared {
		t.Error("SetCurrentMode(EDIT→FULL) should NOT clear grants (upgrade)")
	}

	// Verifica que os grants ainda existem
	grants := ps.ListGrants("sess-1")
	if len(grants) != 1 {
		t.Errorf("expected 1 grant after upgrade, got %d: %+v", len(grants), grants)
	}
}

func TestSameModeKeepsGrants(t *testing.T) {
	ps := newTestStore()
	emitter := &testEmitter2{}

	ps.SetCurrentMode("sess-1", ModeFull, emitter)
	ps.GrantAllowOnce("sess-1", "exec", ModeFull)

	// Mesmo modo
	cleared := ps.SetCurrentMode("sess-1", ModeFull, emitter)
	if cleared {
		t.Error("SetCurrentMode(FULL→FULL) should NOT clear grants")
	}

	grants := ps.ListGrants("sess-1")
	if len(grants) != 1 {
		t.Errorf("expected 1 grant, got %d", len(grants))
	}
}

func TestGrantFromHigherModeInvalidAfterDowngrade(t *testing.T) {
	ps := newTestStore()
	emitter := &testEmitter2{}

	// Concede grant de exec no modo FULL
	ps.SetCurrentMode("sess-1", ModeFull, emitter)
	ps.GrantAllowOnce("sess-1", "exec", ModeFull)

	// Check no FULL → deve passar
	result := ps.Check("sess-1", "exec", `{}`, ModeFull)
	if !result.Allowed {
		t.Errorf("exec em FULL deve ser auto-autorizado, got Allowed=%v", result.Allowed)
	}

	// Downgrade para EDIT
	ps.SetCurrentMode("sess-1", ModeEdit, emitter)

	// Check no EDIT → grant foi limpo, deve pedir permissão
	result2 := ps.Check("sess-1", "exec", `{}`, ModeEdit)
	if result2.Allowed {
		t.Error("exec em EDIT após downgrade deve pedir permissão, got auto-allowed")
	}
	if result2.Request == nil {
		t.Error("exec em EDIT após downgrade deve criar PermissionRequest")
	}
}

func TestNormalizeModeExported(t *testing.T) {
	cases := []struct {
		input string
		want  ChatMode
	}{
		{"ask", ModeAsk},
		{"ASK", ModeAsk},
		{"edit", ModeEdit},
		{"execute", ModeExec},
		{"EXECUTE", ModeExec},
		{"exec", ModeExec},
		{"test", ModeExec},
		{"full", ModeFull},
		{"admin", ModeAdmin},
		{"config", ModeAdmin},
		{"", ModeAsk},
		{"invalid", ModeAsk},
	}
	for _, c := range cases {
		got := NormalizeMode(c.input)
		if got != string(c.want) {
			t.Errorf("NormalizeMode(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestPersistedGrantInvalidAfterDowngrade(t *testing.T) {
	ps := newTestStore()
	emitter := &testEmitter2{}

	// Grant persistido no modo FULL
	ps.SetCurrentMode("sess-1", ModeFull, emitter)
	ps.Grant("sess-1", "exec", "*", "allow_session", TTLSession, ModeFull)

	// Check no FULL → grant funciona
	result := ps.Check("sess-1", "exec", `{}`, ModeFull)
	if !result.Allowed {
		t.Errorf("granted exec em FULL deve passar, got Allowed=%v", result.Allowed)
	}

	// Downgrade para EDIT
	ps.SetCurrentMode("sess-1", ModeEdit, emitter)

	// Check no EDIT → grant foi limpo (downgrade limpa grants persistidos)
	result2 := ps.Check("sess-1", "exec", `{}`, ModeEdit)
	if result2.Allowed {
		t.Error("exec em EDIT após downgrade deve pedir permissão, grant foi limpo")
	}
}

func TestGetCurrentMode(t *testing.T) {
	ps := newTestStore()
	emitter := &testEmitter2{}

	// Modo não definido
	mode := ps.GetCurrentMode("sess-1")
	if mode != "" {
		t.Errorf("GetCurrentMode for unknown session = %q, want ''", mode)
	}

	ps.SetCurrentMode("sess-1", ModeExec, emitter)
	mode = ps.GetCurrentMode("sess-1")
	if mode != ModeExec {
		t.Errorf("GetCurrentMode = %v, want %v", mode, ModeExec)
	}
}

func TestDumpGrantsNoPanic(t *testing.T) {
	ps := newTestStore()
	emitter := &testEmitter2{}

	// Não deve panic com sessão vazia
	ps.DumpGrants("sess-empty")

	// Não deve panic com grants
	ps.SetCurrentMode("sess-1", ModeFull, emitter)
	ps.GrantAllowOnce("sess-1", "exec", ModeFull)
	ps.Grant("sess-1", "write", "*", "allow_session", TTLSession, ModeFull)
	ps.DumpGrants("sess-1")
}
