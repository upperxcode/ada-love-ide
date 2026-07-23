<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import { Separator } from '$lib/components/ui/separator';
	import {
		GetWorkspaces, SetActiveWorkspace, GetWorkers, GetSessions,
		AddWorkerToWorkspace, RemoveWorkerFromWorkspace, ListWorkspaceWorkers,
		CreateSessionWithConfig, CountWorkerSessions, RenameSession, TogglePin, DeleteSession,
	} from '../../../../wailsjs/go/main/App';
	import { toastStore } from '$lib/stores/toast.svelte';

	interface WorkspaceItem {
		id: number; title: string; path: string; icon: string; color: string;
		folders: string[]; spec_wizard_id: string; worker_names?: string[];
	}

	interface WorkerItem {
		id: number; name: string; persona: string; description?: string;
		language?: string; icon?: string; color?: string; connection_type?: string;
	}

	interface ChatItem {
		id: string; title: string; worker_name: string; pinned?: boolean;
	}

	interface SidebarProps {
		class?: string; onOpenSettings?: () => void; onNewWorkspace?: () => void;
		onEditWorkspace?: (ws: Record<string, any>) => void;
		onOpenLogs?: () => void; onOpenGit?: () => void;
		activeWorkspace?: string; activeSessionID?: string;
	}

	let { class: className, onOpenSettings, onNewWorkspace, onEditWorkspace,
		onOpenLogs, onOpenGit,
		activeWorkspace = $bindable(''), activeSessionID = $bindable('') }: SidebarProps = $props();

	let workspaces = $state<WorkspaceItem[]>([]);
	let sessionsMap = $state<Record<string, ChatItem[]>>({});
	let loading = $state(true);

	let popoverOpen = $state(false);
	let popoverWs = $state<WorkspaceItem | null>(null);
	let allWorkers = $state<WorkerItem[]>([]);
	let linkedWorkerNames = $state<string[]>([]);
	let loadingWorkers = $state(false);
	let popoverX = $state(0); let popoverY = $state(0);
	let popoverRef = $state<HTMLDivElement | null>(null);

	let deleteConfirmWs = $state<string | null>(null);
	let deleteConfirmWorker = $state<string | null>(null);
	let deleteConfirmCount = $state(0);
	let deleteTyped = $state('');

	// Chat rename state
	let editingChatID = $state<string | null>(null);
	let editingChatTitle = $state('');

	function startRename(ch: ChatItem) {
		editingChatID = ch.id;
		editingChatTitle = ch.title;
	}

	async function finishRename() {
		const id = editingChatID;
		editingChatID = null;
		if (id && editingChatTitle.trim()) {
			try {
				await RenameSession(id, editingChatTitle.trim());
				await loadSessions();
			} catch (_) {}
		}
	}

	onMount(async () => { await loadAll(); });

	async function loadAll() {
		loading = true;
		try {
			const list = await GetWorkspaces() as WorkspaceItem[];
			workspaces = list || [];
			if (!activeWorkspace && workspaces.length > 0) {
				activeWorkspace = workspaces[0].path;
			}
			await loadSessions();
		} catch (e) {
			console.error('[Sidebar] Failed to load:', e);
			workspaces = [];
		} finally { loading = false; }
	}

	async function loadSessions() {
		const map: Record<string, ChatItem[]> = {};
		for (const ws of workspaces) {
			try {
				const sessions = await GetSessions(ws.path) as ChatItem[];
				for (const s of sessions || []) {
					const key = ws.path + '|' + s.worker_name;
					if (!map[key]) map[key] = [];
					map[key].push(s);
				}
			} catch (_) { /* ignore */ }
		}
		sessionsMap = map;
	}

	async function selectWorkspace(ws: WorkspaceItem) {
		activeWorkspace = ws.path;
		try { await SetActiveWorkspace(ws.path); } catch (_) {}
	}

	async function openPopover(e: MouseEvent, ws: WorkspaceItem) {
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		popoverX = rect.right + 6; popoverY = rect.top;
		popoverWs = ws; popoverOpen = true; loadingWorkers = true; await tick();
		try {
			const [all, linked] = await Promise.all([
				GetWorkers() as Promise<WorkerItem[]>,
				ListWorkspaceWorkers(ws.path) as Promise<WorkerItem[]>,
			]);
			allWorkers = all || [];
			linkedWorkerNames = (linked || []).map((w: WorkerItem) => w.name);
		} catch (_) { allWorkers = []; linkedWorkerNames = []; }
		finally { loadingWorkers = false; }
	}

	async function toggleWorker(workerName: string) {
		if (!popoverWs) return;
		const isLinked = linkedWorkerNames.includes(workerName);
		try {
			isLinked
				? await RemoveWorkerFromWorkspace(popoverWs.path, workerName)
				: await AddWorkerToWorkspace(popoverWs.path, workerName);
			closePopover();
			await loadAll();
		} catch (e: any) { toastStore.error('Erro', e?.message || String(e)); }
	}

	function closePopover() {
		popoverOpen = false; popoverWs = null; allWorkers = []; linkedWorkerNames = [];
	}

	async function createChat(wsPath: string, workerName: string) {
		try {
			const sess = await CreateSessionWithConfig(wsPath, workerName, activeSessionID);
			activeSessionID = sess.id;
			await loadSessions();
		} catch (e: any) { toastStore.error('Failed to create chat', e?.message || String(e)); }
	}

	async function confirmDeleteWorker(wsPath: string, workerName: string) {
		const count = await CountWorkerSessions(wsPath, workerName) as number;
		if (count === 0) { await doRemoveWorker(wsPath, workerName); return; }
		deleteConfirmWs = wsPath; deleteConfirmWorker = workerName;
		deleteConfirmCount = count; deleteTyped = '';
	}

	async function deleteChat(chatID: string) {
		try {
			await DeleteSession(chatID);
			if (activeSessionID === chatID) activeSessionID = '';
			await loadSessions();
		} catch (e: any) {
			toastStore.error('Failed to delete chat', e?.message || String(e));
		}
	}

	async function doRemoveWorker(wsPath: string, workerName: string) {
		try {
			await RemoveWorkerFromWorkspace(wsPath, workerName);
			toastStore.success(`Worker "${workerName}" removed`);
			deleteConfirmWs = null; deleteConfirmWorker = null; deleteTyped = '';
			await loadAll();
		} catch (e: any) { toastStore.error('Erro', e?.message || String(e)); }
	}

	function wsSessions(wsPath: string, wName: string): ChatItem[] {
		return sessionsMap[wsPath + '|' + wName] || [];
	}
</script>

<aside class={cn('flex h-full flex-col shrink-0 border-r border-[var(--border-primary)] bg-[var(--bg-secondary)] w-[280px]', className)}>
	<div class="px-4 pt-4 pb-2">
		<h2 class="text-sm font-semibold tracking-[0.2em] uppercase" style="color: var(--text-muted)">ADA LOVE</h2>
	</div>
	<Separator class="mx-3 w-auto" />

	<div class="flex-1 overflow-y-auto px-2 py-2">
		<div class="flex items-center justify-between px-1 mb-2">
			<span class="text-[11px] font-bold uppercase tracking-[0.2em]" style="color: var(--text-faint)">Workspaces</span>
			<button type="button" onclick={onNewWorkspace} title="New workspace"
				class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-muted)">
				<Icon name="plus" size={16} />
			</button>
		</div>

		{#if loading}
			<p class="px-2 py-6 text-center text-[11px]" style="color: var(--text-faint)">Carregando...</p>
		{:else if workspaces.length === 0}
			<p class="px-2 py-6 text-center text-[11px]" style="color: var(--text-faint)">No workspaces yet</p>
		{:else}
			<div class="flex flex-col gap-0.5">
				{#each workspaces as ws}
					<div class={cn('group flex flex-col w-full px-1 py-1 rounded-lg transition-all border border-transparent',
						activeWorkspace === ws.path ? 'bg-[var(--surface-elevated)] border-[var(--border-primary)]' : 'hover:bg-[var(--surface-hover)]')}>
						<div class="flex items-center gap-1.5 w-full">
							<button type="button" onclick={() => selectWorkspace(ws)}
								class="flex items-center gap-2 flex-1 min-w-0 text-left cursor-pointer">
								<div class="flex items-center justify-center w-[24px] h-[24px] rounded-lg shrink-0 text-base"
									style="background-color: {ws.color || '#3b82f6'}20; color: {ws.color || '#3b82f6'}">{ws.icon || '📁'}</div>
								<div class="flex flex-col min-w-0 flex-1 leading-tight">
									<span class="text-sm font-semibold truncate" style="color: var(--text-primary)">{ws.title}</span>
									<span class="text-[11px] truncate" style="color: var(--text-faint)">{ws.path || ws.folders?.[0] || ''}</span>
								</div>
								{#if ws.spec_wizard_id}
									<div class="shrink-0"><Icon name="wand" size={16} color="var(--accent-primary)" /></div>
								{/if}
							</button>
							<div class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
								<button type="button" onclick={() => onEditWorkspace?.(ws)} title="Edit workspace"
									class="flex items-center justify-center w-7 h-7 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-faint)">
									<Icon name="pencil" size={16} />
								</button>
								<button type="button" onclick={(e) => openPopover(e, ws)} title="Manage workers"
									class="flex items-center justify-center w-7 h-7 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-faint)">
									<Icon name="plus" size={16} />
								</button>
							</div>
						</div>

						{#if ws.worker_names && ws.worker_names.length > 0}
							<div class="flex flex-col pl-[34px] pt-1 gap-0.5">
								{#each ws.worker_names as wName}
									{@const chats = wsSessions(ws.path, wName)}
									<div class="group/wkr">
										<div class="flex items-center gap-1.5 py-0.5 rounded transition-colors hover:bg-[var(--surface-hover)]">
											<div class="flex items-center justify-center w-[20px] h-[20px] rounded shrink-0 text-sm" style="background-color: var(--accent-primary)12; color: var(--accent-primary)">
												<Icon name="bot" size={16} />
											</div>
											<span class="text-[12px] font-medium flex-1 truncate" style="color: var(--text-secondary)">{wName}</span>
											<div class="flex items-center gap-0.5 opacity-0 group-hover/wkr:opacity-100 transition-opacity">
												<button type="button" onclick={() => createChat(ws.path, wName)} title="New chat"
													class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-faint)">
													<Icon name="plus" size={14} />
												</button>
												<button type="button" onclick={() => confirmDeleteWorker(ws.path, wName)} title="Remove worker"
													class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-faint)">
													<Icon name="deleteIcon" size={14} />
												</button>
											</div>
										</div>
										{#if chats.length > 0}
											{@const sorted = [...chats].sort((a, b) => {
												if (a.pinned && !b.pinned) return -1;
												if (!a.pinned && b.pinned) return 1;
												return a.title.localeCompare(b.title);
											})}
											<div class="flex flex-col pl-[30px] gap-0.5 pb-0.5">
												{#each sorted as ch}
													<div class={cn('group/ch flex items-center gap-1 px-1 py-0.5 rounded transition-colors',
														activeSessionID === ch.id ? 'bg-[var(--accent-primary)]/10' : 'hover:bg-[var(--surface-hover)]')}>
														{#if editingChatID === ch.id}
															<input type="text" bind:value={editingChatTitle}
																onblur={finishRename}
																onkeydown={(e) => { if (e.key === 'Enter') finishRename(); if (e.key === 'Escape') editingChatID = null; }}
																class="flex-1 px-1 py-0.5 text-[11px] rounded border border-[var(--accent-primary)] bg-[var(--surface-input)] outline-none min-w-0"
																style="color: var(--text-primary)"
															/>
														{:else}
															<button type="button" onclick={() => { activeWorkspace = ws.path; activeSessionID = ch.id; }}
																ondblclick={() => startRename(ch)}
																class="flex items-center gap-1.5 flex-1 min-w-0 text-left cursor-pointer">
																<span style="background-color: {activeSessionID === ch.id ? '#3b82f6' : 'transparent'}; width: 8px; height: 8px;" class="shrink-0 rounded-full"></span>
																<span class="text-[11px] truncate" style="color: {activeSessionID === ch.id ? 'var(--accent-primary)' : 'var(--text-muted)'}; text-transform: {activeSessionID === ch.id ? 'uppercase' : 'lowercase'}">{ch.title}</span>
															</button>
															{#if ch.pinned}
																<button type="button" onclick={async () => { try { await TogglePin(ch.id); await loadSessions(); } catch { } }} title="Unpin"
																	class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
																	style="color: #22c55e">
																	<Icon name="pinOn" size={14} />
																</button>
															{/if}
															<div class="flex items-center gap-px opacity-0 group-hover/ch:opacity-100 transition-opacity">
																{#if !ch.pinned}
																	<button type="button" onclick={async () => { try { await TogglePin(ch.id); await loadSessions(); } catch { } }} title="Pin"
																		class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
																		style="color: var(--text-faint)">
																		<Icon name="pinOff" size={14} />
																	</button>
																{/if}
																<button type="button" onclick={() => deleteChat(ch.id)} title="Delete chat"
																	class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
																	style="color: var(--text-faint)">
																	<Icon name="deleteIcon" size={14} />
																</button>
															</div>
														{/if}
													</div>
												{/each}
											</div>
										{/if}
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<Separator class="mx-3 w-auto" />
	<div class="flex items-center justify-between px-3 py-2">
		<span class="text-[11px] font-mono" style="color: var(--text-faint)">v1.1.0</span>
		<div class="flex items-center gap-1">
			<button type="button" onclick={onOpenLogs} title="Logs"
				class="flex items-center justify-center w-7 h-7 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-muted)">
				<Icon name="log" size={15} />
			</button>
			<button type="button" onclick={onOpenGit} title="Git"
				class="flex items-center justify-center w-7 h-7 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-muted)">
				<Icon name="git" size={15} />
			</button>
			<button type="button" onclick={onOpenSettings}
				class="flex items-center justify-center w-7 h-7 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-muted)">
				<Icon name="cog" size={16} />
			</button>
		</div>
	</div>
</aside>

{#if popoverOpen}
	<button type="button" class="fixed inset-0 z-[90] cursor-default" aria-label="Close popover" onclick={closePopover}></button>
	<div class="fixed z-[100] w-[240px] bg-[var(--bg-tertiary)] border border-[var(--border-primary)] rounded-xl shadow-2xl overflow-hidden"
		style="left: {popoverX}px; top: {popoverY}px;" bind:this={popoverRef} onclick={(e) => e.stopPropagation()} role="presentation">
		<div class="flex items-center justify-between px-3 py-2 border-b border-[var(--border-primary)]">
			<span class="text-[11px] font-bold uppercase" style="color: var(--text-faint)">{popoverWs?.title || ''} — Workers</span>
			<button type="button" onclick={closePopover}
				class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-faint)">
				<Icon name="x" size={14} />
			</button>
		</div>
		<div class="max-h-[280px] overflow-y-auto py-1">
			{#if loadingWorkers}
				<div class="flex items-center justify-center py-6"><div class="h-5 w-5 rounded-full border-2 border-[var(--border-primary)] border-t-[var(--accent-primary)] animate-spin"></div></div>
			{:else if allWorkers.length === 0}
				<p class="text-center py-6 text-[11px]" style="color: var(--text-faint)">No worker configured</p>
			{:else}
				{#each allWorkers as wkr}
					{@const isLinked = linkedWorkerNames.includes(wkr.name)}
					<button type="button" onclick={() => toggleWorker(wkr.name)}
						class={cn('flex items-center gap-2.5 w-full px-3 py-2 text-left transition-all cursor-pointer',
							isLinked ? 'bg-[var(--accent-primary)]/8' : 'hover:bg-[var(--surface-hover)]')}>
						<div class="flex items-center justify-center w-7 h-7 rounded shrink-0 text-sm"
							style="background-color: {(wkr.color || '#3b82f6')}20; color: {wkr.color || '#3b82f6'}">{wkr.icon || '🤖'}</div>
						<div class="flex flex-col min-w-0 flex-1 leading-tight">
							<div class="flex items-center gap-1.5">
								<span class="text-[13px] font-semibold truncate" style="color: var(--text-primary)">{wkr.name}</span>
								<span class="text-[9px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded" style="background-color: color-mix(in srgb, var(--accent-primary) 12%, transparent); color: var(--accent-primary)">{wkr.connection_type || 'ada'}</span>
							</div>
						</div>
						<div class="shrink-0 w-4 flex items-center justify-center">
							{#if isLinked}<Icon name="check" size={14} color="var(--accent-primary)" />{/if}
						</div>
					</button>
				{/each}
			{/if}
		</div>
	</div>
{/if}

{#if deleteConfirmWs && deleteConfirmWorker}
	<button type="button" class="fixed inset-0 z-[90] cursor-default" aria-label="Close delete confirmation" onclick={() => { deleteConfirmWs = null; deleteConfirmWorker = null; deleteTyped = ''; }}></button>
	<div class="fixed z-[100] w-[280px] bg-[var(--bg-tertiary)] border border-[var(--border-primary)] rounded-xl shadow-2xl overflow-hidden"
		style="left: 50%; top: 50%; transform: translate(-50%, -50%);">
		<div class="px-4 py-3 border-b border-[var(--border-primary)]">
			<p class="text-[13px] font-semibold" style="color: var(--text-primary)">Remover worker "{deleteConfirmWorker}"</p>
			<p class="text-[11px] mt-1" style="color: var(--text-faint)">{deleteConfirmCount} chat{(deleteConfirmCount > 1 ? 's' : '')} ativo{(deleteConfirmCount > 1 ? 's' : '')}. Digite DELETE:</p>
		</div>
		<div class="px-4 py-3">
			<input type="text" bind:value={deleteTyped} placeholder="DELETE"
				class="w-full px-3 py-2 rounded-lg text-[13px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-red-500/30 focus:border-red-500" />
		</div>
		<div class="flex justify-end gap-2 px-4 py-3 border-t border-[var(--border-primary)]">
			<button type="button" onclick={() => { deleteConfirmWs = null; deleteConfirmWorker = null; deleteTyped = ''; }}
				class="px-3 py-1.5 rounded-lg text-[12px] font-medium transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-muted)">Cancelar</button>
			<button type="button" disabled={deleteTyped !== 'DELETE'}
				onclick={() => deleteConfirmWs && deleteConfirmWorker && doRemoveWorker(deleteConfirmWs, deleteConfirmWorker)}
				class="px-3 py-1.5 rounded-lg text-[12px] font-bold transition-colors cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed" style="background-color: #ef4444; color: white;">Remover</button>
		</div>
	</div>
{/if}
