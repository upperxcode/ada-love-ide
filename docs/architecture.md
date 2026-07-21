# 🏗️ Ada Love IDE — Architecture Reference

## Índice
1. [Visão Geral do Ecossistema](#1-visão-geral-do-ecossistema)
2. [ada-love-core](#2-ada-love-core)
3. [Module Graph](#3-module-graph)
4. [Context Composition](#4-context-composition)
5. [Spec Wizard](#5-spec-wizard)
6. [Inference Engine](#6-inference-engine)
7. [Session Isolation](#7-session-isolation)
8. [Workspace Integration](#8-workspace-integration)
9. [MCP Tools](#9-mcp-tools)
10. [Code Indexer](#10-code-indexer)
11. [Chat UI](#11-chat-ui)
12. [Sidebar](#12-sidebar)
13. [Fluxo de Mensagem](#13-fluxo-de-mensagem)
14. [Key Decisions](#14-key-decisions)

---

## 1. Visão Geral do Ecossistema

```
ada-frontend (Svelte 5)
  │ Wails IPC
  ▼
ada-love-ide (main App)
  ├── internal/adapters/    → bridges para módulos externos
  ├── internal/chat/        → SendMessage, streaming, comandos
  ├── internal/commands/    → slash commands (/help, /sync-spec, etc)
  ├── internal/db/          → Store (SQLite via ada-storage-module)
  ├── internal/engine/      → Engine (wiring de dependências)
  ├── internal/specwizardmgr/ → Spec Wizard CRUD + inferência
  └── internal/prompts/     → fábrica de prompts para inferência
      │
      ▼ (import)
ada-love-core (biblioteca)
  ├── orchestrator.go       → pipeline central
  ├── context_builder.go    → CompilePrompt
  ├── interfaces.go         → StorageEngine, LLMClient, etc
  └── types.go              → Message, Session, LLMToken

módulos externos:
  ├── ada-llm-client        → chamada HTTP ao LLM
  ├── ada-context           → compressão de histórico
  ├── ada-stream            → streaming de tokens
  ├── ada-commands          → roteador de slash commands
  ├── ada-executor          → execução de comandos sandbox
  ├── ada-llm-wiki          → busca em wiki local
  ├── ada-code-indexer      → indexador de código
  └── ada-storage-module    → SQLite storage
```

---

## 2. ada-love-core

### Package `core` — biblioteca central

**Localização:** `/home/data/aux/dev/projects/go/ada-love-core/`
**go.mod:** `ada-love-core`
**Uso:** importado via `go.mod` replace no `ada-love-ide`

### Interfaces (`interfaces.go`)

```go
type StorageEngine interface {
    GetMessagesBySession(sessionID string) ([]Message, error)
    SaveMessage(msg Message) error
    GetGreetings() ([]Greeting, error)
    DeleteMessages(sessionID string) error
    GetSession(id string) (*Session, bool)
}

type LLMClient interface {
    Generate(ctx context.Context, prompt string, model string) (<-chan LLMToken, error)
}

type Compactor interface {
    Compact(ctx context.Context, systemPrompt string, history []string, limit int) (string, error)
    CountTokens(text string) int
    CompactWithOverhead(ctx context.Context, history []Message, overheadTokens int) (string, error)
}

type Executor interface {
    ExecuteCommand(ctx context.Context, sessionID, cmd string, args []string) (string, error)
}

type Emitter interface {
    Emit(event string, data ...any)
}

type WikiArticle struct {
    Title   string
    Content string
    Tags    []string
}

type WikiSearcher interface {
    Search(ctx context.Context, query string) []WikiArticle
}
```

### Tipos (`types.go`)

```go
type Message struct {
    ID, SessionID, Role, Content, CreatedAt string
}

type LLMToken struct {
    Token        string `json:"token"`
    Done         bool   `json:"done"`
    FinishReason string `json:"finish_reason,omitempty"`
}

type Session struct {
    ID, WorkspaceID, WorkerName, Title, Summary string
    Model, Provider, Mode, Thinking             string
    Messages        []RawMessage
    CreatedAt, UpdatedAt time.Time
    Pinned              bool
}

type Greeting struct{ Patterns, Response string }
type RawMessage struct{ Role, Content, ToolCallID string; ToolCalls []any; Time time.Time }
type DecisionEvent struct{ SessionID, Reasoning, NextAgent, Task string; SubTasks int }
```

### Orchestrator (`orchestrator.go`)

```go
type Orchestrator struct {
	Storage   StorageEngine
	LLMClient LLMClient
	Compactor Compactor
	Executor  Executor
	Wiki      WikiSearcher                 // interface — ver interfaces.go
	SystemPrompt string                    // estilo de escrita do assistente
	ExtraContext func(ctx, sessionID, userInput) string  // camadas extras
}
```

Métodos:
- `ProcessMessage(ctx, sessionID, userInput, ...model)` → string
- `ProcessMessageStream(ctx, sessionID, userInput, ...model)` → `<-chan LLMToken`
- `CompilePrompt(ctx, sessionID, userInput)` → string (prompt compilado, usa Wiki.Search internamente)
- `ExecuteCommand(ctx, sessionID, cmd, args)` → string
- `SaveMessages(sessionID, userInput, response)`

### System Prompt (`context_builder.go`)

```go
stylePrompt := o.SystemPrompt
if stylePrompt == "" {
    stylePrompt = `You are Ada, an expert software engineering assistant
integrated into the Ada Love IDE. You communicate in a clear, direct,
and professional manner. You prioritize correctness, clarity, and best
practices. You write concise, well-structured code with proper error
handling. You explain your reasoning when necessary, but avoid
unnecessary verbosity. You adapt to the user's language (Portuguese or
English) based on their input.`
}
```

Pode ser sobrescrito via `orch.SystemPrompt` (ex: para usar a persona do worker configurado).

---

## 3. Module Graph

```
ada-love-ide (Wails App)
  ├── app_*.go              → thin Wails bindings
  ├── internal/chat/
  │   ├── chat.go           → Send, streaming, comandos
  │   └── emitter.go        → FrontendEmitter
  ├── internal/adapters/
  │   ├── adapter_compactor.go   → Compactor (ada-context)
  │   ├── adapter_multi_llm.go   → LLMClient (ada-llm-client)
  │   └── adapter_executor.go    → Executor (ada-commands)
  ├── internal/commands/
  │   ├── syncspec_command.go    → /sync-spec
  │   └── ... (health, clear, build, etc)
  ├── internal/db/
  │   ├── adapter.go        → StorageEngine (ada-storage-module)
  │   ├── sessions.go       → CRUD sessões, workspaces, workers
  │   └── collections.go    → Providers, Skills, SpecWizards, MCP
  ├── internal/engine/
  │   └── engine.go         → wiring, ExtraContext, code-indexer, MCP tools
  ├── internal/specwizardmgr/
  │   ├── manager.go        → CRUD, inferência, cache de plugins
  │   ├── inference.go      → validação de dependências
  │   ├── sync.go           → SyncSpecToWorkspace (gera .AGENTS.md)
  │   └── options.go        → Option, Recommendation
  └── internal/prompts/
      ├── prompt.go         → tipos base (Prompt, FieldContext)
      └── specwizard.go     → builders de prompt para inferência
```

---

## 4. Context Composition

`CompilePrompt()` em `ada-love-core/context_builder.go`:

```
1. checkStaticRoutes(userInput) → greetings
2. History → Storage.GetMessagesBySession()
3. ExtraContext → .AGENTS.md + Skills + Knowledge + Workspace dir
4. Wiki → Wiki.Search(userInput) → retorna []WikiArticle (com ContentSize pre-computado)
5. overhead = CountTokens(ExtraContext + Wiki)
6. compactedHistory = Compactor.CompactWithOverhead(history, overhead)
7. Final:
   SystemPrompt (estilo Ada | worker persona)
   + ExtraContext (workspace dir, .AGENTS.md, skills, knowledge)
   + compactedHistory
   + Wiki
   + "User: {input}\nAssistant:"
```

### ExtraContext layers (definido em `engine.go`)

```
=== WORKSPACE ===
You have read/write access to the following directory:
{workspaceDir}

=== ARCHITECTURAL GOLDEN RULES ===
{.AGENTS.md content}

=== SKILLS ===
{skill content}

=== KNOWLEDGE ===
{knowledge items (se < 2000 tokens)}
```

---

## 5. Spec Wizard

### Modelo de dados (`internal/config/specwizard/specwizard.go`)

```go
type SpecWizardConfig struct {
    ID, Name, Description              string
    ExpertLanguagePlugin                string
    PRD                                string
    FunctionalRequirements              []string
    NonFunctionalRequirements           []string
    Persistence, Architecture           string
    EngineeringPhilosophies              []string
    DesignPatterns, DataPatterns        []string
    StackConfig                         []StackItem
    Business                            Business
    Color, Icon                         string
    ArchitectureHealth                  int
    DependencyManifest                  []Dependency
    StackPlugin                         string
    CreatedAt, UpdatedAt                time.Time
}

type Business struct {
    StateManagement, APIContract        string
    CustomizationDetails                 string
    FinalAdjustments                     string
    ArchitectureRecommendations          string
}
```

### Steps do Wizard (SpecWizardDialog.svelte)

| Step | Label | Campos |
|---|---|---|
| 1 | Identity | Name, Expert Language Plugin, Domain & Scope (PRD, Functional, Non-Functional) |
| 2 | Architecture | Select Base Architecture, Persistence Strategy |
| 3 | Patterns | Engineering Philosophies, Design Patterns (GoF), Data Patterns |
| 4 | Stack | Stack Plugin, Dependency Manifest (lib/ver/mandatory) |
| 5 | Business | State Management, API Contract, Customization Details |
| 6 | Advisor | Implementation Instructions, Architecture Recommendations |

---

## 6. Inference Engine

### Validação de Dependências (`internal/specwizardmgr/inference.go`)

| Target | Pré-requisitos |
|---|---|
| `PRD` | `expert_language_plugin` |
| `FUNCTIONAL` | `expert_language_plugin` + `prd` |
| `NON_FUNCTIONAL` | `expert_language_plugin` + `prd` + `functional_requirements` |
| `API_CONTRACT` | Steps 1-4 completos (architecture, persistence, philosophies, patterns, manifest) |
| `CUSTOMIZATION` | Steps 1-4 completos |
| `FINAL_ADJUSTMENTS` | Steps 1-4 + `api_contract` + `customization_details` |

### Fábrica de Prompts (`internal/prompts/specwizard.go`)

| Builder | Temperatura | MaxTokens | Persona |
|---|---|---|---|
| `buildPRD` | 0.2 | 2000 | Elite Product Manager |
| `buildFunctional` | 0.1 | 2000 | System Analyst |
| `buildNonFunctional` | 0.1 | 2000 | Technical Architect |
| `buildAPIContract` | **0.0** | 2000 | Principal Software Architect |
| `buildCustomization` | 0.2 | 2000 | UI/UX Technical Designer |
| `buildFinalAdjustments` | 0.2 | 2000 | Senior Implementation Engineer |

Todos os prompts incluem o sufixo:
```
CRITICAL: Output must be concise, structured, and machine-readable.
This text will be consumed by another AI model — not by humans.
Avoid fluff, marketing language, adjectives, and verbose explanations.
```

### Fluxo de Inferência

```
Usuário clica ✨ Sugerir
  → App.InferField("PRD", formData)
    → ValidarInferencia("PRD", cfg) → valida dependências
    → prompts.Build("PRD", ctx, currentValue) → monta prompt
    → llmFn(ctx, system, user, 0.2, 2000) → modelo "spec"
      → GetFixedModel("spec") → provider + model
      → ada-llm-client → HTTP
    → retorna string → preenche o campo
```

---

## 7. Session Isolation

Cada sessão (chat) tem seu próprio `model`, `provider`, `mode`, `thinking` salvos no banco.

### Fluxo

1. `ChatPanel.loadSession()` carrega sessão via `GetSessionByID()`
2. Restaura `selectedModel` e `selectedMode` dos campos da sessão
3. Ao enviar mensagem, concatena `provider/model` para o backend
4. Backend roteia para o provider correto (sem fallback)

### Salvamento

- `SetSessionConfig(id, model, provider, mode, thinking)` chamado explicitamente:
  - Dentro de `loadSession()` — ao restaurar
  - No clique do modelo no dropdown
  - Em `cycleMode()` — ao alternar modo
  - No auto-select — quando modelo atual não está mais disponível

### Resolução de Provider (`adapter_multi_llm.go`)

```go
func (a *MultiLLMAdapter) resolveClient(model string) (client, resolvedModel) {
    // Model deve estar no formato "provider/model"
    // Se não encontrar o provider exato, retorna nil (erro)
}
```

Sem fallback — se o provider não existe, emite erro para o frontend.

---

## 8. Workspace Integration

### Hierarquia

```
Workspace (Path, Folders[], SpecWizardID)
  └── Worker instances (via workspace_workers join table)
       └── Sessions (chats)
```

### SpecWizardID no Workspace

`WorkspaceConfig.SpecWizardID` vincula um workspace a um Spec Wizard. Resolvido em runtime via `GetSessionWorkspaceSpec(sessionID)`.

### SyncSpecToWorkspace

Quando um workspace é salvo com `SpecWizardID`, gera:

```
workspace/
├── .spec-wizard/
│   ├── config.json    ← configuração do projeto
│   ├── PRD.md         ← Product Requirements Document
│   └── skills.md      ← Golden Rules
└── .AGENTS.md         ← regras de governança para IA
```

Lógica em `internal/specwizardmgr/sync.go`. Chamado também via `/sync-spec` command.

### MCP Server Auto-Registro

`ensureSWMCP()` em `app_workspaces.go` cria MCP server `sw-{workspace-title}` com `WZ_PROJECT_PATH` apontando para o diretório do workspace.

### Worker Management

- `AddWorkerToWorkspace(path, name)` → insere em `workspace_workers`
- `RemoveWorkerFromWorkspace(path, name)` → remove
- `ListWorkspaceWorkers(path)` → lista workers vinculados
- Delete com confirmação: se worker tem chats ativos, usuário deve digitar `DELETE`

---

## 9. MCP Tools

### Tool Definitions

Definidas em `buildToolDefs()` em `engine.go` — converte `AvailableTools()` do Store em `[]llm.ToolDefinition`.

### Injeção no Request

```go
// adapter_multi_llm.go
req := llm.InferenceRequest{
    UserPrompt: prompt,
    Config:     llm.InferenceConfig{...},
    Tools:      a.tools,  // ← injetado quando ada-llm-client suporta
}
```

### Cache

`MultiLLMAdapter.SetTools(tools)` armazena as tool definitions. Chamado no startup do Engine.

---

## 10. Code Indexer

### Inicialização (`engine.go`)

```go
codeIdxStore := codeIndexerStore.NewStore()
if workspaceDir != "" {
    go func() {
        codeIndexer.StartCrawler(workspaceDir, codeIdxStore, 10000)
    }()
}
```

- Roda em background (`go func()`)
- Indexa arquivos `.go` via AST
- Armazena em `*codeIndexerStore.Store` (memória)

### Uso no Contexto

No `ExtraContext`, quando o usuário envia uma mensagem:
```go
symbols := codeIdxStore.Search(userInput)
if len(symbols) > 0 {
    // Injeta até 10 símbolos relevantes
    // Formato: "- Nome (tipo) em path/arquivo.go:42"
}
```

### Engine Struct

```go
type Engine struct {
    // ...
    CodeIndexer *codeIndexerStore.Store
}
```

---

## 11. Chat UI

### Componentes

| Componente | Função |
|---|---|
| `ChatPanel.svelte` | Container principal: mensagens + input + toolbar |
| `MarkdownRenderer.svelte` | Renderiza markdown com `marked` + `highlight.js` |
| `Sidebar.svelte` | Lista de workspaces + workers + chats |
| `ChatLayout.svelte` | Layout: sidebar + settings + chat panel |

### MarkdownRenderer

- Usa `marked` para parsear markdown
- Usa `highlight.js` para syntax highlighting (tema `github-dark`)
- Estilos via `@tailwindcss/typography` com variáveis CSS do tema
- Font monospace: `Geist Mono Variable`

### Auto-scroll

- 3 `$effect`s monitoram: `messages.length`, `streamingContent`, `isLoading`
- Usa `tick()` do Svelte para aguardar DOM antes de scrollar
- Usuário pode scrollar para cima para pausar auto-scroll
- Auto-scroll reativa quando usuário volta ao final

### Eventos de Streaming

| Evento | Quando | Ação |
|---|---|---|
| `chat:delta` | A cada token | Atualiza `streamingContent` + última mensagem |
| `chat:turnEnd` | Fim da resposta | Finaliza streaming, salva mensagens |
| `chat:error` | Erro | Toast de erro + mensagem de erro no chat |
| `chat:commandResult` | Resultado de comando | Exibe no painel de comando |

---

## 12. Sidebar

### Estrutura

```
ADA LOVE (header)
  Workspaces
    ├── Workspace 1       [✏️] [➕]
    │   ├── Worker A      [➕] [🗑️]
    │   │   ├── chat1
    │   │   └── chat2
    │   └── Worker B      [➕] [🗑️]
    │       └── chat1
    └── Workspace 2       [✏️] [➕]
        └── ...
  Settings ⚙️
```

### Funcionalidades

| Ação | Comportamento |
|---|---|
| Clique no workspace | Ativa workspace, cria sessão inicial |
| ✏️ (pencil) | Abre settings na edição do workspace |
| ➕ (na linha do workspace) | Abre popover para adicionar workers |
| ➕ (na linha do worker) | Cria novo chat (`chat1`, `chat2`...) |
| 🗑️ (na linha do worker) | Remove worker (se tiver chats → confirmar DELETE) |
| 📌 (pin verde) | Chat fixado, sempre visível |
| 📌 (pin vermelho, hover) | Pin disponível para fixar |
| 🗑️ (no chat) | Deleta chat |

### Ordenação de Chats

- Fixados (pinned) primeiro, ordem alfabética
- Não-fixados depois, ordem alfabética

---

## 13. Fluxo de Mensagem

```
User digita e envia
  │
  ▼
ChatPanel.handleSend()
  ├── message não vazia? && !isLoading && sessionID?
  │
  ├── modelo = provider + "/" + modelName
  │
  ├── SendMessage(sessionID, text, modelString, "normal", mode)
  │   │
  │   ▼
  │   App.SendMessage()
  │   │
  │   ▼
  │   Chat.Send()
  │   ├── "/comando" → router.Execute() → resposta direta
  │   ├── CheckStaticResponse(text) → saudação? → resposta sem LLM
  │   │
  │   └── orch.CompilePrompt(ctx, sessionID, text)
  │       ├── Storage.GetMessagesBySession()
  │       ├── ExtraContext() → .AGENTS.md + skills + workspace dir
  │       ├── Wiki.Search() (se configurado)
  │       ├── Compactor.CompactWithOverhead()
  │       └── Monta prompt: SystemPrompt + layers + history + user
  │
  └── streamingClient.GenerateStream(ctx, sessionID, prompt, model)
      │
      ▼
      MultiLLMAdapter.GenerateStream()
      ├── resolveClient(model) → provider "lm-studio"
      ├── ada-llm-client.GenerateStream()
      │   └── POST {baseURL}/chat/completions {messages, tools, stream}
      │
      ├── token → StreamToEvents → emit("chat:delta") → frontend
      └── finish → emit("chat:turnEnd") → frontend
```

---

## 14. Key Decisions

### Por que `ada-love-core` como biblioteca separada?

- Separação clara entre core (orquestrador, interfaces, tipos) e shell (Wails, UI)
- Testável independentemente
- Pode ser usado por outros frontends (CLI, servidor HTTP)

### Por que ExtraContext como callback?

- Evita que o core dependa de tipos específicos do `ada-love-ide` (SkillManager, db.Store)
- Engine pode injetar qualquer contexto sem modificar o core
- Fácil de estender: só adicionar mais camadas no callback

### Por que `provider/model` no lugar de model só?

- Roteamento explícito para o provider correto
- Sem fallback silencioso — se o provider não existe, erro claro
- Cada sessão salva `model` + `provider` separadamente

## Context Window Dinâmico

### Problema Original

O `contextLimit` no frontend e o `maxTokens` do compactor eram hardcoded (262K e 8000 respectivamente), sem refletir o modelo real da sessão.

### Solução

```
ModelSettings.ContextSize → persiste no banco → App.GetSessionContextInfo → frontend
     │
     ├── CompilePrompt(sessionID, text, contextSize)
     │     └── Compactor.CompactWithBudget(history, overhead, contextSize)
     │
     └── CompactorAdapter resolve o budget dinamicamente
```

### Fluxo de Dados

1. **Provider config** (`internal/config/provider/provider.go`): `ModelSettings.ContextSize` e `MaxTokens` já existiam mas não eram persistidos
2. **Storage** (`ada-storage-module`): migration v40 adiciona colunas `context_size` / `max_tokens` à tabela `provider_models`
3. **Collections** (`internal/db/collections.go`): `SaveProvider` e `adaptProviderToInternal` agora persistem/leem os campos
4. **Engine** (`internal/engine/engine.go`): `GetSessionContextInfo(sessionID)` → busca o model settings da sessão e estima tokens usados
5. **App binding** (`app_sessions.go`): `GetSessionContextInfo` exposto ao frontend
6. **Frontend** (`ChatPanel.svelte`): substitui constantes hardcoded por chamada a `GetSessionContextInfo()`
7. **CompilePrompt** (`ada-love-core/context_builder.go`): aceita `contextSize` opcional, usa `CompactWithBudget` quando fornecido

### Interface Compactor

```go
type Compactor interface {
    Compact(ctx, systemPrompt, history, limit)  // original
    CountTokens(text) int
    CompactWithOverhead(ctx, history, overhead)  // legacy, delega para CompactWithBudget
    CompactWithBudget(ctx, history, overhead, maxContextSize)  // novo
}
```

`CompactWithBudget` foi adicionada em `ada-context/context.go` e aceita `maxContextSize` explícito, substituindo o `cc.config.MaxTokens` hardcoded.

### Por que WikiSearcher como interface exportada?

- Substitui o antigo `Wiki any` + type assertion privada que era frágil (nunca casava com `[]wiki.Article` ≠ `[]wikiArticle`)
- Permite que qualquer módulo externo implemente a interface sem depender de `ada-love-core`
- A ponte é feita via adapter em `ada-love-ide/internal/adapters/adapter_wiki.go`
- O campo `ContentSize` no `Article` da wiki permite estimativas rápidas de token sem chamar o tokenizer

### Por que sem fallback no resolveClient?

- Usuário escolheu um modelo específico para o chat
- Se o provider não está mais disponível, melhor alertar do que enviar para outro
- Configuração de providers muda em Settings > Models

### Por que o streaming path vai pelo Orchestrator?

- Garante que todo o contexto é compilado (histórico, wiki, skills, .AGENTS.md)
- Evita duplicação de lógica entre streaming e não-streaming
- Mensagens sempre têm o mesmo formato independente do path
