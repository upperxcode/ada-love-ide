<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import {
		DropdownMenu,
		DropdownContent,
	} from '$lib/components/ui/dropdown';
	import Select from '$lib/components/ui/Select.svelte';
	import { providersStore } from '$lib/stores/providers.svelte';
	import { CreateSession, SendMessage } from '../../../../wailsjs/go/main/App';
	import { EventsOn, EventsOff } from '../../../../wailsjs/runtime/runtime';

	interface ModelOption {
		id: string;
		name: string;
		providerName?: string;
		health: number;
	}

	interface ChatPanelProps {
		sidebarOpen: boolean;
		onToggleSidebar: () => void;
		class?: string;
	}

	let { sidebarOpen, onToggleSidebar, class: className }: ChatPanelProps = $props();

	let message = $state('');
	let isLoading = $state(false);
	let streamingContent = $state('');

	// Provider filter & search states
	let selectedProviderFilter = $state('ALL');
	let modelSearchQuery = $state('');

	// Generate a deterministic health percentage (0-100) based on model name
	function getModelHealth(name: string): number {
		let hash = 0;
		for (let i = 0; i < name.length; i++) {
			hash = name.charCodeAt(i) + ((hash << 5) - hash);
		}
		// Deterministic value between 5 and 100 to feel realistic
		return 5 + Math.abs(hash % 96);
	}

	// Derived: List of all unified models
	let allModels = $derived(() => {
		const list: ModelOption[] = [];
		const fallbackModels: ModelOption[] = [
			{ id: 'codestral-latest', name: 'codestral-latest', providerName: 'Mistral', health: 95 },
			{ id: 'claude-sonnet-4', name: 'claude-sonnet-4', providerName: 'Anthropic', health: 100 },
			{ id: 'gpt-4o', name: 'gpt-4o', providerName: 'OpenAI', health: 88 }
		];

		if (!providersStore.loaded || providersStore.providers.length === 0) {
			return fallbackModels;
		}

		providersStore.providers.forEach((p) => {
			const providerModels = providersStore.getModels(p.name);
			providerModels.forEach((m) => {
				list.push({
					id: m.name,
					name: m.name,
					providerName: p.name,
					health: getModelHealth(m.name)
				});
			});
		});

		return list.length > 0 ? list : fallbackModels;
	});

	let selectedModel = $state<ModelOption>({
		id: 'codestral-latest',
		name: 'codestral-latest',
		providerName: 'Mistral',
		health: 95
	});

	// Derived: Unique providers list for selection
	let uniqueProviders = $derived(() => {
		const list = allModels();
		const set = new Set(list.map((m) => m.providerName).filter(Boolean));
		return Array.from(set) as string[];
	});

	// Derived: Filtered models based on selected provider and search query
	let filteredModelsList = $derived(() => {
		return allModels().filter((m) => {
			const matchesProvider = selectedProviderFilter === 'ALL' || m.providerName === selectedProviderFilter;
			const matchesSearch = m.name.toLowerCase().includes(modelSearchQuery.toLowerCase());
			return matchesProvider && matchesSearch;
		});
	});

	// Automatically select the first available model when the list updates and selected doesn't exist
	$effect(() => {
		const list = allModels();
		if (list.length > 0 && !list.find((m) => m.id === selectedModel.id)) {
			selectedModel = list[0];
		}
	});

	interface Message {
		role: 'user' | 'assistant';
		content: string;
	}

	let messages = $state<Message[]>([]);
	let sessionID = $state<string>('');

	let cleanupDelta: (() => void) | null = null;
	let cleanupTurnEnd: (() => void) | null = null;
	let cleanupError: (() => void) | null = null;

	onMount(async () => {
		await providersStore.load();
		try {
			const sess = await CreateSession('default-workspace', 'default-worker');
			sessionID = sess.id;
		} catch (e) {
			console.error('[ChatPanel] Failed to create session:', e);
			sessionID = 'fallback-session';
		}

		// Listen for streaming delta events from the Wails backend
		cleanupDelta = EventsOn('chat:delta', (data: any) => {
			if (!data || !isLoading) return;
			const content = data.content || '';
			if (content) {
				streamingContent = content;
				// Update the last assistant message in-place
				const lastIdx = messages.length - 1;
				if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
					messages[lastIdx] = { role: 'assistant', content: streamingContent };
					messages = [...messages]; // trigger reactivity
				}
			}
		});

		// Listen for turn end events to finalize loading state
		cleanupTurnEnd = EventsOn('chat:turnEnd', (_data: any) => {
			isLoading = false;
			// Ensure the last assistant message has the final content
			const lastIdx = messages.length - 1;
			if (lastIdx >= 0 && messages[lastIdx].role === 'assistant' && streamingContent) {
				messages[lastIdx] = { role: 'assistant', content: streamingContent };
				messages = [...messages];
			}
			streamingContent = '';
		});

		// Listen for error events
		cleanupError = EventsOn('chat:error', (data: any) => {
			isLoading = false;
			const errorMsg = data?.error || 'Unknown error';
			const lastIdx = messages.length - 1;
			if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
				messages[lastIdx] = { role: 'assistant', content: `Erro: ${errorMsg}` };
				messages = [...messages];
			}
			streamingContent = '';
		});
	});

	onDestroy(() => {
		if (cleanupDelta) cleanupDelta();
		if (cleanupTurnEnd) cleanupTurnEnd();
		if (cleanupError) cleanupError();
	});

	// Context Window token usage data
	const contextLimit = 262144; // 262K
	const contextUsed = 153700;  // 153.7K
	const contextPercent = 58.7;
	const contextDetails = [
		{ name: 'Messages', percentage: 90, color: '#3b82f6' },
		{ name: 'System tools', percentage: 7.3, color: '#60a5fa' },
		{ name: 'Skills', percentage: 1.5, color: '#2563eb' },
		{ name: 'MCP tools', percentage: 0.8, color: '#1d4ed8' },
		{ name: 'System prompt', percentage: 0.4, color: '#1e40af' },
		{ name: 'Meta context', percentage: 0, color: '#172554' }
	];

	// ── Chat modes ──
	const modes = ['ASK', 'EDIT', 'PLAN', 'FULL'] as const;
	let selectedMode = $state<'ASK' | 'EDIT' | 'PLAN' | 'FULL'>('ASK');

	async function handleSend() {
		if (!message.trim() || isLoading) return;
		const userText = message.trim();
		message = '';

		messages = [...messages, { role: 'user', content: userText }];
		isLoading = true;
		streamingContent = '';

		// Add a placeholder assistant message that will be updated by streaming
		messages = [...messages, { role: 'assistant', content: '' }];
		const assistantIdx = messages.length - 1;

		try {
			const modeParam = selectedMode.toLowerCase();
			const response = await SendMessage(
				sessionID,
				userText,
				selectedModel.id,
				'normal',
				modeParam
			);

			// If we got a non-empty response (non-streaming path), use it directly
			if (response && response.trim()) {
				messages[assistantIdx] = { role: 'assistant', content: response };
				messages = [...messages]; // trigger reactivity
			} else if (streamingContent) {
				// Streaming already populated via deltas
				messages[assistantIdx] = { role: 'assistant', content: streamingContent };
				messages = [...messages];
			}
			// If both are empty, the chat:delta events will fill it via the listener
		} catch (e) {
			console.error('[ChatPanel] Error sending message:', e);
			messages[assistantIdx] = {
				role: 'assistant',
				content: `Erro ao enviar mensagem: ${String(e)}`
			};
			messages = [...messages];
		} finally {
			// Note: isLoading will be set to false by chat:turnEnd event
			// But as a safety net, if no streaming, set it here after a delay
			if (!streamingContent) {
				isLoading = false;
			}
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSend();
		}
	}

	function cycleMode() {
		const idx = modes.indexOf(selectedMode);
		selectedMode = modes[(idx + 1) % modes.length];
	}
</script>

<div
	class={cn(
		'flex h-full flex-1 flex-col min-w-0',
		'bg-[var(--bg-primary)]',
		className
	)}
>
	<!-- ── Chat Body (Scroll Area) ── -->
	<div class={cn(
		"flex-1 overflow-y-auto px-6 py-6 flex flex-col gap-4 min-h-0",
		messages.length === 0 ? "justify-center items-center" : ""
	)}>
		{#if messages.length === 0}
			<!-- ── Empty State — Centered Logo ── -->
			<div class="flex flex-col items-center gap-4">
				<!-- "A" logo box -->
				<div
					class={cn(
						'flex items-center justify-center',
						'w-16 h-16 rounded-2xl',
						'border border-[var(--border-primary)]',
						'bg-[var(--bg-secondary)]',
						'opacity-20'
					)}
				>
					<span
						class="text-2xl font-bold select-none"
						style="font-family: var(--font-display)"
					>
						A
					</span>
				</div>

				<p
					class="text-xs font-medium tracking-[0.3em] uppercase select-none"
					style="color: var(--text-faint)"
				>
					READY TO CODE
				</p>
			</div>
		{:else}
			{#each messages as msg}
				<div
					class={cn(
						'flex flex-col max-w-[85%] rounded-xl px-4 py-2.5 text-xs font-sans leading-relaxed transition-all shadow-sm border',
						msg.role === 'user'
							? 'self-end bg-[var(--accent-primary)] text-[var(--accent-primary-fg)] border-transparent rounded-tr-none'
							: 'self-start bg-[var(--bg-secondary)] border-[var(--border-primary)] text-[var(--text-primary)] rounded-tl-none'
					)}
				>
					<p class="whitespace-pre-wrap">{msg.content}</p>
				</div>
			{/each}

			{#if isLoading}
				<div class="self-start bg-[var(--bg-secondary)] border border-[var(--border-primary)] text-[var(--text-primary)] rounded-xl rounded-tl-none px-4 py-2.5 text-xs flex items-center gap-2 shadow-sm font-sans">
					<Icon name="loader" size={13} class="animate-spin text-[var(--accent-primary)]" />
					<span class="text-[11px] text-zinc-400">Ada is typing...</span>
				</div>
			{/if}
		{/if}
	</div>

	<!-- ── Unified Bottom Panel: input + toolbar as one visual block ── -->
	<div
		class={cn(
			'mx-4 mb-4 rounded-xl',
			'border border-[var(--border-subtle)]',
			'bg-[var(--surface-input)]',
			'transition-colors',
			'focus-within:border-[var(--border-hover)]'
		)}
	>
		<!-- Textarea row -->
		<div class="px-4 pt-3 pb-1">
			<textarea
				bind:value={message}
				onkeydown={handleKeydown}
				placeholder="message ada..."
				rows="1"
				class={cn(
					'flex-1 w-full resize-none bg-transparent border-none outline-none',
					'text-sm leading-relaxed placeholder:opacity-40',
					'max-h-40 min-h-[24px]'
				)}
				style="color: var(--text-primary)"
			></textarea>
		</div>

		<!-- Toolbar row — integrated inside the same card -->
		<div class="flex items-center justify-between px-2 pb-2 pt-0.5">
			<!-- ── Left: Utility Icons ── -->
			<div class="flex items-center gap-0.5">
				<!-- Zen toggle: sidebar show/hide -->
				<button
					type="button"
					onclick={onToggleSidebar}
					title={sidebarOpen ? 'Zen mode (hide sidebar)' : 'Show sidebar'}
					class={cn(
						'flex items-center justify-center w-8 h-8 rounded-lg',
						'transition-colors cursor-pointer',
						'hover:bg-[var(--surface-hover)]'
					)}
					style="color: var(--text-muted)"
				>
					{#if sidebarOpen}
						<!-- Panel-left-close: sidebar visible, click to hide -->
						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="15"
							height="15"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<rect x="3" y="3" width="18" height="18" rx="2" />
							<path d="M9 3v18" />
							<path d="m16 15-3-3 3-3" />
						</svg>
					{:else}
						<!-- Panel-left-open: sidebar hidden, click to show -->
						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="15"
							height="15"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<rect x="3" y="3" width="18" height="18" rx="2" />
							<path d="M9 3v18" />
							<path d="m14 9 3 3-3 3" />
						</svg>
					{/if}
				</button>

				<button
					type="button"
					title="History"
					class={cn(
						'flex items-center justify-center w-8 h-8 rounded-lg',
						'transition-colors cursor-pointer',
						'hover:bg-[var(--surface-hover)]'
					)}
					style="color: var(--text-muted)"
				>
					<Icon name="history" size={15} />
				</button>

				<button
					type="button"
					title="Attach file"
					class={cn(
						'flex items-center justify-center w-8 h-8 rounded-lg',
						'transition-colors cursor-pointer',
						'hover:bg-[var(--surface-hover)]'
					)}
					style="color: var(--text-muted)"
				>
					<Icon name="attachment" size={15} />
				</button>

				<button
					type="button"
					title="Logs"
					class={cn(
						'flex items-center justify-center w-8 h-8 rounded-lg',
						'transition-colors cursor-pointer',
						'hover:bg-[var(--surface-hover)]'
					)}
					style="color: var(--text-muted)"
				>
					<Icon name="log" size={15} />
				</button>
			</div>

			<!-- ── Right: Status · Mode · Model · Send ── -->
			<div class="flex items-center gap-1.5">
				<!-- Context Window Token indicator -->
				<DropdownMenu>
					{#snippet trigger({ toggle })}
						<button
							type="button"
							onclick={(e) => { e.stopPropagation(); toggle(); }}
							class={cn(
								'flex items-center justify-center w-8 h-8 rounded-full',
								'transition-colors cursor-pointer hover:bg-[var(--surface-hover)]'
							)}
							style="color: var(--text-muted)"
							title="Context windows: {contextPercent}% used"
						>
							<svg class="w-4 h-4" viewBox="0 0 20 20">
								<!-- Background circle -->
								<circle cx="10" cy="10" r="7.5" fill="none" stroke="currentColor" stroke-width="2.5" class="opacity-15" />
								<!-- Progress circle -->
								<circle cx="10" cy="10" r="7.5" fill="none" stroke="#3b82f6" stroke-width="2.5"
										stroke-dasharray="47" stroke-dashoffset={47 - (47 * contextPercent) / 100}
										transform="rotate(-90 10 10)" stroke-linecap="round" />
							</svg>
						</button>
					{/snippet}

					{#snippet content()}
						<DropdownContent align="end" class="bottom-full top-auto mb-1.5 mt-0 w-[300px] p-4 bg-zinc-900 border border-zinc-800 rounded-xl shadow-2xl flex flex-col gap-3">
							<!-- Header -->
							<div class="flex justify-between items-center text-xs font-sans text-zinc-400">
								<span class="font-bold text-zinc-200 text-[13px]">Context windows</span>
								<span class="font-mono text-[11px] text-zinc-400">153.7K/262K ({contextPercent}%)</span>
							</div>

							<!-- Horizontal Progress Bar -->
							<div class="w-full bg-zinc-800 h-2 rounded-full overflow-hidden flex">
								<div class="bg-blue-500 h-full rounded-full" style="width: {contextPercent}%"></div>
							</div>

							<!-- Categories List -->
							<div class="flex flex-col gap-2 font-sans">
								{#each contextDetails as item}
									<div class="flex items-center justify-between text-xs text-zinc-300">
										<div class="flex items-center gap-2">
											<span class="w-2 h-2 rounded-full inline-block shrink-0" style="background-color: {item.color}"></span>
											<span class="text-zinc-400">{item.name}</span>
										</div>
										<span class="font-mono text-zinc-300">{item.percentage}%</span>
									</div>
								{/each}
							</div>
						</DropdownContent>
					{/snippet}
				</DropdownMenu>

				<!-- Loading indicator -->
				<button
					type="button"
					class={cn(
						'flex items-center justify-center w-8 h-8 rounded-full',
						'transition-colors cursor-pointer',
						isLoading ? 'animate-spin' : 'hover:bg-[var(--surface-hover)]'
					)}
					style="color: var(--text-muted)"
					title={isLoading ? 'Generating...' : 'Idle'}
				>
					<Icon name="loader" size={15} />
				</button>

				<!-- Mode Selector (ASK / EDIT / PLAN / FULL) -->
				<button
					type="button"
					onclick={cycleMode}
					class={cn(
						'flex items-center justify-center h-7 px-3 rounded-md',
						'text-[11px] font-semibold tracking-wider cursor-pointer',
						'border transition-colors',
						selectedMode === 'FULL'
							? 'bg-orange-500/10 hover:bg-orange-500/20'
							: 'border-[var(--border-primary)] hover:bg-[var(--surface-hover)] hover:border-[var(--border-hover)]'
					)}
					style={selectedMode === 'FULL' ? 'color: #f97316; border-color: rgba(249, 115, 22, 0.4);' : 'color: var(--text-secondary)'}
				>
					{selectedMode}
				</button>

				<!-- Model Dropdown -->
				<DropdownMenu>
					{#snippet trigger({ toggle })}
						<button
							type="button"
							onclick={(e) => { e.stopPropagation(); toggle(); }}
							class={cn(
								'flex items-center gap-1.5 h-7 px-3 rounded-md',
								'text-[11px] font-mono cursor-pointer',
								'border transition-colors',
								'border-[var(--border-primary)]',
								'hover:bg-[var(--surface-hover)] hover:border-[var(--border-hover)]'
							)}
							style="color: var(--text-secondary)"
						>
							<span
								class="w-2 h-2 rounded-full inline-block shrink-0 shadow-sm border border-black/20"
								style="background-color: hsl({(selectedModel.health / 100) * 120}, 85%, 45%);"
							></span>
							{selectedModel.name}
							<Icon name="chevron-down" size={11} />
						</button>
					{/snippet}

					{#snippet content({ close })}
						<DropdownContent align="end" class="bottom-full top-auto mb-1.5 mt-0 w-[320px] h-[300px] p-2.5 bg-zinc-900 border border-zinc-800 rounded-lg shadow-xl flex flex-col gap-2.5">
							<!-- Provider Selector -->
							<div class="flex flex-col gap-1">
								<span class="text-[9px] uppercase font-bold tracking-wider text-zinc-500 font-sans">Provider</span>
								<Select
									bind:value={selectedProviderFilter}
									options={[
										{ label: 'All Providers', value: 'ALL' },
										...uniqueProviders().map((p) => ({ label: p, value: p }))
									]}
									class="text-xs"
								/>
							</div>

							<!-- Search Input -->
							<div class="flex flex-col gap-1">
								<label for="model-search" class="text-[9px] uppercase font-bold tracking-wider text-zinc-500 font-sans">Search Models</label>
								<div class="relative">
									<input
										id="model-search"
										type="text"
										bind:value={modelSearchQuery}
										placeholder="Type to filter..."
										class="w-full bg-zinc-800 border border-zinc-700 hover:border-zinc-600 text-xs rounded pl-2.5 pr-7 py-1.5 text-zinc-200 placeholder-zinc-500 outline-none transition-colors"
									/>
									{#if modelSearchQuery}
										<button
											type="button"
											onclick={() => modelSearchQuery = ''}
											class="absolute right-2 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-200 cursor-pointer"
											title="Clear search"
										>
											<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
												<line x1="18" y1="6" x2="6" y2="18"></line>
												<line x1="6" y1="6" x2="18" y2="18"></line>
											</svg>
										</button>
									{/if}
								</div>
							</div>

							<!-- Models List -->
							<div class="flex flex-col gap-1 flex-1 min-h-0">
								<span class="text-[9px] uppercase font-bold tracking-wider text-zinc-500 font-sans">Available Models</span>
								<div class="flex-1 overflow-y-auto pr-0.5 flex flex-col gap-1">
									{#if filteredModelsList().length === 0}
										<div class="text-xs text-zinc-500 py-4 text-center font-sans">No models found</div>
									{:else}
										{#each filteredModelsList() as model}
											{@const hue = (model.health / 100) * 120}
											<button
												type="button"
												class={cn(
													'relative flex w-full items-center justify-between rounded px-2.5 py-1.5 text-xs cursor-pointer',
													'font-mono transition-all border border-transparent',
													model.id === selectedModel.id
														? 'bg-zinc-800/80 border-zinc-700 text-zinc-100 font-medium'
														: 'hover:bg-zinc-800/40 text-zinc-300'
												)}
												onclick={() => {
													selectedModel = model;
													close();
												}}
											>
												<div class="flex items-center gap-2 min-w-0">
													<!-- Bolinha de health -->
													<span
														class="w-2.5 h-2.5 rounded-full inline-block shrink-0 shadow-sm border border-black/20"
														style="background-color: hsl({hue}, 85%, 45%);"
														title="Health: {model.health}%"
													></span>
													<span class="truncate">
														{model.name}
													</span>
												</div>
												<span class="text-[9px] opacity-40 font-sans tracking-wide uppercase px-1 rounded bg-zinc-800 text-zinc-400 shrink-0">
													{model.providerName}
												</span>
											</button>
										{/each}
									{/if}
								</div>
							</div>
						</DropdownContent>
					{/snippet}
				</DropdownMenu>

				<!-- Send Button -->
					<button
						type="button"
						onclick={handleSend}
						disabled={isLoading}
						class={cn(
							'flex items-center justify-center w-8 h-8 rounded-lg',
							'cursor-pointer transition-all duration-150',
							'disabled:opacity-30 disabled:cursor-not-allowed',
							'hover:brightness-125 active:scale-95'
						)}
						style="background-color: var(--accent-primary); color: var(--accent-primary-fg)"
						title="Send message"
					>
						<Icon name="arrowUp" size={16} />
					</button>
				</div>
			</div>
		</div>
	</div>
