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
		spec_wizards: [],
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
	let specWizardsOpen = $state(false);

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
			formData = { ...entity };
			// Derive path from folders[0] if available
			if (entity.folders && entity.folders.length > 0) {
				formData.path = entity.folders[0];
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
				spec_wizards: [],
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

	function addAgent() {
		const selectedAgent = agents.find(a => a.id === formData.agents.find((id: string) => id === a.id));
		if (!selectedAgent) return;
		if (!formData.agents.includes(selectedAgent.id)) {
			formData.agents = [...formData.agents, selectedAgent.id];
		}
	}

	function removeAgent(index: number) {
		formData.agents = formData.agents.filter((_: unknown, idx: number) => idx !== index);
	}

	function addSkill() {
		const selectedSkill = skills.find(s => s.id === formData.skills.find((id: string) => id === s.id));
		if (!selectedSkill) return;
		if (!formData.skills.includes(selectedSkill.id)) {
			formData.skills = [...formData.skills, selectedSkill.id];
		}
	}

	function removeSkill(index: number) {
		formData.skills = formData.skills.filter((_: unknown, idx: number) => idx !== index);
	}

	function addTool() {
		const selectedTool = tools.find(t => t.id === formData.tools.find((id: string) => id === t.id));
		if (!selectedTool) return;
		if (!formData.tools.includes(selectedTool.id)) {
			formData.tools = [...formData.tools, selectedTool.id];
		}
	}

	function removeTool(index: number) {
		formData.tools = formData.tools.filter((_: unknown, idx: number) => idx !== index);
	}

	function addSpecWizard() {
		const selectedWizard = specWizards.find(w => w.id === formData.spec_wizards.find((id: string) => id === w.id));
		if (!selectedWizard) return;
		if (!formData.spec_wizards.includes(selectedWizard.id)) {
			formData.spec_wizards = [...formData.spec_wizards, selectedWizard.id];
		}
	}

	function removeSpecWizard(index: number) {
		formData.spec_wizards = formData.spec_wizards.filter((_: unknown, idx: number) => idx !== index);
	}

async function handleOpenDirectory() {
			try {
				const result = await (window as any).go.main.App.OpenDirectoryDialog();
				if (result) {
					// Avoid duplicates
					if (formData.folders.includes(result)) return;
					formData.folders = [...formData.folders, result];
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
					// Avoid duplicates
					if (formData.knowledge_files.includes(result)) return;
					formData.knowledge_files = [...formData.knowledge_files, result];
				}
			} catch (e) {
				console.error('Failed to open file dialog:', e);
			}
		}

	function handleSave() {
		// Force enabled = true
		formData.enabled = true;
		// Ensure path is derived from folders[0]
		if (formData.folders && formData.folders.length > 0) {
			formData.path = formData.folders[0];
		}
		onSave(formData);
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
				<div class="flex flex-col gap-6">
<!-- Field 1: Name -->
						<SettingRow label="Name" description="Display name of the workspace" required>
							<input
								bind:value={formData.name}
								placeholder="Workspace name"
								class="w-full rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]"
							/>
						</SettingRow>

						<!-- Field 2: Description -->
						<SettingRow label="Description" description="Brief summary of the workspace">
							<input
								bind:value={formData.description}
								placeholder="Brief description of the workspace..."
								class="w-full rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]"
							/>
						</SettingRow>

						<!-- Field 3: Path (auto-derived) -->
						<SettingRow label="Path" description="Auto-derived from first folder">
							<input
								bind:value={formData.path}
								placeholder="/path/to/workspace"
								class="w-full rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)] cursor-not-allowed"
								readOnly
							/>
						</SettingRow>

					<!-- Field 4: Folders (Collapsible) -->
					<div class="mt-4 border-t border-[var(--border-primary)] pt-6 flex flex-col gap-4">
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
					<div class="border-t border-[var(--border-primary)] pt-6 flex flex-col gap-4">
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

					<!-- Field 6: Agents (Collapsible) -->
					<div class="border-t border-[var(--border-primary)] pt-6 flex flex-col gap-4">
						<button
							type="button"
							onclick={() => (agentsOpen = !agentsOpen)}
							class="flex w-full items-center justify-between rounded-xl bg-[var(--bg-secondary)] border border-[var(--border-primary)] p-4 transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
						>
							<div class="flex items-center gap-3">
								<div class="w-8 h-8 rounded-lg bg-purple-500/10 flex items-center justify-center text-purple-500">
									<Icon name="robot" size={14} />
								</div>
								<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Agents</h3>
							</div>
							<Icon
								name="chevron-down"
								size={16}
								class={cn('transition-transform duration-300', agentsOpen && 'rotate-180')}
								color="var(--text-faint)"
							/>
						</button>

						{#if agentsOpen}
							<div class="flex flex-col gap-3">
								{#if formData.agents && formData.agents.length > 0}
									<div class="flex flex-col gap-2">
										{#each formData.agents as agentId, i}
											{@const agent = agents.find(a => a.id === agentId)}
											{#if agent}
												<div class="flex items-center gap-2 rounded-lg bg-[var(--surface-input)] border border-[var(--border-primary)] px-3 py-2">
													<span class="flex-1 text-sm text-[var(--text-primary)]">{agent.name || agent.id}</span>
													<button
														type="button"
														onclick={() => removeAgent(i)}
														class="text-[var(--text-faint)] hover:text-red-500 p-1 transition-colors cursor-pointer"
													>
														<Icon name="x" size={14} />
													</button>
												</div>
											{/if}
										{/each}
									</div>
								{:else}
									<div class="py-6 text-center border-2 border-dashed border-[var(--border-primary)] rounded-xl opacity-40">
										<p class="text-xs uppercase font-bold tracking-widest">No agents selected</p>
									</div>
								{/if}

								<div class="flex items-center gap-2">
									<select
										bind:value={formData.agents}
										class="flex-1 rounded-lg px-3 py-2 text-sm border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none focus:ring-1 focus:ring-[var(--accent-primary)]/30"
									>
										<option value="">-- Select Agent --</option>
										{#each agents as agent}
											<option value={agent.id}>{agent.name || agent.id}</option>
										{/each}
									</select>
									<button
										type="button"
										disabled={!formData.agents.includes(agents[0]?.id)}
										onclick={addAgent}
										class="flex items-center justify-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-[0.2em] transition-all cursor-pointer disabled:opacity-30 {formData.agents.length > 0 ? 'bg-[var(--accent-primary)] text-white shadow-lg hover:brightness-110 active:scale-90' : 'bg-[var(--surface-input)] border border-[var(--border-primary)] text-[var(--text-muted)]'}"
									>
										<Icon name="plus" size={12} /> Add
									</button>
								</div>
							</div>
						{/if}
					</div>

					<!-- Field 7: Skills (Collapsible) -->
					<div class="border-t border-[var(--border-primary)] pt-6 flex flex-col gap-4">
						<button
							type="button"
							onclick={() => (skillsOpen = !skillsOpen)}
							class="flex w-full items-center justify-between rounded-xl bg-[var(--bg-secondary)] border border-[var(--border-primary)] p-4 transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
						>
							<div class="flex items-center gap-3">
								<div class="w-8 h-8 rounded-lg bg-cyan-500/10 flex items-center justify-center text-cyan-500">
									<Icon name="code" size={14} />
								</div>
								<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Skills</h3>
							</div>
							<Icon
								name="chevron-down"
								size={16}
								class={cn('transition-transform duration-300', skillsOpen && 'rotate-180')}
								color="var(--text-faint)"
							/>
						</button>

						{#if skillsOpen}
							<div class="flex flex-col gap-3">
								{#if formData.skills && formData.skills.length > 0}
									<div class="flex flex-col gap-2">
										{#each formData.skills as skillId, i}
											{@const skill = skills.find(s => s.id === skillId)}
											{#if skill}
												<div class="flex items-center gap-2 rounded-lg bg-[var(--surface-input)] border border-[var(--border-primary)] px-3 py-2">
													<span class="flex-1 text-sm text-[var(--text-primary)]">{skill.name || skillId}</span>
													<button
														type="button"
														onclick={() => removeSkill(i)}
														class="text-[var(--text-faint)] hover:text-red-500 p-1 transition-colors cursor-pointer"
													>
														<Icon name="x" size={14} />
													</button>
												</div>
											{/if}
										{/each}
									</div>
								{:else}
									<div class="py-6 text-center border-2 border-dashed border-[var(--border-primary)] rounded-xl opacity-40">
										<p class="text-xs uppercase font-bold tracking-widest">No skills selected</p>
									</div>
								{/if}

								<div class="flex items-center gap-2">
									<select
										bind:value={formData.skills}
										class="flex-1 rounded-lg px-3 py-2 text-sm border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none focus:ring-1 focus:ring-[var(--accent-primary)]/30"
									>
										<option value="">-- Select Skill --</option>
										{#each skills as skill}
											<option value={skill.id}>{skill.name || skill.id}</option>
										{/each}
									</select>
									<button
										type="button"
										disabled={!formData.skills.includes(skills[0]?.id)}
										onclick={addSkill}
										class="flex items-center justify-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-[0.2em] transition-all cursor-pointer disabled:opacity-30 {formData.skills.length > 0 ? 'bg-[var(--accent-primary)] text-white shadow-lg hover:brightness-110 active:scale-90' : 'bg-[var(--surface-input)] border border-[var(--border-primary)] text-[var(--text-muted)]'}"
									>
										<Icon name="plus" size={12} /> Add
									</button>
								</div>
							</div>
						{/if}
					</div>

					<!-- Field 8: Tools (Collapsible) -->
					<div class="border-t border-[var(--border-primary)] pt-6 flex flex-col gap-4">
						<button
							type="button"
							onclick={() => (toolsOpen = !toolsOpen)}
							class="flex w-full items-center justify-between rounded-xl bg-[var(--bg-secondary)] border border-[var(--border-primary)] p-4 transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
						>
							<div class="flex items-center gap-3">
								<div class="w-8 h-8 rounded-lg bg-amber-500/10 flex items-center justify-center text-amber-500">
									<Icon name="tool" size={14} />
								</div>
								<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Tools</h3>
							</div>
							<Icon
								name="chevron-down"
								size={16}
								class={cn('transition-transform duration-300', toolsOpen && 'rotate-180')}
								color="var(--text-faint)"
							/>
						</button>

						{#if toolsOpen}
							<div class="flex flex-col gap-3">
								{#if formData.tools && formData.tools.length > 0}
									<div class="flex flex-col gap-2">
										{#each formData.tools as toolId, i}
											{@const tool = tools.find(t => t.id === toolId)}
											{#if tool}
												<div class="flex items-center gap-2 rounded-lg bg-[var(--surface-input)] border border-[var(--border-primary)] px-3 py-2">
													<span class="flex-1 text-sm text-[var(--text-primary)]">{tool.name || toolId}</span>
													<button
														type="button"
														onclick={() => removeTool(i)}
														class="text-[var(--text-faint)] hover:text-red-500 p-1 transition-colors cursor-pointer"
													>
														<Icon name="x" size={14} />
													</button>
												</div>
											{/if}
										{/each}
									</div>
								{:else}
									<div class="py-6 text-center border-2 border-dashed border-[var(--border-primary)] rounded-xl opacity-40">
										<p class="text-xs uppercase font-bold tracking-widest">No tools selected</p>
									</div>
								{/if}

								<div class="flex items-center gap-2">
									<select
										bind:value={formData.tools}
										class="flex-1 rounded-lg px-3 py-2 text-sm border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none focus:ring-1 focus:ring-[var(--accent-primary)]/30"
									>
										<option value="">-- Select Tool --</option>
										{#each tools as tool}
											<option value={tool.id}>{tool.name || tool.id}</option>
										{/each}
									</select>
									<button
										type="button"
										disabled={!formData.tools.includes(tools[0]?.id)}
										onclick={addTool}
										class="flex items-center justify-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-[0.2em] transition-all cursor-pointer disabled:opacity-30 {formData.tools.length > 0 ? 'bg-[var(--accent-primary)] text-white shadow-lg hover:brightness-110 active:scale-90' : 'bg-[var(--surface-input)] border border-[var(--border-primary)] text-[var(--text-muted)]'}"
									>
										<Icon name="plus" size={12} /> Add
									</button>
								</div>
							</div>
						{/if}
					</div>

					<!-- Field 9: Spec Wizards (Collapsible) -->
					<div class="border-t border-[var(--border-primary)] pt-6 flex flex-col gap-4">
						<button
							type="button"
							onclick={() => (specWizardsOpen = !specWizardsOpen)}
							class="flex w-full items-center justify-between rounded-xl bg-[var(--bg-secondary)] border border-[var(--border-primary)] p-4 transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
						>
							<div class="flex items-center gap-3">
								<div class="w-8 h-8 rounded-lg bg-pink-500/10 flex items-center justify-center text-pink-500">
									<Icon name="sparkles" size={14} />
								</div>
								<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Spec Wizards</h3>
							</div>
							<Icon
								name="chevron-down"
								size={16}
								class={cn('transition-transform duration-300', specWizardsOpen && 'rotate-180')}
								color="var(--text-faint)"
							/>
						</button>

						{#if specWizardsOpen}
							<div class="flex flex-col gap-3">
								{#if formData.spec_wizards && formData.spec_wizards.length > 0}
									<div class="flex flex-col gap-2">
										{#each formData.spec_wizards as wizardId, i}
											{@const wizard = specWizards.find(w => w.id === wizardId)}
											{#if wizard}
												<div class="flex items-center gap-2 rounded-lg bg-[var(--surface-input)] border border-[var(--border-primary)] px-3 py-2">
													<span class="flex-1 text-sm text-[var(--text-primary)]">{wizard.name || wizardId}</span>
													<button
														type="button"
														onclick={() => removeSpecWizard(i)}
														class="text-[var(--text-faint)] hover:text-red-500 p-1 transition-colors cursor-pointer"
													>
														<Icon name="x" size={14} />
													</button>
												</div>
											{/if}
										{/each}
									</div>
								{:else}
									<div class="py-6 text-center border-2 border-dashed border-[var(--border-primary)] rounded-xl opacity-40">
										<p class="text-xs uppercase font-bold tracking-widest">No spec wizards selected</p>
									</div>
								{/if}

								<div class="flex items-center gap-2">
									<select
										bind:value={formData.spec_wizards}
										class="flex-1 rounded-lg px-3 py-2 text-sm border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none focus:ring-1 focus:ring-[var(--accent-primary)]/30"
									>
										<option value="">-- Select Spec Wizard --</option>
										{#each specWizards as wizard}
											<option value={wizard.id}>{wizard.name || wizard.id}</option>
										{/each}
									</select>
									<button
										type="button"
										disabled={!formData.spec_wizards.includes(specWizards[0]?.id)}
										onclick={addSpecWizard}
										class="flex items-center justify-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-[0.2em] transition-all cursor-pointer disabled:opacity-30 {formData.spec_wizards.length > 0 ? 'bg-[var(--accent-primary)] text-white shadow-lg hover:brightness-110 active:scale-90' : 'bg-[var(--surface-input)] border border-[var(--border-primary)] text-[var(--text-muted)]'}"
									>
										<Icon name="plus" size={12} /> Add
									</button>
								</div>
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