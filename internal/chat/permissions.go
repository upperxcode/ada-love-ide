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

// ── Tipos de decisão ──

type Decision string

const (
	DecisionAllowOnce    Decision = "allow_once"
	DecisionAllowSession Decision = "allow_session"
	DecisionDeny         Decision = "deny"
)

// ── Grant (permissão persistida) ──

type PermissionGrant struct {
	ID         int64     `json:"id"`
	SessionID  string    `json:"session_id"`
	Action     string    `json:"action"`
	TargetPath string    `json:"target_path"`
	Decision   string    `json:"decision"`
	TTL        string    `json:"ttl"`       // "session" | "task" | "temporary" | "permanent"
	GrantedMode string   `json:"granted_mode"` // modo em que foi concedido (ex: "FULL")
	ExpiresAt  time.Time `json:"expires_at"` // zero time = não expira
	CreatedAt  time.Time `json:"created_at"`
}

func (g *PermissionGrant) IsExpired() bool {
	if g.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(g.ExpiresAt)
}

// ModePowerLevel retorna o nível de poder de um modo.
// Quanto maior, mais permissões o modo tem.
func ModePowerLevel(mode ChatMode) int {
	switch mode {
	case ModeAdmin:
		return 100
	case ModeFull:
		return 80
	case ModeExec:
		return 60
	case ModeEdit:
		return 40
	case ModePlan:
		return 20
	case ModeAsk:
		return 10
	}
	return 0
}

// IsUpgrade retorna true se newMode é mais poderoso ou igual a oldMode.
func IsUpgrade(oldMode, newMode ChatMode) bool {
	return ModePowerLevel(newMode) >= ModePowerLevel(oldMode)
}

// ── PermissionRequest ──

type PermissionRequest struct {
	RequestID  string `json:"request_id"`
	SessionID  string `json:"session_id"`
	ToolName   string `json:"tool_name"`
	Args       string `json:"args"`
	Reason     string `json:"reason"`
	TargetPath string `json:"target_path"`
	Mode       string `json:"mode"`
	Action     string `json:"action"`
	RiskLevel  string `json:"risk_level"`
}

// ── PendingToolCall ──

type PendingToolCall struct {
	ToolName string
	ArgsJSON string
	ToolID   string
	Iter     int
	Index    int
}

// ── CheckResult ──

type CheckResult struct {
	Allowed      bool
	Reason       string
	Request      *PermissionRequest
	RiskLevel    RiskLevel
	Action       ActionClass
	NeedsConfirm bool
}

// ── PermissionStore ──

type PermissionStore struct {
	mu            sync.Mutex
	grants        map[string][]PermissionGrant  // sessionID -> grants persistidos
	sessionGrants map[string]map[string]string   // sessionID -> action -> mode (allow_once dentro do turno)
	pending       map[string]*PermissionRequest
	pendExec      map[string]*PendingToolCall
	pendingChans  map[string]chan string // requestID -> channel de decisão
	db            *sql.DB
	nextID        int64
	currentMode   map[string]ChatMode     // sessionID -> modo atual

	// Callbacks
	onRequest func(req *PermissionRequest)
}

func NewPermissionStore(db *sql.DB) *PermissionStore {
	ps := &PermissionStore{
		grants:        make(map[string][]PermissionGrant),
		sessionGrants: make(map[string]map[string]string),
		pending:       make(map[string]*PermissionRequest),
		pendExec:      make(map[string]*PendingToolCall),
		pendingChans:  make(map[string]chan string),
		currentMode:   make(map[string]ChatMode),
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
			ttl TEXT NOT NULL DEFAULT 'session',
			granted_mode TEXT NOT NULL DEFAULT '',
			expires_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
		if err == nil {
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_perm_grants_session ON permission_grants(session_id, action, target_path)`)
		}
		ps.loadFromDB()
	}
	return ps
}

func (ps *PermissionStore) SetOnRequest(fn func(req *PermissionRequest)) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.onRequest = fn
}

func (ps *PermissionStore) loadFromDB() {
	rows, err := ps.db.Query(`SELECT id, session_id, action, target_path, decision, ttl, COALESCE(granted_mode,''), expires_at, created_at FROM permission_grants`)
	if err != nil {
		fmt.Printf("[PermissionStore] load error: %v\n", err)
		return
	}
	defer rows.Close()
	var expiresAt sql.NullTime
	for rows.Next() {
		var g PermissionGrant
		if err := rows.Scan(&g.ID, &g.SessionID, &g.Action, &g.TargetPath, &g.Decision, &g.TTL, &g.GrantedMode, &expiresAt, &g.CreatedAt); err != nil {
			fmt.Printf("[PermissionStore] scan error: %v\n", err)
			continue
		}
		if expiresAt.Valid {
			g.ExpiresAt = expiresAt.Time
		}
		// Pula grants expirados
		if g.IsExpired() {
			ps.mu.Unlock()
			ps.deleteGrantFromDB(g.ID)
			ps.mu.Lock()
			continue
		}
		ps.grants[g.SessionID] = append(ps.grants[g.SessionID], g)
	}
	fmt.Printf("[PermissionStore] loaded %d permissions from DB\n", len(ps.grants))
}

// Check verifica se uma ferramenta pode ser usada no modo atual.
// Retorna um CheckResult com a decisão.
func (ps *PermissionStore) Check(sessionID string, toolName string, argsJSON string, mode ChatMode) *CheckResult {
	cfg := GetModeConfig(mode)
	result := &CheckResult{Allowed: false}

	// 1. Classificar ação e risco
	action, risk := ClassifyAction(toolName, argsJSON, mode)
	result.Action = action
	result.RiskLevel = risk
	targetPath := extractPath(argsJSON)

	// 2. Verificar grants de sessão (allow_once) — valem para o turno atual
	// Estes grants são concedidos pelo usuário via "Permitir uma vez"
	// Verifica se o grant ainda é compatível com o modo atual
	ps.mu.Lock()
	sessionGrantMode, sessionOk := ps.sessionGrants[sessionID][string(action)]
	ps.mu.Unlock()
	if sessionOk {
		if sessionGrantMode != "" {
			// Se o grant foi dado num modo mais poderoso que o atual, invalida
			grantPower := ModePowerLevel(ChatMode(sessionGrantMode))
			currentPower := ModePowerLevel(mode)
			if currentPower < grantPower {
				fmt.Printf("[PermissionStore] SessionGrant invalidado por downgrade: session=%s action=%s grantMode=%s currentMode=%s\n",
					sessionID, string(action), sessionGrantMode, mode)
				ps.GrantAllowOnceRevoke(sessionID, string(action))
			} else {
				result.Allowed = true
				return result
			}
		} else {
			result.Allowed = true
			return result
		}
	}

	// 3. Verificar grants persistidos (allow_session) — só se AllowSessionGrant = true
	// Estes grants são "Sempre permitir nesta sessão"
	if cfg.AllowSessionGrant && action != "" {
		ps.mu.Lock()
		grant := ps.findValidGrant(sessionID, string(action), targetPath, mode)
		ps.mu.Unlock()
		if grant != nil {
			if grant.IsExpired() {
				ps.deleteGrantFromDB(grant.ID)
			} else {
				result.Allowed = true
				return result
			}
		}
	}

	// 4. Verificar se a ação está na lista de ações negadas do modo
	// DeniedActions são ações inerentemente proibidas no modo.
	// Se o modo permite CanOverrideOnce, o usuário pode aprovar mesmo assim.
	for _, denied := range cfg.DeniedActions {
		if denied == action {
			if cfg.CanOverrideOnce {
				reason := fmt.Sprintf("Ação %s não permitida no modo %s (risco: %s)", action, mode, risk)
				result.Request = ps.createRequest(sessionID, toolName, argsJSON, reason, mode, string(action), risk.String())
				result.Reason = reason
				return result
			}
			result.Reason = fmt.Sprintf("Ação %s bloqueada permanentemente no modo %s", action, mode)
			result.Allowed = false
			return result
		}
	}

	// 5. Verificar se o comando está na lista de comandos bloqueados (DeniedCommands)
	// Aplica-se apenas para ações de execução
	if action == ActionExec || action == ActionExecHighRisk {
		cmd := extractCommand(argsJSON)
		if cmd != "" && len(cfg.DeniedCommands) > 0 {
			for _, denied := range cfg.DeniedCommands {
				if stringsHasPrefix(stringsToLower(stringsTrimSpace(cmd)), denied) {
					result.Reason = fmt.Sprintf("Comando '%s' bloqueado no modo %s (alto risco)", cmd, mode)
					result.Request = ps.createRequest(sessionID, toolName, argsJSON, result.Reason, mode, string(action), risk.String())
					return result
				}
			}
		}
	}

	// 6. Verificar se a capacidade da ferramenta está na lista de permitidas
	cap := GetCapabilityForTool(toolName)
	capAllowed := false
	for _, allowed := range cfg.AllowedCapabilities {
		if allowed == cap {
			capAllowed = true
			break
		}
	}

	if !capAllowed {
		if cfg.CanOverrideOnce {
			reason := fmt.Sprintf("%s (%s) não permitida no modo %s (risco: %s)", toolName, cap, mode, risk)
			result.Request = ps.createRequest(sessionID, toolName, argsJSON, reason, mode, string(action), risk.String())
			result.Reason = reason
			return result
		}
		result.Reason = fmt.Sprintf("%s (%s) não permitida no modo %s", toolName, cap, mode)
		result.Allowed = false
		return result
	}

	// 7. Capacidade permitida — decidir se precisa de confirmação extra
	needsConfirm := cfg.needsConfirmation(risk)

	if needsConfirm {
		reason := fmt.Sprintf("Risco %s: %s (tool=%s, action=%s)", risk, GetRiskDescription(action), toolName, action)
		result.Request = ps.createRequest(sessionID, toolName, argsJSON, reason, mode, string(action), risk.String())
		result.Reason = reason
		result.NeedsConfirm = true
		return result
	}

	// 8. Tudo ok — permitido
	result.Allowed = true
	return result
}

// needsConfirmation determina se uma ação no modo atual requer confirmação do usuário.
func (cfg ModeConfig) needsConfirmation(risk RiskLevel) bool {
	// Risco crítico sempre precisa confirmação (safety net universal)
	if risk == RiskCritical {
		return true
	}

	// ADMIN: modo de administração total (exige ativação explícita)
	// Uma vez em ADMIN, todas as ações são auto-autorizadas
	if cfg.Mode == ModeAdmin {
		return false
	}

	// FULL: modo autônomo (exige ativação explícita)
	// Apenas ações de alto risco (High) precisam confirmação
	if cfg.Mode == ModeFull && risk >= RiskHigh {
		return true
	}
	if cfg.Mode == ModeFull {
		return false
	}

	// EXEC: modo de execução controlada
	// Apenas ações de alto risco (High+) precisam confirmação
	// Comandos destrutivos já foram filtrados por DeniedCommands
	if cfg.Mode == ModeExec && risk >= RiskHigh {
		return true
	}
	if cfg.Mode == ModeExec {
		return false
	}

	// EDIT: modo de edição assistida — toda ação com risco > Low precisa confirmação
	if cfg.Mode == ModeEdit && risk > RiskLow {
		return true
	}

	// ASK/PLAN: apenas leitura, riscos baixos/none — sem confirmação
	return false
}

// createRequest cria uma requisição de permissão pendente.
func (ps *PermissionStore) createRequest(sessionID, toolName, argsJSON, reason string, mode ChatMode, action, riskLevel string) *PermissionRequest {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	id := fmt.Sprintf("perm-%d", ps.nextID)
	ps.nextID++

	req := &PermissionRequest{
		RequestID:  id,
		SessionID:  sessionID,
		ToolName:   toolName,
		Args:       argsJSON,
		Reason:     reason,
		TargetPath: extractPath(argsJSON),
		Mode:       string(mode),
		Action:     action,
		RiskLevel:  riskLevel,
	}
	ps.pending[id] = req

	// Notifica callback (ex: frontend)
	if ps.onRequest != nil {
		ps.onRequest(req)
	}

	return req
}

// GetPending retorna uma requisição pendente pelo ID.
func (ps *PermissionStore) GetPending(requestID string) *PermissionRequest {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.pending[requestID]
}

// RemovePending remove uma requisição pendente.
func (ps *PermissionStore) RemovePending(requestID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.pending, requestID)
}

// Grant concede uma permissão persistente.
// Se grantedMode for fornecido, registra em que modo foi concedida.
func (ps *PermissionStore) Grant(sessionID, action, targetPath, decision string, ttl TTLPolicy, grantedMode ...ChatMode) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	expiresAt := computeExpiry(ttl)
	modeStr := ""
	if len(grantedMode) > 0 {
		modeStr = string(grantedMode[0])
	}

	g := PermissionGrant{
		SessionID:   sessionID,
		Action:      action,
		TargetPath:  targetPath,
		Decision:    decision,
		TTL:         string(ttl),
		GrantedMode: modeStr,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
	}
	ps.grants[sessionID] = append(ps.grants[sessionID], g)

	fmt.Printf("[PermissionStore] Grant: session=%s action=%s target=%s decision=%s ttl=%s mode=%s\n",
		sessionID, action, targetPath, decision, ttl, modeStr)

	if ps.db != nil {
		var expiresPtr *time.Time
		if !expiresAt.IsZero() {
			expiresPtr = &expiresAt
		}
		_, err := ps.db.Exec(
			`INSERT INTO permission_grants (session_id, action, target_path, decision, ttl, granted_mode, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			sessionID, action, targetPath, decision, string(ttl), modeStr, expiresPtr, g.CreatedAt,
		)
		if err != nil {
			fmt.Printf("[PermissionStore] db insert error: %v\n", err)
		}
	}
}

// computeExpiry calcula a data de expiração baseada na política TTL.
func computeExpiry(ttl TTLPolicy) time.Time {
	switch ttl {
	case TTLSession:
		return time.Now().Add(24 * time.Hour) // sessão = 24h no máximo
	case TTLTask:
		return time.Now().Add(1 * time.Hour) // tarefa = 1h
	case TTLAction:
		return time.Time{} // action não expira (dura uma única chamada)
	case TTLTemporary:
		return time.Now().Add(15 * time.Minute) // temporário = 15min
	case TTLPermanent:
		return time.Time{} // permanente = nunca expira
	}
	return time.Time{}
}

// GrantAllowOnce concede permissão temporária para o turno atual.
// Registra o modo em que foi concedida para validação em mudanças de modo.
func (ps *PermissionStore) GrantAllowOnce(sessionID, action string, grantedMode ...ChatMode) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.sessionGrants[sessionID] == nil {
		ps.sessionGrants[sessionID] = make(map[string]string)
	}
	mode := ""
	if len(grantedMode) > 0 {
		mode = string(grantedMode[0])
	}
	ps.sessionGrants[sessionID][action] = mode
	fmt.Printf("[PermissionStore] GrantAllowOnce: session=%s action=%s mode=%s\n", sessionID, action, mode)
}

// GrantAllowOnceRevoke revoga uma permissão allow_once específica.
func (ps *PermissionStore) GrantAllowOnceRevoke(sessionID, action string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if grants, ok := ps.sessionGrants[sessionID]; ok {
		delete(grants, action)
		fmt.Printf("[PermissionStore] GrantAllowOnceRevoke: session=%s action=%s\n", sessionID, action)
	}
}

// ClearSessionGrants limpa as permissões temporárias (allow_once) da sessão.
func (ps *PermissionStore) ClearSessionGrants(sessionID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if grants, ok := ps.sessionGrants[sessionID]; ok && len(grants) > 0 {
		fmt.Printf("[PermissionStore] ClearSessionGrants: session=%s grants=%v\n", sessionID, grants)
	}
	delete(ps.sessionGrants, sessionID)
}

// CleanupExpiredGrants remove grants expirados do map e do DB.
func (ps *PermissionStore) CleanupExpiredGrants() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for sessionID, grants := range ps.grants {
		var valid []PermissionGrant
		for _, g := range grants {
			if g.IsExpired() {
				ps.deleteGrantFromDB(g.ID)
			} else {
				valid = append(valid, g)
			}
		}
		if len(valid) == 0 {
			delete(ps.grants, sessionID)
		} else {
			ps.grants[sessionID] = valid
		}
	}
}

// GetCurrentMode retorna o modo atual de uma sessão.
func (ps *PermissionStore) GetCurrentMode(sessionID string) ChatMode {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if mode, ok := ps.currentMode[sessionID]; ok {
		return mode
	}
	return ""
}

// SetCurrentMode registra o modo atual de uma sessão.
// Se houve downgrade (modo novo é menos poderoso), limpa os grants.
// Retorna true se limpou grants, false caso contrário.
func (ps *PermissionStore) SetCurrentMode(sessionID string, newMode ChatMode, emitter Emitter) bool {
	ps.mu.Lock()
	oldMode, exists := ps.currentMode[sessionID]
	ps.currentMode[sessionID] = newMode
	ps.mu.Unlock()

	if !exists {
		fmt.Printf("[PermissionStore] SetCurrentMode: session=%s mode=%s (first)\n", sessionID, newMode)
		return false
	}

	// Upgrade ou modo igual → mantém grants
	if IsUpgrade(oldMode, newMode) {
		fmt.Printf("[PermissionStore] SetCurrentMode: session=%s %s → %s (upgrade, grants mantidos)\n", sessionID, oldMode, newMode)
		return false
	}

	// Downgrade → limpa grants
	fmt.Printf("[PermissionStore] SetCurrentMode: session=%s %s → %s (DOWNGRADE, limpando grants)\n", sessionID, oldMode, newMode)
	ps.clearAllGrantsForSession(sessionID)

	// Notifica frontend
	if emitter != nil {
		emitter.Emit("chat:grants-cleared", map[string]any{
			"session_id": sessionID,
			"old_mode":   string(oldMode),
			"new_mode":   string(newMode),
			"reason":     "downgrade",
		})
	}

	return true
}

// clearAllGrantsForSession limpa todos os grants (session + persisted) de uma sessão.
func (ps *PermissionStore) clearAllGrantsForSession(sessionID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Limpa session grants (allow_once)
	if grants, ok := ps.sessionGrants[sessionID]; ok && len(grants) > 0 {
		fmt.Printf("[PermissionStore] clearAllGrants: session=%s sessionGrants=%v\n", sessionID, grants)
	}
	delete(ps.sessionGrants, sessionID)

	// Limpa grants persistidos
	if grants, ok := ps.grants[sessionID]; ok && len(grants) > 0 {
		fmt.Printf("[PermissionStore] clearAllGrants: session=%s persistedGrants=%+v\n", sessionID, grants)
		for _, g := range grants {
			if g.ID > 0 {
				ps.deleteGrantFromDB(g.ID)
			}
		}
	}
	delete(ps.grants, sessionID)
}

// ListGrants retorna um resumo de todos os grants ativos (para logging).
func (ps *PermissionStore) ListGrants(sessionID string) []map[string]string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	var result []map[string]string

	// Session grants (allow_once)
	if grants, ok := ps.sessionGrants[sessionID]; ok {
		for action, mode := range grants {
			entry := map[string]string{
				"type":   "session",
				"action": action,
				"mode":   mode,
			}
			result = append(result, entry)
		}
	}

	// Persisted grants (allow_session)
	for _, g := range ps.grants[sessionID] {
		if g.IsExpired() {
			continue
		}
		entry := map[string]string{
			"type":         "persisted",
			"action":       g.Action,
			"target_path":  g.TargetPath,
			"mode":         g.GrantedMode,
			"decision":     g.Decision,
			"ttl":          g.TTL,
		}
		if !g.ExpiresAt.IsZero() {
			entry["expires_at"] = g.ExpiresAt.Format(time.RFC3339)
		}
		result = append(result, entry)
	}

	return result
}

// DumpGrants loga todos os grants de uma sessão no terminal.
func (ps *PermissionStore) DumpGrants(sessionID string) {
	grants := ps.ListGrants(sessionID)
	fmt.Printf("\n=== GRANTS FOR SESSION %s ===\n", sessionID)
	if len(grants) == 0 {
		fmt.Println("  (nenhum grant ativo)")
	} else {
		for i, g := range grants {
			fmt.Printf("  %d. [%s] action=%s", i+1, g["type"], g["action"])
			if mode := g["mode"]; mode != "" {
				fmt.Printf(" mode=%s", mode)
			}
			if path := g["target_path"]; path != "" {
				fmt.Printf(" target=%s", path)
			}
			if ttl := g["ttl"]; ttl != "" {
				fmt.Printf(" ttl=%s", ttl)
			}
			if exp := g["expires_at"]; exp != "" {
				fmt.Printf(" expires=%s", exp)
			}
			fmt.Println()
		}
	}
	fmt.Printf("=== END GRANTS ===\n\n")
}
// ResetSession limpa todos os dados de permissão de uma sessão.
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

// findValidGrant busca um grant válido (não expirado) para a sessão/ação.
// Verifica se o grant é compatível com o modo atual (foi concedido num modo
// de poder igual ou superior ao modo atual).
func (ps *PermissionStore) findValidGrant(sessionID, action, targetPath string, currentMode ChatMode) *PermissionGrant {
	grants := ps.grants[sessionID]
	for _, g := range grants {
		if g.IsExpired() {
			continue
		}
		if g.Action != action {
			continue
		}
		if g.TargetPath != "*" && g.TargetPath != targetPath {
			continue
		}
		// Verifica compatibilidade de modo
		if g.GrantedMode != "" {
			grantedPower := ModePowerLevel(ChatMode(g.GrantedMode))
			currentPower := ModePowerLevel(currentMode)
			// Se o modo atual é MENOS poderoso que o modo do grant, o grant é inválido
			if currentPower < grantedPower {
				fmt.Printf("[PermissionStore] Grant inválido por downgrade: session=%s action=%s grantMode=%s currentMode=%s\n",
					sessionID, action, g.GrantedMode, currentMode)
				continue
			}
		}
		return &g
	}
	return nil
}

// deleteGrantFromDB deleta um grant do banco de dados.
func (ps *PermissionStore) deleteGrantFromDB(id int64) {
	if ps.db != nil {
		_, _ = ps.db.Exec(`DELETE FROM permission_grants WHERE id = ?`, id)
	}
}

// ── Helpers de extração ──

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

// ── Métodos de guarda (PermissionGuard) ──

// StorePendingExec guarda um PendingToolCall para uso posterior.
func (ps *PermissionStore) StorePendingExec(requestID string, tc *PendingToolCall) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.pendExec[requestID] = tc
}

// SendDecision envia uma decisão para o canal de uma requisição pendente.
func (ps *PermissionStore) SendDecision(requestID string, decision string) {
	ps.mu.Lock()
	ch, ok := ps.pendingChans[requestID]
	ps.mu.Unlock()
	if ok {
		ch <- decision
	}
}

// RemovePendingChan limpa o canal de decisão pendente.
func (ps *PermissionStore) RemovePendingChan(requestID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.pendingChans, requestID)
}

// AllGrants retorna todos os grants de uma sessão.
func (ps *PermissionStore) AllGrants(sessionID string) []PermissionGrant {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	grants := ps.grants[sessionID]
	if grants == nil {
		return nil
	}
	// Filtra expirados
	var valid []PermissionGrant
	for _, g := range grants {
		if !g.IsExpired() {
			valid = append(valid, g)
		}
	}
	return valid
}

// GetPendingExec retorna um PendingToolCall pelo ID.
func (ps *PermissionStore) GetPendingExec(requestID string) *PendingToolCall {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.pendExec[requestID]
}

// RemovePendingExec remove um PendingToolCall.
func (ps *PermissionStore) RemovePendingExec(requestID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.pendExec, requestID)
}

// MakeGuard cria um adapters.PermissionGuard que usa o PermissionStore.
func (ps *PermissionStore) MakeGuard(ctx context.Context, sessionID string, mode ChatMode, emitter Emitter) adapters.PermissionGuard {
	return func(toolName, argsJSON string) (bool, string, string) {
		result := ps.Check(sessionID, toolName, argsJSON, mode)

		if result.Allowed {
			return true, "", ""
		}

		if result.Request != nil {
			req := result.Request

			emitter.Emit("chat:permission-request", map[string]any{
				"session_id":  req.SessionID,
				"request_id":  req.RequestID,
				"tool_name":   req.ToolName,
				"args":        req.Args,
				"reason":      req.Reason,
				"target_path": req.TargetPath,
				"mode":        req.Mode,
				"action":      req.Action,
				"risk_level":  req.RiskLevel,
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
				case string(DecisionAllowOnce):
					action := string(result.Action)
					if action != "" {
						ps.GrantAllowOnce(sessionID, action, mode)
					}
					return true, "", ""
				case string(DecisionAllowSession):
					action := string(result.Action)
					if action != "" {
						cfg := GetModeConfig(mode)
						ps.Grant(sessionID, action, "*", string(DecisionAllowSession), cfg.DefaultTTL, mode)
					}
					return true, "", ""
				case string(DecisionDeny):
					return false, "negado pelo usuário", ""
				}
			case <-ctx.Done():
				ps.mu.Lock()
				delete(ps.pendingChans, req.RequestID)
				ps.mu.Unlock()
				return false, "stream cancelado", ""
			}
			return false, result.Reason, ""
		}

		return false, result.Reason, ""
	}
}

// GetRiskDescription retorna a descrição textual de uma ação.
func GetRiskDescription(action ActionClass) string {
	for _, rc := range DefaultRiskMatrix {
		if rc.Action == action {
			return rc.Description
		}
	}
	return string(action)
}

// CreatePermissionRequest cria uma requisição e notifica o callback.
func (ps *PermissionStore) CreatePermissionRequest(sessionID, toolName, argsJSON, reason string, mode ChatMode) *PermissionRequest {
	return ps.createRequest(sessionID, toolName, argsJSON, reason, mode, "", "")
}

// ── Helpers de string (evitam import cycles) ──

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func stringsToLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func stringsTrimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}


