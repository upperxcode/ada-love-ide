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

func TestCheck_ModeFull_AlwaysAllowed(t *testing.T) {
	ps := newTestStore()
	cases := []string{"read", "search", "write", "exec", "plan", "unknown"}
	for _, tool := range cases {
		allowed, _, req := ps.Check("sess-1", tool, `{}`, ModeFull)
		if !allowed {
			t.Errorf("Check(FULL, %q) = false, want true", tool)
		}
		if req != nil {
			t.Errorf("Check(FULL, %q) returned request, want nil", tool)
		}
	}
}

func TestCheck_ToolNotAllowed_Ask(t *testing.T) {
	ps := newTestStore()
	allowed, reason, req := ps.Check("sess-1", "exec", `{"command":"ls"}`, ModeAsk)
	if allowed {
		t.Fatal("Check(ASK, exec) = true, want false")
	}
	if req == nil {
		t.Fatal("Check(ASK, exec) should return PermissionRequest (CanOverrideOnce=true)")
	}
	if reason == "" {
		t.Error("reason should not be empty")
	}
}

func TestCheck_ToolAllowed_Ask(t *testing.T) {
	ps := newTestStore()
	allowed, _, req := ps.Check("sess-1", "read", `{}`, ModeAsk)
	if !allowed {
		t.Error("Check(ASK, read) = false, want true")
	}
	if req != nil {
		t.Error("Check(ASK, read) returned request, want nil")
	}
}

func TestCheck_Edit_Exec_NeedsPerm(t *testing.T) {
	ps := newTestStore()
	allowed, _, req := ps.Check("sess-1", "exec", `{"command":"ls"}`, ModeEdit)
	if allowed {
		t.Fatal("Check(EDIT, exec) = true, want false")
	}
	if req == nil {
		t.Fatal("Check(EDIT, exec) should return PermissionRequest")
	}
}

func TestCheck_Edit_Read_Allowed(t *testing.T) {
	ps := newTestStore()
	allowed, _, req := ps.Check("sess-1", "read", `{}`, ModeEdit)
	if !allowed {
		t.Error("Check(EDIT, read) = false, want true")
	}
	if req != nil {
		t.Error("Check(EDIT, read) returned request, want nil")
	}
}

func TestGrant_AllowSession(t *testing.T) {
	ps := newTestStore()
	ps.Grant("sess-1", "exec", "*", "allow_session")

	// After Grant, Check should return true
	allowed, _, req := ps.Check("sess-1", "exec", `{"command":"ls"}`, ModeEdit)
	if !allowed {
		t.Error("Check after allow_session grant = false, want true")
	}
	if req != nil {
		t.Error("Check after allow_session returned request, want nil")
	}
}

func TestClearSessionGrants(t *testing.T) {
	ps := newTestStore()
	ps.sessionGrants["sess-1"] = map[string]bool{"exec": true}
	ps.ClearSessionGrants("sess-1")

	allowed, _, req := ps.Check("sess-1", "exec", `{"command":"ls"}`, ModeEdit)
	if allowed {
		t.Error("Check after ClearSessionGrants = true, want false")
	}
	if req == nil {
		t.Error("Check after ClearSessionGrants should return request")
	}
}

func TestResetSession(t *testing.T) {
	ps := newTestStore()
	ps.Grant("sess-1", "exec", "*", "allow_session")
	ps.sessionGrants["sess-1"] = map[string]bool{"exec": true}
	req := ps.createRequest("sess-1", "exec", `{}`, "test", ModeEdit)
	if req == nil {
		t.Fatal("createRequest failed")
	}

	ps.ResetSession("sess-1")

	if len(ps.AllGrants("sess-1")) != 0 {
		t.Error("grants not cleared after ResetSession")
	}
	if ps.GetPending(req.RequestID) != nil {
		t.Error("pending not cleared after ResetSession")
	}
	_, sessionOk := ps.sessionGrants["sess-1"]
	if sessionOk {
		t.Error("sessionGrants not cleared after ResetSession")
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

func TestClassifyAction(t *testing.T) {
	cases := []struct {
		tool     string
		path     string
		mode     ChatMode
		want     string
	}{
		{"exec", "ls", ModeEdit, "exec"},
		{"exec", "ls", ModeAsk, "exec"},
		{"exec", "ls", ModeFull, ""},
		{"write", "/outside", ModeEdit, "write_outside"},
		{"write", "/outside", ModeAsk, "write_outside"},
		{"write_file", "/outside", ModeEdit, "write_outside"},
		{"create_dir", "/tmp", ModeEdit, "mkdir_outside"},
		{"read", "/file", ModeEdit, ""},
		{"unknown", "", ModeEdit, ""},
	}
	for _, c := range cases {
		got := classifyAction(c.tool, c.path, c.mode)
		if got != c.want {
			t.Errorf("classifyAction(%q, %q, %v) = %q, want %q", c.tool, c.path, c.mode, got, c.want)
		}
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

func TestSendDecision_AllowOnce(t *testing.T) {
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

	ps.SendDecision(reqID, "allow_once")
	allowed := <-done

	if !allowed {
		t.Error("guard returned false after allow_once, want true")
	}

	ps.mu.Lock()
	_, sessionOk := ps.sessionGrants["sess-1"]["exec"]
	ps.mu.Unlock()
	if !sessionOk {
		t.Error("sessionGrant not set after allow_once")
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

	ps.SendDecision(reqID, "allow_session")
	allowed := <-done

	if !allowed {
		t.Error("guard returned false after allow_session, want true")
	}

	grants := ps.AllGrants("sess-1")
	if len(grants) != 1 || grants[0].Decision != "allow_session" {
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

	ps.SendDecision(reqID, "deny")
	allowed := <-done

	if allowed {
		t.Error("guard returned true after deny, want false")
	}
}

func TestSessionGrants_AllowOnce_SecondCallPasses(t *testing.T) {
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

	ps.SendDecision(reqID, "allow_once")
	firstAllowed := <-done
	if !firstAllowed {
		t.Error("first call: guard should return allowed after decision")
	}

	// Second call — should pass via sessionGrant
	allowed2, _, req2 := ps.Check("sess-1", "exec", `{}`, ModeAsk)
	if !allowed2 {
		t.Error("second call should be allowed via sessionGrant")
	}
	if req2 != nil {
		t.Error("second call should not create new request")
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
	allowed, _, req := ps.Check("sess-1", "search", `{}`, ModePlan)
	if !allowed {
		t.Error("Check(PLAN, search) = false, want true")
	}
	if req != nil {
		t.Error("Check(PLAN, search) returned request, want nil")
	}
}

func TestPlanMode_ExecBlocked(t *testing.T) {
	ps := newTestStore()
	allowed, _, req := ps.Check("sess-1", "exec", `{}`, ModePlan)
	if allowed {
		t.Fatal("Check(PLAN, exec) = true, want false")
	}
	if req == nil {
		t.Fatal("Check(PLAN, exec) should return PermissionRequest")
	}
}
