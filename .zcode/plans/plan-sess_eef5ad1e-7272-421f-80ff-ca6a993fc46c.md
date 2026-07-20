# Plano: Expert plugins em STDIO (dual-mode) — fim de portas e zumbis

## Objetivo
Trocar o transporte HTTP persistente dos expert plugins por invocação CLI via STDIO no `ada-love-ide`, mantendo **dual-mode** nos binários: quando recebem um número de porta (como o `spec-wizard` module faz) sobem o servidor HTTP; quando recebem um subcomando (`options`, `analyze`, `format`) rodam em modo CLI (JSON via stdin/stdout) e encerram. Assim o IDE não aloca portas nem deixa processos zumbis, e o outro repo continua funcionando.

## 1. Plugins — dual-mode (`go-expert/main.go`, `flutter-expert/main.go`)
- Extrair builders reutilizáveis: `buildOptions() map[string]any` (corpo atual do `handleOptions`) e `runAnalyze(input []byte)` / `runFormat(input []byte)`.
- Em `main()`, após o `mux`/handlers:
  - Se `len(os.Args) > 1` e `os.Args[1]` for **numérico** → modo HTTP naquela porta (inalterado, p/ `spec-wizard` module).
  - Se `len(os.Args) > 1` e for um action conhecido (`options`/`analyze`/`format`) → **modo CLI**: ler stdin (se houver), chamar o builder correspondente, `json.NewEncoder(os.Stdout).Encode(result)`, `os.Exit(0)`.
  - Caso contrário (sem args) → HTTP na porta padrão (compatibilidade).
- Handlers HTTP permanecem chamando os mesmos builders (zero impacto no modo HTTP).

## 2. `internal/plugins/manager.go` (ada-love-ide)
- Remover: campo `running`, `EnsureRunning`, `StopAll`, `IsHealthy`, `isHealthy`, `stopLocked`, `findFreePort`.
- Manter: `NewManager`, `List`, `FindByLanguage`.
- Novo transporte (substitui `CallExpert`):
  ```go
  func (m *PluginManager) CallExpert(plugin *ExpertPlugin, action string, input string) (map[string]interface{}, error) {
      binPath := filepath.Join(m.pluginsDir, plugin.StartCommand)
      cmd := exec.Command(binPath, action)
      cmd.Stdin = strings.NewReader(input)
      var out, errb bytes.Buffer
      cmd.Stdout = &out; cmd.Stderr = &errb
      if err := cmd.Run(); err != nil { return nil, fmt.Errorf("plugin %s falhou: %v", plugin.ID, err) }
      var result map[string]interface{}
      if err := json.Unmarshal(out.Bytes(), &result); err != nil {
          return nil, fmt.Errorf("plugin %s resposta inválida: %v", plugin.ID, err)
      }
      return result, nil
  }
  ```
- `Engine.Close()` (`engine.go`): remover `e.Plugins.StopAll()` (nada a parar); manter `e.DB.Close()`.

## 3. `internal/specwizardmgr/manager.go`
- Remover todas as chamadas `m.plugins.EnsureRunning(...)` (em `callOptions`, `aggregateOptions`, `GetStacks`).
- `callOptions`/`aggregateOptions`/`GetStacks` passam a chamar direto `m.plugins.CallExpert(plugin, "options", "")`.
- **Cache por plugin** p/ evitar 7 spawns por open do diálogo:
  - Campo `cache map[string]map[string]any` + `mu sync.Mutex`, chave = `plugin.ID`.
  - `fetchOptions(plugin *ExpertPlugin) (map[string]any, error)`: se em cache retorna; senão `CallExpert(plugin,"options","")`, armazena e retorna.
  - `callOptions(lang)` → `optionsFrom(fetchOptions(plugin), keys...)`
  - `aggregateOptions` → para cada plugin `optionsFrom(fetchOptions(plugin), keys...)`, dedup por id
  - `GetStacks(lang)` → lê `"stack_templates"` de `fetchOptions(plugin)`

## 4. Rebuild & validação dos plugins
- `go build -o plugins/go-expert/expert ./plugins/go-expert` e (no dir do módulo) `cd plugins/flutter-expert && go build -o expert .`
- Validar modo CLI: `echo '{}' | ./expert options` → JSON com os 7 catálogos; e `./expert 8083 &` + `curl localhost:8083/options` → ainda funciona (modo HTTP p/ `spec-wizard` module).
- Informar ao usuário para copiar os 2 binários para `~/.config/ada-love-ide/plugins/spec-wizard/{go-expert,flutter-expert}/expert` (mesmo fluxo de antes).

## 5. Sem mudança em frontend / bindings
As assinaturas dos métodos do `App` (`GetPatterns`, `GetStacks`, `GetStateManagement`, `GetPersistenceOptions`, etc.) continuam idênticas → nenhum ajuste em `SpecWizardDialog.svelte`, `entities.svelte.ts` ou `wailsjs`.

## Verificação
- `cd ada-love-ide && go build ./...` e `go vet ./internal/...`
- `svelte-check` (sem edições no frontend, sem novos erros)
- Manual: abrir Spec Wizard, selecionar expert → catálogos populam via STDIO; encerrar o app → `ps` não mostra processos `expert` zumbis; sem portas alocadas.

## Fora de escopo / notas
- `.codenomad/worktrees/specwizardmgr/` é worktree obsoleto → não modificado.
- `spec-wizard` module continua em HTTP (dual-mode preserva compatibilidade).
