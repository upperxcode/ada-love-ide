# 🛡️ Sistema de Modos e Grants (RBAC)

## Índice
1. [Visão Geral](#1-visão-geral)
2. [Os 6 Modos de Operação](#2-os-6-modos-de-operação)
3. [Matriz de Risco](#3-matriz-de-risco)
4. [Políticas de TTL](#4-políticas-de-ttl)
5. [Fluxo de Permissão](#5-fluxo-de-permissão)
6. [Segurança entre Modos (Downgrade)](#6-segurança-entre-modos-downgrade)
7. [Classificação de Ações](#7-classificação-de-ações)
8. [Comportamento Esperado do LLM](#8-comportamento-esperado-do-llm)
9. [Estrutura do Código](#9-estrutura-do-código)
10. [Testes](#10-testes)

---

## 1. Visão Geral

O sistema de **Grants (RBAC)** do Ada Love IDE é um mecanismo de segurança que controla quais ações o assistente de IA pode executar em cada **Modo de Operação**. Em vez de um sistema binário (permitido/negado), ele usa uma **matriz de risco** que considera:

- **Modo atual** (ASK, PLAN, EDIT, EXECUTE, FULL, ADMIN)
- **Tipo de ação** (read, write, exec, admin, etc.)
- **Nível de risco** da ação naquele modo (none, low, medium, high, critical)
- **Grants concedidos** pelo usuário (allow_once, allow_session)
- **TTL** dos grants (action, task, session, temporary, permanent)

### Princípios

1. **Sem fallback silencioso**: Se um parâmetro obrigatório está ausente, o erro deve explodir. Nunca usar `|| 'default'` ou `?? 'fallback'`.
2. **Defesa em profundidade**: O sistema de modos age em 3 camadas:
   - **Camada 1 — System Prompt**: instrui o LLM sobre o que pode fazer
   - **Camada 2 — Tool Filtering**: `allowedTools()` expõe ou oculta ferramentas
   - **Camada 3 — Permission Guard**: `PermissionGuard` intercepta cada tool call e valida
3. **Upgrade mantém grants, downgrade zera**: segurança contra escalação de privilégio.

---

## 2. Os 6 Modos de Operação

### 2.1 ASK (`ModeAsk = "ASK"`)
- **Propósito**: Consultas e leitura — sem alterações.
- **Poder**: 10 (menos poderoso)
- **Capacidades**: `read`, `search`
- **Ativação explícita**: ❌ Não
- **TTL padrão**: `session`
- **Pode override (uma vez)**: ✅ Sim
- **Permite "Sempre nesta sessão"**: ✅ Sim
- **Ações negadas**: `exec`, `exec_high_risk`, `write_project`, `write_outside`, `write_env`, `admin`, `config_edit`

**System Prompt:**
> Você é um assistente puramente informativo.
> Você pode LER arquivos e PESQUISAR (web/local) para responder perguntas.
> Você NÃO pode modificar, criar ou executar nada.
> Use as ferramentas normally. Se uma ferramenta precisar de permissão, o sistema de segurança vai pedir automaticamente. Não pergunte antes — apenas execute.

### 2.2 PLAN (`ModePlan = "PLAN"`)
- **Propósito**: Análise de dependências, criação de planos de execução, especificações técnicas, diagramas.
- **Poder**: 20
- **Capacidades**: `read`, `search`
- **Ativação explícita**: ❌ Não
- **TTL padrão**: `task` (1 hora)
- **Pode override (uma vez)**: ✅ Sim
- **Permite "Sempre nesta sessão"**: ❌ Não
- **Ações negadas**: `exec`, `exec_high_risk`, `write_project`, `write_outside`, `write_env`, `admin`, `config_edit`

**System Prompt:**
> Você é um arquiteto de software.
> Seu objetivo é ANALISAR o código, EXPLORAR arquivos e CRIAR PLANOS detalhados.
> Você NÃO executa alterações — apenas documenta o que precisa ser feito.
> Use as ferramentas normally. Se uma ferramenta precisar de permissão, o sistema de segurança vai pedir automaticamente. Não pergunte antes — apenas execute.

### 2.3 EDIT (`ModeEdit = "EDIT"`)
- **Propósito**: Edição assistida de código. Criar, modificar e aplicar diffs.
- **Poder**: 40
- **Capacidades**: `read`, `search`, `write`, `plan`
- **Ativação explícita**: ❌ Não
- **TTL padrão**: `action` (uma única ação)
- **Pode override (uma vez)**: ✅ Sim
- **Permite "Sempre nesta sessão"**: ✅ Sim
- **Ações negadas**: `exec`, `exec_high_risk`, `admin`, `config_edit`
- **Regra de confirmação**: Ações com risco > Low precisam confirmação do usuário

**System Prompt:**
> Você é um desenvolvedor editor de código.
> Você pode ler, pesquisar, editar e criar arquivos.
> COMANDOS DE TERMINAL e ESCRITA FORA DO WORKSPACE precisam de confirmação do usuário.
> Use as ferramentas normally (read, write, exec, search, plan). Se um comando precisar de permissão, o sistema de segurança vai pedir automaticamente. Não pergunte antes — apenas execute.

### 2.4 EXECUTE (`ModeExec = "EXECUTE"`)
- **Propósito**: Testes e execução controlada. Intermediário entre EDIT e FULL.
- **Poder**: 60
- **Capacidades**: `read`, `search`, `write`, `exec`, `plan`
- **Ativação explícita**: ❌ Não
- **TTL padrão**: `task` (1 hora)
- **Pode override (uma vez)**: ❌ Não
- **Permite "Sempre nesta sessão"**: ❌ Não (apenas via dialog específico)
- **Ações negadas**: `exec_high_risk`, `write_env`, `write_outside`, `admin`, `config_edit`
- **Comandos bloqueados**: `rm -rf`, `git push --force`, `sudo`, `chmod 777`, `dd`, `docker rm -f`, etc.
- **Regra de confirmação**: Apenas risco High+ precisam confirmação. Comandos seguros (`go test`, `npm test`, `ls`, `cat`, `grep`, etc.) são auto-autorizados.

**System Prompt:**
> Você é um assistente de teste e execução controlada.
> Pode LER, EDITAR arquivos e EXECUTAR uma lista restrita de comandos seguros:
> - go test / npm test / cargo check / pytest
> - go build / npm run build / cargo build
> - git status / git diff / git log
> - ls, cat, head, tail, grep, find
> COMANDOS DESTRUTIVOS (rm -rf, git push --force, sudo) são bloqueados automaticamente.
> Use as ferramentas normally. Comandos seguros serão executados automaticamente. Comandos de alto risco serão bloqueados ou pedirão confirmação. Não pergunte antes — apenas execute.

### 2.5 FULL (`ModeFull = "FULL"`)
- **Propósito**: Agente autônomo completo. Acesso total a todas as ferramentas.
- **Poder**: 80
- **Capacidades**: `read`, `search`, `write`, `exec`, `plan`
- **Ativação explícita**: ✅ Sim (precisa confirmação para entrar)
- **TTL padrão**: `temporary` (15 minutos)
- **Pode override (uma vez)**: ❌ Não
- **Permite "Sempre nesta sessão"**: ❌ Não
- **Ações negadas**: `admin`, `config_edit`
- **Regra de confirmação**: Comandos de alto risco (High) precisam confirmação extra (safety net)

**System Prompt:**
> Você é um agente autônomo completo.
> Pode planejar, explorar, editar arquivos e executar comandos no terminal.
> Todas as ferramentas estão disponíveis.
> ATENÇÃO: Comandos de alto risco (rm -rf, git push --force, sudo) exigem confirmação.
> Seja eficiente e responsável.

### 2.6 ADMIN (`ModeAdmin = "ADMIN"`)
- **Propósito**: Gerenciamento do sistema. Configurar provedores, modelos, MCPs, skills, permissões.
- **Poder**: 100 (mais poderoso)
- **Capacidades**: `read`, `search`, `write`, `exec`, `plan`, `admin`, `config`
- **Ativação explícita**: ✅ Sim (precisa confirmação para entrar — risco alto)
- **TTL padrão**: `temporary` (15 minutos)
- **Pode override (uma vez)**: ❌ Não
- **Permite "Sempre nesta sessão"**: ❌ Não
- **Ações negadas**: Nenhuma (ADMIN pode tudo)

**System Prompt:**
> Você é um administrador do sistema ADA.
> Pode gerenciar configurações de provedores, modelos, MCPs, skills e permissões.
> Ações DESTRUTIVAS (excluir provedores, resetar configurações) exigem confirmação dupla.
> CUIDADO: alterações em system prompts e API keys afetam todo o sistema.

---

## 3. Matriz de Risco

Cada ação tem um nível de risco **diferente por modo**. A matriz completa está em `DefaultRiskMatrix` no código.

| Ação | Descrição | ASK | PLAN | EDIT | EXECUTE | FULL | ADMIN |
|------|-----------|:---:|:----:|:----:|:-------:|:----:|:-----:|
| `read` | Leitura de arquivos | none | none | low | low | none | none |
| `search` | Busca no código/base | none | none | low | low | none | none |
| `write_project` | Escrita no workspace | **high** | **high** | **medium** | low | low | medium |
| `write_env` | Escrita em `.env`, chaves, CI/CD | **critical** | **critical** | **critical** | **critical** | **high** | medium |
| `write_outside` | Escrita fora do workspace | **critical** | **critical** | **critical** | **critical** | **high** | medium |
| `exec` | Execução de comandos | **critical** | **critical** | **high** | medium | low | medium |
| `exec_high_risk` | Comandos destrutivos | **critical** | **critical** | **critical** | **critical** | **high** | **critical** |
| `mkdir` | Criação de diretórios | medium | medium | low | low | low | none |
| `admin` | Alteração de config do sistema | **critical** | **critical** | **critical** | **critical** | **critical** | **high** |
| `config_edit` | Alteração de system prompts, MCPs, API keys | **critical** | **critical** | **critical** | **critical** | **critical** | **high** |
| `network` | Requisições de rede externas | medium | low | low | low | low | medium |

### Níveis de Risco

| Nível | Cor | Comportamento |
|-------|:---:|---------------|
| `none` | 🟢 | Auto-autorizado, sem confirmação |
| `low` | 🟢 | Auto-autorizado em todos os modos |
| `medium` | 🟡 | **ASK/PLAN**: pede confirmação (não é auto só porque é leitura). **EDIT**: pede confirmação. **EXECUTE**: auto para exec, pede para outros. **FULL/ADMIN**: auto |
| `high` | 🟠 | Pede confirmação (exceto ADMIN, que auto-autoriza) |
| `critical` | 🔴 | **Sempre** pede confirmação (safety net universal) |

---

## 4. Políticas de TTL

Cada grant tem um ciclo de vida definido por política:

| Política | Duração | Onde se aplica |
|----------|---------|----------------|
| `session` | 24h | Grants do modo ASK |
| `task` | 1h | Grants dos modos PLAN, EXECUTE |
| `action` | 1 chamada | Grants do modo EDIT (cada diff/arquivo) |
| `temporary` | 15min | Grants dos modos FULL, ADMIN |
| `permanent` | Nunca expira | Apenas com confirmação extra |

Grants expirados são ignorados automaticamente e limpos do banco de dados na próxima verificação.

---

## 5. Fluxo de Permissão

```
LLM quer chamar ferramenta "exec" com args {...}
    │
    ▼
1. PermissionGuard.Check(sessionID, toolName, args, mode)
    │
    ├── 2. Classifica ação e risco
    │     (ClassifyAction → ActionExec, RiskMedium)
    │
    ├── 3. Verifica session grants (allow_once)
    │     └── Se existe e é compatível com modo → ✅ ALLOW
    │
    ├── 4. Verifica persisted grants (allow_session)
    │     └── Se existe, não expirou, e modo é compatível → ✅ ALLOW
    │
    ├── 5. Verifica DeniedActions do modo
    │     └── Se ação é negada e !CanOverrideOnce → 🚫 BLOCK
    │     └── Se ação é negada e CanOverrideOnce → cria PermissionRequest
    │
    ├── 6. Verifica DeniedCommands (apenas exec)
    │     └── Se comando está na blacklist → cria PermissionRequest
    │
    ├── 7. Verifica AllowedCapabilities
    │     └── Se capacidade não é permitida e !CanOverrideOnce → 🚫 BLOCK
    │     └── Se não permitida e CanOverrideOnce → cria PermissionRequest
    │
    ├── 8. needsConfirmation(risk)
    │     ├── critical → sempre pede confirmação
    │     ├── ADMIN → auto (modo confiável)
    │     ├── FULL + high → pede confirmação
    │     ├── EXECUTE + high → pede confirmação
    │     ├── EXECUTE + medium/low → auto (comandos seguros)
    │     ├── EDIT + >low → pede confirmação
    │     ├── ASK/PLAN + none/low → auto (só leitura)
    │     └── ASK/PLAN + medium+ → pede confirmação
    │
    └── 9. ✅ ALLOW ou cria PermissionRequest + emite evento
```

### Quando o PermissionRequest é criado:

1. **Guarda bloqueia** a execução da ferramenta (select/channel)
2. **Evento `chat:permission-request`** é emitido para o frontend
3. **Dialog aparece** com tool, args, risco, modo, ação
4. **Usuário decide**: `allow_once`, `allow_session` (apenas EDIT/EXECUTE), ou `deny`
5. **Decisão é enviada** ao canal → guarda desbloqueia
6. Se `allow_once`: grant registrado com o **modo atual** no `sessionGrants`
7. Se `allow_session`: grant persistido no DB com o **modo atual** em `GrantedMode`
8. Se `deny`: retorna erro "negado pelo usuário"

---

## 6. Segurança entre Modos (Downgrade)

### Hierarquia de Poder

```
ADMIN   (100)   ← mais poderoso
FULL    (80)
EXECUTE (60)
EDIT    (40)
PLAN    (20)
ASK     (10)    ← menos poderoso
```

### Regras

| Transição | Tipo | Grants |
|-----------|------|--------|
| ASK → EDIT | ⬆️ upgrade | Mantidos |
| EDIT → FULL | ⬆️ upgrade | Mantidos |
| FULL → EDIT | ⬇️ **downgrade** | **Zerados** |
| FULL → ASK | ⬇️ **downgrade** | **Zerados** |
| EXECUTE → ADMIN | ⬆️ upgrade | Mantidos |
| ADMIN → FULL | ⬇️ **downgrade** | **Zerados** |

**Por quê?** Se você estava no FULL (poder 80) e ganhou um grant para `exec`, ao migrar para EDIT (poder 40) esse grant não pode mais valer — senão você burla a segurança do EDIT tendo um grant conquistado num modo mais poderoso.

### Implementação

1. Cada `PermissionGrant` tem `GrantedMode` — o modo em que foi concedido
2. `sessionGrants` mapeia `action → mode` (não mais `action → true`)
3. `SetCurrentMode()` compara poder do modo novo com o antigo:
   - **Upgrade ou igual** → mantém
   - **Downgrade** → `clearAllGrantsForSession()` limpa tudo
4. Evento `chat:grants-cleared` notifica o frontend
5. Log completo no terminal via `DumpGrants()`

---

## 7. Classificação de Ações

### Mapeamento Tool → Capacidade

| Tool Name | Capacidade |
|-----------|------------|
| `read`, `read_file`, `cat` | `CapRead` |
| `search`, `grep`, `find`, `explore` | `CapSearch` |
| `write`, `write_file`, `edit`, `patch`, `create` | `CapWrite` |
| `exec`, `execute`, `run`, `shell`, `terminal`, `command` | `CapExec` |
| `plan`, `planning` | `CapPlan` |
| `admin`, `admin_action`, `manage` | `CapAdmin` |
| `config`, `configure`, `settings` | `CapConfig` |
| *qualquer outro* | `CapExec` (fallback seguro) |

### Classificação de Ações (ClassifyAction)

```go
func ClassifyAction(toolName, argsJSON string, mode ChatMode) (ActionClass, RiskLevel)
```

Ações classificadas:
- `ActionRead` — leitura de arquivos
- `ActionSearch` — busca/busca
- `ActionWriteProject` — escrita dentro do workspace
- `ActionWriteEnv` — escrita em arquivos sensíveis (`.env`, `.ssh/`, `credentials`)
- `ActionWriteOutside` — escrita fora do workspace
- `ActionExec` — execução de comandos
- `ActionExecHighRisk` — comandos destrutivos (`rm -rf`, `git push --force`, `sudo`)
- `ActionMkdir` — criação de diretórios
- `ActionAdmin` — ações administrativas
- `ActionConfigEdit` — alteração de configuração da IA
- `ActionNetwork` — requisições de rede

### Comandos de Alto Risco (HighRiskPrefixes)

São detectados por prefixo (case-insensitive):
- `rm `, `rm -rf`, `rm -r`, `rm -f`
- `git push --force`, `git push -f`, `git reset --hard`
- `sudo `, `chmod 777`, `chown `
- `docker rm -f`, `docker system prune`, `docker rmi`
- `dd `, `mkfs.`, `fdisk`, `parted`
- `:(){ :|:& };:` (fork bomb)
- `> /dev/sd`, `> /dev/nvme`
- `mv /`, `cp -rf /`
- `eval `, `source /`, `. /`

### Arquivos Sensíveis (isEnvFile)

Paths que contêm:
- `.env`, `.env.`
- `.gitconfig`, `.ssh/`, `id_rsa`, `id_ed25519`
- `credentials`, `secrets`, `token`, `apikey`
- `.npmrc`, `.netrc`
- `config.json`, `settings.json`
- `ci/`, `.github/`, `gitlab-ci`

---

## 8. Comportamento Esperado do LLM

### Regra de Ouro

> **Nunca pergunte antes de executar.** Apenas chame a ferramenta. Se precisar de permissão, o sistema de segurança vai pedir automaticamente ao usuário. Se for negado, você receberá uma mensagem de erro.

### Cenários Comuns

| Cenário | Modo | Comportamento Esperado |
|---------|------|------------------------|
| "Leia o arquivo main.go" | **ASK** | Chama `read("main.go")` → ✅ auto-autorizado |
| "Leia este arquivo" | **EDIT** | Chama `read("path")` → ✅ auto-autorizado |
| "Execute go test" | **EDIT** | Chama `exec({"command":"go test"})` → 🚫 negado no EDIT. System prompt explica |
| "Execute go test" | **EXECUTE** | Chama `exec({"command":"go test"})` → ✅ auto-autorizado (comando seguro) |
| "Execute rm -rf /" | **EXECUTE** | Chama `exec(...)` → 🚫 bloqueado (DeniedCommands) |
| "Execute rm -rf /" | **FULL** | Chama `exec(...)` → ⚠️ pede confirmação (high risk) |
| "Altere a API key" | **FULL** | Chama `config(...)` → 🚫 negado (DeniedActions do FULL) |
| "Altere a API key" | **ADMIN** | Chama `config(...)` → ✅ auto-autorizado (ADMIN pode) |
| "Liste /etc/passwd" | **EDIT** | Chama `read("/etc/passwd")` → ✅ auto-autorizado (read não é bloqueado) |
| "Edite .env" | **EXECUTE** | Chama `write({"file_path":".env"})` → ⚠️ pede confirmação (write_env é critical) |

### Dicas para o LLM

- **EDIT**: `read` e `search` são sempre auto-autorizados. `write` pede confirmação por diff. `exec` bloqueado.
- **EXECUTE**: comandos seguros (`ls`, `cat`, `go test`, `npm test`, `git status`) são auto-autorizados.
- **FULL**: tudo é permitido, mas comandos destrutivos pedem confirmação extra. Seja responsável.
- **ADMIN**: você tem poder total sobre configurações. Ações destrutivas (excluir provider) pedem confirmação dupla.
- **Se um comando for bloqueado**, não insista. Explique ao usuário por que foi bloqueado e sugira alternativas.
- **Se um comando pedir confirmação**, aguarde o usuário decidir. O sistema bloqueia até a resposta.

---

## 9. Estrutura do Código

### Arquivos Principais

| Arquivo | Descrição |
|---------|-----------|
| `internal/chat/modes.go` | Definição dos modos, matriz de risco, classificação de ações, configuração por modo |
| `internal/chat/permissions.go` | PermissionStore, grants, TTL, fluxo de permissão, guarda, logging |
| `internal/chat/chat.go` | Chat principal, `NormalizeMode`, `HandleModeChange` |
| `app_sessions.go` | `SetSessionConfig` com disparo de mode change |
| `internal/adapters/adapter_multi_llm.go` | `allowedTools()`, integração com PermissionGuard |
| `frontend/src/lib/components/chat/PermissionDialog.svelte` | Dialog de permissão |
| `frontend/src/lib/components/chat/ChatPanel.svelte` | Seletor de modo no toolbar |

### Constantes Importantes

```go
// Modes
ModeAsk    = "ASK"
ModePlan   = "PLAN"
ModeEdit   = "EDIT"
ModeExec   = "EXECUTE"
ModeFull   = "FULL"
ModeAdmin  = "ADMIN"

// Risk Levels
RiskNone     = 0
RiskLow      = 1
RiskMedium   = 2
RiskHigh     = 3
RiskCritical = 4

// Power Levels
ModeAdmin  → 100
ModeFull   → 80
ModeExec   → 60
ModeEdit   → 40
ModePlan   → 20
ModeAsk    → 10
```

### Funções Exportadas Úteis

| Função | Descrição |
|--------|-----------|
| `GetModeConfig(mode)` | Retorna a config completa de um modo |
| `NormalizeMode(mode)` | Normaliza string para modo (ex: "execute" → "EXECUTE") |
| `ModePowerLevel(mode)` | Nível de poder (10-100) |
| `IsUpgrade(old, new)` | True se new ≥ old em poder |
| `ClassifyAction(tool, args, mode)` | Classifica ação + risco |
| `GetRisk(action, mode)` | Nível de risco de uma ação no modo |
| `isHighRiskCommand(cmd)` | True se comando é destrutivo |
| `isEnvFile(path)` | True se path é arquivo sensível |

---

## 10. Testes

Localização: `internal/chat/modes_test.go` e `internal/chat/permissions_test.go`

### Cobertura

- **Configuração**: cada modo tem test de config (`TestGetModeConfig_Ask`, `TestGetModeConfig_Admin`, etc.)
- **Chunk Types**: verifica se cada modo expõe os tipos corretos de streaming
- **Classificação**: todas as ações × modos na matriz são testadas
- **Comandos de alto risco**: `isHighRiskCommand` para rm -rf, git push --force, sudo, etc.
- **Arquivos sensíveis**: `isEnvFile` para .env, .ssh, credentials, etc.
- **Fluxo de permissão**: allow_once, allow_session, deny, contexto cancelado
- **Segurança**: downgrade limpa grants, upgrade mantém
- **NormalizeMode**: todos os modos + variações de case + apelidos (exec, test, config)

### Rodando os Testes

```bash
go test ./internal/chat/... -v -count=1
```
