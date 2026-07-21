<script lang="ts">
	import { onMount, onDestroy, tick } from 'svelte';
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import {
		DropdownMenu,
		DropdownContent,
	} from '$lib/components/ui/dropdown';
	import Select from '$lib/components/ui/Select.svelte';
	import { providersStore } from '$lib/stores/providers.svelte';
	import { CreateSessionWithConfig, GetSessionByID, GetSessionMessages, SendMessage, GetSessionWorkspaceSpec, SetSessionConfig, GetSessionContextInfo } from '../../../../wailsjs/go/main/App';
	import { EventsOn, EventsOff } from '../../../../wailsjs/runtime/runtime';
	import { toastStore } from '$lib/stores/toast.svelte';
	import MarkdownRenderer from './MarkdownRenderer.svelte';

	interface ModelOption {
		id: string; name: string; providerName?: string; health: number;
	}
	interface Message { role: 'user' | 'assistant'; content: string; }

	interface ChatPanelProps {
		sidebarOpen: boolean; activeWorkspace?: string;
		activeSessionID?: string; onToggleSidebar: () => void; class?: string;
	}

	let { sidebarOpen, activeWorkspace = '', activeSessionID = $bindable(''), onToggleSidebar, class: className }: ChatPanelProps = $props();

	let message = $state('');
	let isLoading = $state(false);
	let streamingContent = $state('');
	let selectedProviderFilter = $state('ALL');
	let modelSearchQuery = $state('');
	let thinkingContent = $state('');
	let showThinking = $state(false);

	let timerInterval: any = null;
	let elapsedSeconds = $state(0);
	let responseStartTime = $state(0);
	let responseEndTime = $state(0);

	function getModelHealth(name: string): number {
		let hash = 0;
		for (let i = 0; i < name.length; i++) {
			hash = name.charCodeAt(i) + ((hash << 5) - hash);
		}
		return Math.max(5, Math.abs(hash) % 96);
	}

	let allModels = $derived(() => {
		const list: ModelOption[] = [];
		const fallbackModels: ModelOption[] = [
			{ id: 'codestral-latest', name: 'codestral-latest', providerName: 'Mistral', health: 95 },
			{ id: 'claude-sonnet-4', name: 'claude-sonnet-4', providerName: 'Anthropic', health: 100 },
			{ id: 'gpt-4o', name: 'gpt-4o', providerName: 'OpenAI', health: 88 }
		];
		if (!providersStore.loaded || providersStore.providers.length === 0) return fallbackModels;
		providersStore.providers.forEach((p) => {
			const providerModels = providersStore.getModels(p.name);
			providerModels.forEach((m) => {
				list.push({ id: m.name, name: m.name, providerName: p.name, health: getModelHealth(m.name) });
			});
		});
		return list.length > 0 ? list : fallbackModels;
	});

	let selectedModel = $state<ModelOption>({ id: 'codestral-latest', name: 'codestral-latest', providerName: 'Mistral', health: 95 });

	let uniqueProviders = $derived(() => {
		const list = allModels();
		return Array.from(new Set(list.map((m) => m.providerName).filter(Boolean))) as string[];
	});

	let filteredModelsList = $derived(() => {
		return allModels().filter((m) => {
			const matchesProvider = selectedProviderFilter === 'ALL' || m.providerName === selectedProviderFilter;
			const matchesSearch = m.name.toLowerCase().includes(modelSearchQuery.toLowerCase());
			return matchesProvider && matchesSearch;
		});
	});

	$effect(() => {
		const list = allModels();
		if (list.length > 0 && !list.find((m) => m.id === selectedModel.id)) {
			selectedModel = list[0];
			saveSessionConfig();
		}
	});

	interface Message { role: 'user' | 'assistant'; content: string; }

	let messages = $state<Message[]>([]);
	let sessionID = $state<string>('');
	let chatContainer = $state<HTMLDivElement | null>(null);
	let autoScroll = $state(true);

	let cleanupDelta: (() => void) | null = null;
	let cleanupTurnEnd: (() => void) | null = null;
	let cleanupError: (() => void) | null = null;

	function startTimer() {
		elapsedSeconds = 0;
		responseStartTime = Date.now();
		responseEndTime = 0;
		if (timerInterval) clearInterval(timerInterval);
		timerInterval = setInterval(() => { elapsedSeconds++; }, 1000);
	}

	function stopTimer() {
		if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
		responseEndTime = Date.now();
	}

	function formatTime(seconds: number): string {
		const m = Math.floor(seconds / 60);
		const s = seconds % 60;
		return m > 0 ? `${m}m ${s}s` : `${s}s`;
	}

	function getTotalTime(): string {
		if (!responseStartTime) return '';
		const end = responseEndTime || Date.now();
		return formatTime(Math.floor((end - responseStartTime) / 1000));
	}

	async function copyResponse(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			toastStore.success('Copiado!');
		} catch { toastStore.error('Erro ao copiar'); }
	}

	async function loadSession(sessID: string) {
		sessionID = sessID;
		activeSessionID = sessID;
		try {
			const [sess, rawMessages, spec] = await Promise.all([
				GetSessionByID(sessID),
				GetSessionMessages(sessID),
				GetSessionWorkspaceSpec(sessID).catch(() => null),
			]);
			messages = (rawMessages || []).map((m: any) => ({ role: m.role, content: m.content }));
			if (sess.model) {
				const modelName = sess.model.includes('/') ? sess.model.split('/').pop()! : sess.model;
				selectedModel = { id: modelName, name: modelName, providerName: sess.provider || '', health: 95 };
			}
			if (sess.mode) selectedMode = sess.mode.toUpperCase() as 'ASK' | 'EDIT' | 'PLAN' | 'FULL';
			await saveSessionConfig();
			refreshContextInfo();
			if (spec) console.log('[ChatPanel] Workspace has Spec Wizard:', spec.name);
		} catch (e) { console.error('[ChatPanel] Failed to load session:', e); }
	}

	async function initSession(workspacePath: string, workerName = 'default-worker') {
		try {
			const wsPath = workspacePath || 'default-workspace';
			const sess = await CreateSessionWithConfig(wsPath, workerName, activeSessionID);
			await loadSession(sess.id);
		} catch (e) {
			console.error('[ChatPanel] Failed to create session:', e);
			sessionID = 'fallback-session';
		}
	}

	async function switchSession(sessID: string) {
		if (sessID === sessionID) return;
		await loadSession(sessID);
	}

	onMount(async () => {
		await providersStore.load();

		cleanupDelta = EventsOn('chat:delta', (data: any) => {
			if (!data || !isLoading) return;
			const content = data.content || '';
			if (content) {
				streamingContent = content;
				const lastIdx = messages.length - 1;
				if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
					messages[lastIdx] = { role: 'assistant', content: streamingContent };
					messages = [...messages];
				}
			}
		});

		cleanupTurnEnd = EventsOn('chat:turnEnd', (_data: any) => {
			isLoading = false;
			stopTimer();
			showThinking = false;
			const lastIdx = messages.length - 1;
			if (lastIdx >= 0 && messages[lastIdx].role === 'assistant' && streamingContent) {
				messages[lastIdx] = { role: 'assistant', content: streamingContent };
				messages = [...messages];
			}
		});

		cleanupError = EventsOn('chat:error', (data: any) => {
			isLoading = false;
			stopTimer();
			const errorMsg = data?.error || 'Unknown error';
			toastStore.error('Error', errorMsg);
			const lastIdx = messages.length - 1;
			if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
				messages[lastIdx] = { role: 'assistant', content: `Erro: ${errorMsg}` };
				messages = [...messages];
			}
			streamingContent = '';
		});

		if (!activeSessionID) await initSession(activeWorkspace);
	});

	// Auto-scroll when new messages or tokens arrive
	async function scrollToBottom() {
		if (!autoScroll) return;
		await tick();
		if (chatContainer) {
			chatContainer.scrollTop = chatContainer.scrollHeight;
		}
	}

	$effect(() => {
		if (messages.length > 0) scrollToBottom();
	});

	$effect(() => {
		if (streamingContent) scrollToBottom();
	});

	$effect(() => {
		if (isLoading) scrollToBottom();
	});

	function onScroll() {
		if (!chatContainer) return;
		const { scrollTop, scrollHeight, clientHeight } = chatContainer;
		autoScroll = scrollHeight - scrollTop - clientHeight < 50;
	}

	let prevSessionID = $state('');
	$effect(() => {
		if (activeSessionID && activeSessionID !== prevSessionID) {
			prevSessionID = activeSessionID;
			switchSession(activeSessionID);
		}
	});

	let prevWorkspace = $state('');
	$effect(() => {
		if (activeWorkspace !== prevWorkspace) {
			prevWorkspace = activeWorkspace;
			if (activeWorkspace) initSession(activeWorkspace);
		}
	});

	onDestroy(() => {
		if (cleanupDelta) cleanupDelta();
		if (cleanupTurnEnd) cleanupTurnEnd();
		if (cleanupError) cleanupError();
		if (timerInterval) clearInterval(timerInterval);
	});

	let contextLimit = $state(0);
	let contextUsed = $state(0);
	let contextSystemTokens = $state(0);
	let contextMessagesTokens = $state(0);
	let contextPercent = $derived(contextLimit > 0 ? Math.round((contextUsed / contextLimit) * 100) : 0);

	function formatContextSize(val: number): string {
		if (val <= 0) return '—';
		return val >= 1000 ? `${Math.round(val / 100) / 10}K` : String(val);
	}

	async function refreshContextInfo() {
		if (!sessionID) return;
		try {
			const info = await GetSessionContextInfo(sessionID);
			if (info) {
				contextLimit = info.context_limit || 0;
				contextUsed = info.context_used || 0;
				contextSystemTokens = info.system_tokens || 0;
				contextMessagesTokens = info.messages_tokens || 0;
			}
		} catch (e) {
			console.warn('[ChatPanel] Failed to refresh context info:', e);
		}
	}

	const modes = ['ASK', 'EDIT', 'PLAN', 'FULL'] as const;
	let selectedMode = $state<'ASK' | 'EDIT' | 'PLAN' | 'FULL'>('ASK');

	async function handleSend() {
		if (!message.trim() || isLoading || !sessionID) return;
		const userText = message.trim();
		message = '';
		messages = [...messages, { role: 'user', content: userText }];
		isLoading = true;
		streamingContent = '';
		thinkingContent = '';
		showThinking = true;
		startTimer();
		messages = [...messages, { role: 'assistant', content: '' }];

		try {
			const modeParam = selectedMode.toLowerCase();
			const modelString = selectedModel.providerName ? `${selectedModel.providerName}/${selectedModel.id}` : selectedModel.id;
			const response = await SendMessage(sessionID, userText, modelString, 'normal', modeParam);
			if (response && response.trim()) {
				const lastIdx = messages.length - 1;
				if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
					messages[lastIdx] = { role: 'assistant', content: response };
					messages = [...messages];
				}
			}
		} catch (e) {
			console.error('[ChatPanel] Error sending message:', e);
			const lastIdx = messages.length - 1;
			if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
				messages[lastIdx] = { role: 'assistant', content: `Erro: ${String(e)}` };
				messages = [...messages];
			}
		} finally {
			if (!streamingContent) { isLoading = false; stopTimer(); showThinking = false; }
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend(); }
	}

	function cycleMode() {
		const idx = modes.indexOf(selectedMode);
		selectedMode = modes[(idx + 1) % modes.length];
		saveSessionConfig();
	}

	let lastSavedKey = $state('');

	function saveSessionConfig() {
		if (!sessionID || !selectedModel) return;
		const key = `${selectedModel.id}|${selectedModel.providerName}|${selectedMode}`;
		if (key === lastSavedKey) return;
		lastSavedKey = key;
		SetSessionConfig(sessionID, selectedModel.id, selectedModel.providerName || '', selectedMode.toLowerCase(), '').catch(() => {});
	}
</script>

<div class={cn('flex h-full flex-1 flex-col min-w-0 bg-[var(--bg-primary)]', className)}>
	<div bind:this={chatContainer} onscroll={onScroll} class={cn("flex-1 overflow-y-auto px-4 py-4 flex flex-col gap-6 min-h-0", messages.length === 0 ? "justify-center items-center" : "")}>
		{#if messages.length === 0}
			<div class="flex flex-col items-center gap-4">
				<div class="flex items-center justify-center w-16 h-16 rounded-2xl border border-[var(--border-primary)] bg-[var(--bg-secondary)] opacity-20">
					<span class="text-2xl font-bold select-none" style="font-family: var(--font-display)">A</span>
				</div>
				<p class="text-xs font-medium tracking-[0.3em] uppercase select-none" style="color: var(--text-faint)">READY TO CODE</p>
			</div>
		{:else}
			{#each messages as msg, i}
				{#if msg.role === 'user'}
					<div class="self-end max-w-[75%] rounded-xl px-4 py-2.5 text-xs font-sans leading-relaxed bg-[var(--accent-primary)] text-[var(--accent-primary-fg)] shadow-sm">
						<p class="whitespace-pre-wrap">{msg.content}</p>
					</div>
				{:else}
					<div class="w-full">
						<MarkdownRenderer content={msg.content || ''} />
						<!-- Footer: time + copy -->
						{#if i === messages.length - 1 && !isLoading && msg.content}
							<div class="flex items-center gap-3 mt-2 text-[10px]" style="color: var(--text-faint)">
								<span>{getTotalTime()}</span>
								<button type="button" onclick={() => copyResponse(msg.content)}
									class="flex items-center gap-1 cursor-pointer hover:opacity-80 transition-opacity">
									<Icon name="copy" size={12} />
									Copiar
								</button>
							</div>
						{/if}
					</div>
				{/if}
			{/each}

			{#if isLoading}
				<!-- Thinking box -->
				{#if showThinking}
					<div class="w-full border border-dashed border-[var(--border-primary)] rounded-lg px-4 py-3 bg-[var(--bg-secondary)]/30">
						<div class="flex items-center gap-2 mb-2">
							<Icon name="loader" size={13} class="animate-spin text-[var(--accent-primary)]" />
							<span class="text-[10px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">Thinking... {formatTime(elapsedSeconds)}</span>
						</div>
						{#if thinkingContent}
							<div class="text-[11px] leading-relaxed whitespace-pre-wrap" style="color: var(--text-muted)">{thinkingContent}</div>
						{/if}
					</div>
				{/if}

				<!-- Loading indicator -->
				<div class="flex items-center gap-2 text-xs" style="color: var(--text-faint)">
					<Icon name="loader" size={13} class="animate-spin text-[var(--accent-primary)]" />
					<span>Ada is typing... {formatTime(elapsedSeconds)}</span>
				</div>
			{/if}
		{/if}
	</div>

	<div class={cn('mx-4 mb-4 rounded-xl border border-[var(--border-subtle)] bg-[var(--surface-input)] transition-colors focus-within:border-[var(--border-hover)]')}>
		<div class="px-4 pt-3 pb-1">
			<textarea bind:value={message} onkeydown={handleKeydown}
				placeholder="message ada..."
				rows="1"
				class="flex-1 w-full resize-none bg-transparent border-none outline-none text-sm leading-relaxed placeholder:opacity-40 max-h-40 min-h-[24px]"
				style="color: var(--text-primary)">
			</textarea>
		</div>
		<div class="flex items-center justify-between px-2 pb-2 pt-0.5">
			<div class="flex items-center gap-0.5">
				<button type="button" onclick={onToggleSidebar} title={sidebarOpen ? 'Zen mode (hide sidebar)' : 'Show sidebar'}
					class="flex items-center justify-center w-8 h-8 rounded-lg transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-muted)">
					{#if sidebarOpen}
						<svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M9 3v18"/><path d="m16 15-3-3 3-3"/></svg>
					{:else}
						<svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M9 3v18"/><path d="m14 9 3 3-3 3"/></svg>
					{/if}
				</button>
				<button type="button" title="History" class="flex items-center justify-center w-8 h-8 rounded-lg transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-muted)"><Icon name="history" size={15} /></button>
				<button type="button" title="Attach file" class="flex items-center justify-center w-8 h-8 rounded-lg transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-muted)"><Icon name="attachment" size={15} /></button>
				<button type="button" title="Logs" class="flex items-center justify-center w-8 h-8 rounded-lg transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-muted)"><Icon name="log" size={15} /></button>
			</div>
			<div class="flex items-center gap-1.5">
				<DropdownMenu>
					{#snippet trigger({ toggle })}
						<button type="button" onclick={async (e) => { e.stopPropagation(); await refreshContextInfo(); toggle(); }}
							class="flex items-center justify-center w-8 h-8 rounded-full transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-muted)" title="Context windows: {contextPercent}% used">
							<svg class="w-4 h-4" viewBox="0 0 20 20"><circle cx="10" cy="10" r="7.5" fill="none" stroke="currentColor" stroke-width="2.5" class="opacity-15"/><circle cx="10" cy="10" r="7.5" fill="none" stroke="#3b82f6" stroke-width="2.5" stroke-dasharray="47" stroke-dashoffset={47 - (47 * contextPercent) / 100} transform="rotate(-90 10 10)" stroke-linecap="round"/></svg>
						</button>
					{/snippet}
					{#snippet content()}
						<DropdownContent align="end" class="bottom-full top-auto mb-1.5 mt-0 w-[300px] p-4 bg-zinc-900 border border-zinc-800 rounded-xl shadow-2xl flex flex-col gap-3">
							<div class="flex justify-between items-center text-xs font-sans text-zinc-400"><span class="font-bold text-zinc-200 text-[13px]">Context window</span><span class="font-mono text-[11px] text-zinc-400">{formatContextSize(contextUsed)} / {formatContextSize(contextLimit)} ({contextPercent}%)</span></div>
							<div class="w-full bg-zinc-800 h-2 rounded-full overflow-hidden flex"><div class="bg-blue-500 h-full rounded-full" style="width: {contextPercent}%"></div></div>
							{#if contextLimit > 0}
								<div class="flex flex-col gap-2 font-sans">
									<div class="flex items-center justify-between text-xs text-zinc-300"><div class="flex items-center gap-2"><span class="w-2 h-2 rounded-full inline-block shrink-0 bg-zinc-600"></span><span class="text-zinc-400">Messages</span></div><span class="font-mono text-zinc-300">{formatContextSize(contextMessagesTokens)}</span></div>
									<div class="flex items-center justify-between text-xs text-zinc-300"><div class="flex items-center gap-2"><span class="w-2 h-2 rounded-full inline-block shrink-0 bg-zinc-500"></span><span class="text-zinc-400">System prompt</span></div><span class="font-mono text-zinc-300">{formatContextSize(contextSystemTokens)}</span></div>
								</div>
							{:else}
								<p class="text-[11px] text-zinc-500">Configure o contexto do modelo no provider para exibir o uso.</p>
							{/if}
						</DropdownContent>
					{/snippet}
				</DropdownMenu>
				<button type="button" class="flex items-center justify-center w-8 h-8 rounded-full transition-colors cursor-pointer {isLoading ? 'animate-spin' : 'hover:bg-[var(--surface-hover)]'}" style="color: var(--text-muted)" title={isLoading ? 'Generating...' : 'Idle'}>
					<Icon name="loader" size={15} />
				</button>
				<button type="button" onclick={cycleMode}
					class="flex items-center justify-center h-7 px-3 rounded-md text-[11px] font-semibold tracking-wider cursor-pointer border transition-colors {selectedMode === 'FULL' ? 'bg-orange-500/10 hover:bg-orange-500/20' : 'border-[var(--border-primary)] hover:bg-[var(--surface-hover)] hover:border-[var(--border-hover)]'}"
					style={selectedMode === 'FULL' ? 'color: #f97316; border-color: rgba(249, 115, 22, 0.4);' : 'color: var(--text-secondary)'}>
					{selectedMode}
				</button>
				<DropdownMenu>
					{#snippet trigger({ toggle })}
						<button type="button" onclick={(e) => { e.stopPropagation(); toggle(); }}
							class="flex items-center gap-1.5 h-7 px-3 rounded-md text-[11px] font-mono cursor-pointer border transition-colors border-[var(--border-primary)] hover:bg-[var(--surface-hover)] hover:border-[var(--border-hover)]" style="color: var(--text-secondary)">
							<span class="w-2 h-2 rounded-full inline-block shrink-0 shadow-sm border border-black/20" style="background-color: hsl({(selectedModel.health / 100) * 120}, 85%, 45%);"></span>
							{selectedModel.name}
							<Icon name="chevron-down" size={11} />
						</button>
					{/snippet}
					{#snippet content({ close })}
						<DropdownContent align="end" class="bottom-full top-auto mb-1.5 mt-0 w-[320px] h-[300px] p-2.5 bg-zinc-900 border border-zinc-800 rounded-lg shadow-xl flex flex-col gap-2.5">
							<div class="flex flex-col gap-1"><span class="text-[9px] uppercase font-bold tracking-wider text-zinc-500 font-sans">Provider</span>
								<Select bind:value={selectedProviderFilter} options={[{ label: 'All Providers', value: 'ALL' }, ...uniqueProviders().map((p) => ({ label: p, value: p }))]} class="text-xs" />
							</div>
							<div class="flex flex-col gap-1">
								<label for="model-search" class="text-[9px] uppercase font-bold tracking-wider text-zinc-500 font-sans">Search Models</label>
								<div class="relative">
									<input id="model-search" type="text" bind:value={modelSearchQuery} placeholder="Type to filter..." class="w-full bg-zinc-800 border border-zinc-700 hover:border-zinc-600 text-xs rounded pl-2.5 pr-7 py-1.5 text-zinc-200 placeholder-zinc-500 outline-none transition-colors" />
									{#if modelSearchQuery}
										<button type="button" onclick={() => modelSearchQuery = ''} class="absolute right-2 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-200 cursor-pointer" title="Clear search">
											<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
										</button>
									{/if}
								</div>
							</div>
							<div class="flex flex-col gap-1 flex-1 min-h-0">
								<span class="text-[9px] uppercase font-bold tracking-wider text-zinc-500 font-sans">Available Models</span>
								<div class="flex-1 overflow-y-auto pr-0.5 flex flex-col gap-1">
									{#if filteredModelsList().length === 0}<div class="text-xs text-zinc-500 py-4 text-center font-sans">No models found</div>
									{:else}
										{#each filteredModelsList() as model}
											{@const hue = (model.health / 100) * 120}
											<button type="button" onclick={() => { selectedModel = model; saveSessionConfig(); close(); }}
												class={cn('relative flex w-full items-center justify-between rounded px-2.5 py-1.5 text-xs cursor-pointer font-mono transition-all border border-transparent', model.id === selectedModel.id ? 'bg-zinc-800/80 border-zinc-700 text-zinc-100 font-medium' : 'hover:bg-zinc-800/40 text-zinc-300')}>
												<div class="flex items-center gap-2 min-w-0">
													<span class="w-2.5 h-2.5 rounded-full inline-block shrink-0 shadow-sm border border-black/20" style="background-color: hsl({hue}, 85%, 45%);"></span>
													<span class="truncate">{model.name}</span>
												</div>
												<span class="text-[9px] opacity-40 font-sans tracking-wide uppercase px-1 rounded bg-zinc-800 text-zinc-400 shrink-0">{model.providerName}</span>
											</button>
										{/each}
									{/if}
								</div>
							</div>
						</DropdownContent>
					{/snippet}
				</DropdownMenu>
				<button type="button" onclick={handleSend} disabled={isLoading}
					class="flex items-center justify-center w-8 h-8 rounded-lg cursor-pointer transition-all duration-150 disabled:opacity-30 disabled:cursor-not-allowed hover:brightness-125 active:scale-95"
					style="background-color: var(--accent-primary); color: var(--accent-primary-fg)" title="Send message">
					<Icon name="arrowUp" size={16} />
				</button>
			</div>
		</div>
	</div>
</div>
