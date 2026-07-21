<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import {
		Dialog,
		DialogPortal,
		DialogContent,
		DialogOverlay,
		DialogDescription,
		DialogFooter,
		DialogHeader,
		DialogTitle,
	} from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Switch } from '$lib/components/ui/switch';
	import ThemedSelect from '$lib/components/ui/Select.svelte';
	import ExpandableTextarea from '$lib/components/ui/ExpandableTextarea.svelte';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import EntityHeader from './EntityHeader.svelte';
	import SettingRow from './SettingRow.svelte';
	import { toastStore } from '$lib/stores/toast.svelte';

	interface WorkspaceDialogProps {
		open: boolean;
		onOpenChange: (open: boolean) => void;
		entity: Record<string, any> | null;
		onSave: (data: Record<string, any>) => void;
	}

	let { open = $bindable(), onOpenChange, entity, onSave }: WorkspaceDialogProps = $props();

	// ── Form State ──
	let formData = $state<Record<string, any>>({
		name: '',
		description: '',
		path: '',
		folders: [],
		knowledge_files: [],
		agents: [],
		skills: [],
		tools: [],
		spec_wizard_id: '',
		enabled: true,
		color: '#3b82f6',
		icon: '🏢',
	});

	// ── Option catalogs from backend ──
	let agents = $state<any[]>([]);
	let skills = $state<any[]>([]);
	let tools = $state<any[]>([]);
	let specWizards = $state<any[]>([]);

	// ── Collapsible state ──
	let foldersOpen = $state(false);
	let knowledgeFilesOpen = $state(false);
		let agentsOpen = $state(false);
		let skillsOpen = $state(false);
		let toolsOpen = $state(false);

	// ── New item state (removed: knowledge files only via Browse) ──

	// ── Initialize when opening ──
	let isInitialized = $state(false);

	$effect(() => {
		if (open && !isInitialized) {
			isInitialized = true;
			initializeForm();
		} else if (!open) {
			isInitialized = false;
		}
	});

	async function initializeForm() {
		// Load candidates from backend
		await loadCandidates();

		if (entity) {
			// Normalize: backend uses 'knowledge', UI uses 'knowledge_files'
			const knowledgeFiles = entity.knowledge_files ?? entity.knowledge ?? [];
			formData = {
				...entity,
				folders: entity.folders ?? [],
				knowledge_files: Array.isArray(knowledgeFiles) ? knowledgeFiles : [],
				agents: entity.agents ?? [],
				skills: entity.skills ?? [],
				tools: entity.tools ?? [],
				spec_wizard_id: entity.spec_wizard_id ?? '',
			};
			// Derive path from folders[0] if not already set
			if ((!formData.path || formData.path === '') && formData.folders.length > 0) {
				formData.path = formData.folders[0];
			}
		} else {
			formData = {
				name: '',
				description: '',
				path: '',
				folders: [],
				knowledge_files: [],
				agents: [],
				skills: [],
				tools: [],
				spec_wizard_id: '',
				enabled: true,
				color: '#3b82f6',
				icon: '🏢',
			};
		}
	}

	async function loadCandidates() {
		try {
			agents = await (window as any).go.main.App.GetAgents();
		} catch (e) {
			console.error('Failed to load agents:', e);
		}

		try {
			skills = await (window as any).go.main.App.GetSkills();
		} catch (e) {
			console.error('Failed to load skills:', e);
		}

		try {
			tools = await (window as any).go.main.App.GetAvailableTools();
		} catch (e) {
			console.error('Failed to load tools:', e);
		}

		try {
			specWizards = await (window as any).go.main.App.GetSpecWizards();
		} catch (e) {
			console.error('Failed to load spec wizards:', e);
		}
	}

	function removeFolder(index: number) {
			formData.folders = formData.folders.filter((_: unknown, idx: number) => idx !== index);
			// Re-derive path: keep current primary if still present, else first folder
			if (formData.folders.length > 0) {
				if (!formData.folders.includes(formData.path)) {
					formData.path = formData.folders[0];
				}
			} else {
				formData.path = '';
			}
		}

		function setPrimaryFolder(index: number) {
			formData.path = formData.folders[index];
		}

	function removeKnowledgeFile(index: number) {
		formData.knowledge_files = formData.knowledge_files.filter((_: unknown, idx: number) => idx !== index);
	}

	// Toggle a value in a list field (agents/skills/tools by name)
	function toggleInList(key: 'agents' | 'skills' | 'tools', value: string) {
		const list: string[] = formData[key] ?? [];
		formData[key] = list.includes(value)
			? list.filter((v) => v !== value)
			: [...list, value];
	}

async function handleOpenDirectory() {
			try {
				const result = await (window as any).go.main.App.OpenDirectoryDialog();
				if (result) {
					const list: string[] = formData.folders ?? [];
					// Avoid duplicates
					if (list.includes(result)) return;
					formData.folders = [...list, result];
					// Becomes primary only if none is set yet
					if (!formData.path) {
						formData.path = result;
					}
				}
			} catch (e) {
				console.error('Failed to open directory dialog:', e);
			}
		}

	async function handleOpenFile() {
			try {
				const result = await (window as any).go.main.App.OpenFileDialog();
				if (result) {
					const list: string[] = formData.knowledge_files ?? [];
					// Avoid duplicates
					if (list.includes(result)) return;
					formData.knowledge_files = [...list, result];
				}
			} catch (e) {
				console.error('Failed to open file dialog:', e);
			}
		}

	function handleSave() {
		// Force enabled = true
		formData.enabled = true;
		// Ensure path is set: keep selected primary, fallback to first folder
		if (formData.folders && formData.folders.length > 0) {
			if (!formData.path || !formData.folders.includes(formData.path)) {
				formData.path = formData.folders[0];
			}
		} else {
			formData.path = '';
		}
		// Normalize id: frontend may use a synthetic string ID (toCardData),
		// backend expects an int (0 means "create new").
		let idNum = 0;
		if (typeof formData.id === 'number') {
			idNum = formData.id;
		} else if (typeof formData.id === 'string') {
			const parsed = parseInt(formData.id, 10);
			idNum = isNaN(parsed) ? 0 : parsed;
		}
		// Map UI field name back to backend field name
		const payload = { ...formData, id: idNum, knowledge: formData.knowledge_files ?? [] };
		delete payload.knowledge_files;
		onSave(payload);
		onOpenChange(false);
	}
</script>

<Dialog bind:open onOpenChange={onOpenChange}>
	<DialogPortal>
		<DialogOverlay class="z-[60] bg-black/40 backdrop-blur-sm" />
<DialogContent
				class="z-[70] sm:max-w-[720px] w-[95vw] h-[90dvh] p-0 overflow-hidden flex flex-col bg-[var(--bg-tertiary)] rounded-2xl border border-[var(--border-primary)] shadow-2xl"
				showCloseButton={false}
			>
			<EntityHeader
				icon={formData.icon}
				color={formData.color}
				entityType="Workspace"
				isNew={!entity}
				onIconChange={(emoji) => (formData.icon = emoji)}
				onColorChange={(c) => (formData.color = c)}
				onClose={() => onOpenChange(false)}
			/>

			<!-- ── Form Content ── -->
			<div class="flex-1 overflow-y-auto px-10 py-8 bg-[var(--bg-tertiary)]">
				<div class="flex flex-col gap-2">
						<!-- Field 1: Name -->
						<SettingRow label="Name" description="Display name of the workspace" required>
							<input
								bind:value={formData.name}
								placeholder="Workspace name"
								class="w-[26rem] rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]"
							/>
						</SettingRow>

						<!-- Field 2: Description -->
						<SettingRow label="Description" description="Brief summary of the workspace">
							<input
								bind:value={formData.description}
								placeholder="Brief description of the workspace..."
								class="w-[26rem] rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]"
							/>
						</SettingRow>

						<!-- Field 3: Path (auto-derived) -->
						<SettingRow label="Path" description="Auto-derived from first folder">
							<input
								bind:value={formData.path}
								placeholder="/path/to/workspace"
								class="w-[26rem] rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)] cursor-not-allowed font-mono text-[13px]"
								readOnly
							/>
						</SettingRow>

						<!-- Field 4: Spec Wizard (single select) -->
						<SettingRow label="Spec Wizard" description="Associated Spec Wizard (single)">
							<ThemedSelect
								value={formData.spec_wizard_id ?? ''}
								onValueChange={(v: string) => (formData.spec_wizard_id = v)}
								options={specWizards.map((w: any) => ({ value: w.id, label: w.name || w.id }))}
								placeholder="Select a Spec Wizard"
								class="w-[26rem]"
							/>
						</SettingRow>

					<!-- Field 5: Folders (Collapsible) -->
					<div class="mt-2 border-t border-[var(--border-primary)] pt-4 flex flex-col gap-3">
						<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
						<div
							role="button"
							tabindex="0"
							onclick={() => (foldersOpen = !foldersOpen)}
							onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); foldersOpen = !foldersOpen; } }}
							class="flex w-full items-center justify-between rounded-xl bg-[var(--bg-secondary)] border border-[var(--border-primary)] p-4 transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
						>
							<div class="flex items-center gap-3">
								<div class="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-500">
									<Icon name="folder" size={14} />
								</div>
								<div class="flex flex-col">
									<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Folders</h3>
									<span class="text-[10px] text-[var(--text-faint)]">{formData.folders?.length ?? 0} folder{(formData.folders?.length ?? 0) === 1 ? '' : 's'}</span>
								</div>
							</div>
							<div class="flex items-center gap-2">
								{#if foldersOpen}
									<!-- Browse button inline in header (only when expanded) -->
									<button
										type="button"
										onclick={(e) => { e.stopPropagation(); handleOpenDirectory(); }}
										class="flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-lg text-[11px] font-bold uppercase tracking-[0.15em] transition-all cursor-pointer bg-[var(--accent-primary)] text-white shadow hover:brightness-110 active:scale-95"
									>
										<Icon name="folder-plus" size={12} /> Browse
									</button>
								{/if}
								<Icon
									name="chevron-down"
									size={16}
									class={cn('transition-transform duration-300', foldersOpen && 'rotate-180')}
									color="var(--text-faint)"
								/>
							</div>
						</div>

						{#if foldersOpen}
							<div class="flex flex-col gap-2 pt-1">
								{#if formData.folders && formData.folders.length > 0}
									<div class="flex flex-col gap-1.5">
										{#each formData.folders as folder, i (folder)}
											{@const isPrimary = formData.path === folder}
											<div
												class="flex items-center gap-2.5 rounded-lg bg-[var(--surface-input)] border px-3 py-2 transition-colors {isPrimary ? 'border-[var(--accent-primary)]/50' : 'border-[var(--border-primary)]'}"
											>
												<!-- Primary indicator (clickable) -->
												<button
													type="button"
													onclick={() => setPrimaryFolder(i)}
													title={isPrimary ? 'Primary (defines path)' : 'Set as primary'}
													class="flex items-center justify-center w-5 h-5 rounded-full transition-all cursor-pointer shrink-0 {isPrimary ? 'bg-[var(--accent-primary)] text-white shadow-md' : 'bg-transparent border-2 border-[var(--border-primary)] text-transparent hover:border-[var(--accent-primary)]'}"
												>
													<Icon name="check" size={11} />
												</button>

												<span class="flex-1 text-[12px] font-mono text-[var(--text-primary)] truncate" title={folder}>{folder}</span>

												<button
													type="button"
													onclick={() => removeFolder(i)}
													class="text-[var(--text-faint)] hover:text-red-500 p-1 transition-colors cursor-pointer shrink-0"
													title="Remove folder"
												>
													<Icon name="x" size={14} />
												</button>
											</div>
										{/each}
										<p class="text-[10px] text-[var(--text-faint)] px-1 pt-0.5">
											Click the circle to set which folder defines the workspace path.
										</p>
									</div>
								{:else}
									<div class="py-6 text-center border-2 border-dashed border-[var(--border-primary)] rounded-xl opacity-50">
										<p class="text-xs uppercase font-bold tracking-widest text-[var(--text-muted)]">No folders selected</p>
										<p class="text-[10px] text-[var(--text-faint)] mt-1">Click Browse to add directories</p>
									</div>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Field 5: Knowledge Files (Collapsible) -->
					<div class="border-t border-[var(--border-primary)] pt-4 flex flex-col gap-3">
						<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
						<div
							role="button"
							tabindex="0"
							onclick={() => (knowledgeFilesOpen = !knowledgeFilesOpen)}
							onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); knowledgeFilesOpen = !knowledgeFilesOpen; } }}
							class="flex w-full items-center justify-between rounded-xl bg-[var(--bg-secondary)] border border-[var(--border-primary)] p-4 transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
						>
							<div class="flex items-center gap-3">
								<div class="w-8 h-8 rounded-lg bg-green-500/10 flex items-center justify-center text-green-500">
									<Icon name="book" size={14} />
								</div>
								<div class="flex flex-col">
									<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Knowledge Files</h3>
									<span class="text-[10px] text-[var(--text-faint)]">{formData.knowledge_files?.length ?? 0} file{(formData.knowledge_files?.length ?? 0) === 1 ? '' : 's'}</span>
								</div>
							</div>
							<div class="flex items-center gap-2">
								{#if knowledgeFilesOpen}
									<!-- Browse button inline in header (only when expanded) -->
									<button
										type="button"
										onclick={(e) => { e.stopPropagation(); handleOpenFile(); }}
										class="flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-lg text-[11px] font-bold uppercase tracking-[0.15em] transition-all cursor-pointer bg-[var(--accent-primary)] text-white shadow hover:brightness-110 active:scale-95"
									>
										<Icon name="file-plus" size={12} /> Browse
									</button>
								{/if}
								<Icon
									name="chevron-down"
									size={16}
									class={cn('transition-transform duration-300', knowledgeFilesOpen && 'rotate-180')}
									color="var(--text-faint)"
								/>
							</div>
						</div>

						{#if knowledgeFilesOpen}
							<div class="flex flex-col gap-2 pt-1">
								{#if formData.knowledge_files && formData.knowledge_files.length > 0}
									<div class="flex flex-col gap-1.5">
										{#each formData.knowledge_files as file, i (file)}
											<div class="flex items-center gap-2.5 rounded-lg bg-[var(--surface-input)] border border-[var(--border-primary)] px-3 py-2 transition-colors">
												<span class="flex-1 text-[12px] font-mono text-[var(--text-primary)] truncate" title={file}>{file}</span>
												<button
													type="button"
													onclick={() => removeKnowledgeFile(i)}
													class="text-[var(--text-faint)] hover:text-red-500 p-1 transition-colors cursor-pointer shrink-0"
													title="Remove file"
												>
													<Icon name="x" size={14} />
												</button>
											</div>
										{/each}
									</div>
								{:else}
									<div class="py-6 text-center border-2 border-dashed border-[var(--border-primary)] rounded-xl opacity-50">
										<p class="text-xs uppercase font-bold tracking-widest text-[var(--text-muted)]">No knowledge files selected</p>
										<p class="text-[10px] text-[var(--text-faint)] mt-1">Click Browse to add files</p>
									</div>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Field 6: Agents (Collapsible, toggle grid) -->
					<div class="border-t border-[var(--border-primary)] pt-4 flex flex-col gap-3">
						<button
							type="button"
							onclick={() => (agentsOpen = !agentsOpen)}
							class="flex w-full items-center justify-between rounded-xl bg-[var(--bg-secondary)] border border-[var(--border-primary)] p-4 transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
						>
							<div class="flex items-center gap-3">
								<div class="w-8 h-8 rounded-lg bg-purple-500/10 flex items-center justify-center text-purple-500">
									<Icon name="robot" size={14} />
								</div>
								<div class="flex flex-col">
									<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Agents</h3>
									<span class="text-[10px] text-[var(--text-faint)]">{(formData.agents ?? []).length} of {agents.length} selected</span>
								</div>
							</div>
							<Icon
								name="chevron-down"
								size={16}
								class={cn('transition-transform duration-300', agentsOpen && 'rotate-180')}
								color="var(--text-faint)"
							/>
						</button>

						{#if agentsOpen}
							<div class="pt-1">
								{#if agents.length > 0}
									<div class="grid grid-cols-6 gap-2">
										{#each agents as agent (agent.name)}
											{@const selected = (formData.agents ?? []).includes(agent.name)}
											<button
												type="button"
												onclick={() => toggleInList('agents', agent.name)}
												title={agent.description || ''}
												class="flex flex-col items-start gap-0.5 p-2.5 rounded-lg border text-left transition-all cursor-pointer min-w-0 {selected ? 'bg-[var(--accent-primary)]/15 border-[var(--accent-primary)] shadow-sm' : 'bg-[var(--surface-input)] border-[var(--border-primary)] hover:border-[var(--accent-primary)]/50'}"
											>
												<span class="text-[11px] font-bold truncate w-full" style={selected ? 'color: var(--accent-primary)' : 'color: var(--text-primary)'}>{agent.name}</span>
												<span class="text-[9px] leading-tight line-clamp-2 w-full" style="color: var(--text-faint)">{agent.description || '—'}</span>
											</button>
										{/each}
									</div>
								{:else}
									<div class="py-6 text-center border-2 border-dashed border-[var(--border-primary)] rounded-xl opacity-50">
										<p class="text-xs uppercase font-bold tracking-widest text-[var(--text-muted)]">No agents available</p>
									</div>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Field 7: Skills (Collapsible, toggle grid) -->
					<div class="border-t border-[var(--border-primary)] pt-4 flex flex-col gap-3">
						<button
							type="button"
							onclick={() => (skillsOpen = !skillsOpen)}
							class="flex w-full items-center justify-between rounded-xl bg-[var(--bg-secondary)] border border-[var(--border-primary)] p-4 transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
						>
							<div class="flex items-center gap-3">
								<div class="w-8 h-8 rounded-lg bg-cyan-500/10 flex items-center justify-center text-cyan-500">
									<Icon name="code" size={14} />
								</div>
								<div class="flex flex-col">
									<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Skills</h3>
									<span class="text-[10px] text-[var(--text-faint)]">{(formData.skills ?? []).length} of {skills.length} selected</span>
								</div>
							</div>
							<Icon
								name="chevron-down"
								size={16}
								class={cn('transition-transform duration-300', skillsOpen && 'rotate-180')}
								color="var(--text-faint)"
							/>
						</button>

						{#if skillsOpen}
							<div class="pt-1">
								{#if skills.length > 0}
									<div class="grid grid-cols-6 gap-2">
										{#each skills as skill (skill.name)}
											{@const selected = (formData.skills ?? []).includes(skill.name)}
											<button
												type="button"
												onclick={() => toggleInList('skills', skill.name)}
												title={skill.description || ''}
												class="flex flex-col items-start gap-0.5 p-2.5 rounded-lg border text-left transition-all cursor-pointer min-w-0 {selected ? 'bg-[var(--accent-primary)]/15 border-[var(--accent-primary)] shadow-sm' : 'bg-[var(--surface-input)] border border-[var(--border-primary)] hover:border-[var(--accent-primary)]/50'}"
											>
												<span class="text-[11px] font-bold truncate w-full" style={selected ? 'color: var(--accent-primary)' : 'color: var(--text-primary)'}>{skill.name}</span>
												<span class="text-[9px] leading-tight line-clamp-2 w-full" style="color: var(--text-faint)">{skill.description || '—'}</span>
											</button>
										{/each}
									</div>
								{:else}
									<div class="py-6 text-center border-2 border-dashed border-[var(--border-primary)] rounded-xl opacity-50">
										<p class="text-xs uppercase font-bold tracking-widest text-[var(--text-muted)]">No skills available</p>
									</div>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Field 8: Tools (Collapsible, toggle grid) -->
					<div class="border-t border-[var(--border-primary)] pt-4 flex flex-col gap-3">
						<button
							type="button"
							onclick={() => (toolsOpen = !toolsOpen)}
							class="flex w-full items-center justify-between rounded-xl bg-[var(--bg-secondary)] border border-[var(--border-primary)] p-4 transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
						>
							<div class="flex items-center gap-3">
								<div class="w-8 h-8 rounded-lg bg-amber-500/10 flex items-center justify-center text-amber-500">
									<Icon name="tool" size={14} />
								</div>
								<div class="flex flex-col">
									<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Tools</h3>
									<span class="text-[10px] text-[var(--text-faint)]">{(formData.tools ?? []).length} of {tools.length} selected</span>
								</div>
							</div>
							<Icon
								name="chevron-down"
								size={16}
								class={cn('transition-transform duration-300', toolsOpen && 'rotate-180')}
								color="var(--text-faint)"
							/>
						</button>

						{#if toolsOpen}
							<div class="pt-1">
								{#if tools.length > 0}
									<div class="grid grid-cols-6 gap-2">
										{#each tools as t (t.Name)}
											{@const selected = (formData.tools ?? []).includes(t.Name)}
											<button
												type="button"
												onclick={() => toggleInList('tools', t.Name)}
												title={t.Description || ''}
												class="flex flex-col items-start gap-0.5 p-2.5 rounded-lg border text-left transition-all cursor-pointer min-w-0 {selected ? 'bg-[var(--accent-primary)]/15 border-[var(--accent-primary)] shadow-sm' : 'bg-[var(--surface-input)] border border-[var(--border-primary)] hover:border-[var(--accent-primary)]/50'}"
											>
												<span class="text-[11px] font-bold truncate w-full" style={selected ? 'color: var(--accent-primary)' : 'color: var(--text-primary)'}>{t.Name}</span>
												<span class="text-[9px] leading-tight line-clamp-2 w-full" style="color: var(--text-faint)">{t.Description || '—'}</span>
											</button>
										{/each}
									</div>
								{:else}
									<div class="py-6 text-center border-2 border-dashed border-[var(--border-primary)] rounded-xl opacity-50">
										<p class="text-xs uppercase font-bold tracking-widest text-[var(--text-muted)]">No tools available</p>
									</div>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Field 10: Enabled (always true, hidden) -->
					<!-- This field is always true and not shown in the form -->

					<!-- Field 11: Color & Icon (via EntityHeader) -->
				</div>
			</div>

			<!-- ── Footer Buttons ── -->
			<div class="px-8 py-5 bg-[var(--surface-elevated)] border-t border-[var(--border-primary)] flex justify-end gap-3 shrink-0">
				<Button variant="ghost" onclick={() => onOpenChange(false)}>
					Cancelar
				</Button>
				<Button
					variant="default"
					onclick={handleSave}
					class="bg-[var(--accent-primary)] hover:bg-[var(--accent-primary)]/90 text-white"
				>
					Salvar
				</Button>
			</div>
		</DialogContent>
	</DialogPortal>
</Dialog>