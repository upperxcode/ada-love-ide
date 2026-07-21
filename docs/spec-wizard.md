# 🧙 Spec Wizard — Governança de Arquitetura via IA

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Arquitetura](#2-arquitetura)
3. [Data Model](#3-data-model)
4. [Inference Engine](#4-inference-engine)
5. [Prompts Package](#5-prompts-package)
6. [Workspace Integration](#6-workspace-integration)
7. [MCP Integration](#7-mcp-integration)
8. [File Generation — SyncSpecToWorkspace](#8-file-generation--syncspecToWorkspace)
9. [Frontend — SpecWizardDialog](#9-frontend--specwizarddialog)
10. [API Reference](#10-api-reference)
11. [Fluxos](#11-fluxos)

---

## 1. Visão Geral

O **Spec Wizard** é o cérebro arquitetural do Ada Love IDE. Ele permite:

- **Definir** a arquitetura, padrões, stack e regras de negócio de um projeto via formulário visual
- **Inferir** campos usando IA (modelo "spec") com validação de dependências entre campos
- **Exportar** a configuração para arquivos `.spec-wizard/` no diretório do workspace
- **Integrar** com o MCP server `sw` para execução de tarefas orientadas a especificação

### Hierarquia

```
Workspace (Path, Folders[], SpecWizardID)
  └── SpecWizardConfig
      ├── PRD, FunctionalReqs, NonFunctionalReqs
      ├── Architecture, Persistence
      ├── EngineeringPhilosophies, DesignPatterns, DataPatterns
      ├── StackPlugin, DependencyManifest
      └── Business { StateManagement, APIContract, ... }
```

---

## 2. Arquitetura

```
internal/
├── prompts/                          # Fábrica de prompts para IA
│   ├── prompt.go                     # Tipos: Prompt, FieldContext, PromptBuilder
│   └── specwizard.go                 # Build(target, ctx, currentValue) → Prompt
│
├── specwizardmgr/                    # Orquestração
│   ├── manager.go                    # CRUD, plugin proxy, InferField, ComputeHealth
│   ├── inference.go                  # ValidarInferencia(target, cfg) → error
│   └── options.go                    # Option, Recommendation, optionsFrom()
│
├── config/specwizard/                # Modelo de dados
│   └── specwizard.go                 # SpecWizardConfig, Business, StackItem, Dependency
│
├── engine/                           # Injeção de dependências
│   └── engine.go                     # LLMInferFn, SetLLMFn, ResolveWorkspaceDir
│
app_specwizard.go                     # Wails bindings (15 métodos)
app_sessions.go                       # GetSessionWorkspaceSpec, GetWorkspaceSpec
app_workspaces.go                     # SyncSpecToWorkspace, ensureSWMCP
```

### Fluxo de dados

```
Frontend (SpecWizardDialog)
  │
  ├── CRUD → App.SaveSpecWizard() → specwizardmgr.Manager → db.Store → SQLite
  │
  ├── Inferência → App.InferField() → specwizardmgr.InferField()
  │     ├── ValidarInferencia()            → inference.go
  │     ├── prompts.Build()                → prompts/specwizard.go
  │     └── llmFn (engine.go)             → ada-llm-client → HTTP
  │
  └── Sync → App.SaveWorkspace()
        ├── ensureSWMCP()                 → registra MCP server sw-{title}
        └── SyncSpecToWorkspace()         → gera .spec-wizard/ + .AGENTS.md
```

---

## 3. Data Model

### SpecWizardConfig (`internal/config/specwizard/specwizard.go`)

```go
type SpecWizardConfig struct {
    ID                        string       `json:"id"`
    Name                      string       `json:"name"`
    Description               string       `json:"description"`
    ExpertLanguagePlugin      string       `json:"expert_language_plugin"`
    PRD                       string       `json:"prd"`
    FunctionalRequirements    []string     `json:"functional_requirements"`
    NonFunctionalRequirements []string     `json:"non_functional_requirements"`
    Persistence               string       `json:"persistence"`
    Architecture              string       `json:"architecture"`
    EngineeringPhilosophies   []string     `json:"engineering_philosophies"`
    DesignPatterns            []string     `json:"design_patterns"`
    DataPatterns              []string     `json:"data_patterns"`
    StackConfig               []StackItem  `json:"stack_config"`
    Business                  Business     `json:"business"`
    Color                     string       `json:"color"`
    Icon                      string       `json:"icon"`
    ArchitectureHealth        int          `json:"architecture_health"`
    DependencyManifest        []Dependency `json:"dependency_manifest"`
    StackPlugin               string       `json:"stack_plugin"`
    CreatedAt                 time.Time    `json:"created_at"`
    UpdatedAt                 time.Time    `json:"updated_at"`
}
```

### Business

```go
type Business struct {
    StateManagement             string `json:"state_management"`
    APIContract                 string `json:"api_contract"`
    CustomizationDetails        string `json:"customization_details"`
    FinalAdjustments            string `json:"final_adjustments"`
    ArchitectureRecommendations string `json:"architecture_recommendations"`
}
```

### StackItem / Dependency

```go
type StackItem struct {
    Name    string `json:"name"`
    Example string `json:"example"`
}

type Dependency struct {
    Name      string `json:"lib"`
    Version   string `json:"ver"`
    Mandatory bool   `json:"mandatory"`
}
```

### Persistência (SQLite)

Tabela `spec_wizards` no `ada_love.db`. Campos de slice (`[]string`, `[]Dependency`) são serializados como JSON em colunas de texto. Campos do `Business` são colunas individuais (`business_state_management`, `business_api_contract`, etc.).

---

## 4. Inference Engine

### 4.1 Validação de Dependências (`internal/specwizardmgr/inference.go`)

Antes de chamar a IA, o sistema valida que os pré-requisitos do campo alvo estão preenchidos:

| Target | Pré-requisitos |
|---|---|
| `PRD` | `expert_language_plugin` |
| `FUNCTIONAL` | `expert_language_plugin` + `prd` |
| `NON_FUNCTIONAL` | `expert_language_plugin` + `prd` + `functional_requirements` |
| `API_CONTRACT` | Steps 1-4 completos (architecture, persistence, philosophies, patterns, manifest) |
| `CUSTOMIZATION` | Steps 1-4 completos |
| `FINAL_ADJUSTMENTS` | Steps 1-4 + `api_contract` + `customization_details` |

### 4.2 Execução (`internal/specwizardmgr/manager.go`)

```go
func (m *Manager) InferField(ctx context.Context, target string, cfg SpecWizardConfig) (string, error)
```

1. Verifica se `llmFn` está configurada
2. Chama `ValidarInferencia(target, cfg)` — se falhar, retorna erro
3. Converte `SpecWizardConfig` → `prompts.FieldContext`
4. Extrai `currentValue` do campo alvo
5. Chama `prompts.Build(target, fctx, currentValue)` → `Prompt`
6. Chama `m.llmFn(ctx, system, user, temp, maxTokens)` → LLM

### 4.3 LLM Function (`internal/engine/engine.go`)

A `llmFn` é injetada pelo Engine no startup:

```go
specWizardMgr.SetLLMFn(func(ctx context.Context, systemPrompt, userPrompt string, temperature float64, maxTokens int) (string, error) {
    specProvider, specModel, _ := store.GetFixedModel("spec")
    // look up provider config → create ada-llm-client → call Generate()
})
```

- Usa o **fixed model "spec"** (tabela `fixed_models`)
- Provider configurado em Settings > Models
- Timeout de 60 segundos via `context.WithTimeout`
- Logs em cada etapa para debug (`[engine.llmFn]`, `[specwizardmgr]`, `[inference]`)

### 4.4 Integração com o antigo `SuggestFieldValue`

O método `SuggestFieldValue` (mock) foi substituído por `InferField` com assinatura:

```go
// ANTES
func (a *App) SuggestFieldValue(fieldName, context, currentValue string) string

// DEPOIS
func (a *App) InferField(fieldName string, cfg specwizard.SpecWizardConfig) (string, error)
```

---

## 5. Prompts Package

### 5.1 Tipos Base (`internal/prompts/prompt.go`)

```go
type Prompt struct {
    SystemPrompt string
    UserPrompt   string
    Temperature  float64
    MaxTokens    int
}

type FieldContext struct {
    SpecName              string
    ExpertLanguagePlugin  string
    PRD                   string
    FunctionalReqs        string
    NonFunctionalReqs     string
    Architecture          string
    Persistence           string
    EngineeringPhilosophies string
    DesignPatterns        string
    DataPatterns          string
    StackPlugin           string
    DependencyManifest    string
    StateManagement       string
    APIContract           string
    CustomizationDetails  string
    FinalAdjustments      string
}

type PromptBuilder func(ctx FieldContext, currentValue string) Prompt
```

### 5.2 Builders (`internal/prompts/specwizard.go`)

| Builder | Temperature | MaxTokens | System Prompt |
|---|---|---|---|
| `buildPRD` | 0.2 | 2000 | Elite Product Manager |
| `buildFunctional` | 0.1 | 2000 | System Analyst |
| `buildNonFunctional` | 0.1 | 2000 | Technical Architect |
| `buildAPIContract` | **0.0** | 2000 | Principal Software Architect |
| `buildCustomization` | 0.2 | 2000 | UI/UX Technical Designer |
| `buildFinalAdjustments` | 0.2 | 2000 | Senior Implementation Engineer |

Todos os prompts incluem o sufixo crítico:

```
CRITICAL: Output must be concise, structured, and machine-readable.
This text will be consumed by another AI model — not by humans.
Avoid fluff, marketing language, adjectives, and verbose explanations.
Use short sentences, bullet points, and clear section headers.
```

### 5.3 Temperatura por Target

| Target | Temp | Rationale |
|---|---|---|
| `API_CONTRACT` | **0.0** | Nomes de campos, tipos — determinismo absoluto |
| `FUNCTIONAL` | 0.1 | Checklist estruturado |
| `NON_FUNCTIONAL` | 0.1 | Checklist técnico |
| `PRD` | 0.2 | Expansão criativa controlada |
| `CUSTOMIZATION` | 0.2 | Detalhamento visual |
| `FINAL_ADJUSTMENTS` | 0.2 | Roadmap |

---

## 6. Workspace Integration

### 6.1 SpecWizardID no Workspace

O `WorkspaceConfig` possui:

```go
type WorkspaceConfig struct {
    // ...
    SpecWizard   string   `json:"spec_wizard"`
    SpecWizardID string   `json:"spec_wizard_id"`
    // ...
}
```

Persistido em `workspaces` tabela SQLite, coluna `spec_wizard_id`.

### 6.2 Resolução em Runtime

**`GetSessionWorkspaceSpec(sessionID)`** (`app_sessions.go`):

```
sessionID → Session.WorkspaceID → GetWorkspace(path) → Workspace.SpecWizardID
  → GetWizard(id) → *SpecWizardConfig
  → Se vazio → erro "Workspace sem Spec Wizard"
```

**`GetWorkspaceSpec(workspacePath)`** — mesmo fluxo por path do workspace.

### 6.3 Alerta no Frontend

No `ChatPanel.svelte`, após criar a sessão:

```typescript
const spec = await GetSessionWorkspaceSpec(sessionID);
// Se falhar → toastStore.warning('Workspace sem Spec Wizard', ...)
```

---

## 7. MCP Integration

### 7.1 Auto-Registro do MCP Server `sw`

Quando um workspace é salvo com `SpecWizardID != ""`, o `SaveWorkspace` chama `ensureSWMCP()`:

```go
func ensureSWMCP(store *db.Store, ws workspace.WorkspaceConfig) {
    // Cria MCP server: sw-{workspace-title}
    store.SaveMCPServer("sw-"+ws.Title, mcp.MCPServerUI{
        Command: "/home/john/.local/bin/sw",
        Args:    []string{"mcp"},
        Env: map[string]string{
            "WZ_PROJECT_PATH": dir,  // diretório do workspace
        },
        Enabled: true,
        Icon:    "📋",
        Color:   "#8b5cf6",
    })
}
```

### 7.2 Variável WZ_PROJECT_PATH

O MCP server `sw` usa `WZ_PROJECT_PATH` para saber qual projeto está sendo gerenciado. Essa variável é definida como:

```
WZ_PROJECT_PATH = workspace.Folders[0] ?? workspace.Path
```

### 7.3 Launcher Script (`scripts/sw-launcher.sh`)

O `sw` possui um launcher (shell script) que:

1. Recebe o caminho do projeto como argumento
2. Busca no `~/.spec-wizard/governance.db` os dados do projeto
3. Monta as flags `-resume-content`, `-agents-content`, `-skills-content`
4. Executa `sw mcp` com essas flags

**Mapeamento do launcher:**

```
-resume-content  ← prd + functional_requirements + non_functional_requirements + architecture
-agents-content  ← instructions (fallback: architecture)
-skills-content  ← customization
```

### 7.4 Nosso MCP Server vs Launcher

Atualmente registramos o `sw mcp` diretamente (sem launcher). O launcher depende do `governance.db` do `sw`, enquanto nossos dados estão no `ada_love.db`. Futuramente podemos:

1. **Opção A:** Sincronizar dados do Spec Wizard para o `governance.db` do `sw`
2. **Opção B:** Nosso backend montar as flags e executar `sw mcp` como subprocesso
3. **Opção C:** Usar o launcher diretamente, que já lê do `governance.db`

### 7.5 Tools MCP Expostas pelo `sw`

| Tool | Descrição |
|---|---|
| `context_assembler` | Monta o prompt slim (Clean Window) para IA |
| `context_capture` | Captura resposta da IA na janela deslizante |
| `wz` | Executa comandos do Spec Wizard (`init`, `roadmap`, `refine`, etc.) |
| `db_read` | Lê entidade do banco de governança |
| `db_insert` | Insere entidade no banco |
| `db_remove` | Remove entidade |
| `db_select` | Select genérico |

---

## 8. File Generation — SyncSpecToWorkspace

### 8.1 Disparo

`SyncSpecToWorkspace` é chamado automaticamente em `SaveWorkspace` quando `SpecWizardID` está preenchido. Também exposto via `SyncSpecToWorkspaceBySession(sessionID)` para o frontend.

### 8.2 Arquivos Gerados

No diretório do workspace (primeira pasta do `Folders`, fallback `Path`):

```
workspace/
├── .spec-wizard/
│   ├── config.json       ← Configuração do projeto (sw-compatible)
│   ├── PRD.md           ← Product Requirements Document
│   └── skills.md        ← Golden Rules
└── .AGENTS.md            ← Regras de governança para agentes de IA
```

### 8.3 config.json (formato)

```json
{
  "projectName": "Local Vault",
  "language": "go",
  "architecture": "clean_arch",
  "dataStrategy": "sql",
  "domain": "...",
  "functionalRequirements": ["RF01: ...", "RF02: ..."],
  "nonFunctionalRequirements": ["NF01: ..."],
  "patterns": ["solid", "dry", "repository"],
  "stateManagement": "",
  "apiContract": "...",
  "instructions": "...",
  "customization": "..."
}
```

### 8.4 .AGENTS.md (conteúdo gerado)

```markdown
# 🧙 Spec-Driven Development Rules

## 📐 Context Isolation
1. Ignore raw chat history. Consume only .RESUME.md and .spec-wizard/ files
2. Always consult .spec-wizard/config.json for architecture, patterns, and stack
3. Chain-of-Thought: Plan before coding

## 🏛️ Architecture Enforcement
- **Architecture:** clean_arch
- **Persistence:** sql
- **Patterns:** solid, dry | repository

## ✅ Sensor Validation & Auto-Correction
1. After code changes, run: go vet, go fmt, go test ./...
2. If errors: capture, remount prompt, re-execute (max 3 attempts)
3. After successful compile, update .RESUME.md

## 🧩 Stack
- **Plugin:** go_backend
- **Dependencies:** gin, gorm, viper
```

---

## 9. Frontend — SpecWizardDialog

### 9.1 Localização

`frontend/src/lib/components/settings/SpecWizardDialog.svelte`

### 9.2 Steps

| Step | Label | Campos |
|---|---|---|
| 1 | Identity | Project Name, Description, Expert Language Plugin, Domain & Scope (PRD, Functional, Non-Functional) |
| 2 | Architecture | Select Base Architecture, Persistence Strategy |
| 3 | Patterns | Engineering Philosophies, Design Patterns (GoF), Data Patterns |
| 4 | Stack | Stack Plugin, Dependency Manifest |
| 5 | Business | State Management, API Contract, Customization Details |
| 6 | Advisor | Implementation Instructions, Architecture Recommendations (auto) |

### 9.3 Botões ✨ Sugerir

Presentes em 6 campos:

| Step | Campo | Target |
|---|---|---|
| 1 | PRD | `PRD` |
| 1 | Functional Requirements | `FUNCTIONAL` |
| 1 | Non-Functional Requirements | `NON_FUNCTIONAL` |
| 5 | API Contract | `API_CONTRACT` |
| 5 | Customization Details | `CUSTOMIZATION` |
| 6 | Implementation Instructions | `FINAL_ADJUSTMENTS` |

**Comportamento:**

1. Verifica se `spec_provider` e `spec_model` estão configurados (via `GetAdaConfig()`)
2. Se campo já tem valor → abre dialog de confirmação (padrão do sistema)
3. Chama `App.InferField(target, formData)`
4. Mostra banner "Inferindo {target}..." com spinner durante a chamada
5. Em caso de erro, mostra toast com a mensagem
6. Desabilita todos os botões enquanto uma inferência está rodando

### 9.4 Health Score

O footer exibe **Architecture Health X%** — calculado automaticamente pelo backend (`ComputeHealth()`) sempre que patterns/architecture mudam. O backend também gera **Recommendation cards** (success/warning/critical).

---

## 10. API Reference

### 10.1 Wails Bindings (Go → Frontend)

#### Spec Wizard CRUD

| Método | Retorno | Descrição |
|---|---|---|
| `GetSpecWizards()` | `[]SpecWizardConfig` | Lista todos |
| `GetSpecWizard(id)` | `*SpecWizardConfig` | Por ID |
| `SaveSpecWizard(w)` | `void` | Cria/atualiza |
| `DeleteSpecWizard(id)` | `void` | Remove |

#### Expert Plugin Proxy

| Método | Retorno | Descrição |
|---|---|---|
| `GetExperts()` | `[]map` | Lista plugins instalados |
| `GetPatterns(lang)` | `[]Option` | Arquiteturas por linguagem |
| `GetArchitectures()` | `[]Option` | Agregado de todos plugins |
| `GetStacks(lang)` | `[]map` | Stack templates por linguagem |
| `GetStateManagement(lang)` | `[]Option` | Opções de state management |
| `GetPersistenceOptions(lang)` | `[]Option` | Estratégias de persistência |
| `GetEngineeringPhilosophies(lang)` | `[]Option` | Filosofias de engenharia |
| `GetDesignPatterns(lang)` | `[]Option` | Design patterns (GoF) |
| `GetDataPatterns(lang)` | `[]Option` | Data patterns |

#### Análise Arquitetural

| Método | Retorno | Descrição |
|---|---|---|
| `ComputeHealth(cfg)` | `int` (0-100) | Saúde da arquitetura |
| `GetRecommendations(cfg)` | `[]Recommendation` | Cards de insight |

#### Inferência

| Método | Retorno | Descrição |
|---|---|---|
| `InferField(fieldName, cfg)` | `(string, error)` | Inferência via modelo "spec" |

#### Workspace Integration

| Método | Retorno | Descrição |
|---|---|---|
| `GetSessionWorkspaceSpec(sessionID)` | `(*SpecWizardConfig, error)` | Spec do workspace da sessão |
| `GetWorkspaceSpec(workspacePath)` | `(*SpecWizardConfig, error)` | Spec do workspace por path |
| `GetWorkspaceDir(workspacePath)` | `string` | Diretório de trabalho do workspace |
| `GetSessionDir(sessionID)` | `string` | Diretório de trabalho da sessão |
| `SyncSpecToWorkspaceBySession(sessionID)` | `error` | Gera arquivos .spec-wizard/ + .AGENTS.md |

### 10.2 Option / Recommendation Types

```go
type Option struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
}

type Recommendation struct {
    Level       string `json:"level"` // "success" | "warning" | "critical"
    Title       string `json:"title"`
    Description string `json:"description"`
}
```

---

## 11. Fluxos

### 11.1 Criação de Workspace com Spec Wizard

```
Usuário cria workspace com Spec Wizard vinculado
  │
  └── SaveWorkspace(ws)
      ├── ws.SpecWizardID = "wizard-xxx"
      │
      ├── ensureSWMCP()
      │   └── SaveMCPServer("sw-NomeDoWorkspace", config)
      │       ├── Command: sw mcp
      │       ├── Env: WZ_PROJECT_PATH=/path/to/project
      │       └── Enabled: true
      │
      └── SyncSpecToWorkspace()
          ├── GetWizard("wizard-xxx")
          ├── os.MkdirAll(".spec-wizard/")
          ├── writeFile("config.json", ...)
          ├── writeFile("PRD.md", ...)
          ├── writeFile("skills.md", ...)
          └── writeFile(".AGENTS.md", ...)
```

### 11.2 Sessão de Chat com Spec Wizard

```
ChatPanel: onMount()
  ├── CreateSession("default-workspace", "default-worker")
  ├── GetSessionWorkspaceSpec(sessionID)
  │   ├── Session.WorkspaceID → GetWorkspace(path)
  │   ├── Workspace.SpecWizardID → GetWizard(id)
  │   └── Se erro → toast "Workspace sem Spec Wizard"
  │
  └── (continua normalmente)
```

### 11.3 Inferência de Campo

```
Usuário clica ✨ Sugerir no campo PRD
  ├── specProvider/specModel vazio? → toast erro
  ├── PRD já tem valor? → dialog "Substituir?" → Cancelar? aborta
  │
  └── App.InferField("PRD", formData)
      ├── ValidarInferencia("PRD", cfg)
      │   └── ExpertLanguagePlugin vazio? → erro bloqueante
      │
      ├── fieldContextFrom(cfg)
      ├── currentValueFrom("PRD", cfg) → "texto atual do campo"
      │
      ├── prompts.Build("PRD", ctx, "texto atual")
      │   └── buildPRD()
      │       ├── System: Elite Product Manager
      │       ├── User: SPEC NAME + EXPERT + USER'S IDEA
      │       └── Temp: 0.2, MaxTokens: 2000
      │
      └── llmFn(ctx, systemPrompt, userPrompt, 0.2, 2000)
          ├── GetFixedModel("spec") → ("openai", "gpt-4o")
          ├── ListProviders() → lookup config
          ├── llm.NewClient(...)
          ├── client.Generate(ctx, InferenceRequest{...})
          │   └── POST /chat/completions → resposta
          └── retorna string → setFieldValue("PRD", resposta)
```

---

## Stack Relacionada

| Componente | Localização |
|---|---|
| MCP server `sw` | `/home/data/aux/dev/projects/go/spec-wizard/` |
| `ada-llm-client` | `/home/data/aux/dev/projects/go/ada-llm-client/` |
| Expert Plugins | `~/.config/ada-love-ide/plugins/spec-wizard/` |
| Fixed Models (spec) | Tabela `fixed_models` em `ada_love.db` |
| DB de governança `sw` | `~/.spec-wizard/governance.db` |
