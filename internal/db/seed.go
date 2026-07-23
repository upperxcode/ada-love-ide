package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ada-love-ide/internal/config/worker"

	storage "github.com/ada-love-ai/storage/storage"
)

// seedIfEmpty populates the DB with mock data only if providers table is empty.
func (s *Store) seedIfEmpty(ctx context.Context) {
	if v, err := s.config.GetConfig(ctx, "seed_initialized"); err == nil && v == "true" {
		s.loadConfigDefaults(ctx)
		return
	}

	s.seedWorkers(ctx)
	s.seedAgents(ctx)
	s.seedWorkspaces(ctx)
	s.seedWorkspaceRelations(ctx)
	s.seedProviders(ctx)
	s.seedProfiles()
	_ = s.config.SetConfig(ctx, "seed_initialized", "true")
}

func (s *Store) seedWorkspaces(ctx context.Context) {
	ws := &storage.Workspace{
		Nome:        "Ada",
		Path:        sql.NullString{String: "/tmp/ada", Valid: true},
		Description: sql.NullString{String: "Workspace padrão", Valid: true},
		Icon:        "📂",
		Enabled:     true,
	}
	_ = s.workspaces.CreateWorkspace(ctx, ws)
	s.activeWorkspace = ws.Path.String
	_ = s.config.SetConfig(ctx, "active_workspace", ws.Path.String)
}

func (s *Store) seedWorkspaceRelations(ctx context.Context) {
	ctx = context.Background()
	ws, _ := s.workspaces.GetWorkspaceByPath(ctx, "/tmp/ada")
	workers, _ := s.workers.ListWorkers(ctx)
	agents, _ := s.agents.ListAgents(ctx)

	var workerID, agentID int64
	for _, w := range workers {
		if w.Name == "Ada" {
			workerID = w.ID
			break
		}
	}
	for _, a := range agents {
		if a.Name == "Ada" {
			agentID = a.ID
			break
		}
	}

	if ws != nil && workerID != 0 {
		_ = s.workspaceWorkers.AddWorker(ctx, &storage.WorkspaceWorker{
			WorkspaceID: ws.ID,
			WorkerID:    workerID,
			Enabled:     true,
		})
	}
	if ws != nil && agentID != 0 {
		_ = s.workspaceAgents.AddAgent(ctx, &storage.WorkspaceAgent{
			WorkspaceID: ws.ID,
			AgentID:     agentID,
			Enabled:     true,
		})
	}
}

func (s *Store) seedWorkers(ctx context.Context) {
	w := worker.New("Ada")
	w.Persona = "Assistant"
	w.Color = "#3b82f6"
	sw := &storage.Worker{
		Name:             w.Name,
		Persona:          sql.NullString{String: w.Persona, Valid: true},
		ResponseLanguage: w.Language,
		ConnectionType:   w.ConnectionType,
		Color:            w.Color,
		Icon:             w.Icon,
	}
	_ = s.workers.CreateWorker(ctx, sw)
}

func (s *Store) seedAgents(ctx context.Context) {
	a := &storage.Agent{
		Name:         "Ada",
		Description:  sql.NullString{String: "Your AI assistant", Valid: true},
		Type:         storage.AgentType("assistant"),
		MaxIteration: 10,
		Temperature:  0.7,
		SystemPrompt: sql.NullString{String: "You are a helpful assistant.", Valid: true},
		Color:        "#3b82f6",
		Icon:         "🤖",
	}
	_ = s.agents.CreateAgent(ctx, a)
}

func (s *Store) seedProviders(ctx context.Context) {
	sp := &storage.Provider{
		Name:            "mock",
		APIURL:          sql.NullString{String: "http://localhost:8080/v1", Valid: true},
		ConnectionTypes: sql.NullString{String: "openai", Valid: true},
		Color:           "#10b981",
		Icon:            "🧪",
	}
	_ = s.providers.CreateProvider(ctx, sp)

	p, _ := s.providers.GetProviderByName(ctx, "mock")
	models := []storage.ProviderModel{
		{ProviderID: p.ID, Model: "echo-1", Free: true, Health: 100},
		{ProviderID: p.ID, Model: "echo-2", Tool: true, Health: 100},
		{ProviderID: p.ID, Model: "thinker", Thinking: true, Health: 100},
	}
	for _, m := range models {
		_ = s.models.CreateProviderModel(ctx, &m)
	}
}

func (s *Store) seedProfiles() {
	ctx := context.Background()
	sp := &storage.ToolProfile{
		Name:  "Default",
		Color: "#3b82f6",
		Icon:  "⚙️",
	}
	if err := s.toolProfiles.CreateProfile(ctx, sp); err != nil {
		fmt.Printf("[DB] seedProfiles: error creating default profile: %v\n", err)
	}
	// Seed default tools for the Default profile
	defaultTools := []string{"read", "write", "search"}
	for _, toolName := range defaultTools {
		_ = s.toolProfiles.AddTool(ctx, &storage.ToolProfileTool{
			ProfileID: sp.ID,
			ToolName:  toolName,
		})
	}

	s.mu.Lock()
	s.nextTemplateID = 1
	s.workerCategories = []string{"Geral"}
	s.agentCategories = []string{}
	s.mu.Unlock()
}

func (s *Store) seedGreetings(ctx context.Context) {
	greetings := []storage.Greeting{
		{
			Keyword:  "hello,hi,hey,oi,olá",
			Response: "Olá! Como posso ajudar?",
		},
		{
			Keyword:  "bom dia,boa tarde,boa noite",
			Response: "Olá! Como posso ajudar?",
		},
	}
	for _, g := range greetings {
		_ = s.greetings.CreateGreeting(ctx, &g)
	}
}

func (s *Store) ResetToFactoryDefaults(ctx context.Context) {
	db := s.engine.DB()
	_, _ = db.ExecContext(ctx, `
		DELETE FROM workspace_agents;
		DELETE FROM workspace_workers;
		DELETE FROM workspace_tools;
		DELETE FROM workspace_knowledge;
		DELETE FROM workspace_skills;
		DELETE FROM workspace_folders;
		DELETE FROM workspaces;
		DELETE FROM providers;
		DELETE FROM agents;
		DELETE FROM workers;
		DELETE FROM greetings;
		DELETE FROM session_attachments;
		DELETE FROM sessions;
		DELETE FROM memories;
	`)
	_ = s.config.DeleteConfig(ctx, "seed_initialized")
	s.seedIfEmpty(ctx)
}

// suppress unused import
var _ = time.Now
