package chat

import (
	"context"
	"sync"
	"testing"
	"time"
)

type mockEmitter struct {
	mu     sync.Mutex
	events []string
}

func (m *mockEmitter) Emit(event string, data ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
}

func newTestStore() *PermissionStore {
	return NewPermissionStore(nil)
}

func TestNewPermissionStore_NoDB(t *testing.T) {
	ps := NewPermissionStore(nil)
	if ps == nil {
		t.Fatal("NewPermissionStore(nil) returned nil")
	}
	if ps.db != nil {
		t.Error("db should be nil")
	}
}

func TestCheck_ModeFull_AlwaysAllowedForLowRisk(t *testing.T) {
	ps := newTestStore()
	cases := []string{"read", "search", "write", "exec", "plan"}
	for _, tool := range cases {
		result := ps.Check("sess-1", tool, `{}`, ModeFull)
		if !result.Allowed {
			t.Errorf("Check(FULL, %q) = %+v, want Allowed=true", tool, result)
		}
	}
}

func TestCheck_ModeFull_HighRiskNeedsConfirm(t *testing.T) {
	ps := newTestStore()
	// rm -rf é high risk no FULL
	result := ps.Check("sess-1", "exec", `{"command":"rm -rf /tmp"}`, ModeFull)
	if result.Allowed {
		t.Error("Check(FULL, rm -rf) should not be allowed without confirm")
	}
	if !result.NeedsConfirm {
		t.Error("Check(FULL, rm -rf) should need confirm")
	}
	if result.Request == nil {
		t.Error("Check(FULL, rm -rf) should return a PermissionRequest")
	}
}

func TestCheck_ModeFull_GitPushForceNeedsConfirm(t *testing.T) {
	ps := newTestStore()
	result := ps.Check("sess-1", "exec", `{"command":"git push --force origin main"}`, ModeFull)
	if result.Allowed {
		t.Error("Check(FULL, git push --force) should not be allowed without confirm")
	}
	if !result.NeedsConfirm {
		t.Error("Check(FULL, git push --force) should need confirm")
	}
}

func TestCheck_ToolNotAllowed_Ask(t *testing.T) {
	ps := newTestStore()
	result := ps.Check("sess-1", "exec", `{"command":"ls"}`, ModeAsk)
	if result.Allowed {
		t.Fatal("Check(ASK, exec) = true, want false")
	}
	if result.Request == nil {
		t.Fatal("Check(ASK, exec) should return PermissionRequest (CanOverrideOnce=true)")
	}
	if result.Reason == "" {
		t.Error("reason should not be empty")
	}
}

func TestCheck_ToolAllowed_Ask(t *testing.T) {
	ps := newTestStore()
	result := ps.Check("sess-1", "read", `{}`, ModeAsk)
	if !result.Allowed {
		t.Error("Check(ASK, read) = false, want true")
	}
	if result.Request != nil {
		t.Error("Check(ASK, read) returned request, want nil")
	}
}

func TestCheck_Edit_Exec_NeedsPerm(t *testing.T) {
	ps := newTestStore()
	result := ps.Check("sess-1", "exec", `{"command":"ls"}`, ModeEdit)
	if result.Allowed {
		t.Fatal("Check(EDIT, exec) = true, want false")
	}
	if result.Request == nil {
		t.Fatal("Check(EDIT, exec) should return PermissionRequest")
	}
}

func TestCheck_Edit_Read_Allowed(t *testing.T) {
	ps := newTestStore()
	result := ps.Check("sess-1", "read", `{}`, ModeEdit)
	if !result.Allowed {
		t.Error("Check(EDIT, read) = false, want true")
	}
	if result.Request != nil {
		t.Error("Check(EDIT, read) returned request, want nil")
	}
}

func TestGrant_AllowSession(t *testing.T) {
	ps := newTestStore()
	ps.Grant("sess-1", "exec", "*", string(DecisionAllowSession), TTLSession)

	result := ps.Check("sess-1", "exec", `{"command":"ls"}`, ModeEdit)
	if !result.Allowed {
		t.Errorf("Check after allow_session grant = false, want true: %+v", result)
	}
	if result.Request != nil {
		t.Error("Check after allow_session returned request, want nil")
	}
}

func TestGrant_TTLExpiry(t *testing.T) {
	ps := newTestStore()
	// Grant com TTL muito curto (já expirado)
	ps.Grant("sess-1", "exec", "*", string(DecisionAllowSession), TTLAction)
	// TTLAction não expira (time.Time{}), então ainda deve funcionar
	result := ps.Check("sess-1", "exec", `{"command":"ls"}`, ModeEdit)
	if !result.Allowed {
		t.Errorf("Check after TTLAction grant should be allowed: %+v", result)
	}
}

func TestClearSessionGrants(t *testing.T) {
	ps := newTestStore()
	ps.GrantAllowOnce("sess-1", "exec")
	ps.ClearSessionGrants("sess-1")

	result := ps.Check("sess-1", "exec", `{"command":"ls"}`, ModeEdit)
	if result.Allowed {
		t.Error("Check after ClearSessionGrants = true, want false")
	}
	if result.Request == nil {
		t.Error("Check after ClearSessionGrants should return request")
	}
}

func TestResetSession(t *testing.T) {
	ps := newTestStore()
	ps.Grant("sess-1", "exec", "*", string(DecisionAllowSession), TTLSession)
	ps.GrantAllowOnce("sess-1", "exec")
	req := ps.CreatePermissionRequest("sess-1", "exec", `{}`, "test", ModeEdit)
	if req == nil {
		t.Fatal("CreatePermissionRequest failed")
	}

	ps.ResetSession("sess-1")

	if len(ps.AllGrants("sess-1")) != 0 {
		t.Error("grants not cleared after ResetSession")
	}
	if ps.GetPending(req.RequestID) != nil {
		t.Error("pending not cleared after ResetSession")
	}
}

func TestExtractPath(t *testing.T) {
	cases := []struct {
		args string
		want string
	}{
		{`{"file_path":"/tmp/x"}`, "/tmp/x"},
		{`{"path":"src/main.go"}`, "src/main.go"},
		{`{"command":"ls -la"}`, "ls -la"},
		{`{}`, ""},
		{"", ""},
		{`invalid json`, ""},
	}
	for _, c := range cases {
		got := extractPath(c.args)
		if got != c.want {
			t.Errorf("extractPath(%q) = %q, want %q", c.args, got, c.want)
		}
	}
}

func TestCheck_ExecMode_BlocksHighRiskCommands(t *testing.T) {
	ps := newTestStore()
	// Modo EXECUTE deve bloquear rm -rf
	result := ps.Check("sess-1", "exec", `{"command":"rm -rf /tmp"}`, ModeExec)
	if result.Allowed {
		t.Error("Check(EXEC, rm -rf) should be blocked")
	}
	// Modo EXECUTE deve permitir go test
	result2 := ps.Check("sess-1", "exec", `{"command":"go test ./..."}`, ModeExec)
	if !result2.Allowed {
		t.Errorf("Check(EXEC, go test) should be allowed: %+v", result2)
	}
}

func TestCheck_ExecMode_AllowsWrite(t *testing.T) {
	ps := newTestStore()
	result := ps.Check("sess-1", "write", `{"file_path":"src/main.go"}`, ModeExec)
	if !result.Allowed {
		t.Errorf("Check(EXEC, write src/main.go) should be allowed: %+v", result)
	}
}

func TestCheck_AdminMode_AllowsCapabilities(t *testing.T) {
	ps := newTestStore()
	for _, tool := range []string{"read", "search", "write", "exec", "plan", "config", "admin"} {
		result := ps.Check("sess-1", tool, `{}`, ModeAdmin)
		if !result.Allowed {
			t.Errorf("Check(ADMIN, %q) should be allowed: %+v", tool, result)
		}
	}
}

func TestCheck_WriteEnv_AlwaysCritical(t *testing.T) {
	ps := newTestStore()
	modeCases := []ChatMode{ModeAsk, ModeEdit, ModeExec}
	for _, mode := range modeCases {
		result := ps.Check("sess-1", "write", `{"file_path":".env"}`, mode)
		// Deve pedir permissão (NeedsConfirm) ou ser bloqueado (depende do modo)
		if result.Allowed {
			t.Errorf("Check(%v, write .env) should not be auto-allowed", mode)
		}
	}
}

func TestGrantAllowOnce_SecondCallPasses(t *testing.T) {
	ps := newTestStore()
	emitter := &mockEmitter{}
	ctx := context.Background()
	guard := ps.MakeGuard(ctx, "sess-1", ModeAsk, emitter)

	done := make(chan bool)
	go func() { a, _, _ := guard("exec", `{}`); done <- a; close(done) }()

	reqID := pendingReqID(ps)
	if reqID == "" {
		t.Fatal("no pending channel found")
	}

	ps.SendDecision(reqID, string(DecisionAllowOnce))
	firstAllowed := <-done
	if !firstAllowed {
		t.Error("first call: guard should return allowed after decision")
	}

	// Second call — should pass via sessionGrant
	result := ps.Check("sess-1", "exec", `{}`, ModeAsk)
	if !result.Allowed {
		t.Error("second call should be allowed via sessionGrant")
	}
	if result.Request != nil {
		t.Error("second call should not create new request")
	}
}

func TestSendDecision_AllowSession(t *testing.T) {
	ps := newTestStore()
	emitter := &mockEmitter{}
	ctx := context.Background()
	guard := ps.MakeGuard(ctx, "sess-1", ModeEdit, emitter)

	done := make(chan bool)
	go func() { a, _, _ := guard("exec", `{}`); done <- a; close(done) }()

	reqID := pendingReqID(ps)
	if reqID == "" {
		t.Fatal("no pending channel found")
	}

	ps.SendDecision(reqID, string(DecisionAllowSession))
	allowed := <-done

	if !allowed {
		t.Error("guard returned false after allow_session, want true")
	}

	grants := ps.AllGrants("sess-1")
	if len(grants) != 1 || grants[0].Decision != string(DecisionAllowSession) {
		t.Errorf("grants = %+v, want 1 grant with allow_session", grants)
	}
}

func TestSendDecision_Deny(t *testing.T) {
	ps := newTestStore()
	emitter := &mockEmitter{}
	ctx := context.Background()
	guard := ps.MakeGuard(ctx, "sess-1", ModeEdit, emitter)

	done := make(chan bool)
	go func() { a, _, _ := guard("exec", `{}`); done <- a; close(done) }()

	reqID := pendingReqID(ps)
	if reqID == "" {
		t.Fatal("no pending channel found")
	}

	ps.SendDecision(reqID, string(DecisionDeny))
	allowed := <-done

	if allowed {
		t.Error("guard returned true after deny, want false")
	}
}

func TestContextCancel_UnblocksGuard(t *testing.T) {
	ps := newTestStore()
	emitter := &mockEmitter{}
	ctx, cancel := context.WithCancel(context.Background())
	guard := ps.MakeGuard(ctx, "sess-1", ModeEdit, emitter)

	done := make(chan bool)
	go func() { a, _, _ := guard("exec", `{}`); done <- a; close(done) }()

	reqID := pendingReqID(ps)
	if reqID == "" {
		t.Fatal("no pending channel found")
	}

	cancel()

	allowed := <-done
	if allowed {
		t.Error("guard returned true after ctx cancel, want false")
	}
}

func TestPlanMode_SearchAllowed(t *testing.T) {
	ps := newTestStore()
	result := ps.Check("sess-1", "search", `{}`, ModePlan)
	if !result.Allowed {
		t.Error("Check(PLAN, search) = false, want true")
	}
	if result.Request != nil {
		t.Error("Check(PLAN, search) returned request, want nil")
	}
}

func TestPlanMode_ExecBlocked(t *testing.T) {
	ps := newTestStore()
	result := ps.Check("sess-1", "exec", `{}`, ModePlan)
	if result.Allowed {
		t.Fatal("Check(PLAN, exec) = true, want false")
	}
	if result.Request == nil {
		t.Fatal("Check(PLAN, exec) should return PermissionRequest")
	}
}

func TestCheck_ResultHasCorrectRiskLevel(t *testing.T) {
	ps := newTestStore()
	// exec no ASK deve ser RiskCritical
	result := ps.Check("sess-1", "exec", `{"command":"ls"}`, ModeAsk)
	if result.RiskLevel != RiskCritical {
		t.Errorf("RiskLevel = %v, want %v", result.RiskLevel, RiskCritical)
	}
	// read no ASK deve ser RiskNone
	result2 := ps.Check("sess-1", "read", `{}`, ModeAsk)
	if result2.RiskLevel != RiskNone {
		t.Errorf("RiskLevel = %v, want %v", result2.RiskLevel, RiskNone)
	}
}

func TestCheck_ExecMode_NpmTestAllowed(t *testing.T) {
	ps := newTestStore()
	result := ps.Check("sess-1", "exec", `{"command":"npm test"}`, ModeExec)
	if !result.Allowed {
		t.Errorf("Check(EXEC, npm test) should be allowed: %+v", result)
	}
}

func TestCheck_ExplicitActivationModes(t *testing.T) {
	// FULL e ADMIN exigem ativação explícita
	cfgFull := GetModeConfig(ModeFull)
	if !cfgFull.RequiresExplicitActivation {
		t.Error("ModeFull deve exigir ativação explícita")
	}
	cfgAdmin := GetModeConfig(ModeAdmin)
	if !cfgAdmin.RequiresExplicitActivation {
		t.Error("ModeAdmin deve exigir ativação explícita")
	}
	cfgAsk := GetModeConfig(ModeAsk)
	if cfgAsk.RequiresExplicitActivation {
		t.Error("ModeAsk não deve exigir ativação explícita")
	}
}

func TestGetRiskDescription(t *testing.T) {
	desc := GetRiskDescription(ActionRead)
	if desc == "" {
		t.Error("GetRiskDescription(ActionRead) returned empty")
	}
	desc = GetRiskDescription(ActionExec)
	if desc == "" {
		t.Error("GetRiskDescription(ActionExec) returned empty")
	}
	// Ação desconhecida deve retornar o nome
	desc = GetRiskDescription("unknown")
	if desc != "unknown" {
		t.Errorf("GetRiskDescription('unknown') = %q, want 'unknown'", desc)
	}
}

func pendingReqID(ps *PermissionStore) string {
	for i := 0; i < 20; i++ {
		ps.mu.Lock()
		for id := range ps.pendingChans {
			ps.mu.Unlock()
			return id
		}
		ps.mu.Unlock()
		time.Sleep(50 * time.Millisecond)
	}
	return ""
}
