<script lang="ts">
	import { onMount, onDestroy, tick } from 'svelte';
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import {
		DropdownMenu,
		DropdownContent,
	} from '$lib/components/ui/dropdown';
	import Select from '$lib/components/ui/Select.svelte';
	import { theme } from '$lib/stores/theme.svelte';
	import {
		colorThemes,
		fontThemes,
		fontSizePresets,
		type FontSizePreset,
	} from '$lib/themes/definitions';
	import { providersStore } from '$lib/stores/providers.svelte';
	import { CreateSessionWithConfig, GetSessionByID, GetSessionMessages, SendMessage, RespondPermission, GetSessionWorkspaceSpec, SetSessionConfig, GetSessionContextInfo } from '../../../../wailsjs/go/main/App';
	import { EventsOn, EventsOff } from '../../../../wailsjs/runtime/runtime';
	import { toastStore } from '$lib/stores/toast.svelte';
	import MessageBubble from './MessageBubble.svelte';
	import PermissionDialog from './PermissionDialog.svelte';
	import type { ChatMessage, ActionLog, ThinkingSection } from './chat-types';
	import { parseActionPayload } from './chat-types';

	interface ModelOption {
		id: string; name: string; providerName?: string; health: number;
	}

	interface ChatPanelProps {
		sidebarOpen: boolean; activeWorkspace?: string;
		activeSessionID?: string; onToggleSidebar: () => void; 
		class?: string;
	}

	let { sidebarOpen, activeWorkspace = '', activeSessionID = $bindable(''), onToggleSidebar, class: className }: ChatPanelProps = $props();

	// ── Chat theme application ──
	// Builds a reactive CSS string for the root div's style attribute.
	// Also builds CSS override rules because Tailwind v4 generates
	// text-* classes with HARDCODED values (e.g., .text-xs { font-size: 11px })
	// instead of referencing CSS variables. So we must override them directly.
	let chatStyle = $state('');

	// ── Style override engine ──
	// Tailwind v4 generates HARDCODED font-size values (e.g. .text-xs { font-size: 11px }).
	// font-family works via inheritance because most elements don't have explicit font-family.
	// font-size DOESN'T work via inheritance because EVERY element has text-xs/sm/etc.
	// So we inject a <style> element into <head> with !important overrides.
	let _styleEl: HTMLStyleElement | null = null;

	function applyChatThemeOverrides(fs: FontSizePreset, ft: { sans: string; mono: string }) {
		// Map each Tailwind text-* class to the corresponding chat theme size.
		// Tailwind v4 generates HARDCODED values, so we MUST override with !important.
		const sizeMap: Record<string, string> = {
			// Preset sizes
			'text-xs': fs.xs,
			'text-sm': fs.sm,
			'text-base': fs.base,
			'text-lg': fs.lg,
			// Arbitrary sizes used in chat components
			'text-\\[9px\\]': `calc(${fs.xs} * 0.85)`,
			'text-\\[10px\\]': `calc(${fs.xs} * 0.95)`,
			'text-\\[11px\\]': fs.xs,
			'text-\\[12px\\]': fs.sm,
			'text-\\[13px\\]': fs.base,
			// Larger sizes (proportional scaling)
			'text-2xl': `calc(${fs.base} * 1.5)`,
		};

		const rules: string[] = [];

		for (const [cls, size] of Object.entries(sizeMap)) {
			rules.push(`.chat-root .${cls} { font-size: ${size} !important; }`);
		}

		// Also override font-family classes to use chat theme fonts
		rules.push(`.chat-root .font-sans { font-family: ${ft.sans} !important; }`);
		rules.push(`.chat-root .font-mono { font-family: ${ft.mono} !important; }`);

		/* 
		 * OVERRIDE html/body font-size for the chat subtree.
		 * The SYSTEM sets `html { font-size: var(--text-base) }` globally.
		 * We force ALL elements in the chat to use the chat theme font-size
		 * by default (unless they have a specific text-* override above).
		 */
		rules.push(`.chat-root, .chat-root * { font-size: ${fs.base} !important; }`);

		const css = rules.join('\n');

		if (!_styleEl) {
			_styleEl = document.createElement('style');
			_styleEl.id = 'chat-theme-override';
			document.head.appendChild(_styleEl);
		}
		_styleEl.textContent = css;
	}

	// Rebuild whenever chat theme version bumps (cross-module reactivity bridge)
	$effect(() => {
		// Read version COUNTER (primitive number) for guaranteed signal tracking
		const _v = theme.chatThemeVersion;
		void _v;

		// Now read the actual values — they'll be fresh because version changed
		const ct = colorThemes[theme.chatColorThemeId] ?? colorThemes.zinc;
		const ft = fontThemes[theme.chatFontThemeId] ?? fontThemes.geist;
		const fs = fontSizePresets[theme.chatFontSizeId] ?? fontSizePresets.default;

		// ── Build inline style (CSS variables + properties) ──
		const parts: string[] = [];

		// CSS variables (for custom access)
		for (const [key, value] of Object.entries(ct.vars)) {
			parts.push(`${key}: ${value}`);
		}
		// Also expose as --chat-* vars so the override <style> can reference them
		parts.push(`--chat-font-sans: ${ft.sans}`);
		parts.push(`--chat-font-mono: ${ft.mono}`);
		parts.push(`--chat-font-display: ${ft.display}`);
		parts.push(`--font-sans: ${ft.sans}`);
		parts.push(`--font-mono: ${ft.mono}`);
		parts.push(`--font-display: ${ft.display}`);

		// Actual CSS properties (override inherited computed values)
		parts.push(`font-family: ${ft.sans}`);
		parts.push(`font-size: ${fs.base}`);
		parts.push(`line-height: ${fs.lineHeight}`);

		chatStyle = parts.join('; ');

		// ── Apply ALL chat theme CSS overrides via DOM <style> element ──
		applyChatThemeOverrides(fs, { sans: ft.sans, mono: ft.mono });
	});

	let message = $state('');
	let isLoading = $state(false);
	let streamingContent = $state('');
	let selectedProviderFilter = $state('ALL');
	let modelSearchQuery = $state('');

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

	let messages = $state<ChatMessage[]>([]);
	let sessionID = $state<string>('');
	let chatContainer = $state<HTMLDivElement | null>(null);
	let autoScroll = $state(true);

	let actionLogs = $state<ActionLog[]>([]);
	let persistedActionLogs = $state<Set<string>>(new Set());

	$effect(() => {
		if (messages.length > 0) {
			const lastMsg = messages[messages.length - 1];
			if (lastMsg.actions && lastMsg.actions.length > 0) {
				const existingMap = new Map(actionLogs.map(a => [a.id, a]));
				const persisted = new Set([...persistedActionLogs].filter(id => lastMsg.actions.some(a => a.id === id)));
				const merged = lastMsg.actions.map(a => {
					const existing = existingMap.get(a.id);
					if (existing) {
						const status = persisted.has(a.id) ? a.status : (existing.status || a.status || 'done');
						return { ...a, status };
					}
					const status = persisted.has(a.id) ? a.status : (a.status || 'done');
					persisted.add(a.id);
					return { ...a, status };
				});
				actionLogs = merged;
				persistedActionLogs = persisted;
			}
		}
	})

	let permRequest = $state<{ toolName: string; args: string; reason: string; targetPath: string; mode: string; requestID: string } | null>(null);

	let cleanupStreamChunk: (() => void) | null = null;
	let cleanupDelta: (() => void) | null = null;
	let cleanupTurnEnd: (() => void) | null = null;
	let cleanupError: (() => void) | null = null;
	let cleanupThinking: (() => void) | null = null;
	let cleanupPermission: (() => void) | null = null;
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

	async function handlePermissionDecision(decision: 'allow_once' | 'allow_session' | 'deny') {
		if (permRequest) {
			try {
				await RespondPermission(permRequest.requestID, decision);
			} catch (e) {
				console.error('[ChatPanel] Failed to respond permission:', e);
			}
			permRequest = null;
		}
	}

	function handleToggleAction(id: string) {
		// Update live streaming actions
		actionLogs = actionLogs.map(l =>
			l.id === id ? { ...l, status: l.status === 'expanded' ? 'done' as const : 'expanded' as const } : l
		);
		// Update the message that owns this action
		for (let i = messages.length - 1; i >= 0; i--) {
			const msg = messages[i];
			if (msg.role !== 'assistant' || !msg.actions) continue;
			const found = msg.actions.find(a => a.id === id);
			if (found) {
				messages[i] = {
					...msg,
					actions: msg.actions.map(a =>
						a.id === id ? { ...a, status: a.status === 'expanded' ? 'done' as const : 'expanded' as const } : a
					)
				};
				messages = [...messages];
				break;
			}
		}
	}

	async function copyResponse(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			toastStore.success('Copied!');
		} catch { toastStore.error('Failed to copy'); }
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
 // Process messages and associate thinking content
  let loadedMessages: ChatMessage[] = (rawMessages || []).map((m: any) => ({ 
  role: m.role, 
  content: m.content,
  thinkingContent: m.thinking_content || ''
  }));
			// Associate thinking messages with their parent assistant messages
			for (let i = loadedMessages.length - 1; i >= 0; i--) {
				if (loadedMessages[i]?.role === 'thinking') {
					if (i > 0 && loadedMessages[i - 1]?.role === 'assistant') {
						const raw = loadedMessages[i].content;
						let text = raw;
						let sections: ThinkingSection[] | undefined;
						try {
							const parsed = JSON.parse(raw);
							if (parsed && typeof parsed === 'object' && parsed.text) {
								text = parsed.text;
								sections = parsed.sections;
							}
						} catch {}
						loadedMessages[i - 1].thinkingContent = text;
						if (sections) loadedMessages[i - 1].thinkingSections = sections;
					}
					loadedMessages.splice(i, 1);
				}
			}
			messages = loadedMessages;
			if (sess.model) {
				selectedModel = { id: sess.model, name: sess.model, providerName: sess.provider || '', health: 95 };
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
			if (!data) return;
			const content = data.content || '';
			console.log('[DEBUG] chat:delta', { len: content.length, preview: content.slice(0, 120) });
			if (content) {
				streamingContent = content;
				const lastIdx = messages.length - 1;
				if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
					messages[lastIdx] = { ...messages[lastIdx], content: streamingContent };
					messages = [...messages];
				}
			}
		});

cleanupStreamChunk = EventsOn('stream:chunk', (data: any) => {
						if (!data) return;
						const type = data.type;
						const payload = data.payload || '';
						const id = data.id || '';
						const meta = data.meta || '';
						console.log('[DEBUG] stream:chunk', { type, payloadLen: payload.length, payloadPreview: payload.slice(0, 100), id, metaLen: meta.length, metaPreview: meta.slice(0, 80) });

						// 基于 type 分流处理，兼容旧格式 eventType 与新格式 message
						if (type === 'action' || type === 'message') {
							// 优先从 payload 中解析完整的 message（包含 actions）
							let incomingMessage: any = null;
							try {
								incomingMessage = JSON.parse(payload);
							} catch {}
							if (incomingMessage?.actions?.length) {
								// 使用 message 中的 actions 作为准实时来源，覆盖 actionLogs
								const existingMap = new Map(actionLogs.map(a => [a.id, a]));
								const persisted = new Set([...persistedActionLogs].filter(id => incomingMessage.actions.some((a: any) => a.id === id)));
								const merged = incomingMessage.actions.map((a: any) => {
									const existing = existingMap.get(a.id);
									const status = persisted.has(a.id)
										? a.status
										: (existing?.status ?? a.status ?? 'done');
									persisted.add(a.id);
									return { ...a, status };
								});
								actionLogs = merged;
								persistedActionLogs = persisted;
							}
						} else if (type === 'thought') {
							const reasoning = data.payload || '';
							if (reasoning) {
								const lastIdx = messages.length - 1;
								if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
									messages[lastIdx] = { ...messages[lastIdx], thinkingContent: reasoning };
									messages = [...messages];
								}
							}
						}
					});

cleanupTurnEnd = EventsOn('chat:turnEnd', (_data: any) => {
					isLoading = false;
					stopTimer();
					const lastIdx = messages.length - 1;
					if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
						const thinkingDuration = responseStartTime > 0 ? Math.ceil((Date.now() - responseStartTime) / 1000) : 0;
						// 最终化 actions：将未持久化且 status 为 pending 的标记为 done，并持久化所有已完成
						const finalizedActions = actionLogs.map(l =>
							l.status === 'pending' ? { ...l, status: 'done' as const } : l
						);
						finalizedActions.forEach(a => persistedActionLogs.add(a.id));
						messages[lastIdx] = { ...messages[lastIdx], content: streamingContent, thinkingDuration, actions: finalizedActions, thinkingContent: messages[lastIdx].thinkingContent, thinkingSections: messages[lastIdx].thinkingSections };
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
					messages[lastIdx] = { ...messages[lastIdx], content: `Error: ${errorMsg}` };
					messages = [...messages];
				}
				streamingContent = '';
			});

			cleanupThinking = EventsOn('chat:thinking', (data: any) => {
				if (!data) return;
				const delta = data.content || '';
				const sectionType = data.type || 'text';
				if (delta) {
					const lastIdx = messages.length - 1;
					if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
						const prevContent = messages[lastIdx].thinkingContent || '';
						const prevSections = messages[lastIdx].thinkingSections || [];
						const lastSec = prevSections[prevSections.length - 1];
						if (lastSec && lastSec.type === sectionType) {
							lastSec.content += delta;
							messages[lastIdx] = {
								...messages[lastIdx],
								thinkingContent: prevContent + delta,
								thinkingSections: [...prevSections]
							};
						} else {
							messages[lastIdx] = {
								...messages[lastIdx],
								thinkingContent: prevContent + delta,
								thinkingSections: [...prevSections, { type: sectionType, content: delta }]
							};
						}
						messages = [...messages];
					}
				}
			});

			cleanupPermission = EventsOn('chat:permission-request', (data: any) => {
				if (!data) return;
				permRequest = {
					requestID: data.request_id || '',
					toolName: data.tool_name || '',
					args: data.args || '',
					reason: data.reason || '',
					targetPath: data.target_path || '',
					mode: data.mode || '',
				};
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
			if (cleanupStreamChunk) cleanupStreamChunk();
			if (cleanupDelta) cleanupDelta();
			if (cleanupTurnEnd) cleanupTurnEnd();
			if (cleanupError) cleanupError();
			if (cleanupThinking) cleanupThinking();
			if (cleanupPermission) cleanupPermission();
			if (timerInterval) clearInterval(timerInterval);
		});

	let contextLimit = $state(0);
	let contextUsed = $state(0);
	let contextBreakdown = $state<{ name: string; tokens: number; color: string }[]>([]);
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
				contextBreakdown = info.breakdown || [];
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
		messages = [...messages, { role: 'user' as const, content: userText }];
			isLoading = true;
			streamingContent = '';
			actionLogs = [];
			startTimer();
			messages = [...messages, { role: 'assistant' as const, content: '', actions: [] }];

		try {
			const modeParam = selectedMode.toLowerCase();
			if (!selectedModel?.id || !selectedModel?.providerName) {
				toastStore.warning('No model selected', 'Select a model before sending a message.');
				isLoading = false; stopTimer();
				messages = messages.slice(0, -1);
				return;
			}
			const modelString = `${selectedModel.providerName}/${selectedModel.id}`;
			const response = await SendMessage(sessionID, userText, modelString, 'normal', modeParam);
			if (response && response.trim()) {
				const lastIdx = messages.length - 1;
				if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
					messages[lastIdx] = { role: 'assistant' as const, content: response };
					messages = [...messages];
				}
				isLoading = false;
				stopTimer();
			}
		} catch (e) {
			console.error('[ChatPanel] Error sending message:', e);
			const lastIdx = messages.length - 1;
			if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
				messages[lastIdx] = { role: 'assistant' as const, content: `Error: ${String(e)}` };
				messages = [...messages];
			}
			isLoading = false;
			stopTimer();
		} finally {
			// streaming path: isLoading is managed by chat:turnEnd / chat:error events
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

<div class={cn('chat-root flex h-full flex-1 flex-col min-w-0 bg-[var(--bg-primary)]', className)} style={chatStyle}>
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
						{@const isLastAssistant = msg.role === 'assistant' && i === messages.length - 1}
						{@const liveActions = isLastAssistant && isLoading ? actionLogs : msg.actions}

						<MessageBubble
							message={liveActions ? { ...msg, actions: liveActions } : msg}
							{isLastAssistant}
							isStreaming={isLoading && isLastAssistant}
							responseTime={isLastAssistant ? getTotalTime() : ''}
							onCopy={copyResponse}
							onToggleAction={handleToggleAction}
						/>
					{/each}
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
				<button type="button" title="Attach file" class="flex items-center justify-center w-8 h-8 rounded-lg transition-colors cursor-pointer hover:bg-[var(--surface-hover)]" style="color: var(--text-muted)"><Icon name="attachment" size={15} /></button>
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
									{#each contextBreakdown as item}
										<div class="flex items-center justify-between text-xs text-zinc-300"><div class="flex items-center gap-2"><span class="w-2 h-2 rounded-full inline-block shrink-0" style="background: {item.color}"></span><span class="text-zinc-400">{item.name}</span></div><span class="font-mono text-zinc-300">{formatContextSize(item.tokens)}</span></div>
									{/each}
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

<PermissionDialog
	open={permRequest !== null}
	toolName={permRequest?.toolName || ''}
	args={permRequest?.args || ''}
	reason={permRequest?.reason || ''}
	targetPath={permRequest?.targetPath || ''}
	mode={permRequest?.mode || ''}
	onDecision={handlePermissionDecision}
/>
