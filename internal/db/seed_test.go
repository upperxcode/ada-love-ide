package db

import (
	"context"
	"os"
	"testing"

	"ada-love-ide/internal/config/mcp"
	"ada-love-ide/internal/config/workspace"

	storage "github.com/ada-love-ai/storage/storage"
)

func TestSeedIfEmpty(t *testing.T) {
	tmpFile := "/tmp/test_seed.db"
	os.Remove(tmpFile)
	defer os.Remove(tmpFile)

	engine, _, _, err := storage.Init(context.Background(), tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	db := &Store{
		engine:           engine,
		providers:        storage.NewProviderStore(engine.DB()),
		models:           storage.NewProviderModelStore(engine.DB()),
		workspaces:       storage.NewWorkspaceStore(engine.DB()),
		agents:           storage.NewAgentStore(engine.DB()),
		workers:          storage.NewWorkerStore(engine.DB()),
		skills:           storage.NewSkillStore(engine.DB()),
		config:           storage.NewConfigStore(engine.DB()),
		greetings:        storage.NewGreetingStore(engine.DB()),
		folders:          storage.NewWorkspaceFolderStore(engine.DB()),
		knowledge:        storage.NewWorkspaceKnowledgeStore(engine.DB()),
		workspaceSkills:  storage.NewWorkspaceSkillsStore(engine.DB()),
		tools:            storage.NewWorkspaceToolsStore(engine.DB()),
		workspaceWorkers: storage.NewWorkspaceWorkersStore(engine.DB()),
		workspaceAgents:  storage.NewWorkspaceAgentsStore(engine.DB()),
		memories:         storage.NewMemoryStore(engine.DB()),
		sessions:         storage.NewSessionStore(engine.DB()),
		toolProfiles:     storage.NewToolProfileStore(engine.DB()),
		mcps:             storage.NewMcpStore(engine.DB()),
		specWizards:      storage.NewSpecWizardStore(engine.DB()),
		templates:        map[int]workspace.WorkspaceTemplate{},
		mcpServers:       map[string]mcp.MCPServerUI{},
	}

	ctx := context.Background()
	db.seedIfEmpty(ctx)

	providers, _ := db.providers.ListProviders(ctx)
	if len(providers) == 0 {
		t.Error("Expected providers to be seeded")
	}

	workers, _ := db.workers.ListWorkers(ctx)
	if len(workers) == 0 {
		t.Error("Expected workers to be seeded")
	}

	agents, _ := db.agents.ListAgents(ctx)
	t.Logf("Agents count: %d", len(agents))
	for _, a := range agents {
		t.Logf("  Agent: %s", a.Name)
	}
}

func TestSeedIfEmptyIdempotent(t *testing.T) {
	tmpFile := "/tmp/test_seed_idempotent.db"
	os.Remove(tmpFile)
	defer os.Remove(tmpFile)

	engine, _, _, err := storage.Init(context.Background(), tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	db := &Store{
		engine:           engine,
		providers:        storage.NewProviderStore(engine.DB()),
		models:           storage.NewProviderModelStore(engine.DB()),
		workspaces:       storage.NewWorkspaceStore(engine.DB()),
		agents:           storage.NewAgentStore(engine.DB()),
		workers:          storage.NewWorkerStore(engine.DB()),
		skills:           storage.NewSkillStore(engine.DB()),
		config:           storage.NewConfigStore(engine.DB()),
		greetings:        storage.NewGreetingStore(engine.DB()),
		folders:          storage.NewWorkspaceFolderStore(engine.DB()),
		knowledge:        storage.NewWorkspaceKnowledgeStore(engine.DB()),
		workspaceSkills:  storage.NewWorkspaceSkillsStore(engine.DB()),
		tools:            storage.NewWorkspaceToolsStore(engine.DB()),
		workspaceWorkers: storage.NewWorkspaceWorkersStore(engine.DB()),
		workspaceAgents:  storage.NewWorkspaceAgentsStore(engine.DB()),
		memories:         storage.NewMemoryStore(engine.DB()),
		sessions:         storage.NewSessionStore(engine.DB()),
		toolProfiles:     storage.NewToolProfileStore(engine.DB()),
		mcps:             storage.NewMcpStore(engine.DB()),
		specWizards:      storage.NewSpecWizardStore(engine.DB()),
		templates:        map[int]workspace.WorkspaceTemplate{},
		mcpServers:       map[string]mcp.MCPServerUI{},
	}

	ctx := context.Background()
	db.seedIfEmpty(ctx)

	initialWorkers, _ := db.workers.ListWorkers(ctx)
	initialAgents, _ := db.agents.ListAgents(ctx)

	db.seedIfEmpty(ctx)

	workers, _ := db.workers.ListWorkers(ctx)
	agents, _ := db.agents.ListAgents(ctx)

	if len(workers) != len(initialWorkers) {
		t.Errorf("Workers count changed after second seed: initial=%d, after=%d", len(initialWorkers), len(workers))
	}

	if len(agents) != len(initialAgents) {
		t.Errorf("Agents count changed after second seed: initial=%d, after=%d", len(initialAgents), len(agents))
	}
}
