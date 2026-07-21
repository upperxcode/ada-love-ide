package specwizardmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	specwizardmodel "ada-love-ide/internal/config/specwizard"
	"ada-love-ide/internal/config/workspace"
	"ada-love-ide/internal/db"
)

// SyncSpecToWorkspace gera os arquivos .spec-wizard/ e .AGENTS.md no diretório
// do workspace com base nos dados do Spec Wizard vinculado.
func SyncSpecToWorkspace(store *db.Store, ws workspace.WorkspaceConfig) {
	wiz, ok := store.GetWizard(ws.SpecWizardID)
	if !ok {
		fmt.Printf("[specwizard] SyncSpecToWorkspace: Spec Wizard %s not found\n", ws.SpecWizardID)
		return
	}

	dir := ws.Path
	if len(ws.Folders) > 0 && ws.Folders[0] != "" {
		dir = ws.Folders[0]
	}

	specDir := filepath.Join(dir, ".spec-wizard")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		fmt.Printf("[specwizard] SyncSpecToWorkspace: mkdir error: %v\n", err)
		return
	}

	writeFile := func(name, content string) {
		path := filepath.Join(specDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Printf("[specwizard] SyncSpecToWorkspace: write %s error: %v\n", name, err)
		}
	}

	writeFile("config.json", generateSpecConfigJSON(wiz))
	writeFile("PRD.md", generatePRDMarkdown(wiz))
	writeFile("skills.md", generateSkillsMarkdown(wiz))

	agentsPath := filepath.Join(dir, ".AGENTS.md")
	agentsContent := generateAgentsMarkdown(wiz)
	if err := os.WriteFile(agentsPath, []byte(agentsContent), 0644); err != nil {
		fmt.Printf("[specwizard] SyncSpecToWorkspace: write .AGENTS.md error: %v\n", err)
	}

	fmt.Printf("[specwizard] Spec synced to %s\n", dir)
}

func generateSpecConfigJSON(wiz specwizardmodel.SpecWizardConfig) string {
	cfg := map[string]interface{}{
		"projectName":               wiz.Name,
		"language":                  wiz.ExpertLanguagePlugin,
		"architecture":              wiz.Architecture,
		"dataStrategy":              wiz.Persistence,
		"domain":                    truncate(wiz.PRD, 500),
		"functionalRequirements":    wiz.FunctionalRequirements,
		"nonFunctionalRequirements": wiz.NonFunctionalRequirements,
		"patterns":                  append(append(wiz.EngineeringPhilosophies, wiz.DesignPatterns...), wiz.DataPatterns...),
		"stateManagement":           wiz.Business.StateManagement,
		"apiContract":               truncate(wiz.Business.APIContract, 500),
		"instructions":              truncate(wiz.Business.FinalAdjustments, 1000),
		"customization":             truncate(wiz.Business.CustomizationDetails, 500),
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}

func generatePRDMarkdown(wiz specwizardmodel.SpecWizardConfig) string {
	return fmt.Sprintf(`# 📄 PRODUCT REQUIREMENT DOCUMENT (PRD)

## 🎯 Identity & Purpose
- **Project:** %s
- **Expert Plugin:** %s
- **Architecture:** %s
- **Persistence:** %s

## 🚀 Business Requirements

### PRD
%s

### ✅ Functional
%s

### ⚙️ Non-Functional
%s
`,
		wiz.Name, wiz.ExpertLanguagePlugin, wiz.Architecture, wiz.Persistence,
		wiz.PRD,
		joinLines(wiz.FunctionalRequirements),
		joinLines(wiz.NonFunctionalRequirements),
	)
}

func generateSkillsMarkdown(wiz specwizardmodel.SpecWizardConfig) string {
	return fmt.Sprintf(`# 📏 SKILLS & GOLDEN RULES

## ⚓ Core Technical Rules
### Architecture
- Architecture: %s
- Persistence: %s

### Patterns
- Engineering Philosophies: %s
- Design Patterns: %s
- Data Patterns: %s

### Stack
- Plugin: %s
- Dependencies: %s
`,
		wiz.Architecture, wiz.Persistence,
		joinComma(wiz.EngineeringPhilosophies),
		joinComma(wiz.DesignPatterns),
		joinComma(wiz.DataPatterns),
		wiz.StackPlugin,
		depsSummary(wiz.DependencyManifest),
	)
}

func generateAgentsMarkdown(wiz specwizardmodel.SpecWizardConfig) string {
	return fmt.Sprintf(`# 🧙 Spec-Driven Development Rules

## 📐 Context Isolation
1. Ignore raw chat history. Consume only .RESUME.md and .spec-wizard/ files
2. Always consult .spec-wizard/config.json for architecture, patterns, and stack
3. Chain-of-Thought: Plan before coding

## 🏛️ Architecture Enforcement
- **Architecture:** %s
- **Persistence:** %s
- **Philosophies:** %s
- **Patterns:** %s | %s

## ✅ Sensor Validation & Auto-Correction
1. After code changes, run: go vet, go fmt, go test ./...
2. If errors: capture, remount prompt, re-execute (max 3 attempts)
3. After successful compile, update .RESUME.md

## 🧩 Stack
- **Plugin:** %s
- **Dependencies:** %s
`,
		wiz.Architecture, wiz.Persistence,
		joinComma(wiz.EngineeringPhilosophies),
		joinComma(wiz.DesignPatterns), joinComma(wiz.DataPatterns),
		wiz.StackPlugin,
		depsSummary(wiz.DependencyManifest),
	)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func joinLines(items []string) string {
	out := ""
	for _, item := range items {
		out += "- " + item + "\n"
	}
	return out
}

func joinComma(items []string) string {
	out := ""
	for i, item := range items {
		if i > 0 {
			out += ", "
		}
		out += item
	}
	return out
}

func depsSummary(deps []specwizardmodel.Dependency) string {
	if len(deps) == 0 {
		return "(none)"
	}
	out := ""
	for i, d := range deps {
		if i > 0 {
			out += ", "
		}
		out += d.Name
		if d.Version != "" && d.Version != "latest" {
			out += "@" + d.Version
		}
		if d.Mandatory {
			out += " (mandatory)"
		}
	}
	return out
}
