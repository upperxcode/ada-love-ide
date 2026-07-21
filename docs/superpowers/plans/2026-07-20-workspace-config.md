# Workspace Configuration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Configurar o formulário de Workspace na UI (13 campos, `path` auto-derivado, `enabled` sempre `true`) e sincronizar o `ada-storage-module` adicionando a tabela `workspace_knowledge` que estava faltando.

**Architecture:** O backend já persiste workspaces via `ada-storage-module` (`storage.WorkspaceStore` + stores de junção); só falta a migration da tabela `workspace_knowledge`. A UI ganha um `WorkspaceDialog.svelte` dedicado (espelhando `SpecWizardDialog`) ligado no `SettingsPanel`, consumindo endpoints já existentes (`GetAgents`, `GetSkills`, `GetAvailableTools`, `GetSpecWizards`, `OpenDirectoryDialog`, `OpenFileDialog`).

**Tech Stack:** Go + `ada-storage-module` (SQLite/mattn-sqlite3); Svelte 5 (runes `$state`/`$props`/`$effect`), Tailwind, componentes `ui/dialog`, `ui/collapsible`, `ui/Select`, `ui/switch`, `ExpandableTextarea`, `EntityHeader`, `SettingRow`; Wails bindings.

---

## File Structure

- **Create** `ada-storage-module/storage/migrations_workspace_knowledge_test.go` — teste determinístico (DB temporário) que falha porque a tabela não existe.
- **Modify** `ada-storage-module/storage/migrations_workspaces.go` — adiciona a constante `workspaceKnowledgeTable`.
- **Modify** `ada-storage-module/storage/migrations.go` — registra a migration `v39`.
- **Modify** `frontend/src/lib/stores/entities.svelte.ts` — adiciona `GetAvailableTools()` à interface, `spec_wizard_id` ao `WorkspaceConfig`, remove a entrada `workspaces` de `FIELD_CONFIGS`.
- **Create** `frontend/src/lib/components/settings/WorkspaceDialog.svelte` — o novo formulário.
- **Modify** `frontend/src/lib/components/settings/SettingsPanel.svelte` — branch `workspaces` renderizando `WorkspaceDialog`.
- **Modify** `frontend/src/lib/components/settings/SettingsPanel.svelte` (import) — importa `WorkspaceDialog`.

Nenhuma mudança em `app_workspaces.go` (o `upsertWorkspace` já grava todos os campos/junções).

---

### Task 1: ada-storage — migration `v39` para `workspace_knowledge` (TDD)

**Files:**
- Create: `ada-storage-module/storage/migrations_workspace_knowledge_test.go`
- Modify: `ada-storage-module/storage/migrations_workspaces.go`
- Modify: `ada-storage-module/storage/migrations.go`

- [ ] **Step 1: Escrever o teste que falha (tabela ausente)**

Crie `ada-storage-module/storage/migrations_workspace_knowledge_test.go`:

```go
package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestWorkspaceKnowledge_TableCreatedByMigration garante que rodar todas as
// migrations cria a tabela workspace_knowledge (faltava antes da v39).
func TestWorkspaceKnowledge_TableCreatedByMigration(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "migration_check.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(context.Background(), db); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}

	var name string
	err = db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='workspace_knowledge'",
	).Scan(&name)
	if err != nil {
		t.Fatalf("workspace_knowledge table missing after migrations: %v", err)
	}
	if name != "workspace_knowledge" {
		t.Fatalf("unexpected table name: %s", name)
	}
}
```

- [ ] **Step 2: Rodar o teste para confirmar que falha**

Run: `cd /home/data/aux/dev/projects/go/ada-storage-module && go test ./storage/ -run TestWorkspaceKnowledge_TableCreatedByMigration -v`
Expected: FAIL — `workspace_knowledge table missing after migrations: sql: no rows in result set`

- [ ] **Step 3: Adicionar a constante da tabela**

Em `ada-storage-module/storage/migrations_workspaces.go`, dentro do `const ( ... )` existente, adicione antes do `)` final:

```go
	// workspaceKnowledgeTable creates the workspace_knowledge junction table
	workspaceKnowledgeTable = `
	CREATE TABLE IF NOT EXISTS workspace_knowledge (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		knowledge_item TEXT NOT NULL,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE ON UPDATE NO ACTION
	);
	`
```

- [ ] **Step 4: Registrar a migration `v39`**

Em `ada-storage-module/storage/migrations.go`, na slice de migrations, após `{38, addStrategyToProviders},` adicione:

```go
		{39, workspaceKnowledgeTable},
```

- [ ] **Step 5: Rodar o teste para confirmar que passa**

Run: `cd /home/data/aux/dev/projects/go/ada-storage-module && go test ./storage/ -run TestWorkspaceKnowledge_TableCreatedByMigration -v`
Expected: PASS

- [ ] **Step 6: Rodar a suíte de testes do módulo e o build**

Run: `cd /home/data/aux/dev/projects/go/ada-storage-module && go test ./storage/ 2>&1 | tail -20 && go build ./...`
Expected: testes do storage passam (incluindo `TestWorkspaceKnowledge_Create`/`_DeleteAll` que antes falhavam por tabela ausente) e build ok.

- [ ] **Step 7: Commit (repo ada-storage-module)**

```bash
cd /home/data/aux/dev/projects/go/ada-storage-module && git add storage/migrations_workspace_knowledge_test.go storage/migrations_workspaces.go storage/migrations.go && git commit -m "feat(storage): add workspace_knowledge migration (v39)"
```

---

### Task 2: `entities.svelte.ts` — interface, `spec_wizard_id` e limpeza de `FIELD_CONFIGS`

**Files:**
- Modify: `frontend/src/lib/stores/entities.svelte.ts`

- [ ] **Step 1: Adicionar `GetAvailableTools()` na interface `WailsApp`**

Após a linha `GetSpecWizards(): Promise<any[]>;` (por volta da linha 33), adicione:

```ts
		GetAvailableTools(): Promise<any[]>;
```

- [ ] **Step 2: Adicionar `spec_wizard_id` na interface `WorkspaceConfig`**

Na interface `WorkspaceConfig` (termina com `agents: string[];`), adicione ao final:

```ts
		spec_wizard_id?: string;
```

- [ ] **Step 3: Remover a entrada `workspaces` de `FIELD_CONFIGS`**

Substitua o bloco:

```ts
		workspaces: [
			{ key: 'title', label: 'Title', description: 'Display name of the workspace', type: 'text', required: true, placeholder: 'Workspace title' },
			{ key: 'path', label: 'Path', description: 'Local filesystem path', type: 'text', placeholder: '/path/to/project' },
			{ key: 'description', label: 'Description', description: 'Brief summary of the project', type: 'text', placeholder: 'Workspace description' },
			{ key: 'personality', label: 'Personality', description: 'Custom traits for AI in this context', type: 'textarea', placeholder: 'AI personality traits', fullWidth: true, expandable: true },
			{ key: 'enabled', label: 'Enabled', description: 'Enable or disable this workspace', type: 'toggle' },
		],
```

por (remoção completa — o `WorkspaceDialog` passa a tratar workspaces):

```ts
		// workspaces são editados via WorkspaceDialog.svelte (não via EntityEditDialog)
```

- [ ] **Step 4: Verificar tipo (type-check do frontend)**

Run: `cd /home/data/aux/dev/projects/go/ada-love-ide/frontend && npm run check 2>&1 | tail -20`
Expected: sem erros de tipo referentes a `entities.svelte.ts`. (Se `npm run check` não existir, use `npm run build`.)

- [ ] **Step 5: Commit**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ide && git add frontend/src/lib/stores/entities.svelte.ts && git commit -m "refactor(frontend): prepara entities p/ WorkspaceDialog (GetAvailableTools, spec_wizard_id, remove FIELD_CONFIGS workspaces)"
```

---

### Task 3: `WorkspaceDialog.svelte` — novo componente

**Files:**
- Create: `frontend/src/lib/components/settings/WorkspaceDialog.svelte`

- [ ] **Step 1: Criar o componente completo**

Crie `frontend/src/lib/components/settings/WorkspaceDialog.svelte` com o conteúdo abaixo:

```svelte
<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import {
		Dialog,
		DialogPortal,
		DialogContent,
	} from '$lib/components/ui/dialog';
	import { Switch } from '$lib/components/ui/switch';
	import ThemedSelect from '$lib/components/ui/Select.svelte';
	import ExpandableTextarea from '$lib/components/ui/ExpandableTextarea.svelte';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import EntityHeader from './EntityHeader.svelte';
	import SettingRow from './SettingRow.svelte';

	interface WorkspaceDialogProps {
		open: boolean;
		onOpenChange: (open: boolean) => void;
		entity: Record<string, any> | null;
		onSave: (data: Record<string, any>) => void;
	}

	let { open = $bindable(), onOpenChange, entity, onSave }: WorkspaceDialogProps = $props();

	let formData = $state<Record<string, any>>({});
	let agents = $state<any[]>([]);
	let skills = $state<any[]>([]);
	let tools = $state<any[]>([]);
	let specWizards = $state<any[]>([]);

	let foldersOpen = $state(true);
	let knowledgeOpen = $state(false);
	let agentsOpen = $state(false);
	let skillsOpen = $state(false);
	let toolsOpen = $state(false);

	const inputBase =
		'rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]';

	function getApp(): any {
		return (window as any).go?.main?.App ?? {};
	}

	$effect(() => {
		if (open) {
			if (entity) {
				formData = { ...entity };
			} else {
				formData = {
					title: '',
					description: '',
					personality: '',
					routing_rules: '',
					max_prompt_send: 0,
					max_context_length: 0,
					commit_changes: true,
					folders: [],
					knowledge: [],
					agents: [],
					skills: [],
					tools: [],
					spec_wizard_id: '',
					color: '#3f3f46',
					icon: '📁',
					enabled: true,
					path: '',
				};
			}
			loadCandidates();
		}
	});

	async function loadCandidates() {
		const app = getApp();
		try {
			const [a, s, t, w] = await Promise.all([
				app.GetAgents ? app.GetAgents() : Promise.resolve([]),
				app.GetSkills ? app.GetSkills() : Promise.resolve([]),
				app.GetAvailableTools ? app.GetAvailableTools() : Promise.resolve([]),
				app.GetSpecWizards ? app.GetSpecWizards() : Promise.resolve([]),
			]);
			agents = a ?? [];
			skills = s ?? [];
			tools = t ?? [];
			specWizards = w ?? [];
		} catch (e) {
			console.error('[WorkspaceDialog] Failed to load candidates:', e);
		}
	}

	function updateField(key: string, value: any) {
		formData = { ...formData, [key]: value };
	}

	async function addFolder() {
		try {
			const path = await getApp().OpenDirectoryDialog();
			if (!path) return;
			const folders = [...(formData.folders ?? []), path];
			formData = { ...formData, folders, path: folders[0] };
		} catch (e) {
			console.error('[WorkspaceDialog] OpenDirectoryDialog failed:', e);
		}
	}

	function removeFolder(idx: number) {
		const folders = (formData.folders ?? []).filter((_: any, i: number) => i !== idx);
		formData = { ...formData, folders, path: folders[0] ?? '' };
	}

	async function addKnowledge() {
		try {
			const path = await getApp().OpenFileDialog();
			if (!path) return;
			const knowledge = [...(formData.knowledge ?? []), path];
			formData = { ...formData, knowledge };
		} catch (e) {
			console.error('[WorkspaceDialog] OpenFileDialog failed:', e);
		}
	}

	function removeKnowledge(idx: number) {
		const knowledge = (formData.knowledge ?? []).filter((_: any, i: number) => i !== idx);
		formData = { ...formData, knowledge };
	}

	function toggleInList(key: string, value: string) {
		const list: string[] = formData[key] ?? [];
		const next = list.includes(value) ? list.filter((v) => v !== value) : [...list, value];
		formData = { ...formData, [key]: next };
	}

	function handleSave() {
		const data = {
			...formData,
			path: (formData.folders && formData.folders[0]) || formData.path || '',
			enabled: true,
		};
		onSave(data);
	}
</script>

<Dialog bind:open onOpenChange={onOpenChange}>
	<DialogPortal>
		<DialogContent class="sm:max-w-[640px] max-h-[85dvh] flex flex-col p-0 overflow-hidden" showCloseButton={false}>
			<EntityHeader
				icon={formData.icon ?? '📁'}
				color={formData.color ?? '#3f3f46'}
				entityType="Workspace"
				isNew={!entity}
				onIconChange={(emoji: string) => updateField('icon', emoji)}
				onColorChange={(c: string) => updateField('color', c)}
				onClose={() => onOpenChange(false)}
			/>

			<div class="flex-1 overflow-y-auto">
				<!-- 1. Name -->
				<div class="px-5 py-3">
					<SettingRow label="Name" description="Display name of the workspace">
						<input
							type="text"
							value={formData.title ?? ''}
							oninput={(e: Event) => updateField('title', (e.target as HTMLInputElement).value)}
							placeholder="Workspace name"
							class={cn(inputBase, 'w-full')}
						/>
					</SettingRow>
				</div>

				<!-- 2. Description -->
				<div class="px-5 py-3">
					<SettingRow label="Description" description="Brief summary of the project">
						<input
							type="text"
							value={formData.description ?? ''}
							oninput={(e: Event) => updateField('description', (e.target as HTMLInputElement).value)}
							placeholder="Workspace description"
							class={cn(inputBase, 'w-full')}
						/>
					</SettingRow>
				</div>

				<!-- 3. Personality (full width) -->
				<div class="px-5 pb-3">
					<ExpandableTextarea
						id="personality"
						label="Personality"
						value={formData.personality ?? ''}
						oninput={(v: string) => updateField('personality', v)}
						placeholder="AI personality traits"
						minHeight={80}
						class="w-full"
						textareaClass="w-full"
					/>
				</div>

				<!-- 4. Routing Rules (full width) -->
				<div class="px-5 pb-3">
					<ExpandableTextarea
						id="routing_rules"
						label="Routing Rules"
						value={formData.routing_rules ?? ''}
						oninput={(v: string) => updateField('routing_rules', v)}
						placeholder="Rules that decide how requests are routed"
						minHeight={80}
						class="w-full"
						textareaClass="w-full"
					/>
				</div>

				<!-- 5. Max Prompts Send (int) -->
				<div class="px-5 py-3">
					<SettingRow label="Max Prompts Send" description="Maximum number of prompts to send">
						<input
							type="text"
							inputmode="numeric"
							value={formData.max_prompt_send ?? 0}
							oninput={(e: Event) => {
								const raw = (e.target as HTMLInputElement).value.replace(/[^0-9]/g, '');
								updateField('max_prompt_send', raw === '' ? 0 : parseInt(raw, 10));
							}}
							class={cn(inputBase, 'w-24 text-right font-mono')}
						/>
					</SettingRow>
				</div>

				<!-- 6. Max Context Length (int) -->
				<div class="px-5 py-3">
					<SettingRow label="Max Context Length" description="Maximum context length in tokens">
						<input
							type="text"
							inputmode="numeric"
							value={formData.max_context_length ?? 0}
							oninput={(e: Event) => {
								const raw = (e.target as HTMLInputElement).value.replace(/[^0-9]/g, '');
								updateField('max_context_length', raw === '' ? 0 : parseInt(raw, 10));
							}}
							class={cn(inputBase, 'w-24 text-right font-mono')}
						/>
					</SettingRow>
				</div>

				<!-- 7. Commit Changes -->
				<div class="px-5 py-3">
					<SettingRow label="Commit Changes" description="Automatically commit changes to the workspace">
						<Switch
							checked={!!formData.commit_changes}
							onCheckedChange={(v: boolean) => updateField('commit_changes', v)}
						/>
					</SettingRow>
				</div>

				<!-- 8. Folders (collapsible, native dir picker) -->
				<div class="px-5 pb-3">
					<Collapsible.Root bind:open={foldersOpen} class="w-full bg-[var(--bg-secondary)] rounded-xl border border-[var(--border-primary)] p-4 shadow-sm">
						<Collapsible.Trigger>
							{#snippet child({ props })}
								<button {...props} class="flex w-full items-center justify-between transition-colors cursor-pointer group">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 rounded-lg bg-[var(--accent-primary)]/10 flex items-center justify-center text-[var(--accent-primary)]">
											<Icon name="folder" size={14} />
										</div>
										<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Folders</h3>
									</div>
									<Icon name="chevron-down" size={16} class={cn('transition-transform duration-300', foldersOpen && 'rotate-180')} color="var(--text-faint)" />
								</button>
							{/snippet}
						</Collapsible.Trigger>
						<Collapsible.Content class="pt-4 flex flex-col gap-3">
							<button type="button" onclick={addFolder} class="flex items-center gap-2 px-3 py-2 rounded-lg border border-[var(--border-primary)] bg-[var(--surface-input)] text-[12px] font-medium cursor-pointer hover:bg-[var(--surface-hover)] w-fit">
								<Icon name="folder-plus" size={14} /> Add Folder
							</button>
							{#each formData.folders ?? [] as folder, idx (folder)}
								<div class="flex items-center justify-between gap-2 px-3 py-2 rounded-lg bg-[var(--surface-input)] border border-[var(--border-primary)]">
									<span class="text-[12px] truncate">{folder}</span>
									<button type="button" onclick={() => removeFolder(idx)} class="text-[var(--text-muted)] hover:text-red-500 cursor-pointer">
										<Icon name="trash" size={14} />
									</button>
								</div>
							{/each}
						</Collapsible.Content>
					</Collapsible.Root>
				</div>

				<!-- 9. Knowledge Items (collapsible, native file picker) -->
				<div class="px-5 pb-3">
					<Collapsible.Root bind:open={knowledgeOpen} class="w-full bg-[var(--bg-secondary)] rounded-xl border border-[var(--border-primary)] p-4 shadow-sm">
						<Collapsible.Trigger>
							{#snippet child({ props })}
								<button {...props} class="flex w-full items-center justify-between transition-colors cursor-pointer group">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 rounded-lg bg-[var(--accent-primary)]/10 flex items-center justify-center text-[var(--accent-primary)]">
											<Icon name="file" size={14} />
										</div>
										<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Knowledge Items</h3>
									</div>
									<Icon name="chevron-down" size={16} class={cn('transition-transform duration-300', knowledgeOpen && 'rotate-180')} color="var(--text-faint)" />
								</button>
							{/snippet}
						</Collapsible.Trigger>
						<Collapsible.Content class="pt-4 flex flex-col gap-3">
							<button type="button" onclick={addKnowledge} class="flex items-center gap-2 px-3 py-2 rounded-lg border border-[var(--border-primary)] bg-[var(--surface-input)] text-[12px] font-medium cursor-pointer hover:bg-[var(--surface-hover)] w-fit">
								<Icon name="file-plus" size={14} /> Add File
							</button>
							{#each formData.knowledge ?? [] as item, idx (item)}
								<div class="flex items-center justify-between gap-2 px-3 py-2 rounded-lg bg-[var(--surface-input)] border border-[var(--border-primary)]">
									<span class="text-[12px] truncate">{item}</span>
									<button type="button" onclick={() => removeKnowledge(idx)} class="text-[var(--text-muted)] hover:text-red-500 cursor-pointer">
										<Icon name="trash" size={14} />
									</button>
								</div>
							{/each}
						</Collapsible.Content>
					</Collapsible.Root>
				</div>

				<!-- 10. Agents (collapsible multiselect) -->
				<div class="px-5 pb-3">
					<Collapsible.Root bind:open={agentsOpen} class="w-full bg-[var(--bg-secondary)] rounded-xl border border-[var(--border-primary)] p-4 shadow-sm">
						<Collapsible.Trigger>
							{#snippet child({ props })}
								<button {...props} class="flex w-full items-center justify-between transition-colors cursor-pointer group">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 rounded-lg bg-[var(--accent-primary)]/10 flex items-center justify-center text-[var(--accent-primary)]">
											<Icon name="bot" size={14} />
										</div>
										<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Agents</h3>
									</div>
									<Icon name="chevron-down" size={16} class={cn('transition-transform duration-300', agentsOpen && 'rotate-180')} color="var(--text-faint)" />
								</button>
							{/snippet}
						</Collapsible.Trigger>
						<Collapsible.Content class="pt-4 flex flex-col gap-1">
							{#each agents as agent (agent.name)}
								<label class="flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-[var(--surface-hover)] cursor-pointer">
									<input type="checkbox" checked={(formData.agents ?? []).includes(agent.name)} onchange={() => toggleInList('agents', agent.name)} class="accent-[var(--accent-primary)]" />
									<span class="text-[13px]">{agent.name}</span>
								</label>
							{/each}
						</Collapsible.Content>
					</Collapsible.Root>
				</div>

				<!-- 11. Skills (collapsible multiselect) -->
				<div class="px-5 pb-3">
					<Collapsible.Root bind:open={skillsOpen} class="w-full bg-[var(--bg-secondary)] rounded-xl border border-[var(--border-primary)] p-4 shadow-sm">
						<Collapsible.Trigger>
							{#snippet child({ props })}
								<button {...props} class="flex w-full items-center justify-between transition-colors cursor-pointer group">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 rounded-lg bg-[var(--accent-primary)]/10 flex items-center justify-center text-[var(--accent-primary)]">
											<Icon name="sparkles" size={14} />
										</div>
										<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Skills</h3>
									</div>
									<Icon name="chevron-down" size={16} class={cn('transition-transform duration-300', skillsOpen && 'rotate-180')} color="var(--text-faint)" />
								</button>
							{/snippet}
						</Collapsible.Trigger>
						<Collapsible.Content class="pt-4 flex flex-col gap-1">
							{#each skills as skill (skill.name)}
								<label class="flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-[var(--surface-hover)] cursor-pointer">
									<input type="checkbox" checked={(formData.skills ?? []).includes(skill.name)} onchange={() => toggleInList('skills', skill.name)} class="accent-[var(--accent-primary)]" />
									<span class="text-[13px]">{skill.name}</span>
								</label>
							{/each}
						</Collapsible.Content>
					</Collapsible.Root>
				</div>

				<!-- 12. Tools (collapsible multiselect, system tools) -->
				<div class="px-5 pb-3">
					<Collapsible.Root bind:open={toolsOpen} class="w-full bg-[var(--bg-secondary)] rounded-xl border border-[var(--border-primary)] p-4 shadow-sm">
						<Collapsible.Trigger>
							{#snippet child({ props })}
								<button {...props} class="flex w-full items-center justify-between transition-colors cursor-pointer group">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 rounded-lg bg-[var(--accent-primary)]/10 flex items-center justify-center text-[var(--accent-primary)]">
											<Icon name="wrench" size={14} />
										</div>
										<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Tools</h3>
									</div>
									<Icon name="chevron-down" size={16} class={cn('transition-transform duration-300', toolsOpen && 'rotate-180')} color="var(--text-faint)" />
								</button>
							{/snippet}
						</Collapsible.Trigger>
						<Collapsible.Content class="pt-4 flex flex-col gap-1">
							{#each tools as tool (tool.Name)}
								<label class="flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-[var(--surface-hover)] cursor-pointer">
									<input type="checkbox" checked={(formData.tools ?? []).includes(tool.Name)} onchange={() => toggleInList('tools', tool.Name)} class="accent-[var(--accent-primary)]" />
									<span class="text-[13px]">{tool.Name}</span>
								</label>
							{/each}
						</Collapsible.Content>
					</Collapsible.Root>
				</div>

				<!-- 13. Spec Wizard (select) -->
				<div class="px-5 py-3">
					<SettingRow label="Spec Wizard" description="Associated Spec Wizard">
						<ThemedSelect
							value={formData.spec_wizard_id ?? ''}
							onValueChange={(v: string) => updateField('spec_wizard_id', v)}
							options={specWizards.map((s: any) => ({ value: s.id, label: s.name }))}
							placeholder="Select a Spec Wizard"
							class="w-52"
						/>
					</SettingRow>
				</div>
			</div>

			<!-- Footer -->
			<div class="flex items-center justify-between px-5 py-3 border-t border-[var(--border-primary)] bg-[var(--surface-elevated)]">
				<div></div>
				<div class="flex items-center gap-2">
					<button
						type="button"
						onclick={() => onOpenChange(false)}
						class="flex items-center px-4 py-2 rounded-lg text-[11px] font-medium cursor-pointer transition-colors hover:bg-[var(--surface-hover)]"
						style="color: var(--text-muted)"
					>
						Cancel
					</button>
					<button
						type="button"
						onclick={handleSave}
						class="flex items-center px-4 py-2 rounded-lg text-[11px] font-semibold cursor-pointer transition-all hover:brightness-110 active:scale-[0.97]"
						style="background-color: var(--accent-primary); color: var(--accent-primary-fg)"
					>
						{entity ? 'Save' : 'Create'}
					</button>
				</div>
			</div>
		</DialogContent>
	</DialogPortal>
</Dialog>
```

- [ ] **Step 2: Verificar build do frontend (compilação do novo componente)**

Run: `cd /home/data/aux/dev/projects/go/ada-love-ide/frontend && npm run build 2>&1 | tail -25`
Expected: build conclui sem erros de Svelte/TS referente a `WorkspaceDialog.svelte`. (Se `npm run build` não existir, use `npm run check`.)

- [ ] **Step 3: Commit**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ide && git add frontend/src/lib/components/settings/WorkspaceDialog.svelte && git commit -m "feat(frontend): novo WorkspaceDialog com 13 campos e listas colapsáveis"
```

---

### Task 4: `SettingsPanel.svelte` — ligar o `WorkspaceDialog`

**Files:**
- Modify: `frontend/src/lib/components/settings/SettingsPanel.svelte`

- [ ] **Step 1: Adicionar o import**

Após a linha `import SpecWizardDialog from './SpecWizardDialog.svelte';` (linha 7), adicione:

```ts
import WorkspaceDialog from './WorkspaceDialog.svelte';
```

- [ ] **Step 2: Substituir o bloco de diálogo para incluir o branch `workspaces`**

Substitua:

```svelte
				{#if activeCategory !== 'general' && FIELD_CONFIGS[activeCategory]}
					{#if activeCategory === 'spec-wizard'}
						<SpecWizardDialog
							bind:open={dialogOpen}
							onOpenChange={(val) => (dialogOpen = val)}
							entity={dialogEntity}
							onSave={handleSave}
						/>
					{:else}
						<EntityEditDialog
							bind:open={dialogOpen}
							onOpenChange={(val) => (dialogOpen = val)}
							entity={dialogEntity}
							entityType={categories.find((c) => c.id === activeCategory)?.label ?? 'item'}
							fields={FIELD_CONFIGS[activeCategory]}
							onSave={handleSave}
						/>
					{/if}
				{/if}
```

por:

```svelte
				{#if activeCategory !== 'general'}
					{#if activeCategory === 'spec-wizard'}
						<SpecWizardDialog
							bind:open={dialogOpen}
							onOpenChange={(val) => (dialogOpen = val)}
							entity={dialogEntity}
							onSave={handleSave}
						/>
					{:else if activeCategory === 'workspaces'}
						<WorkspaceDialog
							bind:open={dialogOpen}
							onOpenChange={(val) => (dialogOpen = val)}
							entity={dialogEntity}
							onSave={handleSave}
						/>
					{:else if FIELD_CONFIGS[activeCategory]}
						<EntityEditDialog
							bind:open={dialogOpen}
							onOpenChange={(val) => (dialogOpen = val)}
							entity={dialogEntity}
							entityType={categories.find((c) => c.id === activeCategory)?.label ?? 'item'}
							fields={FIELD_CONFIGS[activeCategory]}
							onSave={handleSave}
						/>
					{/if}
				{/if}
```

- [ ] **Step 3: Verificar build do frontend**

Run: `cd /home/data/aux/dev/projects/go/ada-love-ide/frontend && npm run build 2>&1 | tail -25`
Expected: build conclui sem erros; `WorkspaceDialog` está ligado em `SettingsPanel`.

- [ ] **Step 4: Commit**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ide && git add frontend/src/lib/components/settings/SettingsPanel.svelte && git commit -m "feat(frontend): liga WorkspaceDialog no SettingsPanel"
```

---

### Task 5: Verificação final (build + round-trip)

**Files:**
- (nenhum novo; validação cruzada)

- [ ] **Step 1: Build do backend principal (Go)**

Run: `cd /home/data/aux/dev/projects/go/ada-love-ide && go build ./... 2>&1 | tail -20`
Expected: compila sem erros (confirma que `app_workspaces.go` continua compatível — nenhuma mudança necessária).

- [ ] **Step 2: Build do ada-storage + testes**

Run: `cd /home/data/aux/dev/projects/go/ada-storage-module && go build ./... && go test ./storage/ 2>&1 | tail -20`
Expected: build ok e todos os testes do storage passam (incluindo `workspace_knowledge`).

- [ ] **Step 3: Build do frontend**

Run: `cd /home/data/aux/dev/projects/go/ada-love-ide/frontend && npm run build 2>&1 | tail -20`
Expected: build ok.

- [ ] **Step 4: Teste manual (round-trip)**

1. `wails dev` (ou app compilado) → abrir Settings → Workspaces.
2. Clicar em **New**: confirmar que não há campo `Path` nem `Enabled`.
3. Preencher **Name**, **Description**, **Personality**, **Routing Rules**.
4. **Folders** → Add Folder → escolher um diretório; confirmar que o `path` interno passa a ser esse diretório (verificável ao reabrir/salvar).
5. **Knowledge Items** → Add File → escolher um arquivo.
6. Marcar alguns **Agents**, **Skills**, **Tools** e selecionar um **Spec Wizard**.
7. Salvar → reabrir o workspace → confirmar que **todos** os campos voltaram, especialmente **Knowledge Items** (que antes não persistia por falta da tabela).

---

## Self-Review (checklist interno)

1. **Cobertura do spec:** (1) remover `path` do form → Task 3 não renderiza path e deriva em `handleSave`; (2) remover `enabled` → não renderizado, enviado `true`; (3) 13 campos → Task 3; (4) Folders via `OpenDirectoryDialog` → Task 3; (5) Knowledge via `OpenFileDialog` → Task 3; (6) Agents/Skills/Tools multiselect → Task 3; (7) Spec Wizard select → Task 3; (8) verificar/sincronizar ada-storage → Task 1 (workspace_knowledge v39). Tudo coberto.
2. **Sem placeholders:** todo step tem código/conteúdo completo; verificações usam comandos reais (`go test`, `go build`, `npm run build`).
3. **Consistência de tipos/nomes:** `spec_wizard_id` definido na interface (Task 2) e usado no diálogo (Task 3); `GetAvailableTools` declarado na interface (Task 2) e chamado no diálogo (Task 3); nomes de campos (`title`, `folders`, `knowledge`, `agents`, `skills`, `tools`, `max_prompt_send`, `max_context_length`, `commit_changes`, `routing_rules`, `personality`) batem entre `workspace.WorkspaceConfig`, `upsertWorkspace` e o diálogo. `path`/`enabled` tratados como derivados/em `true` em todo lugar.
