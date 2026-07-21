package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ada-love-ide/internal/adapters"
	"ada-love-ide/internal/chat"
	"ada-love-ide/internal/commands"
	"ada-love-ide/internal/configfile"
	"ada-love-ide/internal/config/provider"
	core "ada-love-core"
	"ada-love-ide/internal/db"
	"ada-love-ide/internal/icons"
	"ada-love-ide/internal/modelselect"
	"ada-love-ide/internal/plugins"
	"ada-love-ide/internal/sessionfetch"
	"ada-love-ide/internal/sessionstore"
	"ada-love-ide/internal/skillmanager"
	"ada-love-ide/internal/specwizardmgr"

	llm "github.com/upperxcode/ada-llm-client"
	codeIndexer "ada-code-indexer/core"
	codeIndexerStore "ada-code-indexer/storage"
	adaCommands "github.com/upperxcode/ada-commands"
	executor "github.com/upperxcode/ada-executor"
	wiki "github.com/upperxcode/ada-llm-wiki"
	stream "github.com/upperxcode/ada-stream"
)

type Engine struct {
	DB           *db.Store
	Saver        *sessionstore.Saver
	Fetcher      *sessionfetch.Fetcher
	Models       *modelselect.Selector
	Chat         *chat.Chat
	Skills       *skillmanager.Manager
	SkillReg     *skillmanager.RegistryManager
	Orch         *core.Orchestrator
	Plugins      *plugins.PluginManager
	SpecWizardMgr *specwizardmgr.Manager
	Router       *adaCommands.CommandRouter
	Executor     *executor.TaskExecutor
	WorkspaceDir string
	CodeIndexer  *codeIndexerStore.Store

	ctx context.Context
}

func New() (*Engine, error) {
	home, _ := os.UserHomeDir()
	dbDir := filepath.Join(home, ".config", "ada-love-ide")
	os.MkdirAll(dbDir, 0o755)
	dbPath := filepath.Join(dbDir, "ada_love.db")

	store, err := db.New(dbPath)
	if err != nil {
		return nil, err
	}
	cfg := configfile.Load()
	if cfg.IconTheme != "" {
		icons.SetTheme(cfg.IconTheme)
	}
	_ = cfg.FontTheme // applied client-side via data-font-theme
	saver := sessionstore.New(store)
	fetcher := sessionfetch.New(store)
	selector := modelselect.New(store)

	adapter := db.NewStorageAdapter(store)

	var llmClient core.LLMClient
	var streamingClient adapters.LLMStreamingClient
	providers := store.ListProviders()

	validProviders := make(map[string]adapters.ProviderConfig)
	var defaultModel string

	for name, p := range providers {
		if p.APIURL == "" || p.TypeConnection == "" {
			continue
		}

		apiKey := ""
		if len(p.APIKeys) > 0 {
			apiKey = p.APIKeys[0].Key
		}

		validProviders[name] = adapters.ProviderConfig{
			Type:    p.TypeConnection,
			BaseURL: p.APIURL,
			APIKey:  apiKey,
		}

		if defaultModel == "" {
			for m, settings := range p.Models {
				if !settings.Embedding {
					defaultModel = name + "/" + m
					break
				}
			}
			if defaultModel == "" {
				for m := range p.Models {
					defaultModel = name + "/" + m
					break
				}
			}
		}
	}

	frontendEmitter := chat.NewFrontendEmitter() // Create it once

	if len(validProviders) > 0 {
		// Define the adapter function to bridge chat.Emitter to stream.EventEmitter
		streamEmitterAdapter := func(eventName string, optionalData ...interface{}) {
			frontendEmitter.Emit(eventName, optionalData...) // Use the created frontendEmitter
		}

		llmAdapter := adapters.NewMultiLLMAdapterWithEmitter(validProviders, defaultModel, streamEmitterAdapter)
		llmAdapter.SetTools(buildToolDefs(store))
		llmClient = llmAdapter
		streamingClient = llmAdapter

		fmt.Printf("[Engine] Using MultiLLMAdapter with providers: %v, default: %s\n", len(validProviders), defaultModel)
	} else {
		fmt.Println("[Engine] WARNING: No providers configured - chat will not work")
		llmAdapter := adapters.NewMultiLLMAdapterWithEmitter(nil, "", nil)
		llmClient = llmAdapter
		streamingClient = llmAdapter
	}

	compactor := adapters.NewCompactorAdapter(8000, 5, "You are a helpful AI assistant.")

	// Initialize ada-executor with workspace directory from active workspace
	// Get the active workspace and use its first folder as the working directory
	workspaceDir := ""
	activeWorkspacePath := store.ActiveWorkspace()
	if activeWorkspacePath != "" {
		workspaces := store.ListWorkspaces()
		for _, ws := range workspaces {
			if ws.Path == activeWorkspacePath && len(ws.Folders) > 0 {
				workspaceDir = ws.Folders[0]
				fmt.Printf("[Engine] Using workspace directory: %s\n", workspaceDir)
				break
			}
		}
	}

	// If no active workspace or folders, use current directory but executor commands will fail gracefully
	if workspaceDir == "" {
		workspaceDir = "."
		fmt.Printf("[Engine] WARNING: No active workspace found, using current directory\n")
	}

	executorCfg := executor.ExecutorConfig{
		AllowedWorkspaceDir: workspaceDir,
		DefaultTimeout:      30 * time.Second,
	}
	taskExecutor, err := executor.NewTaskExecutor(executorCfg)
	if err != nil {
		fmt.Printf("[Engine] WARNING: Failed to init executor: %v\n", err)
		taskExecutor = nil
	}

	router := adaCommands.NewCommandRouter()
	router.Register(adaCommands.NewHelpCommand(router))
	router.Register(commands.NewClearCommand(store, ""))
	router.Register(&adaCommands.WorkspaceCommand{})

	// Register build/test/run commands if executor is available
	if taskExecutor != nil {
		router.Register(commands.NewBuildCommand(taskExecutor, workspaceDir))
		router.Register(commands.NewTestCommand(taskExecutor, workspaceDir))
		router.Register(commands.NewRunCommand(taskExecutor, workspaceDir))
		router.Register(commands.NewIndexCommand(taskExecutor, workspaceDir))
		router.Register(commands.NewSearchCommand(taskExecutor, workspaceDir))
		router.Register(commands.NewDocCommand(taskExecutor, workspaceDir))

		// Register tool commands (read, write, shell, plan)
		router.Register(commands.NewReadCommand(taskExecutor, workspaceDir))
		router.Register(commands.NewWriteCommand(taskExecutor, workspaceDir))
		router.Register(commands.NewShellCommand(taskExecutor, workspaceDir))
		router.Register(commands.NewPlanCommand(taskExecutor, workspaceDir))
	}

	realExecutor := adapters.NewExecutorAdapter(router)
	orch := core.NewOrchestrator(adapter, llmClient, compactor, realExecutor)

	wikiDir := filepath.Join(home, ".config", "ada-love-ide", "wiki")
	os.MkdirAll(wikiDir, 0o755)
	wikiMgr := wiki.NewWikiManager(wikiDir)
	if err := wikiMgr.LoadArticles(); err != nil {
		fmt.Printf("[Engine] WARNING: Failed to load wiki articles: %v\n", err)
	}
	orch.Wiki = adapters.NewWikiAdapter(wikiMgr)

	// Pass the frontendEmitter to the Chat
	ch := chat.New(orch, frontendEmitter) // Pass the frontendEmitter

	if streamingClient != nil {
		ch.SetStreamingClient(streamingClient)
	}

	home2, _ := os.UserHomeDir()
	skillsDir := filepath.Join(home2, ".opencode", "skills")
	os.MkdirAll(skillsDir, 0o755)
	skm := skillmanager.New(skillsDir)

	// Extra context layers (.AGENTS.md, skills, knowledge, code-indexer)
	orch.ExtraContext = func(ctx context.Context, sessionID, userInput string) string {
		sess, ok := store.GetSession(sessionID)
		if !ok || sess.WorkspaceID == "" {
			return ""
		}
		ws, err := store.GetWorkspace(sess.WorkspaceID)
		if err != nil {
			return ""
		}
		wsDir := ws.Path
		if len(ws.Folders) > 0 && ws.Folders[0] != "" {
			wsDir = ws.Folders[0]
		}
		var layers strings.Builder

		// Workspace directory (so the LLM knows where it can create files)
		layers.WriteString("=== WORKSPACE ===\n")
		layers.WriteString("You have read/write access to the following directory:\n")
		layers.WriteString(wsDir)
		layers.WriteString("\n\n")

		// .AGENTS.md
		agentsPath := filepath.Join(wsDir, ".AGENTS.md")
		if data, err := os.ReadFile(agentsPath); err == nil && len(data) > 0 {
			layers.WriteString("=== ARCHITECTURAL GOLDEN RULES ===\n")
			layers.Write(data)
			layers.WriteString("\n")
		}

		// Workspace skills
		if len(ws.Skills) > 0 {
			layers.WriteString("=== SKILLS ===\n")
			for _, name := range ws.Skills {
				if info, err := skm.GetInfo(name); err == nil && info != nil && info.Markdown != "" {
					layers.WriteString("--- ")
					layers.WriteString(name)
					layers.WriteString(" ---\n")
					layers.WriteString(info.Markdown)
					layers.WriteString("\n")
				}
			}
		}

		// Knowledge (if small enough)
		if len(ws.Knowledge) > 0 {
			kb := strings.Join(ws.Knowledge, "\n")
			if compactor.CountTokens(kb) < 2000 {
				layers.WriteString("=== KNOWLEDGE ===\n")
				layers.WriteString(kb)
				layers.WriteString("\n")
			}
		}

		return layers.String()
	}

	// Initialize code indexer (background crawl)
	codeIdxStore := codeIndexerStore.NewStore()
	if workspaceDir != "" && workspaceDir != "." {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("[Engine] Code indexer panic: %v\n", r)
				}
			}()
			fmt.Printf("[Engine] Starting code indexer for %s\n", workspaceDir)
			if err := codeIndexer.StartCrawler(workspaceDir, codeIdxStore, 10000); err != nil {
				fmt.Printf("[Engine] Code indexer error: %v\n", err)
			}
			fmt.Printf("[Engine] Code indexer ready: %d symbols\n", codeIdxStore.Size())
		}()
	}

	// Add code-indexer search to extra context (update closure)
	origExtra := orch.ExtraContext
	orch.ExtraContext = func(ctx context.Context, sessionID, userInput string) string {
		base := origExtra(ctx, sessionID, userInput)
		if codeIdxStore.Size() > 0 && userInput != "" {
			symbols := codeIdxStore.Search(userInput)
			if len(symbols) > 0 {
				var sb strings.Builder
				sb.WriteString("=== RELEVANT CODE SYMBOLS ===\n")
				maxSym := 10
				if len(symbols) < maxSym {
					maxSym = len(symbols)
				}
				for _, s := range symbols[:maxSym] {
					sb.WriteString(fmt.Sprintf("- %s (%s) in %s:%d\n", s.Name, s.Type, s.FilePath, s.StartLine))
				}
				sb.WriteString("\n")
				return base + sb.String()
			}
		}
		return base
	}

	regMgr := skillmanager.NewRegistryManager()
	regMgr.AddRegistry(skillmanager.NewClawHubRegistry("https://clawhub.ai", ""))
	regMgr.AddRegistry(skillmanager.NewGitHubRegistry("https://github.com", ""))

	home3, _ := os.UserHomeDir()
	pluginsDir := filepath.Join(home3, ".config", "ada-love-ide", "plugins", "spec-wizard")
	pluginMgr, err := plugins.NewManager(pluginsDir)
	if err != nil {
		fmt.Printf("[Engine] WARNING: Failed to load plugins: %v\n", err)
		pluginMgr = &plugins.PluginManager{}
	}

	specWizardMgr := specwizardmgr.New(store, pluginMgr)
	specWizardMgr.SetLLMFn(func(ctx context.Context, systemPrompt, userPrompt string, temperature float64, maxTokens int) (string, error) {
		fmt.Printf("[engine.llmFn] Looking up fixed model 'spec'\n")
		specProvider, specModel, _ := store.GetFixedModel("spec")
		fmt.Printf("[engine.llmFn] spec -> provider=%q model=%q\n", specProvider, specModel)
		if specProvider == "" || specModel == "" {
			return "", errors.New("[ALERTA] Modelo 'spec' não configurado. Configure um provider e modelo para 'spec' em Settings > Models.")
		}

		providers := store.ListProviders()
		pCfg, ok := providers[specProvider]
		fmt.Printf("[engine.llmFn] providers found=%d looking for %q -> ok=%v\n", len(providers), specProvider, ok)
		if !ok {
			return "", fmt.Errorf("[ALERTA] Provider '%s' configurado para o modelo 'spec' não foi encontrado", specProvider)
		}
		fmt.Printf("[engine.llmFn] provider cfg: type=%q baseURL=%q hasAPIKey=%v\n",
			pCfg.TypeConnection, pCfg.APIURL, len(pCfg.APIKeys) > 0)

		apiKey := ""
		if len(pCfg.APIKeys) > 0 {
			apiKey = pCfg.APIKeys[0].Key
		}

		client := llm.NewClient(llm.ConnectionConfig{
			Type:    llm.ConnectionType(pCfg.TypeConnection),
			BaseURL: pCfg.APIURL,
			APIKey:  apiKey,
		})

		// Create a context with timeout so the call doesn't hang forever
		llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		fmt.Printf("[engine.llmFn] Calling Generate model=%q temp=%.1f max=%d\n", specModel, temperature, maxTokens)
		resp, err := client.Generate(llmCtx, llm.InferenceRequest{
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
			Config: llm.InferenceConfig{
				Model:       specModel,
				Temperature: temperature,
				MaxTokens:   maxTokens,
			},
		})
		if err != nil {
			fmt.Printf("[engine.llmFn] Generate FAILED: %v\n", err)
			return "", fmt.Errorf("[ALERTA] Modelo 'spec' não respondeu: %w", err)
		}
		fmt.Printf("[engine.llmFn] Generate OK resp len=%d\n", len(resp))
		return resp, nil
	})

	// Register health command after skm and pluginMgr are available
	router.Register(&commands.SyncSpecCommand{
		SyncFn: func(workspacePath string) error {
			ws, err := store.GetWorkspace(workspacePath)
			if err != nil {
				return err
			}
			if ws.SpecWizardID == "" {
				return fmt.Errorf("workspace %s não possui Spec Wizard", ws.Title)
			}
			specwizardmgr.SyncSpecToWorkspace(store, ws)
			return nil
		},
	})
	router.Register(commands.NewHealthCommand(
		commands.HealthCheck{
			Component: "configs",
			Check: func() (status, details string, count int) {
				providers := store.ListProviders()
				if len(providers) == 0 {
					return "critical", "No LLM providers configured - chat will not work", 0
				}
				return "ok", fmt.Sprintf("%d LLM providers configured", len(providers)), len(providers)
			},
		},
		commands.HealthCheck{
			Component: "workspace",
			Check: func() (status, details string, count int) {
				ws := store.ActiveWorkspace()
				if ws == "" {
					return "critical", "No active workspace selected", 0
				}
				return "ok", fmt.Sprintf("Active workspace: %s", ws), 1
			},
		},
		commands.HealthCheck{
			Component: "workers",
			Check: func() (status, details string, count int) {
				workers := store.ListWorkers()
				if len(workers) == 0 {
					return "critical", "No workers configured - system may not route requests properly", 0
				}
				return "ok", fmt.Sprintf("%d workers configured", len(workers)), len(workers)
			},
		},
		commands.HealthCheck{
			Component: "agents",
			Check: func() (status, details string, count int) {
				agents := store.ListAgents()
				if len(agents) == 0 {
					return "critical", "No agents configured - multi-agent features disabled", 0
				}
				return "ok", fmt.Sprintf("%d agents configured", len(agents)), len(agents)
			},
		},
		commands.HealthCheck{
			Component: "skills",
			Check: func() (status, details string, count int) {
				skills := skm.ListInstalled()
				if len(skills) == 0 {
					return "warning", "No skills installed - advanced features may be limited", 0
				}
				return "ok", fmt.Sprintf("%d skills installed", len(skills)), len(skills)
			},
		},
		commands.HealthCheck{
			Component: "tools",
			Check: func() (status, details string, count int) {
				tools := store.ListProfiles()
				if len(tools) == 0 {
					return "warning", "No tool profiles configured", 0
				}
				return "ok", fmt.Sprintf("%d tool profiles configured", len(tools)), len(tools)
			},
		},
		commands.HealthCheck{
			Component: "spec-wizard",
			Check: func() (status, details string, count int) {
				wizards := store.ListWizards()
				if len(wizards) == 0 {
					return "warning", "No spec-wizards installed - spec generation features may be limited", 0
				}
				return "ok", fmt.Sprintf("%d spec-wizards available", len(wizards)), len(wizards)
			},
		},
		commands.HealthCheck{
			Component: "mcp-servers",
			Check: func() (status, details string, count int) {
				mcp := store.ListMCPServers()
				if len(mcp) == 0 {
					return "warning", "No MCP servers configured - external tool integration limited", 0
				}
				return "ok", fmt.Sprintf("%d MCP servers available", len(mcp)), len(mcp)
			},
		},
	))

	return &Engine{
		DB:           store,
		Saver:        saver,
		Fetcher:      fetcher,
		Models:       selector,
		Chat:         ch,
		Skills:       skm,
		SkillReg:     regMgr,
		Orch:         orch,
		Plugins:      pluginMgr,
		SpecWizardMgr: specWizardMgr,
		Router:       router,
		Executor:     taskExecutor,
		WorkspaceDir: workspaceDir,
		CodeIndexer:  codeIdxStore,
	}, nil
}

func buildToolDefs(store *db.Store) []llm.ToolDefinition {
	tools := store.AvailableTools()
	defs := make([]llm.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		if !t.Enabled {
			continue
		}
		defs = append(defs, llm.ToolDefinition{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		})
	}
	return defs
}

func (e *Engine) SetContext(ctx context.Context) { e.ctx = ctx }

func (e *Engine) Context() context.Context { return e.ctx }

func (e *Engine) Close() error {
	// Expert plugins are invoked on-demand via STDIO (no long-running server),
	// so there is nothing to stop here.
	return e.DB.Close()
}

func (e *Engine) StreamManager() *stream.StreamManager {
	return stream.NewStreamManager()
}

func (e *Engine) ResolveWorkspaceDir(workspacePath string) string {
	ws, err := e.DB.GetWorkspace(workspacePath)
	if err != nil {
		fmt.Printf("[engine] ResolveWorkspaceDir: workspace %q not found, using fallback\n", workspacePath)
		return e.WorkspaceDir
	}
	if len(ws.Folders) > 0 && ws.Folders[0] != "" {
		return ws.Folders[0]
	}
	if ws.Path != "" {
		return ws.Path
	}
	return e.WorkspaceDir
}

func (e *Engine) ResolveSessionDir(sessionID string) string {
	sess, ok := e.DB.GetSession(sessionID)
	if !ok || sess.WorkspaceID == "" {
		fmt.Printf("[engine] ResolveSessionDir: session %q not found, using fallback\n", sessionID)
		return e.WorkspaceDir
	}
	return e.ResolveWorkspaceDir(sess.WorkspaceID)
}

func (e *Engine) GetWorkspaceBySession(sessionID string) (string, error) {
	sess, ok := e.DB.GetSession(sessionID)
	if !ok {
		return "", fmt.Errorf("session %q not found", sessionID)
	}
	return sess.WorkspaceID, nil
}

// ContextInfo retorna o uso de contexto de uma sessão.
type ContextInfo struct {
	ContextLimit  int     `json:"context_limit"`
	ContextUsed   int     `json:"context_used"`
	SystemTokens  int     `json:"system_tokens"`
	MessagesTokens int    `json:"messages_tokens"`
}

func (e *Engine) GetSessionContextInfo(sessionID string) ContextInfo {
	sess, ok := e.DB.GetSession(sessionID)
	if !ok {
		return ContextInfo{}
	}

	// 1. Context limit: model settings > workspace config
	ctxLimit := 0
	if sess.Provider != "" && sess.Model != "" {
		modelString := sess.Provider + "/" + sess.Model
		if ms, ok := e.DB.GetModelSettings(modelString); ok && ms.ContextSize > 0 {
			ctxLimit = ms.ContextSize
		}
	}
	if ctxLimit <= 0 {
		ws, err := e.DB.GetWorkspace(sess.WorkspaceID)
		if err == nil && ws.MaxContextLength > 0 {
			ctxLimit = ws.MaxContextLength
		}
	}

	// 2. Token usage — count real tokens via compactor
	msgs := e.DB.GetMessages(sessionID)
	sysTokens := 120
	msgsTokens := 0
	for _, m := range msgs {
		msgsTokens += e.Orch.Compactor.CountTokens(m.Content)
	}
	used := sysTokens + msgsTokens

	return ContextInfo{
		ContextLimit:   ctxLimit,
		ContextUsed:    used,
		SystemTokens:   sysTokens,
		MessagesTokens: msgsTokens,
	}
}

func (e *Engine) GetModelSettings(modelString string) (provider.ModelSettings, bool) {
	return e.DB.GetModelSettings(modelString)
}
