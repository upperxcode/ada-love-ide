package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ada-love-ide/internal/config/agent"
	"ada-love-ide/internal/config/mcp"
	"ada-love-ide/internal/config/provider"
	"ada-love-ide/internal/config/skill"
	"ada-love-ide/internal/config/specwizard"
	"ada-love-ide/internal/config/tool"
	"ada-love-ide/internal/config/workspace"
	iprovider "ada-love-ide/internal/provider"

	storage "github.com/ada-love-ai/storage/storage"
)

// ── Providers ──────────────────────────────────────────────────

func (s *Store) ListProviders() map[string]provider.ProviderConfig {
	ctx := context.Background()
	raw, _ := s.providers.ListProviders(ctx)
	out := make(map[string]provider.ProviderConfig)
	for _, p := range raw {
		out[p.Name] = adaptProviderToInternal(ctx, s, &p)
	}
	return out
}

func (s *Store) SaveProvider(name string, cfg provider.ProviderConfig) {
	ctx := context.Background()

	// upsert provider
	existing, err := s.providers.GetProviderByName(ctx, name)
	var providerID int64
	if err == nil {
		providerID = existing.ID
			existing.APIURL = sql.NullString{String: cfg.APIURL, Valid: cfg.APIURL != ""}
			existing.ConnectionTypes = sql.NullString{String: cfg.TypeConnection, Valid: cfg.TypeConnection != ""}
			existing.Strategy = sql.NullString{String: cfg.Strategy, Valid: cfg.Strategy != ""}
			existing.Color = cfg.Color
			existing.Icon = cfg.Icon
			_ = s.providers.UpdateProvider(ctx, existing)
		} else {
			sp := &storage.Provider{
				Name:            name,
				APIURL:          sql.NullString{String: cfg.APIURL, Valid: cfg.APIURL != ""},
				ConnectionTypes: sql.NullString{String: cfg.TypeConnection, Valid: cfg.TypeConnection != ""},
				Strategy:        sql.NullString{String: cfg.Strategy, Valid: cfg.Strategy != ""},
				Color:           cfg.Color,
				Icon:            cfg.Icon,
			}
		_ = s.providers.CreateProvider(ctx, sp)
		p, _ := s.providers.GetProviderByName(ctx, name)
		providerID = p.ID
	}

	// replace models
	models, _ := s.models.GetProviderModels(ctx, providerID)
	for _, m := range models {
		_ = s.models.DeleteProviderModel(ctx, m.ID)
	}
	for modelName, ms := range cfg.Models {
		_ = s.models.CreateProviderModel(ctx, &storage.ProviderModel{
			ProviderID:  providerID,
			Model:       modelName,
			Free:        ms.Free,
			Thinking:    ms.Thinking,
			Tool:        ms.Tools,
			Embedding:   ms.Embedding,
			Vision:      ms.Vision,
			Health:      100,
			ContextSize: ms.ContextSize,
			MaxTokens:   ms.MaxTokens,
		})
	}

	// save API keys
	keys, _ := s.config.GetAPIKeys(ctx, providerID)
	for _, k := range keys {
		_ = s.config.DeleteAPIKey(ctx, providerID, k)
	}
	for _, ak := range cfg.APIKeys {
		_ = s.config.SetAPIKey(ctx, providerID, ak.Key)
	}
}

func (s *Store) DeleteProvider(name string) {
	ctx := context.Background()
	p, err := s.providers.GetProviderByName(ctx, name)
	if err != nil {
		return
	}
	_ = s.providers.DeleteProvider(ctx, p.ID)
}

	func adaptProviderToInternal(ctx context.Context, s *Store, p *storage.Provider) provider.ProviderConfig {
		cfg := provider.ProviderConfig{
			Icon:           p.Icon,
			Color:          p.Color,
			APIURL:         p.APIURL.String,
			TypeConnection: p.ConnectionTypes.String,
			Strategy:       p.Strategy.String,
			APIKeys:        []provider.ProviderAPIKey{},
			Models:         map[string]provider.ModelSettings{},
		}

	// load models
	models, _ := s.models.GetProviderModels(ctx, p.ID)
	for _, m := range models {
		cfg.Models[m.Model] = provider.ModelSettings{
			Free:        m.Free,
			Thinking:    m.Thinking,
			Tools:       m.Tool,
			Embedding:   m.Embedding,
			Vision:      m.Vision,
			ContextSize: m.ContextSize,
			MaxTokens:   m.MaxTokens,
		}
	}

	// load API keys
	keys, _ := s.config.GetAPIKeys(ctx, p.ID)
	for _, k := range keys {
		cfg.APIKeys = append(cfg.APIKeys, provider.ProviderAPIKey{Key: k})
	}

	return cfg
}

// ── Tool Profiles ──────────────────────────────────────────────

func (s *Store) ListProfiles() []tool.ToolProfile {
	ctx := context.Background()
	storageProfiles, err := s.toolProfiles.ListProfiles(ctx)
	if err != nil {
		return nil
	}
	out := make([]tool.ToolProfile, 0, len(storageProfiles))
	for _, sp := range storageProfiles {
		// Load tools for this profile
		tools, _ := s.toolProfiles.ListTools(ctx, sp.ID)
		toolNames := make([]string, 0, len(tools))
		for _, t := range tools {
			toolNames = append(toolNames, t.ToolName)
		}
		out = append(out, tool.ToolProfile{
			ID:    int(sp.ID),
			Name:  sp.Name,
			Color: sp.Color,
			Icon:  sp.Icon,
			Tools: toolNames,
		})
	}
	return out
}

func (s *Store) PutProfile(p tool.ToolProfile) {
	ctx := context.Background()
	desc := sql.NullString{String: "", Valid: false}
	sp := &storage.ToolProfile{
		ID:          int64(p.ID),
		Name:        p.Name,
		Description: desc,
		Color:       p.Color,
		Icon:        p.Icon,
	}

	if p.ID > 0 {
		_ = s.toolProfiles.UpdateProfile(ctx, sp)
	} else {
		_ = s.toolProfiles.CreateProfile(ctx, sp)
		p.ID = int(sp.ID)
	}

	// Replace associated tools
	existingTools, _ := s.toolProfiles.ListTools(ctx, sp.ID)
	for _, t := range existingTools {
		_ = s.toolProfiles.RemoveTool(ctx, sp.ID, t.ToolName)
	}
	for _, toolName := range p.Tools {
		_ = s.toolProfiles.AddTool(ctx, &storage.ToolProfileTool{
			ProfileID: sp.ID,
			ToolName:  toolName,
		})
	}
}

func (s *Store) DeleteProfile(id int) {
	ctx := context.Background()
	_ = s.toolProfiles.DeleteProfile(ctx, int64(id))
}

func (s *Store) NextProfileID() int {
	return 0
}

// ── Spec Wizards ───────────────────────────────────────────────

func (s *Store) ListWizards() []specwizard.SpecWizardConfig {
	ctx := context.Background()
	raw, err := s.specWizards.List(ctx)
	if err != nil {
		return []specwizard.SpecWizardConfig{}
	}
	out := make([]specwizard.SpecWizardConfig, 0, len(raw))
	for _, w := range raw {
		out = append(out, convertStorageToSpecWizard(&w))
	}
	return out
}

func (s *Store) PutWizard(w specwizard.SpecWizardConfig) {
	ctx := context.Background()
	// Always refresh UpdatedAt so a caller that forgets to set it never
	// persists a zero timestamp (0001-01-01) over a real one.
	if w.UpdatedAt.IsZero() {
		w.UpdatedAt = time.Now()
	}
	sw := convertSpecWizardToStorage(&w)
	existing, _ := s.specWizards.Get(ctx, w.ID)
	if existing != nil {
		if err := s.specWizards.Update(ctx, sw); err != nil {
			fmt.Printf("[db] PutWizard Update(%s) failed: %v\n", w.ID, err)
		}
	} else {
		if w.CreatedAt.IsZero() {
			w.CreatedAt = w.UpdatedAt
			sw.CreatedAt = w.CreatedAt
		}
		if err := s.specWizards.Create(ctx, sw); err != nil {
			fmt.Printf("[db] PutWizard Create(%s) failed: %v\n", w.ID, err)
		}
	}
}

func (s *Store) GetWizard(id string) (specwizard.SpecWizardConfig, bool) {
	ctx := context.Background()
	w, err := s.specWizards.Get(ctx, id)
	if err != nil || w == nil {
		return specwizard.SpecWizardConfig{}, false
	}
	return convertStorageToSpecWizard(w), true
}

func (s *Store) DeleteWizard(id string) {
	ctx := context.Background()
	_ = s.specWizards.Delete(ctx, id)
}

	func convertStorageToSpecWizard(sw *storage.SpecWizard) specwizard.SpecWizardConfig {
		var manifest []specwizard.Dependency
		if sw.DependencyManifest != "" {
			_ = json.Unmarshal([]byte(sw.DependencyManifest), &manifest)
		}

		return specwizard.SpecWizardConfig{
			ID:                        sw.ID,
			Name:                      sw.Name,
			Description:               sw.Description,
			ExpertLanguagePlugin:      sw.ExpertLanguagePlugin,
			PRD:                       sw.PRD,
			FunctionalRequirements:    storage.UnmarshalStringSlice(sw.FunctionalRequirements),
			NonFunctionalRequirements: storage.UnmarshalStringSlice(sw.NonFunctionalRequirements),
			Persistence:               sw.Persistence,
			Architecture:              sw.Architecture,
			EngineeringPhilosophies:   storage.UnmarshalStringSlice(sw.EngineeringPhilosophies),
			DesignPatterns:            storage.UnmarshalStringSlice(sw.DesignPatterns),
			DataPatterns:              storage.UnmarshalStringSlice(sw.DataPatterns),
			StackConfig:               convertStackItems(sw.StackConfig),
			DependencyManifest:        manifest,
			StackPlugin:               sw.StackPlugin,
			Business: specwizard.Business{
			StateManagement:             sw.BusinessStateManagement,
			APIContract:                 sw.BusinessAPIContract,
			CustomizationDetails:        sw.BusinessCustomizationDetails,
			FinalAdjustments:            sw.BusinessFinalAdjustments,
			ArchitectureRecommendations: sw.BusinessArchitectureRecommendations,
		},
		Color:              sw.Color,
		Icon:               sw.Icon,
		ArchitectureHealth: sw.ArchitectureHealth,
		CreatedAt:          sw.CreatedAt,
		UpdatedAt:          sw.UpdatedAt,
	}
}

	func convertSpecWizardToStorage(w *specwizard.SpecWizardConfig) *storage.SpecWizard {
		manifestJSON, _ := json.Marshal(w.DependencyManifest)

		return &storage.SpecWizard{
			ID:                                  w.ID,
			Name:                                w.Name,
			Description:                         w.Description,
			ExpertLanguagePlugin:                w.ExpertLanguagePlugin,
			PRD:                                 w.PRD,
			FunctionalRequirements:              storage.MarshalStringSlice(w.FunctionalRequirements),
			NonFunctionalRequirements:           storage.MarshalStringSlice(w.NonFunctionalRequirements),
			Persistence:                         w.Persistence,
			Architecture:                        w.Architecture,
			EngineeringPhilosophies:             storage.MarshalStringSlice(w.EngineeringPhilosophies),
			DesignPatterns:                      storage.MarshalStringSlice(w.DesignPatterns),
			DataPatterns:                        storage.MarshalStringSlice(w.DataPatterns),
			StackConfig:                         storage.MarshalStackConfig(w.StackConfig),
			BusinessStateManagement:             w.Business.StateManagement,
			BusinessAPIContract:                 w.Business.APIContract,
			BusinessCustomizationDetails:        w.Business.CustomizationDetails,
			BusinessFinalAdjustments:            w.Business.FinalAdjustments,
			BusinessArchitectureRecommendations: w.Business.ArchitectureRecommendations,
			Color:                               w.Color,
			Icon:                                w.Icon,
			ArchitectureHealth:                  w.ArchitectureHealth,
			DependencyManifest:                  string(manifestJSON),
			StackPlugin:                         w.StackPlugin,
			CreatedAt:                           w.CreatedAt,
			UpdatedAt:                           w.UpdatedAt,
		}
	}

func convertStackItems(s string) []specwizard.StackItem {
	raw := storage.UnmarshalStackConfig(s)
	out := make([]specwizard.StackItem, 0, len(raw))
	for _, r := range raw {
		item := specwizard.StackItem{}
		if name, ok := r["name"].(string); ok {
			item.Name = name
		}
		if example, ok := r["example"].(string); ok {
			item.Example = example
		}
		out = append(out, item)
	}
	return out
}

// ── Workspace Templates ────────────────────────────────────────

func (s *Store) ListTemplates() []workspace.WorkspaceTemplate {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]workspace.WorkspaceTemplate, 0, len(s.templates))
	for _, t := range s.templates {
		out = append(out, t)
	}
	return out
}

func (s *Store) PutTemplate(t workspace.WorkspaceTemplate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[t.ID] = t
}

func (s *Store) DeleteTemplate(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.templates, id)
}

func (s *Store) NextTemplateID() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextTemplateID
	s.nextTemplateID++
	return id
}

// ── MCP Servers ────────────────────────────────────────────────

func (s *Store) ListMCPServers() map[string]mcp.MCPServerUI {
	ctx := context.Background()
	rows, err := s.mcps.ListMcps(ctx)
	if err != nil {
		fmt.Printf("[Backend] ListMCPServers: error listing from SQL: %v\n", err)
		return map[string]mcp.MCPServerUI{}
	}
	fmt.Printf("[Backend] ListMCPServers: found %d rows in SQL\n", len(rows))
	out := map[string]mcp.MCPServerUI{}
	for _, r := range rows {
		args := []string{}
		if r.Arguments.Valid {
			if err := json.Unmarshal([]byte(r.Arguments.String), &args); err != nil {
				fmt.Printf("[Backend] ListMCPServers: error unmarshaling args for %s: %v\n", r.Name, err)
			}
		}
		headers := []mcp.HeaderEntry{}
		timeout := r.Timeout
		oauthClientID := r.OAuthClientID

			env := map[string]string{}
			if r.Environment.Valid && r.Environment.String != "" && r.Environment.String != "{}" {
				// Tenta desserializar o novo payload JSON (com timeout, headers e env)
				type envPayload struct {
					Timeout       int               `json:"timeout"`
					OAuthClientID string            `json:"oauth_client_id"`
					Headers       []mcp.HeaderEntry `json:"headers"`
					Env           map[string]string `json:"env"`
				}
				var payload envPayload
				if err := json.Unmarshal([]byte(r.Environment.String), &payload); err == nil && (len(payload.Headers) > 0 || payload.Timeout > 0 || len(payload.Env) > 0) {
					headers = payload.Headers
					env = payload.Env
					if payload.Timeout > 0 {
						timeout = payload.Timeout
					}
					if payload.OAuthClientID != "" {
						oauthClientID = payload.OAuthClientID
					}
				} else {
					// Fallback para o formato antigo (pode ser map string string)
					var legacyEnv map[string]string
					if err := json.Unmarshal([]byte(r.Environment.String), &legacyEnv); err == nil {
						env = legacyEnv
					}
				}
			}

		url := ""
		if r.URL.Valid {
			url = r.URL.String
		}
		command := ""
		if r.Command.Valid {
			command = r.Command.String
		}

			out[r.Name] = mcp.MCPServerUI{
				Command:       command,
				Args:          args,
				Env:           env,
				URL:           url,
				Enabled:       r.Enabled,
				Icon:          r.Icon,
				Color:         r.Color,
				Timeout:       timeout,
				OAuthClientID: oauthClientID,
				Headers:       headers,
			}
	}
	return out
}

func (s *Store) SaveMCPServer(name string, srv mcp.MCPServerUI) {
	ctx := context.Background()
	connectType := "cli_command"
	if srv.URL != "" {
		connectType = "url"
	}
	argsJSON, _ := json.Marshal(srv.Args)
	// Serializar timeout + headers + env juntos no environment para compatibilidade
	type envPayload struct {
		Timeout       int               `json:"timeout"`
		OAuthClientID string            `json:"oauth_client_id"`
		Headers       []mcp.HeaderEntry `json:"headers"`
		Env           map[string]string `json:"env"`
	}
	payload := envPayload{
		Timeout:       srv.Timeout,
		OAuthClientID: srv.OAuthClientID,
		Headers:       srv.Headers,
		Env:           srv.Env,
	}
	envJSON, _ := json.Marshal(payload)

	storageMcp := &storage.Mcp{
		Name:          name,
		ConnectType:   connectType,
		Command:       sql.NullString{String: srv.Command, Valid: srv.Command != ""},
		Arguments:     sql.NullString{String: string(argsJSON), Valid: true},
		Environment:   sql.NullString{String: string(envJSON), Valid: true},
		URL:           sql.NullString{String: srv.URL, Valid: srv.URL != ""},
		Enabled:       srv.Enabled,
		Timeout:       srv.Timeout,
		OAuthClientID: srv.OAuthClientID,
		Color:         srv.Color,
		Icon:          srv.Icon,
	}

	// Upsert: check if exists first
	existing, err := s.mcps.GetMcpByName(ctx, name)
	if err == nil {
		storageMcp.ID = existing.ID
		_ = s.mcps.UpdateMcp(ctx, storageMcp)
	} else {
		_ = s.mcps.CreateMcp(ctx, storageMcp)
	}

	// Keep in-memory map in sync
	s.mu.Lock()
	s.mcpServers[name] = srv
	s.mu.Unlock()
}

func (s *Store) DeleteMCPServer(name string) {
	ctx := context.Background()
	m, err := s.mcps.GetMcpByName(ctx, name)
	if err == nil {
		_ = s.mcps.DeleteMcp(ctx, m.ID)
	}
	s.mu.Lock()
	delete(s.mcpServers, name)
	s.mu.Unlock()
}

func (s *Store) ReplaceMCPServers(servers map[string]mcp.MCPServerUI) {
	ctx := context.Background()
	existing, _ := s.mcps.ListMcps(ctx)
	for _, m := range existing {
		_ = s.mcps.DeleteMcp(ctx, m.ID)
	}
	for name, srv := range servers {
		s.SaveMCPServer(name, srv)
	}
}

// ── Agents ─────────────────────────────────────────────────────

func (s *Store) ListAgents() []agent.AgentConfig {
	ctx := context.Background()
	raw, err := s.agents.ListAgents(ctx)
	if err != nil {
		return []agent.AgentConfig{}
	}

	// Resolve provider_id/model_id to names using s.ListProviders() logic
	providerIDToName := make(map[int64]string)
	modelIDToName := make(map[int64]string)

	rawProviders, _ := s.providers.ListProviders(ctx)
	for _, p := range rawProviders {
		providerIDToName[p.ID] = p.Name
		rawModels, _ := s.models.GetProviderModels(ctx, p.ID)
		for _, m := range rawModels {
			modelIDToName[m.ID] = m.Model
		}
	}

	out := make([]agent.AgentConfig, 0, len(raw))
	for _, a := range raw {
		out = append(out, agent.AgentConfig{
			ID:            a.ID,
			Name:          a.Name,
			Description:   a.Description.String,
			Provider:      providerIDToName[a.ProviderID.Int64],
			Model:         modelIDToName[a.ModelID.Int64],
			Type:          string(a.Type),
			Icon:          a.Icon,
			Color:         a.Color,
			MaxIterations: a.MaxIteration,
			Temperature:   a.Temperature,
			SystemPrompt:  a.SystemPrompt.String,
		})
	}
	return out
}

func (s *Store) PutAgent(a agent.AgentConfig) error {
	ctx := context.Background()

	// Resolve provider name to ID
	p, err := s.providers.GetProviderByName(ctx, a.Provider)
	if err != nil {
		return fmt.Errorf("provider %s not found", a.Provider)
	}

	// Resolve model name to ID
	var modelID int64
	models, _ := s.models.GetProviderModels(ctx, p.ID)
	for _, m := range models {
		if m.Model == a.Model {
			modelID = m.ID
			break
		}
	}

	sa := &storage.Agent{
		ID:           a.ID,
		Name:         a.Name,
		Description:  sql.NullString{String: a.Description, Valid: a.Description != ""},
		ProviderID:   sql.NullInt64{Int64: p.ID, Valid: true},
		ModelID:      sql.NullInt64{Int64: modelID, Valid: modelID > 0},
		Type:         storage.AgentType(a.Type),
		Icon:         a.Icon,
		Color:        a.Color,
		MaxIteration: a.MaxIterations,
		Temperature:  a.Temperature,
		SystemPrompt: sql.NullString{String: a.SystemPrompt, Valid: a.SystemPrompt != ""},
	}

	if a.ID > 0 {
		return s.agents.UpdateAgent(ctx, sa)
	} else {
		return s.agents.CreateAgent(ctx, sa)
	}
}

func (s *Store) DeleteAgent(id int64) error {
	ctx := context.Background()
	return s.agents.DeleteAgent(ctx, id)
}

func (s *Store) PatchAgentField(id int64, field string, value string) error {
	ctx := context.Background()
	a, err := s.agents.GetAgent(ctx, id)
	if err != nil {
		return fmt.Errorf("agent %d not found", id)
	}
	switch field {
	case "color":
		a.Color = value
	case "icon":
		a.Icon = value
	default:
		return fmt.Errorf("unsupported field: %s", field)
	}
	return s.agents.UpdateAgent(ctx, a)
}

func (s *Store) SetAgents(list []agent.AgentConfig) error {
	ctx := context.Background()
	existing, _ := s.agents.ListAgents(ctx)
	incomingIDs := make(map[int64]bool)

	for _, a := range list {
		if a.ID > 0 {
			incomingIDs[a.ID] = true
		}
		if err := s.PutAgent(a); err != nil {
			return err
		}
	}

	for _, e := range existing {
		if !incomingIDs[e.ID] {
			if err := s.DeleteAgent(e.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

// ── Tools ──────────────────────────────────────────────────────

type ToolInfo struct {
	Name        string
	Description string
	Category    string
	Enabled     bool
}

func (s *Store) AvailableTools() []ToolInfo {
	return []ToolInfo{
		{"read", "Lê conteúdo", "io", true},
		{"write", "Escreve conteúdo", "io", true},
		{"search", "Busca arquivos", "fs", true},
		{"exec", "Executa comando", "shell", false},
		{"plan", "Planeja mudanças", "agent", true},
	}
}

// ── Skills ───────────────────────────────────────────────────────

// skillState stores the customization fields for skills (Color, Icon, Active)
type skillState struct {
	ID     int64
	Color  string
	Icon   string
	Active bool
}

// skillStates is an in-memory map for skill customization (Color, Icon, Active)
// This is a temporary solution until proper DB migration is added
var skillStates = map[int64]skillState{}

// ListSkills returns all skills from the database with their state
func (s *Store) ListSkills() []skill.SkillConfig {
	ctx := context.Background()
	raw, err := s.skills.ListSkills(ctx)
	if err != nil {
		return []skill.SkillConfig{}
	}

	out := make([]skill.SkillConfig, 0, len(raw))
	for _, sk := range raw {
		state, hasState := skillStates[sk.ID]
		result := skill.SkillConfig{
			ID:          sk.ID,
			Name:        sk.Name,
			Description: sk.Description.String,
			Tags:        sk.Tags.String,
			Content:     sk.Content,
			Color:       sk.Color,
			Icon:        sk.Icon,
			Active:      state.Active,
		}
		if !hasState {
			// Set defaults for new skills
			if result.Color == "" {
				result.Color = "#4c5578"
			}
			if result.Icon == "" {
				result.Icon = "🤖"
			}
			result.Active = true
			skillStates[sk.ID] = skillState{Color: result.Color, Icon: result.Icon, Active: result.Active}
		} else {
			result.Color = state.Color
			result.Icon = state.Icon
		}
		out = append(out, result)
	}
	return out
}

// PutSkill creates or updates a skill
func (s *Store) PutSkill(sk skill.SkillConfig) error {
	ctx := context.Background()

	storageSkill := &storage.Skill{
		ID:          sk.ID,
		Name:        sk.Name,
		Description: sql.NullString{String: sk.Description, Valid: sk.Description != ""},
		Tags:        sql.NullString{String: sk.Tags, Valid: sk.Tags != ""},
		Content:     sk.Content,
		Color:       sk.Color,
		Icon:        sk.Icon,
	}

	var err error
	if sk.ID > 0 {
		err = s.skills.UpdateSkill(ctx, storageSkill)
	} else {
		err = s.skills.CreateSkill(ctx, storageSkill)
	}
	if err != nil {
		return err
	}

	// Save state in memory
	id := storageSkill.ID // after create/update, this holds the correct ID
	skillStates[id] = skillState{
		Color:  sk.Color,
		Icon:   sk.Icon,
		Active: sk.Active,
	}

	return nil
}

// DeleteSkill removes a skill by ID
func (s *Store) DeleteSkill(id int64) error {
	ctx := context.Background()
	err := s.skills.DeleteSkill(ctx, id)
	if err == nil {
		delete(skillStates, id)
	}
	return err
}

// GetSkill returns a skill by ID
func (s *Store) GetSkill(id int64) (skill.SkillConfig, error) {
	ctx := context.Background()
	raw, err := s.skills.GetSkill(ctx, id)
	if err != nil {
		return skill.SkillConfig{}, err
	}

	state, hasState := skillStates[raw.ID]
	result := skill.SkillConfig{
		ID:          raw.ID,
		Name:        raw.Name,
		Description: raw.Description.String,
		Tags:        raw.Tags.String,
		Content:     raw.Content,
		Color:       raw.Color,
		Icon:        raw.Icon,
	}

	if hasState {
		result.Color = state.Color
		result.Icon = state.Icon
		result.Active = state.Active
	} else {
		if result.Color == "" {
			result.Color = "#4c5578"
		}
		if result.Icon == "" {
			result.Icon = "🤖"
		}
		result.Active = true
		skillStates[raw.ID] = skillState{Color: result.Color, Icon: result.Icon, Active: result.Active}
	}

	return result, nil
}

// ListSkillStates returns the enabled/disabled state of skills
func (s *Store) ListSkillStates() map[string]bool {
	skills := s.ListSkills()
	result := make(map[string]bool)
	for _, sk := range skills {
		result[sk.Name] = sk.Active
	}
	return result
}

// ── Icon Theme ──────────────────────────────────────────────────

func (s *Store) GetIconTheme() string {
	ctx := context.Background()
	v, err := s.config.GetConfig(ctx, "icon_theme")
	if err != nil || v == "" {
		return "lucide"
	}
	return v
}

func (s *Store) SetIconTheme(theme string) error {
	ctx := context.Background()
	return s.config.SetConfig(ctx, "icon_theme", theme)
}

func (s *Store) GetFixedModel(name string) (provider, model string, tools []string) {
	ctx := context.Background()
	m, err := s.fixedModels.GetFixedModel(ctx, name)
	if err != nil {
		return "", "", nil
	}
	t, _ := s.fixedModels.ListTools(ctx, m.ID)
	tools = make([]string, len(t))
	for i, tool := range t {
		tools[i] = tool.Tool
	}
	return m.Provider, m.Model, tools
}

func (s *Store) SaveFixedModel(name, provider, model string, tools []string) {
	ctx := context.Background()
	m, err := s.fixedModels.GetFixedModel(ctx, name)
	if err == nil {
		// Storage has no UpdateFixedModel; delete-then-recreate is the upsert.
		// The fixed_model_tools FK cascades on delete, so old tools are wiped.
		_ = s.fixedModels.DeleteFixedModel(ctx, m.ID)
	}

	newM := &storage.FixedModel{
		Name:     name,
		Provider: provider,
		Model:    model,
	}
	_ = s.fixedModels.CreateFixedModel(ctx, newM)

	// Refresh the ID (CreateFixedModel does not set it on the arg) and add tools.
	created, _ := s.fixedModels.GetFixedModel(ctx, name)
	for _, toolName := range tools {
		_ = s.fixedModels.AddTool(ctx, &storage.FixedModelTool{
			FixedModelID: created.ID,
			Tool:         toolName,
		})
	}
}

// ── Tools ──────────────────────────────────────────────────────

// GetModelSettings retorna as configurações (context_size, max_tokens etc.)
// de um modelo no formato "provider/model".
func (s *Store) GetModelSettings(modelString string) (provider.ModelSettings, bool) {
	if modelString == "" || !strings.Contains(modelString, "/") {
		return provider.ModelSettings{}, false
	}
	parts := strings.SplitN(modelString, "/", 2)
	providerName := parts[0]
	modelName := parts[1]

	providers := s.ListProviders()
	p, ok := providers[providerName]
	if !ok {
		return provider.ModelSettings{}, false
	}
	ms, ok := p.Models[modelName]
	if !ok {
		return provider.ModelSettings{}, false
	}
	// Fallback: inferir context_size do nome se não foi configurado
	if ms.ContextSize <= 0 {
		if inferred := iprovider.InferContextSize(modelName); inferred > 0 {
			ms.ContextSize = inferred
		}
	}
	return ms, true
}

// ── Tools ──────────────────────────────────────────────────────
