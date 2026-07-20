package specwizardmgr

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	specwizardmodel "ada-love-ide/internal/config/specwizard"
	"ada-love-ide/internal/db"
	"ada-love-ide/internal/plugins"
	"ada-love-ide/internal/prompts"
)

// LLMInferFn is the function that performs the actual LLM call for inference.
// It receives system prompt, user prompt, temperature, and max tokens, and
// returns the generated text or an error.
type LLMInferFn func(ctx context.Context, systemPrompt, userPrompt string, temperature float64, maxTokens int) (string, error)

// Manager is the single integration point for everything related to the
// Spec Wizard: persistence, expert-plugin proxying, option catalogs,
// architecture analysis, and LLM-backed field inference.
type Manager struct {
	db      *db.Store
	plugins *plugins.PluginManager
	llmFn   LLMInferFn

	cacheMu sync.Mutex
	cache   map[string]map[string]any
}

func New(db *db.Store, pluginMgr *plugins.PluginManager) *Manager {
	return &Manager{
		db:      db,
		plugins: pluginMgr,
		cache:   make(map[string]map[string]any),
	}
}

// SetLLMFn injects the LLM inference function. Must be called before InferField.
func (m *Manager) SetLLMFn(fn LLMInferFn) {
	m.llmFn = fn
}

// ── Persistence ──────────────────────────────────────────────────────────

func (m *Manager) List() []specwizardmodel.SpecWizardConfig {
	return m.db.ListWizards()
}

func (m *Manager) Get(id string) (*specwizardmodel.SpecWizardConfig, error) {
	w, ok := m.db.GetWizard(id)
	if !ok {
		return nil, nil
	}
	return &w, nil
}

func (m *Manager) Save(w specwizardmodel.SpecWizardConfig) {
	if w.ID == "" {
		w.ID = time.Now().Format("20060102150405")
	}
	w.UpdatedAt = time.Now()
	m.db.PutWizard(w)
}

func (m *Manager) Delete(id string) {
	m.db.DeleteWizard(id)
}

// ── Experts / plugin proxy (STDIO) ──────────────────────────────────────────

func (m *Manager) GetExperts() []map[string]any {
	plugins := m.plugins.List()
	result := make([]map[string]any, 0, len(plugins))
	for _, p := range plugins {
		result = append(result, map[string]any{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"language":    p.Language,
		})
	}
	return result
}

// fetchOptions invokes the plugin's `options` action once (cached per plugin)
// and returns the raw response map.
func (m *Manager) fetchOptions(plugin *plugins.ExpertPlugin) (map[string]any, error) {
	m.cacheMu.Lock()
	if cached, ok := m.cache[plugin.ID]; ok {
		m.cacheMu.Unlock()
		return cached, nil
	}
	m.cacheMu.Unlock()

	resp, err := m.plugins.CallExpert(plugin, "options", "")
	if err != nil {
		return nil, err
	}

	m.cacheMu.Lock()
	m.cache[plugin.ID] = resp
	m.cacheMu.Unlock()
	return resp, nil
}

// callOptions extracts the given keys from a plugin's /options response. When
// lang is empty it aggregates across all experts, de-duplicating by option ID.
func (m *Manager) callOptions(lang string, keys ...string) []Option {
	if lang == "" {
		return m.aggregateOptions(keys...)
	}
	plugin, ok := m.plugins.FindByLanguage(lang)
	if !ok {
		return []Option{}
	}
	resp, err := m.fetchOptions(plugin)
	if err != nil {
		fmt.Printf("[specwizardmgr] failed to call expert %s: %v\n", plugin.ID, err)
		return []Option{}
	}
	return optionsFrom(resp, keys...)
}

func (m *Manager) aggregateOptions(keys ...string) []Option {
	seen := make(map[string]bool)
	var out []Option
	for _, plugin := range m.plugins.List() {
		resp, err := m.fetchOptions(plugin)
		if err != nil {
			continue
		}
		for _, opt := range optionsFrom(resp, keys...) {
			if opt.ID != "" && !seen[opt.ID] {
				seen[opt.ID] = true
				out = append(out, opt)
			}
		}
	}
	return out
}

func (m *Manager) GetArchitectures() []Option {
	return m.aggregateOptions("architectures")
}

func (m *Manager) GetPatterns(lang string) []Option {
	return m.callOptions(lang, "architectures")
}

func (m *Manager) GetStacks(lang string) []map[string]any {
	plugin, ok := m.plugins.FindByLanguage(lang)
	if !ok {
		return []map[string]any{}
	}
	resp, err := m.fetchOptions(plugin)
	if err != nil {
		fmt.Printf("[specwizardmgr] failed to call expert %s: %v\n", plugin.ID, err)
		return []map[string]any{}
	}
	stacks := make([]map[string]any, 0)
	if templates, ok := resp["stack_templates"].([]interface{}); ok {
		for _, t := range templates {
			if m, ok := t.(map[string]interface{}); ok {
				stacks = append(stacks, map[string]any{
					"id":        m["id"],
					"name":      m["name"],
					"libraries": m["libraries"],
				})
			}
		}
	}
	return stacks
}

// GetStateManagement resolves BOTH the correct plural key (`state_managements`)
// and the legacy singular one (`state_management`) used by older plugins.
func (m *Manager) GetStateManagement(lang string) []Option {
	return m.callOptions(lang, "state_managements", "state_management")
}

// ── Catalogs sourced from the expert plugins ────────────────────────────────

func (m *Manager) GetPersistenceOptions(lang string) []Option {
	return m.callOptions(lang, "data_strategies", "persistence")
}

func (m *Manager) GetEngineeringPhilosophies(lang string) []Option {
	return m.callOptions(lang, "philosophies", "engineering_philosophies")
}

func (m *Manager) GetDesignPatterns(lang string) []Option {
	return m.callOptions(lang, "design_patterns")
}

func (m *Manager) GetDataPatterns(lang string) []Option {
	return m.callOptions(lang, "data_patterns")
}

// ── Architecture analysis ───────────────────────────────────────────────────

// ComputeHealth returns a 0–100 health score derived from the selected
// patterns. It rewards testability and penalizes coupling and overload.
func (m *Manager) ComputeHealth(cfg specwizardmodel.SpecWizardConfig) int {
	philos := toSet(cfg.EngineeringPhilosophies)
	data := toSet(cfg.DataPatterns)

	score := 70

	if contains(data, "repository") {
		score += 12
	}
	if contains(philos, "solid") {
		score += 10
	}
	if contains(philos, "dry") {
		score += 4
	}

	// Fat-model danger: business rules + persistence tightly coupled.
	if contains(data, "active_record") {
		score -= 18
	}

	// Architecture overload: too many patterns for a typical project.
	total := len(cfg.DesignPatterns) + len(cfg.DataPatterns) + len(cfg.EngineeringPhilosophies)
	if total > 8 {
		score -= (total - 8) * 6
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

// GetRecommendations generates the insight cards previously hardcoded in the
// frontend, now driven by the actual selections.
func (m *Manager) GetRecommendations(cfg specwizardmodel.SpecWizardConfig) []Recommendation {
	var recs []Recommendation

	philos := toSet(cfg.EngineeringPhilosophies)
	data := toSet(cfg.DataPatterns)

	if contains(data, "repository") && contains(philos, "solid") {
		recs = append(recs, Recommendation{
			Level:       "success",
			Title:       "High Testability",
			Description: "Repositories with SOLID facilitate mocking and dependency injection, enabling easy unit and integration testing.",
		})
	} else {
		recs = append(recs, Recommendation{
			Level:       "warning",
			Title:       "Low Testability",
			Description: "Consider adding Repository patterns and SOLID principles to improve mocking and dependency injection for easier testing.",
		})
	}

	if contains(data, "active_record") {
		recs = append(recs, Recommendation{
			Level:       "warning",
			Title:       "'Fat' Model Danger",
			Description: "Business rules and persistence seem tightly coupled. Watch out for giant classes that violate the Single Responsibility Principle.",
		})
	}

	total := len(cfg.DesignPatterns) + len(cfg.DataPatterns) + len(cfg.EngineeringPhilosophies)
	if total > 8 {
		recs = append(recs, Recommendation{
			Level:       "critical",
			Title:       "CRITICAL: Architecture Overload",
			Description: fmt.Sprintf("Your architecture is overloaded with %d patterns for a project of this size, which will cause significant slowdown in development and maintenance.", total),
		})
	}

	return recs
}

// ── LLM Inference ───────────────────────────────────────────────────────────

// InferField uses the LLM to generate a suggestion for the given target field
// based on the current SpecWizardConfig. It validates prerequisites first,
// builds a structured prompt, and calls the injected LLM function.
func (m *Manager) InferField(ctx context.Context, target string, cfg specwizardmodel.SpecWizardConfig) (string, error) {
	fmt.Printf("[specwizardmgr] InferField start target=%q name=%q expert=%q\n", target, cfg.Name, cfg.ExpertLanguagePlugin)

	if m.llmFn == nil {
		return "", fmt.Errorf("[ALERTA] Modelo 'spec' não configurado — LLM inference function not set")
	}
	fmt.Printf("[specwizardmgr] llmFn is set\n")

	if err := ValidarInferencia(target, cfg); err != nil {
		fmt.Printf("[specwizardmgr] validation FAILED: %v\n", err)
		return "", err
	}
	fmt.Printf("[specwizardmgr] validation passed\n")

	fctx := fieldContextFrom(cfg)
	currentValue := currentValueFrom(target, cfg)
	fmt.Printf("[specwizardmgr] currentValue len=%d\n", len(currentValue))

	prompt, err := prompts.Build(target, fctx, currentValue)
	if err != nil {
		fmt.Printf("[specwizardmgr] prompt build FAILED: %v\n", err)
		return "", err
	}
	fmt.Printf("[specwizardmgr] prompt built: sys=%d user=%d temp=%.1f max=%d\n",
		len(prompt.SystemPrompt), len(prompt.UserPrompt), prompt.Temperature, prompt.MaxTokens)

	result, err := m.llmFn(ctx, prompt.SystemPrompt, prompt.UserPrompt, prompt.Temperature, prompt.MaxTokens)
	if err != nil {
		fmt.Printf("[specwizardmgr] llmFn FAILED: %v\n", err)
		return "", err
	}
	fmt.Printf("[specwizardmgr] llmFn OK result len=%d\n", len(result))
	return result, nil
}

func fieldContextFrom(cfg specwizardmodel.SpecWizardConfig) prompts.FieldContext {
	return prompts.FieldContext{
		SpecName:                cfg.Name,
		ExpertLanguagePlugin:    cfg.ExpertLanguagePlugin,
		PRD:                     cfg.PRD,
		FunctionalReqs:          strings.Join(cfg.FunctionalRequirements, "\n"),
		NonFunctionalReqs:       strings.Join(cfg.NonFunctionalRequirements, "\n"),
		Architecture:            cfg.Architecture,
		Persistence:             cfg.Persistence,
		EngineeringPhilosophies: strings.Join(cfg.EngineeringPhilosophies, ", "),
		DesignPatterns:          strings.Join(cfg.DesignPatterns, ", "),
		DataPatterns:            strings.Join(cfg.DataPatterns, ", "),
		StackPlugin:             cfg.StackPlugin,
		DependencyManifest:      depsToString(cfg.DependencyManifest),
		StateManagement:         cfg.Business.StateManagement,
		APIContract:             cfg.Business.APIContract,
		CustomizationDetails:    cfg.Business.CustomizationDetails,
		FinalAdjustments:        cfg.Business.FinalAdjustments,
	}
}

func currentValueFrom(target string, cfg specwizardmodel.SpecWizardConfig) string {
	switch strings.ToUpper(target) {
	case "PRD":
		return cfg.PRD
	case "FUNCTIONAL":
		return strings.Join(cfg.FunctionalRequirements, "\n")
	case "NON_FUNCTIONAL":
		return strings.Join(cfg.NonFunctionalRequirements, "\n")
	case "API_CONTRACT":
		return cfg.Business.APIContract
	case "CUSTOMIZATION":
		return cfg.Business.CustomizationDetails
	case "FINAL_ADJUSTMENTS":
		return cfg.Business.FinalAdjustments
	}
	return ""
}

func depsToString(deps []specwizardmodel.Dependency) string {
	if len(deps) == 0 {
		return "(none)"
	}
	var b strings.Builder
	for i, d := range deps {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(d.Name)
		if d.Version != "" && d.Version != "latest" {
			b.WriteString("@")
			b.WriteString(d.Version)
		}
		if d.Mandatory {
			b.WriteString(" (mandatory)")
		}
	}
	return b.String()
}

// ── helpers ────────────────────────────────────────────────────────────────

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, it := range items {
		s[strings.ToLower(strings.TrimSpace(it))] = true
	}
	return s
}

func contains(set map[string]bool, key string) bool {
	return set[strings.ToLower(strings.TrimSpace(key))]
}
