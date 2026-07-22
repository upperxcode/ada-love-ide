<script lang="ts">
	import { Icon } from '$lib/components/icon';
	import {
		GitInit, GitStatus, GitBranchList, GitBranchCreate, GitBranchCheckout,
		GitAdd, GitCommit, GitPull, GitPush, GitRemoteAdd, GitRemoteList,
		GitDiff, GitLog
	} from '../../../../wailsjs/go/main/App';

	interface Props { onClose?: () => void; workspacePath?: string; }

	let { onClose, workspacePath = '' }: Props = $props();

	let repoPath = $state(workspacePath || '');
	let output = $state('');
	let error = $state('');
	let loading = $state(false);
	let statusLines = $state<{ staging: string; worktree: string; path: string }[]>([]);
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

	async function checkGit() {
		if (!repoPath) return;
		try {
			const status = await GitStatus(repoPath);
			hasGit = !status.includes('failed to open repo');
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
			if (statusRaw === 'Working tree clean') {
				statusLines = [];
			} else {
				statusLines = statusRaw.split('\n').filter(Boolean).map(l => ({
					staging: l[0] || ' ',
					worktree: l[1] || ' ',
					path: l.slice(2).trim()
				}));
			}
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

	async function doBranchCreate() {
		if (!branchName) return;
		try { setOut(await GitBranchCreate(repoPath, branchName)); branchName = ''; await refresh(); }
		catch (e: any) { setErr(String(e)); }
	}

	async function doBranchCheckout(name: string) {
		try { setOut(await GitBranchCheckout(repoPath, name)); await refresh(); }
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

	async function doPull() {
		try { setOut(await GitPull(repoPath, remoteName, pullBranch, authToken)); await refresh(); }
		catch (e: any) { setErr(String(e)); }
	}

	async function doPush() {
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

	function statusIcon(staging: string, worktree: string): string {
		if (staging === 'M' || worktree === 'M') return 'M';
		if (staging === 'A' || worktree === 'A') return 'A';
		if (staging === 'D' || worktree === 'D') return 'D';
		if (staging === '?' || worktree === '?') return '?';
		return ' ';
	}
</script>

<div class="flex flex-col h-full w-[380px] shrink-0 border-l" style="background: var(--bg-primary); border-color: var(--border-primary)">
	<div class="flex items-center justify-between px-4 py-2.5 border-b" style="border-color: var(--border-primary)">
		<div class="flex items-center gap-2">
			<Icon name="git" size={15} style="color: var(--text-secondary)" />
			<h3 class="text-[11px] font-bold uppercase tracking-widest" style="color: var(--text-faint)">Git</h3>
		</div>
		<button type="button" onclick={onClose}
			class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-faint)" title="Close">
			<Icon name="x" size={12} />
		</button>
	</div>

	<div class="flex-1 overflow-y-auto text-[11px]" style="color: var(--text-secondary)">
		<div class="p-3 border-b" style="border-color: var(--border-primary)">
			<div class="flex gap-2">
				<input type="text" bind:value={repoPath} placeholder="Path do repositório"
					class="flex-1 px-2.5 py-1.5 rounded-lg text-[11px] font-mono outline-none transition-colors"
					style="background: var(--surface-input); color: var(--text-primary); border: 1px solid var(--border-primary)" />
				<button type="button" onclick={checkGit}
					class="px-2.5 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-colors hover:opacity-80"
					style="background: var(--bg-tertiary); color: var(--text-secondary); border: 1px solid var(--border-subtle)">Open</button>
			</div>
			{#if !hasGit && repoPath}
				<button type="button" onclick={doInit} disabled={loading}
					class="mt-2 w-full px-3 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-all disabled:opacity-30"
					style="background: var(--accent-primary); color: var(--accent-primary-fg); border: none">
					{loading ? 'Initializing...' : 'Init Repository'}
				</button>
			{/if}
		</div>

		{#if hasGit}
			<div class="flex border-b" style="border-color: var(--border-primary)">
				<button type="button" onclick={() => tab = 'status'}
					class="flex-1 px-3 py-2 text-[10px] font-semibold uppercase tracking-wider cursor-pointer transition-colors"
					style="color: {tab === 'status' ? 'var(--accent-primary)' : 'var(--text-faint)'}; border-bottom: {tab === 'status' ? '2px solid var(--accent-primary)' : '2px solid transparent'}">Status</button>
				<button type="button" onclick={() => { tab = 'diff'; doDiff(); }}
					class="flex-1 px-3 py-2 text-[10px] font-semibold uppercase tracking-wider cursor-pointer transition-colors"
					style="color: {tab === 'diff' ? 'var(--accent-primary)' : 'var(--text-faint)'}; border-bottom: {tab === 'diff' ? '2px solid var(--accent-primary)' : '2px solid transparent'}">Diff</button>
				<button type="button" onclick={() => tab = 'log'}
					class="flex-1 px-3 py-2 text-[10px] font-semibold uppercase tracking-wider cursor-pointer transition-colors"
					style="color: {tab === 'log' ? 'var(--accent-primary)' : 'var(--text-faint)'}; border-bottom: {tab === 'log' ? '2px solid var(--accent-primary)' : '2px solid transparent'}">Log</button>
			</div>

			<div class="p-3">
				{#if tab === 'status'}
					<div class="flex flex-col gap-3">
						<div class="flex items-center gap-2">
							<span class="text-[10px] uppercase tracking-wider font-semibold opacity-60">Branch</span>
							<span class="text-[12px] font-mono font-bold" style="color: var(--accent-primary)">{branches.find(b => b.current)?.name || '?'}</span>
						</div>

						{#if statusLines.length > 0}
							<div class="flex flex-col gap-0.5">
								{#each statusLines as line}
									<div class="flex items-center gap-2 px-2 py-1 rounded font-mono text-[10px]" style="background: var(--bg-tertiary)">
										<span class="w-4 text-center font-bold" style="color: {statusIcon(line.staging, line.worktree) === 'M' ? 'var(--status-warning)' : statusIcon(line.staging, line.worktree) === '?' ? 'var(--text-faint)' : 'var(--status-success)'}">
											{statusIcon(line.staging, line.worktree)}
										</span>
										<span>{line.path}</span>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-[11px] opacity-50 italic py-2">Working tree clean</p>
						{/if}

						<div class="flex gap-2">
							<button type="button" onclick={doAddAll} disabled={loading}
								class="flex-1 px-3 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-colors hover:opacity-80 disabled:opacity-30"
								style="background: var(--bg-tertiary); color: var(--text-secondary); border: 1px solid var(--border-subtle)">Add All</button>
						</div>

						<div class="flex gap-2">
							<input type="text" bind:value={commitMsg} placeholder="Commit message"
								class="flex-1 px-2.5 py-1.5 rounded-lg text-[11px] font-mono outline-none transition-colors"
								style="background: var(--surface-input); color: var(--text-primary); border: 1px solid var(--border-primary)" />
							<button type="button" onclick={doCommit} disabled={loading || !commitMsg}
								class="px-3 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-all disabled:opacity-30 hover:brightness-110"
								style="background: var(--accent-primary); color: var(--accent-primary-fg); border: none">Commit</button>
						</div>

						<div class="flex flex-col gap-2 pt-2 border-t" style="border-color: var(--border-primary)">
							<p class="text-[10px] uppercase tracking-wider font-semibold opacity-60">Branch</p>
							<div class="flex gap-2">
								<input type="text" bind:value={branchName} placeholder="New branch name"
									class="flex-1 px-2.5 py-1.5 rounded-lg text-[11px] font-mono outline-none transition-colors"
									style="background: var(--surface-input); color: var(--text-primary); border: 1px solid var(--border-primary)" />
								<button type="button" onclick={doBranchCreate} disabled={loading || !branchName}
									class="px-3 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-colors disabled:opacity-30 hover:opacity-80"
									style="background: var(--bg-tertiary); color: var(--text-secondary); border: 1px solid var(--border-subtle)">Create</button>
							</div>
							{#each branches as b}
								<button type="button" onclick={() => doBranchCheckout(b.name)}
									class="flex items-center gap-2 px-2.5 py-1.5 rounded-lg text-[11px] font-mono cursor-pointer transition-colors hover:bg-[var(--surface-hover)]"
									style="color: {b.current ? 'var(--accent-primary)' : 'var(--text-secondary)'}">
									<span class="text-[10px]">{b.current ? '*' : ' '}</span>
									<span>{b.name}</span>
								</button>
							{/each}
						</div>

						<div class="flex flex-col gap-2 pt-2 border-t" style="border-color: var(--border-primary)">
							<p class="text-[10px] uppercase tracking-wider font-semibold opacity-60">Remote</p>
							<div class="flex gap-2">
								<input type="text" bind:value={remoteUrl} placeholder="Remote URL"
									class="flex-1 px-2.5 py-1.5 rounded-lg text-[11px] font-mono outline-none transition-colors"
									style="background: var(--surface-input); color: var(--text-primary); border: 1px solid var(--border-primary)" />
								<button type="button" onclick={doRemoteAdd} disabled={loading || !remoteUrl}
									class="px-3 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-colors disabled:opacity-30 hover:opacity-80"
									style="background: var(--bg-tertiary); color: var(--text-secondary); border: 1px solid var(--border-subtle)">Add</button>
							</div>
							<div class="flex gap-2">
								<div class="flex-1 flex gap-1">
									<input type="text" bind:value={pullBranch} placeholder="Branch"
										class="w-20 px-2 py-1.5 rounded-lg text-[10px] font-mono outline-none transition-colors"
										style="background: var(--surface-input); color: var(--text-primary); border: 1px solid var(--border-primary)" />
									<input type="password" bind:value={authToken} placeholder="Token (optional)"
										class="flex-1 px-2 py-1.5 rounded-lg text-[10px] font-mono outline-none transition-colors"
										style="background: var(--surface-input); color: var(--text-primary); border: 1px solid var(--border-primary)" />
								</div>
								<button type="button" onclick={doPull} disabled={loading}
									class="px-3 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-colors hover:opacity-80 disabled:opacity-30"
									style="background: var(--bg-tertiary); color: var(--text-secondary); border: 1px solid var(--border-subtle)">Pull</button>
								<button type="button" onclick={doPush} disabled={loading}
									class="px-3 py-1.5 rounded-lg text-[10px] font-semibold cursor-pointer transition-colors hover:opacity-80 disabled:opacity-30"
									style="background: var(--bg-tertiary); color: var(--text-secondary); border: 1px solid var(--border-subtle)">Push</button>
							</div>
						</div>
					</div>

				{:else if tab === 'diff'}
					<pre class="text-[10px] font-mono leading-[1.6] whitespace-pre-wrap overflow-x-auto max-h-[500px]" style="color: var(--text-faint)">
						{diffOutput || 'No diff output'}
					</pre>

				{:else if tab === 'log'}
					<div class="flex flex-col gap-1">
						{#each commits as c}
							<div class="px-2.5 py-1.5 rounded-lg text-[11px] font-mono" style="background: var(--bg-tertiary)">
								<div class="flex items-center gap-2">
									<span class="text-[10px] font-bold" style="color: var(--accent-primary)">{c.hash}</span>
									<span class="flex-1 truncate">{c.msg}</span>
								</div>
								<p class="text-[9px] opacity-50 mt-0.5">{c.date}</p>
							</div>
						{:else}
							<p class="text-[11px] opacity-50 italic py-2">No commits yet</p>
						{/each}
					</div>
				{/if}
			</div>

			{#if output}
				<div class="mx-3 mb-3 px-3 py-2 rounded-lg text-[10px]" style="background: color-mix(in srgb, var(--status-success) 10%, transparent); color: var(--status-success)">{output}</div>
			{/if}
			{#if error}
				<div class="mx-3 mb-3 px-3 py-2 rounded-lg text-[10px]" style="background: color-mix(in srgb, var(--status-error) 10%, transparent); color: var(--status-error)">{error}</div>
			{/if}
		{/if}
	</div>
</div>
