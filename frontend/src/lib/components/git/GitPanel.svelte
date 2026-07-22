<script lang="ts">
	import { onMount } from 'svelte';
	import { Icon } from '$lib/components/icon';
	import { toastStore } from '$lib/stores/toast.svelte';
	import {
		GitInit, GitStatus, GitBranchList, GitBranchCreate, GitBranchCheckout,
		GitAdd, GitCommit, GitPull, GitPush, GitRemoteAdd, GitRemoteList,
		GitDiff, GitLog, GitInferCommitMessage, GetSessionByID, GetWorkspaceDir
	} from '../../../../wailsjs/go/main/App';

	interface Props { onClose?: () => void; workspacePath?: string; activeSessionID?: string; }

	let { onClose, workspacePath = '', activeSessionID = '' }: Props = $props();

	let repoPath = $state(workspacePath || '');
	let output = $state('');
	let error = $state('');
	let loading = $state(false);
	let inferring = $state(false);
	let newBranchDialog = $state(false);

	interface GitFile { path: string; letter: string; folder: string; file: string; }
	let staged = $state<GitFile[]>([]);
	let unstaged = $state<GitFile[]>([]);
	let untracked = $state<GitFile[]>([]);

	let branches = $state<{ current: boolean; name: string }[]>([]);
	let commits = $state<{ hash: string; msg: string; date: string }[]>([]);
	let diffOutput = $state('');
	let branchName = $state('');
	let commitMsg = $state('');
	let remoteUrl = $state('');
	let remoteName = $state('origin');
	let authToken = $state('');
	let pullBranch = $state('main');
	let tab = $state<'status' | 'diff' | 'log'>('status');

	let hasGit = $state(false);
	let stagedOpen = $state(true);
	let unstagedOpen = $state(true);

	onMount(async () => {
		if (!repoPath && activeSessionID) {
			try {
				const sess: any = await GetSessionByID(activeSessionID);
				if (sess?.workspace_id) {
					const dir = await GetWorkspaceDir(sess.workspace_id);
					if (dir) repoPath = dir;
				}
			} catch (_) {}
		}
		if (repoPath) checkGit();
	});

	$effect(() => {
		if (workspacePath && workspacePath !== repoPath) {
			repoPath = workspacePath;
			checkGit();
		}
	});

	async function checkGit() {
		if (!repoPath) return;
		try {
			await GitBranchList(repoPath);
			hasGit = true;
			await refresh();
		} catch { hasGit = false; }
	}

	function setOut(msg: string) { output = msg; error = ''; }
	function setErr(msg: string) { error = msg; output = ''; }

	async function doInit() {
		loading = true;
		try { setOut(await GitInit(repoPath)); hasGit = true; await refresh(); }
		catch (e: any) { setErr(String(e)); }
		finally { loading = false; }
	}

	async function refresh() {
		if (!repoPath) return;
		loading = true;
		try {
			const statusRaw = await GitStatus(repoPath);
			function gf(path: string, letter: string): GitFile {
				const idx = path.lastIndexOf('/');
				return { path, letter, folder: idx >= 0 ? path.slice(0, idx) : '', file: idx >= 0 ? path.slice(idx + 1) : path };
			}
			const s: GitFile[] = [];
			const u: GitFile[] = [];
			const ut: GitFile[] = [];
			for (const line of statusRaw.split('\n').filter(Boolean)) {
				const staging = line[0];
				const worktree = line[1];
				const path = line.slice(2).trim();
				if (!path) continue;
				const isModified = (c: string) => c !== ' ' && c !== '?';
				if (isModified(staging)) {
					s.push(gf(path, staging));
				}
				if (isModified(worktree)) {
					u.push(gf(path, worktree));
				}
				if (staging === '?' || worktree === '?') {
					ut.push(gf(path, '?'));
				}
			}
			staged = s; unstaged = u; untracked = ut;

			const branchRaw = await GitBranchList(repoPath);
			branches = branchRaw.split('\n').filter(Boolean).map(b => ({
				current: b.startsWith('*'),
				name: b.replace('* ', '')
			}));
			const logRaw = await GitLog(repoPath, 10);
			commits = logRaw.split('\n').filter(Boolean).map(l => {
				const m = l.match(/^(\S+)\s(.+?)\s\((.+)\)$/);
				return m ? { hash: m[1], msg: m[2], date: m[3] } : { hash: '', msg: l, date: '' };
			});
		} catch (e: any) { setErr(String(e)); }
		finally { loading = false; }
	}

	async function doStage(path: string) {
		try { await GitAdd(repoPath, path); await refresh(); }
		catch (e: any) { setErr(String(e)); }
	}

	async function doUnstage(path: string) {
		try { await GitAdd(repoPath, 'UNSTAGE:' + path); await refresh(); }
		catch (e: any) { setErr(String(e)); }
	}

	async function doAddAll() {
		try { setOut(await GitAdd(repoPath, '')); await refresh(); }
		catch (e: any) { setErr(String(e)); }
	}

	async function doCommit() {
		if (!commitMsg) return;
		try { setOut(await GitCommit(repoPath, commitMsg)); commitMsg = ''; await refresh(); }
		catch (e: any) { setErr(String(e)); }
	}

	async function doBranchCreate() {
		if (!branchName) return;
		try { setOut(await GitBranchCreate(repoPath, branchName)); branchName = ''; await refresh(); }
		catch (e: any) { setErr(String(e)); }
	}

	async function doBranchCheckout(name: string) {
		try { setOut(await GitBranchCheckout(repoPath, name)); await refresh(); }
		catch (e: any) { setErr(String(e)); }
	}

	async function doPull() {
		if (staged.length > 0 || unstaged.length > 0 || untracked.length > 0) {
			setErr('Há alterações pendentes. Commit ou stash antes de pull.');
			return;
		}
		try { setOut(await GitPull(repoPath, remoteName, pullBranch, authToken)); await refresh(); }
		catch (e: any) { setErr(String(e)); }
	}

	async function doPush() {
		if (staged.length > 0 || unstaged.length > 0 || untracked.length > 0) {
			setErr('Há alterações pendentes. Commit ou stash antes de push.');
			return;
		}
		try { setOut(await GitPush(repoPath, remoteName, branches.find(b => b.current)?.name || 'main', authToken)); await refresh(); }
		catch (e: any) { setErr(String(e)); }
	}

	async function doRemoteAdd() {
		if (!remoteUrl) return;
		try { setOut(await GitRemoteAdd(repoPath, remoteName, remoteUrl)); remoteUrl = ''; await refresh(); }
		catch (e: any) { setErr(String(e)); }
	}

	async function doDiff() {
		try { diffOutput = await GitDiff(repoPath); }
		catch (e: any) { diffOutput = String(e); }
	}

	function letterColor(letter: string): string {
		if (letter === 'M') return 'var(--status-warning)';
		if (letter === 'A') return 'var(--status-success)';
		if (letter === 'D') return 'var(--status-error)';
		return 'var(--text-faint)';
	}


</script>

<div class="flex flex-col h-full w-[380px] shrink-0 border-l" style="background: var(--bg-primary); border-color: var(--border-primary)">

	<div class="flex items-center justify-between px-4 py-2.5 border-b" style="border-color: var(--border-primary)">
		<div class="flex items-center gap-2">
			<Icon name="git" size={15} style="color: var(--text-secondary)" />
			<h3 class="text-[11px] font-bold uppercase tracking-widest" style="color: var(--text-faint)">Git</h3>
		</div>
		<div class="flex items-center">
			<button type="button" onclick={doPull} disabled={loading}
				class="flex items-center justify-center w-7 h-7 rounded transition-colors cursor-pointer hover:opacity-80 disabled:opacity-30"
				style="color: var(--accent-primary); padding: 0; border: none; background: none" title="Pull">
				<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"><g fill="none"><path d="M21 18a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0M8 6a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0m0 12a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0"/><path stroke="currentColor" stroke-linecap="square" stroke-width="2" d="M5.5 9v6M8 6a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0Zm0 12a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0Zm10.5-3V6H13m2-3l-3 3l3 3m6 9a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0Z"/></g></svg>
			</button>
			<button type="button" onclick={doPush} disabled={loading}
				class="flex items-center justify-center w-7 h-7 rounded transition-colors cursor-pointer hover:opacity-80 disabled:opacity-30"
				style="color: var(--accent-primary); padding: 0; border: none; background: none" title="Push">
				<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 21 21"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" d="m6.5 10.5l4-4l4 4m-4-4v11m-7-14h14"/></svg>
			</button>
			<button type="button" onclick={() => newBranchDialog = true}
				class="flex items-center justify-center w-7 h-7 rounded transition-colors cursor-pointer hover:opacity-80"
				style="color: var(--accent-primary); padding: 0; border: none; background: none" title="New Branch">
				<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M5 18a2 2 0 1 0 4 0a2 2 0 1 0-4 0M5 6a2 2 0 1 0 4 0a2 2 0 1 0-4 0m10 0a2 2 0 1 0 4 0a2 2 0 1 0-4 0M7 8v8m2 2h6a2 2 0 0 0 2-2v-5"/><path d="m14 14l3-3l3 3"/></g></svg>
			</button>
			<span class="w-px h-4 mx-0.5" style="background: var(--border-primary)"></span>
			<span class="text-[10px] font-mono font-bold" style="color: var(--accent-primary)">{branches.find(b => b.current)?.name || ''}</span>
			<button type="button" onclick={onClose}
				class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-faint)" title="Close">
				<Icon name="x" size={12} />
			</button>
		</div>
	</div>

	{#if !hasGit}
		<div class="p-3">
			<div class="flex gap-2">
				<input type="text" bind:value={repoPath} placeholder="Path"
					class="flex-1 px-2.5 py-1.5 rounded-lg text-[11px] font-mono outline-none transition-colors"
					style="background: var(--surface-input); color: var(--text-primary); border: 1px solid var(--border-primary)" />
				<button type="button" onclick={async () => { await checkGit(); if (hasGit) refresh(); }}
					class="px-2.5 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-colors hover:opacity-80"
					style="background: var(--bg-tertiary); color: var(--text-secondary); border: 1px solid var(--border-subtle)">Open</button>
			</div>
			{#if repoPath}
				<button type="button" onclick={doInit} disabled={loading}
					class="mt-2 w-full px-3 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-all disabled:opacity-30"
					style="background: var(--accent-primary); color: var(--accent-primary-fg); border: none">
					{loading ? '…' : 'Init Repository'}
				</button>
			{/if}
		</div>
	{:else}
		<div class="flex border-b" style="border-color: var(--border-primary)">
			<button type="button" onclick={() => tab = 'status'}
				class="flex-1 px-3 py-2 text-[10px] font-semibold uppercase tracking-wider cursor-pointer transition-colors"
				style="color: {tab === 'status' ? 'var(--accent-primary)' : 'var(--text-faint)'}; border-bottom: {tab === 'status' ? '2px solid var(--accent-primary)' : '2px solid transparent'}">Changes</button>
			<button type="button" onclick={() => { tab = 'diff'; doDiff(); }}
				class="flex-1 px-3 py-2 text-[10px] font-semibold uppercase tracking-wider cursor-pointer transition-colors"
				style="color: {tab === 'diff' ? 'var(--accent-primary)' : 'var(--text-faint)'}; border-bottom: {tab === 'diff' ? '2px solid var(--accent-primary)' : '2px solid transparent'}">Diff</button>
			<button type="button" onclick={() => tab = 'log'}
				class="flex-1 px-3 py-2 text-[10px] font-semibold uppercase tracking-wider cursor-pointer transition-colors"
				style="color: {tab === 'log' ? 'var(--accent-primary)' : 'var(--text-faint)'}; border-bottom: {tab === 'log' ? '2px solid var(--accent-primary)' : '2px solid transparent'}">Log</button>
		</div>

		{#if tab === 'status'}
			<div class="flex flex-col gap-2 px-3 pt-3 pb-2 border-b" style="border-color: var(--border-primary)">
				<div class="flex gap-2 items-center">
					<input type="text" bind:value={commitMsg} placeholder="Mensagem de commit"
						class="flex-1 px-2.5 py-1.5 rounded-lg text-[11px] font-mono outline-none transition-colors"
						style="background: var(--surface-input); color: var(--text-primary); border: 1px solid var(--border-primary)" />
					<button type="button" onclick={async () => {
						if (inferring) return;
						inferring = true;
						try {
							const msg = await GitInferCommitMessage(repoPath);
							commitMsg = msg;
						} catch (e: any) {
							toastStore.error('Inferência', String(e?.message ?? e));
						}
						inferring = false;
					}}
						class="shrink-0 flex items-center justify-center w-7 h-7 rounded-lg cursor-pointer transition-colors hover:opacity-80"
						style="color: var(--accent-primary)" title="{inferring ? 'Inferindo...' : 'Inferir mensagem'}">
						{#if inferring}
							<svg class="animate-spin" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10" opacity=".25"/><path d="M12 2a10 10 0 0 1 10 10" stroke-linecap="round"/></svg>
						{:else}
							<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"><path fill="currentColor" d="m9.96 9.137l.886-3.099c.332-1.16 1.976-1.16 2.308 0l.885 3.099a1.2 1.2 0 0 0 .824.824l3.099.885c1.16.332 1.16 1.976 0 2.308l-3.099.885a1.2 1.2 0 0 0-.824.824l-.885 3.099c-.332 1.16-1.976 1.16-2.308 0l-.885-3.099a1.2 1.2 0 0 0-.824-.824l-3.099-.885c-1.16-.332-1.16-1.976 0-2.308l3.099-.885a1.2 1.2 0 0 0 .824-.824m8.143 7.37c.289-.843 1.504-.844 1.792 0l.026.087l.296 1.188l1.188.297c.96.24.96 1.602 0 1.842l-1.188.297l-.296 1.188c-.24.959-1.603.959-1.843 0l-.297-1.188l-1.188-.297c-.96-.24-.96-1.603 0-1.842l1.188-.297l.297-1.188zm.896 2.29a1 1 0 0 1-.203.203a1 1 0 0 1 .203.203a1 1 0 0 1 .203-.203a1 1 0 0 1-.203-.204M4.104 2.506c.298-.871 1.585-.842 1.818.087l.296 1.188l1.188.297c.96.24.96 1.602 0 1.842l-1.188.297l-.296 1.188c-.24.959-1.603.959-1.843 0l-.297-1.188l-1.188-.297c-.96-.24-.96-1.603 0-1.842l1.188-.297l.297-1.188zM5 4.797a1 1 0 0 1-.203.202A1 1 0 0 1 5 5.203a1 1 0 0 1 .203-.204A1 1 0 0 1 5 4.796"/></svg>
						{/if}
					</button>
					<button type="button" onclick={doCommit} disabled={loading || inferring || !commitMsg}
						class="shrink-0 flex items-center justify-center w-7 h-7 rounded-lg cursor-pointer transition-all disabled:opacity-30 hover:opacity-80"
						style="color: var(--accent-primary)" title="Commit">
						<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 512 512"><path fill="currentColor" d="M448 224h-68a128 128 0 0 0-247.9 0H64a32 32 0 0 0 0 64h68.05A128 128 0 0 0 380 288h68a32 32 0 0 0 0-64m-192 96a64 64 0 1 1 64-64a64.07 64.07 0 0 1-64 64"/></svg>
					</button>
				</div>
			</div>
		{/if}

		<div class="flex-1 overflow-y-auto text-[11px]" style="color: var(--text-secondary)">
			{#if tab === 'status'}
				{#if staged.length > 0}
					<div class="select-none">
						<div class="flex items-center w-full px-3 py-1.5">
							<button type="button" onclick={() => stagedOpen = !stagedOpen}
								class="flex items-center gap-1 text-[10px] font-semibold uppercase tracking-wider cursor-pointer hover:opacity-80"
								style="color: var(--text-faint)">
								<span class="transition-transform {stagedOpen ? 'rotate-90' : ''}">▶</span>
								<span>Staged Changes</span>
							</button>
							<span class="ml-1.5 text-[10px] font-mono opacity-50">{staged.length}</span>
							<div class="ml-auto flex items-center">
								<button type="button" onclick={(e) => { e.stopPropagation(); doAddAll(); }}
									class="flex items-center justify-center w-7 h-7 rounded cursor-pointer hover:opacity-80 transition-opacity"
									style="color: var(--text-faint); padding: 0; border: none; background: none" title="Add All">
									<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 17h7m5-1h3m0 0h3m-3 0v3m0-3v-3M3 12h11M3 7h11"/></svg>
								</button>
								<button type="button" onclick={(e) => { e.stopPropagation(); doPull(); }}
									class="flex items-center justify-center w-7 h-7 rounded cursor-pointer hover:opacity-80 transition-opacity"
									style="color: var(--text-faint); padding: 0; border: none; background: none" title="Pull">
									<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"><g fill="none"><path d="M21 18a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0M8 6a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0m0 12a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0"/><path stroke="currentColor" stroke-linecap="square" stroke-width="2" d="M5.5 9v6M8 6a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0Zm0 12a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0Zm10.5-3V6H13m2-3l-3 3l3 3m6 9a2.5 2.5 0 1 1-5 0a2.5 2.5 0 0 1 5 0Z"/></g></svg>
								</button>
							</div>
						</div>
						{#if stagedOpen}
							{#each staged as item}
								<div class="group flex items-center gap-2 px-3 py-1 hover:bg-[var(--surface-hover)] cursor-pointer"
									onclick={() => doStage(item.path)}>
									<span class="w-6 text-right text-[10px] font-bold font-mono shrink-0" style="color: {letterColor(item.letter)}">{item.letter}</span>
									<div class="flex-1 min-w-0 flex items-center gap-1.5 text-[11px] font-mono">
										{#if item.folder}
											<span class="truncate text-[10px] opacity-40">{item.folder}/</span>
										{/if}
										<span class="shrink-0">{item.file}</span>
									</div>
									<button type="button" onclick={(e) => { e.stopPropagation(); doUnstage(item.path); }}
										class="opacity-0 group-hover:opacity-100 shrink-0 w-5 h-5 flex items-center justify-center rounded hover:bg-[var(--bg-tertiary)] transition-all"
										style="color: var(--status-error)" title="Unstage">
										<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="5" y1="12" x2="19" y2="12"/></svg>
									</button>
								</div>
							{/each}
						{/if}
					</div>
				{/if}

				{#if unstaged.length > 0 || untracked.length > 0}
					{@const allUnstaged = [...unstaged, ...untracked]}
					<div class="select-none">
						<button type="button" onclick={() => unstagedOpen = !unstagedOpen}
							class="flex items-center gap-1 w-full px-3 py-1.5 text-[10px] font-semibold uppercase tracking-wider cursor-pointer hover:bg-[var(--surface-hover)]"
							style="color: var(--text-faint)">
							<span class="transition-transform {unstagedOpen ? 'rotate-90' : ''}">▶</span>
							Changes
							<span class="ml-auto text-[10px] font-mono opacity-50">{allUnstaged.length}</span>
						</button>
						{#if unstagedOpen}
							{#each allUnstaged as item}
								<div class="group flex items-center gap-2 px-3 py-1 hover:bg-[var(--surface-hover)]"
									style="cursor: default">
									<span class="w-6 text-right text-[10px] font-bold font-mono shrink-0" style="color: {letterColor(item.letter)}">{item.letter}</span>
									<div class="flex-1 min-w-0 flex items-center gap-1.5 text-[11px] font-mono">
										{#if item.folder}
											<span class="truncate text-[10px] opacity-40">{item.folder}/</span>
										{/if}
										<span class="shrink-0">{item.file}</span>
									</div>
									<button type="button" onclick={(e) => { e.stopPropagation(); doStage(item.path); }}
										class="opacity-0 group-hover:opacity-100 shrink-0 w-5 h-5 flex items-center justify-center rounded hover:bg-[var(--bg-tertiary)] transition-all"
										style="color: var(--status-success)" title="Stage">
										<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
									</button>
								</div>
							{/each}
						{/if}
					</div>
				{/if}

				{#if staged.length === 0 && unstaged.length === 0 && untracked.length === 0}
					<div class="flex flex-col items-center justify-center py-12 opacity-40">
						<Icon name="git" size={24} />
						<p class="text-[11px] mt-2 italic" style="color: var(--text-faint)">Working tree clean</p>
					</div>
				{/if}

			{:else if tab === 'diff'}
				<div class="p-3">
					<pre class="text-[10px] font-mono leading-[1.6] whitespace-pre-wrap overflow-x-auto max-h-[500px]" style="color: var(--text-faint)">{diffOutput || 'No changes'}</pre>
				</div>

			{:else if tab === 'log'}
				<div class="p-3 flex flex-col gap-1">
					{#each commits as c}
						<div class="px-2.5 py-1.5 rounded-lg text-[11px] font-mono" style="background: var(--bg-tertiary)">
							<div class="flex items-center gap-2">
								<span class="text-[10px] font-bold shrink-0" style="color: var(--accent-primary)">{c.hash}</span>
								<span class="flex-1 truncate">{c.msg}</span>
							</div>
							<p class="text-[9px] opacity-50 mt-0.5">{c.date}</p>
						</div>
					{:else}
						<p class="text-[11px] opacity-50 italic py-2">No commits yet</p>
					{/each}
		</div>

		{#if newBranchDialog}
			<div class="px-3 py-3 border-t flex flex-col gap-2" style="border-color: var(--border-primary)">
				<div class="flex items-center gap-2">
					<input type="text" bind:value={branchName} placeholder="Branch name"
						class="flex-1 px-2.5 py-1.5 rounded-lg text-[11px] font-mono outline-none transition-colors"
						style="background: var(--surface-input); color: var(--text-primary); border: 1px solid var(--border-primary)" />
					<button type="button" onclick={async () => { await doBranchCreate(); newBranchDialog = false; }}
						disabled={loading || !branchName}
						class="px-3 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-all disabled:opacity-30 hover:brightness-110"
						style="background: var(--accent-primary); color: var(--accent-primary-fg); border: none">Confirm</button>
					<button type="button" onclick={() => { newBranchDialog = false; branchName = ''; }}
						class="px-3 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-colors hover:opacity-80"
						style="background: transparent; color: var(--text-secondary); border: 1px solid var(--border-subtle)">Cancel</button>
				</div>
			</div>
		{/if}
	{/if}
		</div>

		{#if output || error}
			<div class="px-3 pb-2">
				{#if output}
					<div class="px-3 py-2 rounded-lg text-[10px]" style="background: color-mix(in srgb, var(--status-success) 10%, transparent); color: var(--status-success)">{output}</div>
				{/if}
				{#if error}
					<div class="px-3 py-2 rounded-lg text-[10px]" style="background: color-mix(in srgb, var(--status-error) 10%, transparent); color: var(--status-error)">{error}</div>
				{/if}
			</div>
		{/if}


	{/if}
</div>
