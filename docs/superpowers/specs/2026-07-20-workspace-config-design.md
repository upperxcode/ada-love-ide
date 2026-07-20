# Design: Configuração do Workspace (formulário + sincronia de storage)

**Data:** 2026-07-20
**Status:** Aprovado

## 1. Contexto

O app já persiste workspaces através do módulo `ada-storage-module`. O `internal/db.Store`
usa `storage.WorkspaceStore` + stores de junção (`folders`, `knowledge`, `skills`, `tools`,
`agents`, `workers`). As funções `upsertWorkspace` / `adaptWorkspaceToInternal` já fazem o
round-trip de todos os campos entre o `workspace.WorkspaceConfig` (interno) e o storage.

O objetivo deste trabalho é:
1. Configurar o formulário de Workspace na UI com os campos corretos, **removendo** `path`
   (auto-derivado do primeiro item de Folders) e `enabled` (sempre `true`).
2. Verificar se o `ada-storage-module` está em sincronia com as tabelas do banco e corrigir
   a lacuna encontrada.

## 2. Decisões aprovadas

- **Max Prompts Send = INTEGER (int).** O DDL (`max_prompt_send INTEGER`) e o struct Go
  (`MaxPromptSend int`) permanecem; o input numérico não aceita decimais. Sem migration de
  alteração de tipo.
- **Formulário = `WorkspaceDialog.svelte` dedicado**, espelhando o padrão do `SpecWizardDialog`
  (nosso padrão para entidades complexas). Não se estende o `EntityEditDialog` genérico.
- **Tools (tools do sistema)** são populados a partir de `GetAvailableTools()` (router commands),
  confirmado pelo usuário.

## 3. Verificação do ada-storage (sincronia com as tabelas)

Comparando o SQL fornecido com `ada-storage-module/storage/migrations_*.go`:

| Tabela | Situação |
|--------|----------|
| `workspaces` | ✅ corresponde |
| `workspace_agents` | ✅ corresponde |
| `workspace_folders` | ✅ corresponde |
| `workspace_skills` | ✅ corresponde |
| `workspace_tools` | ✅ corresponde |
| `workspace_workers` | ✅ corresponde |
| `workspace_templates` | ✅ corresponde (`migrations_other.go`) |
| `workspace_knowledge` | ❌ **FALTANDO** — o store `workspace_knowledge_store.go` executa `INSERT INTO workspace_knowledge …`, mas nenhuma migration cria a tabela, então os itens de conhecimento falham silenciosamente ao persistir hoje. |

### Correção
- Adicionar a constante `workspaceKnowledgeTable` (idêntica ao DDL fornecido) em
  `ada-storage-module/storage/migrations_workspaces.go` (ou `migrations_other.go`).
- Registrá-la como a próxima versão de migration (após a atual máxima `v38` → **`v39`**) em
  `ada-storage-module/storage/migrations.go`.

```sql
CREATE TABLE IF NOT EXISTS workspace_knowledge (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workspace_id INTEGER NOT NULL,
    knowledge_item TEXT NOT NULL,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE ON UPDATE NO ACTION
);
```

Bancos existentes (em `v38`) executarão `v39` e criarão a tabela; bancos novos a recebem em
sequência. Nenhuma coluna existente é alterada.

## 4. Frontend: `WorkspaceDialog.svelte`

Novo componente em `frontend/src/lib/components/settings/WorkspaceDialog.svelte`, com a mesma
estrutura do `SpecWizardDialog` (`EntityHeader` + `SettingRow` + `Collapsible` + `ThemedSelect` +
`ExpandableTextarea` + `Switch`).

### Mapa de campos

| # | Campo (label) | Tipo widget | Bind (WorkspaceConfig) | Origem / comportamento |
|---|---------------|-------------|------------------------|------------------------|
| 1 | Name | text | `title` | input texto |
| 2 | Description | text | `description` | input texto |
| 3 | Personality | textarea | `personality` | `ExpandableTextarea` |
| 4 | Routing Rules | textarea | `routing_rules` | `ExpandableTextarea` |
| 5 | Max Prompts Send | number (int) | `max_prompt_send` | input numérico, sem decimais |
| 6 | Max Context Length | int | `max_context_length` | input numérico |
| 7 | Commit Changes | bool | `commit_changes` | `Switch` |
| 8 | Folders | lista colapsável | `folders` | "Add Folder" → `App.OpenDirectoryDialog()`; lista + remover. **O 1º item define `path`** |
| 9 | Knowledge Items | lista colapsável | `knowledge` | "Add File" → `App.OpenFileDialog()`; lista + remover |
| 10 | Agents | lista colapsável (multiselect) | `agents` | checkbox list de `GetAgents()` (nomes) |
| 11 | Skills | lista colapsável (multiselect) | `skills` | checkbox list de `GetSkills()` (nomes) |
| 12 | Tools | lista colapsável (multiselect) | `tools` | checkbox list de `GetAvailableTools()` (nomes) |
| 13 | Spec Wizard | select | `spec_wizard_id` | `ThemedSelect` de `GetSpecWizards()` (value=id, label=name) |

**Removidos do formulário:** `path` (auto-derivado: `path = folders[0]`) e `enabled` (sempre
enviado como `true`).

### Comportamento de `path`
Sempre que a lista `folders` muda, `path = folders[0] ?? ''`. O fallback de `path` vazio no
backend (`SaveWorkspace` gera `workspace-<timestamp>`) permanece como rede de segurança.

## 5. Wiring (integração)

- `frontend/src/lib/stores/entities.svelte.ts`:
  - Adicionar `GetAvailableTools(): Promise<any[]>` na interface `WailsApp`.
  - Adicionar `spec_wizard_id?: string` na interface `WorkspaceConfig`.
  - Remover a entrada `workspaces` de `FIELD_CONFIGS` (o `EntityEditDialog` genérico não trata
    mais workspaces; `WorkspaceDialog` o substitui).
- `frontend/src/lib/components/settings/SettingsPanel.svelte`:
  - Adicionar branch `workspaces` que renderiza `<WorkspaceDialog>` (mesmo formato do branch
    `spec-wizard` existente), passando `entity`, `onSave`, `bind:open`.
- `app_workspaces.go` (backend): `SaveWorkspace(ws)` já encaminha `workspace.WorkspaceConfig`
  para `upsertWorkspace`; nenhuma mudança de assinatura. `path` vem de `formData.path` (definido
  pelo diálogo); o fallback de path vazio no backend permanece.

## 6. Fluxo de dados no save

1. Diálogo coleta campos → `formData`.
2. Antes de `onSave`: `formData.path = folders[0] ?? ''` e `formData.enabled = true`.
3. `entityStore.saveWorkspace` → `SaveWorkspace` → `upsertWorkspace` grava a linha principal +
   todas as linhas de junção (`folders`, `knowledge`, `skills`, `tools`, `agents`). Após o
   conserto da migration, `knowledge` passa a persistir de fato.

## 7. Testes / verificação

- `go build ./...` (app principal) e compilação do `ada-storage-module`; rodar os testes
  existentes do módulo ada-storage.
- Frontend: `npm run check` / build.
- Manual: Settings → Workspaces → New → adicionar uma pasta (confirmar que `path` autopreenche)
  → adicionar um arquivo de conhecimento → selecionar Agents/Skills/Tools/Spec Wizard → Save →
  reabrir → confirmar round-trip de todos os valores (especialmente `knowledge`, que hoje não
  persiste).

## 8. Fora de escopo (YAGNI)

- Alteração de tipo de `max_prompt_send` para float/REAL.
- Refactor da camada de persistência (já está em ada-storage).
- Novos endpoints de backend além dos já existentes (`GetAvailableTools`, `GetAgents`,
  `GetSkills`, `GetSpecWizards`, `OpenDirectoryDialog`, `OpenFileDialog`).
- Edição de `enabled` em qualquer lugar (sempre `true`).
