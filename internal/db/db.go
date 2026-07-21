package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"

	"ada-love-ide/internal/config/mcp"
	"ada-love-ide/internal/config/workspace"

	storage "github.com/ada-love-ai/storage/storage"
)

type Store struct {
	mu sync.Mutex

	engine           *storage.StorageEngine
	sessions         *storage.SessionStore
	providers        *storage.ProviderStore
	models           *storage.ProviderModelStore
	workspaces       *storage.WorkspaceStore
	agents           *storage.AgentStore
	workers          *storage.WorkerStore
	skills           *storage.SkillStore
	config           *storage.ConfigStore
	greetings        *storage.GreetingStore
	memories         *storage.MemoryStore
	folders          *storage.WorkspaceFolderStore
	knowledge        *storage.WorkspaceKnowledgeStore
	workspaceSkills  *storage.WorkspaceSkillsStore
	tools            *storage.WorkspaceToolsStore
	workspaceWorkers *storage.WorkspaceWorkersStore
	workspaceAgents  *storage.WorkspaceAgentsStore
	toolProfiles     *storage.ToolProfileStore
	mcps             *storage.McpStore
	fixedModels      *storage.FixedModelStore
	specWizards      *storage.SpecWizardStore

	templates  map[int]workspace.WorkspaceTemplate
	mcpServers map[string]mcp.MCPServerUI

	workerCategories []string
	agentCategories  []string
	activeWorkspace  string
	activeSessionId  string
	sidebarVisible   bool

	nextTemplateID int
}

func New(dbPath string) (*Store, error) {
	ctx := context.Background()

	engine, sessionStore, configStore, err := storage.Init(ctx, dbPath)
	if err != nil {
		return nil, err
	}

	db := engine.DB()
	s := &Store{
		engine:           engine,
		sessions:         sessionStore,
		providers:        storage.NewProviderStore(db),
		models:           storage.NewProviderModelStore(db),
		workspaces:       storage.NewWorkspaceStore(db),
		agents:           storage.NewAgentStore(db),
		workers:          storage.NewWorkerStore(db),
		skills:           storage.NewSkillStore(db),
		config:           configStore,
		greetings:        storage.NewGreetingStore(db),
		memories:         storage.NewMemoryStore(db),
		folders:          storage.NewWorkspaceFolderStore(db),
		knowledge:        storage.NewWorkspaceKnowledgeStore(db),
		workspaceSkills:  storage.NewWorkspaceSkillsStore(db),
		tools:            storage.NewWorkspaceToolsStore(db),
		workspaceWorkers: storage.NewWorkspaceWorkersStore(db),
		workspaceAgents:  storage.NewWorkspaceAgentsStore(db),
		toolProfiles:     storage.NewToolProfileStore(db),
		mcps:             storage.NewMcpStore(db),
		fixedModels:      storage.NewFixedModelStore(db),
		specWizards:      storage.NewSpecWizardStore(db),
		templates:        map[int]workspace.WorkspaceTemplate{},
		mcpServers:       map[string]mcp.MCPServerUI{},
		sidebarVisible:   true,
	}

	s.loadConfigDefaults(ctx)
	s.seedIfEmpty(ctx)
	return s, nil
}

func (s *Store) DB() *sql.DB { return s.engine.DB() }

func (s *Store) Close() error { return s.engine.Close() }

func (s *Store) Greetings() *storage.GreetingStore { return s.greetings }

func (s *Store) Sessions() *storage.SessionStore { return s.sessions }

func (s *Store) Folders() *storage.WorkspaceFolderStore { return s.folders }

func (s *Store) Knowledge() *storage.WorkspaceKnowledgeStore { return s.knowledge }

func (s *Store) WorkspaceSkills() *storage.WorkspaceSkillsStore { return s.workspaceSkills }

func (s *Store) Tools() *storage.WorkspaceToolsStore { return s.tools }

func (s *Store) WorkspaceWorkers() *storage.WorkspaceWorkersStore { return s.workspaceWorkers }

func (s *Store) WorkspaceAgents() *storage.WorkspaceAgentsStore { return s.workspaceAgents }

func (s *Store) ToolProfiles() *storage.ToolProfileStore { return s.toolProfiles }

func (s *Store) Mcps() *storage.McpStore { return s.mcps }

func (s *Store) FixedModels() *storage.FixedModelStore { return s.fixedModels }

func (s *Store) WorkspaceStore() *storage.WorkspaceStore { return s.workspaces }

func (s *Store) WorkerStore() *storage.WorkerStore { return s.workers }

func (s *Store) loadConfigDefaults(ctx context.Context) {
	if v, err := s.config.GetConfig(ctx, "active_workspace"); err == nil {
		s.activeWorkspace = v
	}
	if v, err := s.config.GetConfig(ctx, "worker_categories"); err == nil {
		s.workerCategories = parseStringSlice(v)
	}
	if v, err := s.config.GetConfig(ctx, "agent_categories"); err == nil {
		s.agentCategories = parseStringSlice(v)
	}
	if v, err := s.config.GetConfig(ctx, "active_session_id"); err == nil {
		s.activeSessionId = v
	}
	if v, err := s.config.GetConfig(ctx, "sidebar_visible"); err == nil {
		s.sidebarVisible = v != "false"
	}
	// Carrega servidores MCP do banco para o mapa em memória
	s.loadMCPServersFromSQL(ctx)
}

func (s *Store) loadMCPServersFromSQL(ctx context.Context) {
	servers := s.ListMCPServers()
	s.mu.Lock()
	s.mcpServers = servers
	s.mu.Unlock()
}

func parseStringSlice(v string) []string {
	if v == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(v), &out); err != nil {
		return []string{}
	}
	return out
}

func saveStringSlice(ctx context.Context, cfg *storage.ConfigStore, key string, vals []string) {
	b, _ := json.Marshal(vals)
	_ = cfg.SetConfig(ctx, key, string(b))
}
