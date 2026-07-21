package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ada-love-ide/internal/config/agent"
	"ada-love-ide/internal/config/worker"
	"ada-love-ide/internal/config/workspace"
	core "ada-love-core"

	storage "github.com/ada-love-ai/storage/storage"
)

// ── Sessions ──────────────────────────────────────────────────

func (s *Store) InsertSession(sess *core.Session) {
	ctx := context.Background()
	ss := &storage.Session{
		ID:              sess.ID,
		WorkspacePath:   sql.NullString{String: sess.WorkspaceID, Valid: sess.WorkspaceID != ""},
		Title:           sql.NullString{String: sess.Title, Valid: sess.Title != ""},
		Pinned:          sess.Pinned,
		WorkerName:      sess.WorkerName,
		ParentSessionID: sess.ParentSessionID,
		Model:           sess.Model,
		Provider:        sess.Provider,
		Mode:            sess.Mode,
		Thinking:        sql.NullString{String: sess.Thinking, Valid: sess.Thinking != ""},
		Summary:         sql.NullString{String: sess.Summary, Valid: sess.Summary != ""},
		CreatedAt:       sess.CreatedAt,
		UpdatedAt:       sess.UpdatedAt,
	}

	// Tenta atualizar se já existe, senão cria
	_, err := s.sessions.GetSession(ctx, sess.ID)
	if err == nil {
		_ = s.sessions.UpdateSession(ctx, ss)
	} else {
		_ = s.sessions.CreateSession(ctx, ss)
	}
}

func (s *Store) PutSession(sess *core.Session) {
	s.InsertSession(sess)
}

func (s *Store) GetSession(id string) (*core.Session, bool) {
	ctx := context.Background()
	ss, err := s.sessions.GetSession(ctx, id)
	if err != nil {
		return nil, false
	}
	sess := adaptSessionToInternal(ss)
	return &sess, true
}

func (s *Store) DeleteSession(id string) {
	_ = s.sessions.DeleteSession(context.Background(), id)
}

func (s *Store) ListSessions(workspaceID string) []*core.Session {
	ctx := context.Background()
	raw, _ := s.sessions.ListSessions(ctx, workspaceID)
	out := make([]*core.Session, 0, len(raw))
	for _, ss := range raw {
		sess := adaptSessionToInternal(&ss)
		out = append(out, &sess)
	}
	return out
}

func adaptSessionToInternal(ss *storage.Session) core.Session {
	return core.Session{
		ID:              ss.ID,
		WorkspaceID:     ss.WorkspacePath.String,
		WorkerName:      ss.WorkerName,
		ParentSessionID: ss.ParentSessionID,
		Title:           ss.Title.String,
		Summary:         ss.Summary.String,
		Model:           ss.Model,
		Provider:        ss.Provider,
		Mode:            ss.Mode,
		Thinking:        ss.Thinking.String,
		CreatedAt:       ss.CreatedAt,
		UpdatedAt:       ss.UpdatedAt,
		Pinned:          ss.Pinned,
	}
}

// ── Messages ──────────────────────────────────────────────────

func (s *Store) AppendMessage(sessionID string, msg core.RawMessage) {
	ctx := context.Background()
	sm := &storage.Message{
		SessionID: sessionID,
		Role:      msg.Role,
		Content:   msg.Content,
		Tokens:    0,
		Time:      msg.Time,
	}
	_ = s.sessions.SaveMessage(ctx, sm)
}

func (s *Store) GetMessages(sessionID string) []core.RawMessage {
	ctx := context.Background()
	raw, _ := s.sessions.GetMessages(ctx, sessionID)
	out := make([]core.RawMessage, 0, len(raw))
	for _, m := range raw {
		out = append(out, core.RawMessage{
			Role:    m.Role,
			Content: m.Content,
			Time:    m.Time,
		})
	}
	return out
}

// DeleteMessages removes all messages for a session from the database.
func (s *Store) DeleteMessages(ctx context.Context, sessionID string) error {
	return s.sessions.DeleteMessages(ctx, sessionID)
}

// ── Workspaces ─────────────────────────────────────────────────

func (s *Store) ListWorkspaces() []workspace.WorkspaceConfig {
	ctx := context.Background()
	raw, _ := s.workspaces.ListWorkspaces(ctx)
	out := make([]workspace.WorkspaceConfig, 0, len(raw))
	for i := range raw {
		out = append(out, adaptWorkspaceToInternal(ctx, s, &raw[i]))
	}
	return out
}

func (s *Store) SetWorkspaces(list []workspace.WorkspaceConfig) {
	ctx := context.Background()
	existing, _ := s.workspaces.ListWorkspaces(ctx)
	incomingPaths := make(map[string]bool)

	for _, ws := range list {
		incomingPaths[ws.Path] = true
		s.upsertWorkspace(ctx, ws)
	}

	// Cleanup
	for _, e := range existing {
		if e.Path.Valid && !incomingPaths[e.Path.String] {
			_ = s.workspaces.DeleteWorkspace(ctx, e.ID)
		}
	}
}

func (s *Store) AddWorkspace(ws workspace.WorkspaceConfig) {
	s.upsertWorkspace(context.Background(), ws)
}

func (s *Store) DeleteWorkspace(path string) {
	ctx := context.Background()
	w, err := s.workspaces.GetWorkspaceByPath(ctx, path)
	if err != nil {
		return
	}
	_ = s.workspaces.DeleteWorkspace(ctx, w.ID)
}

// GetWorkspace returns a workspace by its path.
func (s *Store) GetWorkspace(path string) (workspace.WorkspaceConfig, error) {
	ctx := context.Background()
	w, err := s.workspaces.GetWorkspaceByPath(ctx, path)
	if err != nil {
		return workspace.WorkspaceConfig{}, err
	}
	return adaptWorkspaceToInternal(ctx, s, w), nil
}

// UpdateWorkspace updates an existing workspace.
func (s *Store) UpdateWorkspace(ws workspace.WorkspaceConfig) error {
	s.upsertWorkspace(context.Background(), ws)
	return nil
}

func (s *Store) SetActiveWorkspace(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeWorkspace = path
	_ = s.config.SetConfig(context.Background(), "active_workspace", path)
}

func (s *Store) ActiveWorkspace() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeWorkspace
}

func (s *Store) ActiveSessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeSessionId
}

func (s *Store) SetActiveSessionID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeSessionId = id
	_ = s.config.SetConfig(context.Background(), "active_session_id", id)
}

func (s *Store) SidebarVisible() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sidebarVisible
}

func (s *Store) SetSidebarVisible(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sidebarVisible = v
	val := "true"
	if !v {
		val = "false"
	}
	_ = s.config.SetConfig(context.Background(), "sidebar_visible", val)
}

func (s *Store) upsertWorkspace(ctx context.Context, ws workspace.WorkspaceConfig) {
	sw := &storage.Workspace{
		Nome:             ws.Title,
		Description:      sql.NullString{String: ws.Description, Valid: ws.Description != ""},
		Path:             sql.NullString{String: ws.Path, Valid: ws.Path != ""},
		Enabled:          ws.Enabled,
		Color:            ws.Color,
		Icon:             ws.Icon,
		MaxPromptSend:    ws.MaxPromptSend,
		CommitChanges:    ws.CommitChanges,
		MaxContextLength: ws.MaxContextLength,
		Personality:      sql.NullString{String: ws.Personality, Valid: ws.Personality != ""},
		RoutingRules:     sql.NullString{String: ws.RoutingRules, Valid: ws.RoutingRules != ""},
		SpecWizardID:     sql.NullString{String: ws.SpecWizardID, Valid: ws.SpecWizardID != ""},
	}

	// Tenta carregar pelo path primeiro
	existing, err := s.workspaces.GetWorkspaceByPath(ctx, ws.Path)
	if err == nil {
		sw.ID = existing.ID
		fmt.Printf("[DB] upsertWorkspace: found existing by path %s, ID %d\n", ws.Path, sw.ID)
		_ = s.workspaces.UpdateWorkspace(ctx, sw)
	} else {
		// Se não encontrou pelo path, tenta pelo nome (fallback)
		rows, _ := s.workspaces.ListWorkspaces(ctx)
		for _, r := range rows {
			if r.Nome == ws.Title {
				sw.ID = r.ID
				fmt.Printf("[DB] upsertWorkspace: found existing by name %s, ID %d\n", ws.Title, sw.ID)
				_ = s.workspaces.UpdateWorkspace(ctx, sw)
				err = nil
				break
			}
		}
		if err != nil {
			_ = s.workspaces.CreateWorkspace(ctx, sw)
			// Recupera o ID gerado
			created, err := s.workspaces.GetWorkspaceByPath(ctx, ws.Path)
			if err == nil {
				sw.ID = created.ID
				fmt.Printf("[DB] upsertWorkspace: created new workspace, ID %d\n", sw.ID)
			} else {
				fmt.Printf("[DB] upsertWorkspace: FAILED to recover ID for new workspace %s\n", ws.Title)
			}
		}
	}

	_ = s.folders.DeleteAllByWorkspace(ctx, sw.ID)
	for _, folder := range ws.Folders {
		_ = s.folders.Create(ctx, &storage.WorkspaceFolder{
			WorkspaceID: sw.ID,
			FolderPath:  folder,
		})
	}

	_ = s.knowledge.DeleteAllKnowledge(ctx, sw.ID)
	for _, item := range ws.Knowledge {
		_ = s.knowledge.AddKnowledge(ctx, &storage.WorkspaceKnowledge{
			WorkspaceID:   sw.ID,
			KnowledgeItem: item,
		})
	}

	_ = s.workspaceSkills.DeleteByWorkspace(ctx, sw.ID)
	skills, _ := s.skills.ListSkills(ctx)
	skillMap := make(map[string]int64)
	for _, sk := range skills {
		skillMap[sk.Name] = sk.ID
	}
	for _, skillName := range ws.Skills {
		skillID, exists := skillMap[skillName]
		if !exists {
			newSkill := &storage.Skill{
				Name: skillName,
			}
			_ = s.skills.CreateSkill(ctx, newSkill)
			skillID = newSkill.ID
		}
		if skillID != 0 {
			_ = s.workspaceSkills.Create(ctx, &storage.WorkspaceSkill{
				WorkspaceID: sw.ID,
				SkillID:     skillID,
				Enabled:     true,
			})
		}
	}

	_ = s.tools.DeleteAllTools(ctx, sw.ID)
	for _, toolName := range ws.Tools {
		_ = s.tools.AddTool(ctx, &storage.WorkspaceTool{
			WorkspaceID: sw.ID,
			ToolName:    toolName,
			Enabled:     true,
		})
	}

	_ = s.workspaceWorkers.DeleteAllWorkers(ctx, sw.ID)
	workers, _ := s.workers.ListWorkers(ctx)
	workerMap := make(map[string]int64)
	for _, w := range workers {
		workerMap[w.Name] = w.ID
	}
	fmt.Printf("[DB] upsertWorkspace: workspaceID=%d, workerNames=%v\n", sw.ID, ws.WorkerNames)
	for _, workerName := range ws.WorkerNames {
		if workerID, ok := workerMap[workerName]; ok {
			fmt.Printf("[DB] upsertWorkspace: linking worker %s (ID %d) to workspace %d\n", workerName, workerID, sw.ID)
			_ = s.workspaceWorkers.AddWorker(ctx, &storage.WorkspaceWorker{
				WorkspaceID: sw.ID,
				WorkerID:    workerID,
				Enabled:     true,
			})
		} else {
			fmt.Printf("[DB] upsertWorkspace: worker %s not found in global list\n", workerName)
		}
	}

	_ = s.workspaceAgents.DeleteAllAgents(ctx, sw.ID)
	agents, _ := s.agents.ListAgents(ctx)
	agentMap := make(map[string]int64)
	for _, a := range agents {
		agentMap[a.Name] = a.ID
	}
	for _, agentName := range ws.Agents {
		if agentID, ok := agentMap[agentName]; ok {
			_ = s.workspaceAgents.AddAgent(ctx, &storage.WorkspaceAgent{
				WorkspaceID: sw.ID,
				AgentID:     agentID,
				Enabled:     true,
			})
		}
	}
}

func adaptWorkspaceToInternal(ctx context.Context, s *Store, w *storage.Workspace) workspace.WorkspaceConfig {
	folders, _ := s.folders.ListByWorkspace(ctx, w.ID)
	folderPaths := make([]string, len(folders))
	for i, f := range folders {
		folderPaths[i] = f.FolderPath
	}

	knowledge, _ := s.knowledge.ListKnowledge(ctx, w.ID)
	knowledgeItems := make([]string, len(knowledge))
	for i, k := range knowledge {
		knowledgeItems[i] = k.KnowledgeItem
	}

	workspaceSkills, _ := s.workspaceSkills.ListByWorkspace(ctx, w.ID)
	skillNames := make([]string, 0, len(workspaceSkills))
	for _, ws := range workspaceSkills {
		if skill, err := s.skills.GetSkill(ctx, ws.SkillID); err == nil {
			skillNames = append(skillNames, skill.Name)
		}
	}

	tools, _ := s.tools.ListTools(ctx, w.ID)
	toolNames := make([]string, len(tools))
	for i, t := range tools {
		toolNames[i] = t.ToolName
	}

	workspaceWorkers, _ := s.workspaceWorkers.ListWorkers(ctx, w.ID)
	workerNames := make([]string, 0, len(workspaceWorkers))
	fmt.Printf("[DB] adaptWorkspace: workspaceID=%d, found %d links in SQL\n", w.ID, len(workspaceWorkers))
	for _, ww := range workspaceWorkers {
		if worker, err := s.workers.GetWorker(ctx, ww.WorkerID); err == nil {
			workerNames = append(workerNames, worker.Name)
		} else {
			fmt.Printf("[DB] adaptWorkspace: failed to get worker %d: %v\n", ww.WorkerID, err)
		}
	}

	workspaceAgents, _ := s.workspaceAgents.ListAgents(ctx, w.ID)
	agentNames := make([]string, 0, len(workspaceAgents))
	for _, wa := range workspaceAgents {
		if agent, err := s.agents.GetAgent(ctx, wa.AgentID); err == nil {
			agentNames = append(agentNames, agent.Name)
		}
	}

	return workspace.WorkspaceConfig{
		Title:            w.Nome,
		Description:      w.Description.String,
		Path:             w.Path.String,
		Folders:          folderPaths,
		Personality:      w.Personality.String,
		RoutingRules:     w.RoutingRules.String,
		WorkerNames:      workerNames,
		Knowledge:        knowledgeItems,
		Skills:           skillNames,
		Tools:            toolNames,
		Enabled:          w.Enabled,
		Color:            w.Color,
		Icon:             w.Icon,
		MaxPromptSend:    w.MaxPromptSend,
		CommitChanges:    w.CommitChanges,
		MaxContextLength: w.MaxContextLength,
		SpecWizard:       "",
		SpecWizardID:     w.SpecWizardID.String,
		Agents:           agentNames,
	}
}

// ── Workers ────────────────────────────────────────────────────

func (s *Store) ListWorkers() []worker.WorkerConfig {
	ctx := context.Background()
	raw, _ := s.workers.ListWorkers(ctx)
	out := make([]worker.WorkerConfig, 0, len(raw))
	for i := range raw {
		out = append(out, adaptWorkerToInternal(&raw[i]))
	}
	return out
}

func (s *Store) SetWorkers(list []worker.WorkerConfig) {
	ctx := context.Background()
	existing, _ := s.workers.ListWorkers(ctx)
	incomingNames := make(map[string]bool)

	for _, wc := range list {
		incomingNames[wc.Name] = true
		s.PutWorker(wc)
	}

	// Cleanup: deleta os que não vieram na nova lista
	// Mas bloqueia se o worker tiver cópias em workspaces
	for _, e := range existing {
		if !incomingNames[e.Name] {
			if s.workerHasCopies(ctx, e.ID) {
				fmt.Printf("[db] SetWorkers: worker %q has workspace copies, skipping delete\n", e.Name)
				continue
			}
			_ = s.workers.DeleteWorker(ctx, e.ID)
		}
	}
}

func (s *Store) workerHasCopies(ctx context.Context, workerID int64) bool {
	allWorkspaces, _ := s.workspaces.ListWorkspaces(ctx)
	for _, ws := range allWorkspaces {
		links, err := s.workspaceWorkers.ListWorkers(ctx, ws.ID)
		if err != nil {
			continue
		}
		for _, link := range links {
			if link.WorkerID == workerID {
				return true
			}
		}
	}
	return false
}

func (s *Store) DeleteWorker(name string) error {
	ctx := context.Background()
	raw, err := s.workers.GetWorkerByName(ctx, name)
	if err != nil {
		return err
	}
	if s.workerHasCopies(ctx, raw.ID) {
		return fmt.Errorf("worker %q possui cópias em workspaces e não pode ser apagado", name)
	}
	return s.workers.DeleteWorker(ctx, raw.ID)
}

func (s *Store) PutWorker(wc worker.WorkerConfig) {
	ctx := context.Background()
	sw := &storage.Worker{
		ID:                   wc.ID,
		Name:                 wc.Name,
		Persona:              sql.NullString{String: wc.Persona, Valid: wc.Persona != ""},
		ResponseLanguage:     wc.Language,
		ConnectionType:       wc.ConnectionType,
		Command:              sql.NullString{String: wc.ConnectionConfig, Valid: wc.ConnectionConfig != ""},
		Color:                wc.Color,
		Icon:                 wc.Icon,
		InheritanceFolders:   wc.InheritFolders,
		InheritanceSkills:    wc.InheritSkills,
		InheritancePersona:   wc.InheritPersona,
		InheritanceKnowledge: wc.InheritKnowledge,
		InheritanceTools:     wc.InheritTools,
	}

	if wc.ID > 0 {
		_ = s.workers.UpdateWorker(ctx, sw)
	} else {
		_ = s.workers.CreateWorker(ctx, sw)
	}
}

func (s *Store) GetWorker(name string) (worker.WorkerConfig, error) {
	ctx := context.Background()
	raw, err := s.workers.GetWorkerByName(ctx, name)
	if err != nil {
		return worker.WorkerConfig{}, err
	}
	return adaptWorkerToInternal(raw), nil
}

func adaptWorkerToInternal(w *storage.Worker) worker.WorkerConfig {
	return worker.WorkerConfig{
		ID:               w.ID,
		Name:             w.Name,
		Persona:          w.Persona.String,
		Language:         w.ResponseLanguage,
		Icon:             w.Icon,
		Color:            w.Color,
		ConnectionType:   w.ConnectionType,
		ConnectionConfig: w.Command.String,
		InheritFolders:   w.InheritanceFolders,
		InheritKnowledge: w.InheritanceKnowledge,
		InheritSkills:    w.InheritanceSkills,
		InheritTools:     w.InheritanceTools,
		InheritPersona:   w.InheritancePersona,
	}
}

func adaptAgentToInternal(a *storage.Agent) agent.AgentConfig {
	return agent.AgentConfig{
		ID:            a.ID,
		Name:          a.Name,
		Description:   a.Description.String,
		Type:          string(a.Type),
		Icon:          a.Icon,
		Color:         a.Color,
		MaxIterations: a.MaxIteration,
		Temperature:   a.Temperature,
		SystemPrompt:  a.SystemPrompt.String,
	}
}

// ── Categories ─────────────────────────────────────────────────

func (s *Store) WorkerCategories() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.workerCategories))
	copy(out, s.workerCategories)
	return out
}
func (s *Store) SetWorkerCategories(c []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workerCategories = append([]string{}, c...)
	saveStringSlice(context.Background(), s.config, "worker_categories", c)
}

func (s *Store) AgentCategories() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.agentCategories))
	copy(out, s.agentCategories)
	return out
}
func (s *Store) SetAgentCategories(c []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agentCategories = append([]string{}, c...)
	saveStringSlice(context.Background(), s.config, "agent_categories", c)
}

// ── ensure time import used ───────────────────────────────────
var _ = time.Now
