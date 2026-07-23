# Workers

Workers represent AI personalities/behaviors. Each worker has a `connection_type` that determines how it communicates with the backend.

## Connection Types

| Type | Description | Backend Package | Handlers |
|---|---|---|---|
| `ada` | Native Ada engine | — | `SendMessage` (`app_chat.go`) |
| `cli` | Subprocess CLI (opencode run, etc.) | `internal/cliworker/` | `SendCLIMessage` |
| `url` | Generic HTTP API | `internal/urlworker/` | `SendURLMessage` |
| `opencode_serve` | OpenCode Server (session-based API) | `internal/urlworker/` | `SendOpenCodeMessage` |

## Backend Architecture

### `internal/cliworker/`

Purpose: Execute CLI tools as subprocesses and stream output.

**worker.go** — `Runtime` struct:
```go
type Runtime struct {
    Config       worker.WorkerConfig
    WorkspaceDir string
}
```

Key methods:
- `BuildCommand(message, model string) Command` — builds the full CLI command
- `BuildListModelsCommand() Command` — builds command to list models

The command is built as: `{command} {arguments} --model {model} "{message}"`

**executor.go** — Command execution:
- `Execute(cmd Command) (string, error)` — runs command, returns output
- `ExecuteStream(cmd Command, emit EmitterFn) (string, error)` — runs with live SSE events (`chat:delta`, `chat:thinking`)
- `ListModels(rt *Runtime) ([]ModelInfo, error)` — runs `{command} {models_command}` and parses `provider/model` per line

**Model discovery**: runs `{command} {models_command}` (default: `models`). Parses `provider/model` lines.

**Encoding**: `Arguments` and `ModelsCommand` are stored as `{args}\n{models_command}` in the DB `arguments` column.

### `internal/urlworker/`

Purpose: Send HTTP requests to API servers with SSE streaming support.

**worker.go** — `Runtime` struct:
```go
type Runtime struct {
    Config   worker.WorkerConfig
    BaseURL  string
    URLPaths worker.URLPaths
}
```

Key methods:
- `BuildChatRequest(message, model string) Request` — builds HTTP request with configurable body template
- `BuildModelsRequest() Request` — builds model listing request

The chat body uses Go templates with variables `{{.Message}}`, `{{.Model}}`, `{{.Stream}}`.

**executor.go** — HTTP execution:
- `ExecuteChat(req Request, emit EmitterFn) (string, error)` — sends HTTP request with optional SSE streaming
- `FetchModels(req Request) ([]ModelInfo, error)` — fetches and parses model list from API
- `parseProvidersArrayJSON` / `parseProvidersJSON` / `parseModelLines` — three parsers for different response formats

**opencode.go** — OpenCode Server specific:
- `OpenCodeCreateSession() (string, error)` — `POST /session`
- `OpenCodeSendMessage(sessionID, message, model, emit) (string, error)` — async message via SSE:
  1. Connects to `GET /event` SSE stream
  2. Sends `POST /session/{id}/prompt_async`
  3. Reads `message.part.delta` events:
     - `type: "reasoning"` → `chat:thinking` (cadeia de pensamento)
     - `type: "text"` → `chat:delta` (texto da resposta)
  4. Completes on `message.updated` with `finish: "stop"` or `session.status` idle
- `OpenCodeListModels() ([]ModelInfo, error)` — `GET /config/providers`
- `OpenCodeSessionManager` — in-memory map `adaSessionID → opencodeSessionID`

**server.go** — Process manager for URL workers:
- `ServerManager.Start(cfg Runtime) error` — starts server in background
- `ServerManager.Stop(workerName string) error` — kills server process
- `ServerManager.IsRunning(workerName string) bool` — health check via TCP + HTTP
- `ServerManager.Status(workerName string) (running, port, baseURL, uptime)`

### `internal/config/worker/worker.go`

**WorkerConfig** struct:
```go
type WorkerConfig struct {
    ID               int64   `json:"id"`
    Name             string  `json:"name"`
    Persona          string  `json:"persona"`
    Language         string  `json:"language"`
    Icon             string  `json:"icon"`
    Color            string  `json:"color"`
    ConnectionType   string  `json:"connection_type"`   // ada | cli | url | opencode_serve
    ConnectionName   string  `json:"connection_name"`
    Command          string  `json:"command"`            // CLI: binary path, URL/OC: base URL
    Arguments        string  `json:"arguments"`          // CLI: flags, URL: JSON config
    ModelsCommand    string  `json:"models_command"`     // CLI: list command, URL/OC: models URL
    Environment      string  `json:"environment"`        // JSON env vars or auth headers
    InheritFolders   bool    `json:"inherit_folders"`
    InheritKnowledge bool    `json:"inherit_knowledge"`
    InheritSkills    bool    `json:"inherit_skills"`
    InheritTools     bool    `json:"inherit_tools"`
    InheritPersona   bool    `json:"inherit_persona"`
}
```

**URLPaths** struct (stored as JSON in `Arguments` for URL/OC workers):
```go
type URLPaths struct {
    ChatPath         string `json:"chat_path"`          // default: /v1/chat/completions
    ChatBodyTemplate string `json:"chat_body_template"` // Go template with {{.Message}}, {{.Model}}, {{.Stream}}
    ModelsPath       string `json:"models_path"`        // default: /config/providers
    ModelsFormat     string `json:"models_format"`      // providers_obj | providers_arr | json_array | flat
    Stream           bool   `json:"stream"`              // SSE streaming enabled
    StartCommand     string `json:"start_command"`       // command to start the server
}
```

**Encoding:** The `Arguments` field stores different data per connection type:
- **CLI**: `{run_args}\n{models_command}` (newline separated)
- **URL/OC**: `{"chat_path":"...","chat_body_template":"...","models_path":"...","models_format":"...","stream":true,"start_command":"..."}` (JSON)

### `internal/db/sessions.go`

Key functions:
- `PutWorker(wc WorkerConfig)` — saves worker, calls `wc.EncodeArguments()` to serialize
- `adaptWorkerToInternal(w *storage.Worker) WorkerConfig` — loads worker, calls `DecodeArguments()` to parse CLI encoding, or `DecodeURLPaths()` for URL/OC
- `connectionNameForType(ct string) string` — maps types to display names

## Frontend Architecture

### Entity Store (`frontend/src/lib/stores/entities.svelte.ts`)

**FIELD_CONFIGS.workers** — defines form fields per connection type:
- Fields with `cliOnly: true` appear only when `connection_type === 'cli'`
- Fields with `urlOnly: true` appear only when `connection_type === 'url'`
- Fields with `opencodeServeOnly: true` appear only when `connection_type === 'opencode_serve'`

**saveWorker(data)** — encodes type-specific fields before saving:
- `url` → encodes `chat_path`, `chat_body_template`, `models_url`, `models_format`, `stream_enabled`, `start_command` into `arguments` JSON
- `opencode_serve` → encodes `start_command` into `arguments` JSON
- `cli` → preserves `models_command`

**toCardData(raw)** — decodes `arguments` JSON back into flat fields for URL/OC workers.

### ChatPanel (`frontend/src/lib/components/chat/ChatPanel.svelte`)

**Worker detection** in `loadSession`:
1. Gets `sess.worker_name` from loaded session
2. Fetches workers via `GetWorkers()`
3. Matches by `name` to find `connection_type`
4. Sets `currentWorkerType` and `currentWorkerName`

**Reactive adapters**:
- `availableModes` — derived from `currentWorkerType`:
  - `cli`/`url`/`opencode_serve`: only `['ASK']`
  - `ada`: all 6 modes
- `isNonAdaWorker` — `$derived` for conditional rendering
- `allModels` — returns CLI/API-specific models for non-Ada workers, providerStore models for Ada
- Model-fetching `$effect` — calls `GetCLIModels`, `GetURLModels`, or `GetOpenCodeModels` based on type
- Model selector badge: `CLI`, `API`, or `OC`

**handleSend** — dispatches to the correct backend:
- `cli` → `SendCLIMessage`
- `url` → `SendURLMessage`
- `opencode_serve` → `SendOpenCodeMessage`
- `ada` → `SendMessage`

The message is persisted with **both** user message and assistant response for all non-Ada workers.

### EntityEditDialog (`frontend/src/lib/components/settings/EntityEditDialog.svelte`)

- `opencodeServeOnly` field type for conditional visibility
- Collapsible `<details>` section for URL/OC configuration with server controls (Start/Stop/Status)

## OpenCode Server API Integration

### SSE Event Format

```
data: {"id":"evt_...","type":"server.connected","properties":{...}}
data: {"id":"evt_...","type":"session.updated","properties":{...}}
data: {"id":"evt_...","type":"session.status","properties":{"status":{"type":"busy"}}}
data: {"id":"evt_...","type":"message.updated","properties":{...}}
data: {"id":"evt_...","type":"message.part.updated","properties":{"part":{"id":"prt_...","type":"reasoning|text","text":"..."}}}
data: {"id":"evt_...","type":"message.part.delta","properties":{"partID":"prt_...","field":"text","delta":"chunk"}}
```

### Endpoints Used

| Endpoint | Method | Purpose |
|---|---|---|
| `/session` | POST | Create session |
| `/session/{id}/prompt_async` | POST | Send message async |
| `/event` | GET | SSE event stream |
| `/config/providers` | GET | List models |

### Response Parsing

Models are parsed from `/config/providers` response. The parser extracts `id`, `providerID`, and model name from each model entry. Three formats supported:

- **providers_obj**: `{"provider": {"name": {"models": {"modelName": {...}}}}}`
- **providers_arr**: `{"providers": [{"name":"X","models":{"modelName":{"id":"x/y","providerID":"x"}}}]}`
- **flat**: lines of `provider/model`

## Adding a New Worker Type

1. **Go**: Add case to `connectionNameForType()` in `internal/db/sessions.go`
2. **Go**: Create runtime & executor package in `internal/{type}worker/`
3. **Go**: Add App handler methods in `app_sessions.go`
4. **Frontend**: Add connection type option in `FIELD_CONFIGS.workers` in `entities.svelte.ts`
5. **Frontend**: Add `{type}Only` field config + visibility in `EntityEditDialog.svelte`
6. **Frontend**: Wire up in `ChatPanel.svelte` — `availableModes`, `isNonAdaWorker`, model fetching, `handleSend`
7. **Bindings**: Add TypeScript declarations in `App.d.ts` and JS functions in `App.js`
