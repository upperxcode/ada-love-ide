package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"ada-love-ide/internal/adapters"
)

type PermissionGrant struct {
	SessionID  string    `json:"session_id"`
	Action     string    `json:"action"`
	TargetPath string    `json:"target_path"`
	Decision   string    `json:"decision"`
	CreatedAt  time.Time `json:"created_at"`
}

type PermissionRequest struct {
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
	ToolName  string `json:"tool_name"`
	Args      string `json:"args"`
	Reason    string `json:"reason"`
	TargetPath string `json:"target_path"`
	Mode      string `json:"mode"`
}

type PendingToolCall struct {
	ToolName string
	ArgsJSON string
	ToolID   string
	Iter     int
	Index    int
}

type PermissionStore struct {
	mu           sync.Mutex
	grants       map[string][]PermissionGrant // key = sessionID, persisted allow_session
	sessionGrants map[string]map[string]bool  // sessionID -> action -> true (turnt-level allow_once)
	pending      map[string]*PermissionRequest
	pendExec     map[string]*PendingToolCall
	pendingChans map[string]chan string // requestID -> decision channel
	db           *sql.DB
	nextID       int64
}

func NewPermissionStore(db *sql.DB) *PermissionStore {
	ps := &PermissionStore{
		grants:        make(map[string][]PermissionGrant),
		sessionGrants: make(map[string]map[string]bool),
		pending:       make(map[string]*PermissionRequest),
		pendExec:      make(map[string]*PendingToolCall),
		pendingChans:  make(map[string]chan string),
		db:            db,
		nextID:        1,
	}
	if db != nil {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS permission_grants (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			action TEXT NOT NULL,
			target_path TEXT NOT NULL DEFAULT '*',
			decision TEXT NOT NULL CHECK(decision IN ('allow_session','deny_session')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
		if err == nil {
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_perm_grants_session ON permission_grants(session_id, action, target_path)`)
		}
		ps.loadFromDB()
	}
	return ps
}

func (ps *PermissionStore) loadFromDB() {
	rows, err := ps.db.Query(`SELECT session_id, action, target_path, decision, created_at FROM permission_grants`)
	if err != nil {
		fmt.Printf("[PermissionStore] load error: %v\n", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var g PermissionGrant
		if err := rows.Scan(&g.SessionID, &g.Action, &g.TargetPath, &g.Decision, &g.CreatedAt); err != nil {
			fmt.Printf("[PermissionStore] scan error: %v\n", err)
			continue
		}
		ps.grants[g.SessionID] = append(ps.grants[g.SessionID], g)
	}
	fmt.Printf("[PermissionStore] loaded %d permissions from DB\n", len(ps.grants))
}

func (ps *PermissionStore) Check(sessionID string, toolName string, argsJSON string, mode ChatMode) (bool, string, *PermissionRequest) {
	cfg := GetModeConfig(mode)

	if mode == ModeFull {
		return true, "", nil
	}

	targetPath := extractPath(argsJSON)
	action := classifyAction(toolName, targetPath, mode)

	// Check session-level grants first (allow_once within turn)
	ps.mu.Lock()
	_, sessionOk := ps.sessionGrants[sessionID][action]
	ps.mu.Unlock()
	if sessionOk {
		return true, "", nil
	}

	// Check persisted allow_session grants
	if action != "" {
		ps.mu.Lock()
		grant := ps.findGrant(sessionID, action, targetPath)
		ps.mu.Unlock()
		if grant != nil && grant.Decision == "allow_session" {
			return true, "", nil
		}
	}

	// Check if tool is in AllowedTools
	toolAllowed := false
	for _, t := range cfg.AllowedTools {
		if t == toolName {
			toolAllowed = true
			break
		}
	}
	if !toolAllowed {
		if cfg.CanOverrideOnce {
			reason := fmt.Sprintf("%s não permitida no modo %s", toolName, mode)
			req := ps.createRequest(sessionID, toolName, argsJSON, reason, mode)
			return false, reason, req
		}
		return false, fmt.Sprintf("tool %s não permitida no modo %s", toolName, mode), nil
	}

	// Tool is allowed, but check if it needs additional permission (e.g., exec in EDIT)
	if action != "" {
		reason := fmt.Sprintf("action=%s tool=%s path=%s", action, toolName, targetPath)
		req := ps.createRequest(sessionID, toolName, argsJSON, reason, mode)
		return false, reason, req
	}

	return true, "", nil
}

func (ps *PermissionStore) createRequest(sessionID, toolName, argsJSON, reason string, mode ChatMode) *PermissionRequest {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	id := fmt.Sprintf("perm-%d", ps.nextID)
	ps.nextID++

	req := &PermissionRequest{
		RequestID: id,
		SessionID: sessionID,
		ToolName:  toolName,
		Args:      argsJSON,
		Reason:    reason,
		TargetPath: extractPath(argsJSON),
		Mode:      string(mode),
	}
	ps.pending[id] = req
	return req
}

func (ps *PermissionStore) GetPending(requestID string) *PermissionRequest {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.pending[requestID]
}

func (ps *PermissionStore) RemovePending(requestID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.pending, requestID)
}

func (ps *PermissionStore) Grant(sessionID, action, targetPath, decision string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	g := PermissionGrant{
		SessionID:  sessionID,
		Action:     action,
		TargetPath: targetPath,
		Decision:   decision,
		CreatedAt:  time.Now(),
	}
	ps.grants[sessionID] = append(ps.grants[sessionID], g)

	if ps.db != nil {
		_, err := ps.db.Exec(
			`INSERT INTO permission_grants (session_id, action, target_path, decision, created_at) VALUES (?, ?, ?, ?, ?)`,
			sessionID, action, targetPath, decision, g.CreatedAt,
		)
		if err != nil {
			fmt.Printf("[PermissionStore] db insert error: %v\n", err)
		}
	}
}

func (ps *PermissionStore) ClearSessionGrants(sessionID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.sessionGrants, sessionID)
}

func (ps *PermissionStore) ResetSession(sessionID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	delete(ps.grants, sessionID)
	delete(ps.sessionGrants, sessionID)

	if ps.db != nil {
		_, _ = ps.db.Exec(`DELETE FROM permission_grants WHERE session_id = ?`, sessionID)
	}

	for id, req := range ps.pending {
		if req.SessionID == sessionID {
			delete(ps.pending, id)
		}
	}
}

func (ps *PermissionStore) findGrant(sessionID, action, targetPath string) *PermissionGrant {
	grants := ps.grants[sessionID]
	for _, g := range grants {
		if g.Action != action {
			continue
		}
		if g.TargetPath == "*" || g.TargetPath == targetPath {
			return &g
		}
	}
	return nil
}

func extractPath(argsJSON string) string {
	if argsJSON == "" {
		return ""
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &data); err != nil {
		return ""
	}
	if path, ok := data["file_path"].(string); ok {
		return path
	}
	if path, ok := data["path"].(string); ok {
		return path
	}
	if cmd, ok := data["command"].(string); ok {
		return cmd
	}
	return ""
}

func classifyAction(toolName, targetPath string, mode ChatMode) string {
	if mode == ModeFull {
		return ""
	}
	switch toolName {
	case "write_file", "write":
		if mode == ModeEdit || mode == ModeAsk || mode == ModePlan {
			return "write_outside"
		}
		return ""
	case "run_terminal", "execute", "exec", "run":
		return "exec"
	case "create_dir", "mkdir":
		return "mkdir_outside"
	default:
		return ""
	}
}

func (ps *PermissionStore) StorePendingExec(requestID string, tc *PendingToolCall) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.pendExec[requestID] = tc
}

func (ps *PermissionStore) SendDecision(requestID, decision string) {
	ps.mu.Lock()
	ch, ok := ps.pendingChans[requestID]
	ps.mu.Unlock()
	if ok {
		ch <- decision
	}
}

func (ps *PermissionStore) RemovePendingChan(requestID string) {
	ps.mu.Lock()
	delete(ps.pendingChans, requestID)
	ps.mu.Unlock()
}

func (ps *PermissionStore) AllGrants(sessionID string) []PermissionGrant {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.grants[sessionID]
}

func (ps *PermissionStore) MakeGuard(ctx context.Context, sessionID string, mode ChatMode, emitter Emitter) adapters.PermissionGuard {
	return func(toolName, argsJSON string) (bool, string, string) {
		allowed, reason, req := ps.Check(sessionID, toolName, argsJSON, mode)
		if allowed {
			return true, "", ""
		}
		if req != nil {
			emitter.Emit("chat:permission-request", map[string]any{
				"session_id":  sessionID,
				"request_id":  req.RequestID,
				"tool_name":   req.ToolName,
				"args":        req.Args,
				"reason":      req.Reason,
				"target_path": req.TargetPath,
				"mode":        req.Mode,
			})

			ch := make(chan string, 1)
			ps.mu.Lock()
			ps.pendingChans[req.RequestID] = ch
			ps.mu.Unlock()

			select {
			case decision := <-ch:
				ps.mu.Lock()
				delete(ps.pendingChans, req.RequestID)
				ps.mu.Unlock()

				switch decision {
				case "allow_once":
					targetPath := extractPath(argsJSON)
					action := classifyAction(toolName, targetPath, mode)
					if action != "" {
						ps.mu.Lock()
						if ps.sessionGrants[sessionID] == nil {
							ps.sessionGrants[sessionID] = make(map[string]bool)
						}
						ps.sessionGrants[sessionID][action] = true
						ps.mu.Unlock()
					}
					return true, "", ""
				case "allow_session":
					targetPath := extractPath(argsJSON)
					action := classifyAction(toolName, targetPath, mode)
					if action != "" {
						ps.Grant(sessionID, action, "*", "allow_session")
					}
					return true, "", ""
				case "deny":
					return false, "negado pelo usuário", ""
				}
			case <-ctx.Done():
				ps.mu.Lock()
				delete(ps.pendingChans, req.RequestID)
				ps.mu.Unlock()
				return false, "stream cancelado", ""
			}
			return false, reason, ""
		}
		return false, reason, ""
	}
}

func (ps *PermissionStore) GetPendingExec(requestID string) *PendingToolCall {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.pendExec[requestID]
}

func (ps *PermissionStore) RemovePendingExec(requestID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.pendExec, requestID)
}
