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
	import { CreateSessionWithConfig, GetSessionByID, GetSessionMessages, SendMessage, SendCLIMessage, SendURLMessage, SendOpenCodeMessage, RespondPermission, GetSessionWorkspaceSpec, SetSessionConfig, GetSessionContextInfo, AddSessionAttachment, RemoveSessionAttachment, ListSessionAttachments, CheckAttachmentExists, OpenFileDialog, StopGeneration, GetAdaConfig, SetAdaConfig, GetWorkers, GetCLIModels, GetURLModels, GetOpenCodeModels } from '../../../../wailsjs/go/main/App';
	import { EventsOn, EventsOff } from '../../../../wailsjs/runtime/runtime';
	import { toastStore } from '$lib/stores/toast.svelte';
	import MessageBubble from './MessageBubble.svelte';
	import PermissionDialog from './PermissionDialog.svelte';
	import type { ChatMessage, ActionLog, ThinkingSection, SessionAttachment } from './chat-types';
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
		// For CLI and URL workers, use their specific models
		if (isNonAdaWorker) {
			if (cliModels.length > 0) return cliModels;
			const fallbackNonAda: ModelOption[] = [
				{ id: 'claude-sonnet-4', name: 'claude-sonnet-4', providerName: currentWorkerType === 'url' ? 'API' : 'CLI', health: 100 },
				{ id: 'gpt-4o', name: 'gpt-4o', providerName: currentWorkerType === 'url' ? 'API' : 'CLI', health: 88 },
			];
			return fallbackNonAda;
		}
		// For Ada workers, use providersStore
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
	let cliModels = $state<ModelOption[]>([]);

	// Fetch non-Ada models when worker type changes.
	// First tries to get models from DB (providers filtered by worker),
	// then falls back to external API calls.
	$effect(() => {
		if (isNonAdaWorker && currentWorkerName) {
			const app = (window as any).go?.main?.App;
			// Try DB first
			if (app?.GetProvidersByWorker) {
				app.GetProvidersByWorker(currentWorkerName).then((providers: Record<string, any>) => {
					if (providers && Object.keys(providers).length > 0) {
						const models: ModelOption[] = [];
						for (const [provName, provCfg] of Object.entries(providers)) {
							const pCfg = provCfg as any;
							if (pCfg.models) {
								for (const modelName of Object.keys(pCfg.models)) {
									const ms = pCfg.models[modelName];
									models.push({
										id: modelName,
										name: modelName,
										providerName: provName,
										health: ms.health ?? 95,
									});
								}
							}
						}
						if (models.length > 0) {
							console.log(`[ChatPanel] using ${models.length} DB-stored models for worker ${currentWorkerName}`);
							cliModels = models;
							if (!selectedModel?.id) {
								selectedModel = models[0];
							}
							return; // DB models found, skip external fetch
						}
					}
					// Fallback: no DB models found, fetch from external API
					fetchExternalModels();
				}).catch((_e: any) => {
					console.warn(`[ChatPanel] GetProvidersByWorker failed, falling back to external fetch`);
					fetchExternalModels();
				});
			} else {
				fetchExternalModels();
			}

			function fetchExternalModels() {
				const fetchFn = currentWorkerType === 'url' ? GetURLModels : currentWorkerType === 'opencode_serve' ? GetOpenCodeModels : GetCLIModels;
				console.log(`[ChatPanel] fetching ${currentWorkerType} models for`, currentWorkerName);
				fetchFn(currentWorkerName).then((list: any[]) => {
					console.log(`[ChatPanel] ${currentWorkerType} models received`, list?.length);
					const providerLabel = currentWorkerType === 'url' ? 'API' : 'CLI';
					cliModels = (list || []).map((m: any) => ({
						id: m.id,
						name: m.name,
						providerName: m.provider_name || providerLabel,
						health: 95,
					}));
					if (cliModels.length > 0 && !selectedModel?.id) {
						selectedModel = cliModels[0];
					}
				}).catch((e: any) => {
					console.error(`[ChatPanel] ${currentWorkerType} models fetch failed`, e);
					cliModels = [];
				});
			}
		} else if (!isNonAdaWorker) {
			cliModels = [];
		}
	});

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
	let sessionAttachments = $state<SessionAttachment[]>([]);
	let showAttachmentsPanel = $state(false);
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

	let permRequest = $state<{ toolName: string; args: string; reason: string; targetPath: string; mode: string; requestID: string; action?: string; riskLevel?: string } | null>(null);

	// ── Worker context ──
	let currentWorkerType = $state<string>('ada');
	let currentWorkerName = $state<string>('');
	let _workersCache = $state<any[]>([]);

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
		try {
			const [sess, rawMessages, spec] = await Promise.all([
				GetSessionByID(sessID),
				GetSessionMessages(sessID),
				GetSessionWorkspaceSpec(sessID).catch(() => null),
			]);
			sessionID = sessID;
			activeSessionID = sessID;
			// Sync workspace to the session's workspace so Sidebar stays in sync
			if (sess.workspace_id) {
				activeWorkspace = sess.workspace_id;
			}
 // Process messages and associate thinking content
  let loadedMessages: ChatMessage[] = (rawMessages || []).map((m: any) => ({ 
  role: m.role, 
  content: m.content,
  thinkingContent: m.thinking_content || ''
  }));
			// Associate thinking messages with their parent assistant messages
			// Supports both orders: [..., assistant, thinking] and [..., thinking, assistant]
			for (let i = loadedMessages.length - 1; i >= 0; i--) {
				if (loadedMessages[i]?.role === 'thinking') {
					// Try next (i+1) or previous (i-1) for the assistant
					let targetIdx = -1;
					if (i + 1 < loadedMessages.length && loadedMessages[i + 1]?.role === 'assistant') {
						targetIdx = i + 1;
					} else if (i > 0 && loadedMessages[i - 1]?.role === 'assistant') {
						targetIdx = i - 1;
					}
					if (targetIdx >= 0) {
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
						loadedMessages[targetIdx].thinkingContent = text;
						if (sections) loadedMessages[targetIdx].thinkingSections = sections;
					}
					loadedMessages.splice(i, 1);
				}
			}
			messages = loadedMessages;
			if (sess.model) {
				selectedModel = { id: sess.model, name: sess.model, providerName: sess.provider || '', health: 95 };
			}
			if (sess.mode) selectedMode = sess.mode.toUpperCase();
			// Resolve worker type from session's worker_name
			if (sess.worker_name) {
				currentWorkerName = sess.worker_name;
				// Always fetch fresh workers to ensure latest config
				try {
					_workersCache = await GetWorkers();
				} catch {
					if (_workersCache.length === 0) _workersCache = [];
				}
				const match = _workersCache.find((w: any) => w.name === sess.worker_name);
				currentWorkerType = match?.connection_type || 'ada';
				console.log('[ChatPanel] worker resolved', { workerName: sess.worker_name, cacheSize: _workersCache.length, matchName: match?.name, matchType: match?.connection_type, resolvedType: currentWorkerType });
			} else {
				currentWorkerName = '';
				currentWorkerType = 'ada';
				console.log('[ChatPanel] no worker_name on session', { sessId: sessID });
			}
			await saveSessionConfig();
			refreshContextInfo();
			await loadSessionAttachments(sessID);
			if (spec) console.log('[ChatPanel] Workspace has Spec Wizard:', spec.name);
		} catch (e) {
			console.error('[ChatPanel] Failed to load session:', e);
			if (sessID) toastStore.warning('Session not found', 'Saved session was deleted. Starting a new one.');
			sessionID = '';
			activeSessionID = '';
			messages = [];
		}
	}

	async function loadSessionAttachments(sessID: string) {
		try {
			const list = await ListSessionAttachments(sessID);
			sessionAttachments = list || [];
		} catch (e) {
			console.error('[ChatPanel] Failed to load attachments:', e);
			sessionAttachments = [];
		}
	}

	async function toggleAttachmentsPanel() {
		if (!sessionID) return;
		showAttachmentsPanel = !showAttachmentsPanel;
		if (showAttachmentsPanel) {
			await loadSessionAttachments(sessionID);
		}
	}

	async function addAttachment() {
		if (!sessionID) return;
		try {
			const filePath = await OpenFileDialog();
			if (!filePath) return;
			const exists = await CheckAttachmentExists(sessionID, filePath);
			if (exists) {
				const name = filePath.split('/').pop() || filePath.split('\\').pop() || filePath;
				toastStore.warning('Already attached', `${name} is already attached`);
				return;
			}
			await AddSessionAttachment(sessionID, filePath);
			await loadSessionAttachments(sessionID);
		} catch (e) {
			console.error('[ChatPanel] Failed to attach file:', e);
			toastStore.error('Failed to attach file', String(e));
		}
	}

	async function removeAttachment(filePath: string) {
		if (!sessionID) return;
		try {
			await RemoveSessionAttachment(sessionID, filePath);
			await loadSessionAttachments(sessionID);
		} catch (e) {
			console.error('[ChatPanel] Failed to remove attachment:', e);
		}
	}

	async function initSession(workspacePath: string, workerName = 'default-worker') {
		try {
			const sess = await CreateSessionWithConfig(workspacePath, workerName, activeSessionID);
			await loadSession(sess.id);
		} catch (e) {
			console.error('[ChatPanel] Failed to create session:', e);
		}
	}

	async function switchSession(sessID: string) {
		if (sessID === sessionID) return;
		await loadSession(sessID);
	}

	onMount(async () => {
		await providersStore.load();
		// Preload workers list for worker type detection
		GetWorkers().then(list => { _workersCache = list; }).catch(() => {});

cleanupDelta = EventsOn('chat:delta', (data: any) => {
			if (!data || !isLoading) return;
			const content = data.content || '';
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
					action: data.action || '',
					riskLevel: data.risk_level || '',
				};
			});

			// ── Restore saved session on startup ──
			// We set _restoringSession = true so that the $effect for
			// activeWorkspace does NOT auto-create a new session while we
			// are reading saved config and restoring the last session.
			_restoringSession = true;
			try {
				const cfg = await GetAdaConfig();
				const savedID = cfg.active_session_id;
				const savedWorkspace = cfg.active_workspace_path;

				// Set workspace FIRST so Sidebar's bind sees it already
				if (savedWorkspace) {
					activeWorkspace = savedWorkspace;
				}

				if (savedID) {
					try {
						await loadSession(savedID);
					} catch {
						toastStore.warning('Session not synced', 'Saved session was deleted. Starting a new one.');
						await initSession(activeWorkspace);
					}
				} else {
					await initSession(activeWorkspace);
				}
			} catch {
				await initSession(activeWorkspace);
			} finally {
				_restoringSession = false;
			}
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

	// ── Restoring session guard ──
	// Prevents $effect(activeWorkspace) from auto-creating a session
	// while ChatPanel.onMount is restoring the saved session from config.
	let _restoringSession = $state(false);

	// ── Batch-change detection ──
	// When user clicks a chat in a different workspace, both activeSessionID
	// AND activeWorkspace change in the same event. The session $effect runs
	// first (defined first) and sets this flag. The workspace $effect checks
	// it: if both changed together, we must NOT create a new session — the
	// clicked chat should be loaded instead.
	let _sessionChangedInBatch = false;

	let prevSessionID = $state('');
	let persistedSessionID = $state('');
	$effect(() => {
		// Reset at the start — only set to true if session changes in THIS batch
		_sessionChangedInBatch = false;

		if (activeSessionID && activeSessionID !== prevSessionID) {
			_sessionChangedInBatch = true;
			prevSessionID = activeSessionID;
			switchSession(activeSessionID);
		}
		if (activeSessionID && activeSessionID !== persistedSessionID) {
			persistedSessionID = activeSessionID;
			// Merge with existing config to avoid overwriting sidebar / fixed models
			GetAdaConfig().then(cfg => {
				cfg.active_session_id = activeSessionID;
				SetAdaConfig(cfg);
			}).catch(() => {
				SetAdaConfig({ active_session_id: activeSessionID } as any);
			});
		}
	});

	let prevWorkspace = $state('');
	$effect(() => {
		if (activeWorkspace !== prevWorkspace) {
			prevWorkspace = activeWorkspace;
			// Only auto-create a session when the user explicitly changes workspace,
			// NOT during initial restore from saved config,
			// and NOT when a specific chat was also selected (both changed together).
			if (activeWorkspace && !_restoringSession && !_sessionChangedInBatch) {
				initSession(activeWorkspace);
			}
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

	const ALL_MODES = ['ASK', 'EDIT', 'PLAN', 'EXECUTE', 'FULL', 'ADMIN'] as const;
	const CLI_MODES = ['ASK'] as const;

	let availableModes = $derived<string[]>(
		currentWorkerType === 'cli' || currentWorkerType === 'url' || currentWorkerType === 'opencode_serve'
			? [...CLI_MODES]
			: [...ALL_MODES]
	);
	let isNonAdaWorker = $derived(currentWorkerType === 'cli' || currentWorkerType === 'url' || currentWorkerType === 'opencode_serve');
	let selectedMode = $state<string>('ASK');

	// Reset mode if current selection is no longer available
	$effect(() => {
		if (!availableModes.includes(selectedMode)) {
			selectedMode = availableModes[0] || 'ASK';
		}
	});

	const modeColors: Record<string, { color: string; border: string }> = {
		ASK: { color: 'var(--text-faint)', border: 'var(--border-primary)' },
		EDIT: { color: 'var(--accent-primary)', border: 'rgba(59,130,246,0.4)' },
		PLAN: { color: '#8b5cf6', border: 'rgba(139,92,246,0.4)' },
		EXECUTE: { color: '#f97316', border: 'rgba(249,115,22,0.4)' },
		FULL: { color: '#ef4444', border: 'rgba(239,68,68,0.4)' },
		ADMIN: { color: '#dc2626', border: 'rgba(220,38,38,0.5)' },
	};
	const currentModeColor = $derived(modeColors[selectedMode] || modeColors.ASK);
	const modeIcons: Record<string, string> = {
		ASK: 'search',
		EDIT: 'pencil',
		PLAN: 'layers',
		EXECUTE: 'terminal',
		FULL: 'bot',
		ADMIN: 'settings',
	};
	const modeDescriptions: Record<string, string> = {
		ASK: 'Consultas e leitura — sem alterações',
		EDIT: 'Edição assistida de código',
		PLAN: 'Análise e planejamento arquitetural',
		EXECUTE: 'Testes e comandos seguros',
		FULL: 'Agente autônomo completo',
		ADMIN: 'Gerenciamento do sistema',
	};

	async function handleSend() {
		if (!message.trim() || isLoading || !sessionID) return;
		const userText = message.trim();
		message = '';
		messages = [...messages, { role: 'user' as const, content: userText, attachments: sessionAttachments }];
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
			const modelString = currentWorkerType === 'opencode_serve'
				? selectedModel.id
				: `${selectedModel.providerName}/${selectedModel.name}`;

			if (currentWorkerType === 'cli') {
				const response = await SendCLIMessage(sessionID, userText, modelString);
				if (response && response.trim()) {
					const lastIdx = messages.length - 1;
					if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
						const existing = messages[lastIdx];
						messages[lastIdx] = { ...existing, content: response, role: 'assistant' as const };
						messages = [...messages];
					}
				}
				isLoading = false;
				stopTimer();
			} else if (currentWorkerType === 'url') {
				const response = await SendURLMessage(sessionID, userText, modelString);
				if (response && response.trim()) {
					const lastIdx = messages.length - 1;
					if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
						const existing = messages[lastIdx];
						messages[lastIdx] = { ...existing, content: response, role: 'assistant' as const };
						messages = [...messages];
					}
				}
				isLoading = false;
				stopTimer();
			} else if (currentWorkerType === 'opencode_serve') {
				const response = await SendOpenCodeMessage(sessionID, userText, modelString);
				if (response && response.trim()) {
					const lastIdx = messages.length - 1;
					if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
						const existing = messages[lastIdx];
						messages[lastIdx] = { ...existing, content: response, role: 'assistant' as const };
						messages = [...messages];
					}
				}
				isLoading = false;
				stopTimer();
			} else {
				const response = await SendMessage(sessionID, userText, modelString, 'normal', modeParam);
				if (response && response.trim()) {
					const lastIdx = messages.length - 1;
					if (lastIdx >= 0 && messages[lastIdx].role === 'assistant') {
						const existing = messages[lastIdx];
						messages[lastIdx] = { ...existing, content: response, role: 'assistant' as const };
						messages = [...messages];
					}
					isLoading = false;
					stopTimer();
				}
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

	let textareaEl = $state<HTMLTextAreaElement | null>(null);

	function autoResize() {
		if (!textareaEl) return;
		textareaEl.style.height = 'auto';
		textareaEl.style.height = textareaEl.scrollHeight + 'px';
	}

	$effect(() => {
		if (textareaEl) {
			message;
			autoResize();
		}
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend(); }
	}

	function handleInput() {
		autoResize();
	}

	function cycleMode() {
		const idx = availableModes.indexOf(selectedMode);
		selectedMode = availableModes[(idx + 1) % availableModes.length];
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

	<div class={cn('mx-4 mb-4 rounded-xl border border-[var(--border-subtle)] bg-[var(--surface-input)] relative transition-colors focus-within:border-[var(--border-hover)]')}>
		<div class="px-4 pt-3 pb-1">
			<textarea bind:value={message} onkeydown={handleKeydown} oninput={handleInput} bind:this={textareaEl}
				placeholder="message {currentWorkerType === 'cli' ? currentWorkerName || 'CLI' : 'ada'}...  (  /  for commands ·  @  for files )"
				class="flex-1 w-full resize-none bg-transparent border-none outline-none text-base leading-relaxed placeholder:opacity-40 max-h-56 min-h-[28px] overflow-y-auto"
				style="color: var(--text-primary)">
			</textarea>
		</div>
		{#if showAttachmentsPanel}
			<div class="absolute bottom-[40px] left-2 w-[300px] rounded-lg border border-[var(--border-subtle)] bg-[var(--surface-input)] shadow-lg overflow-hidden z-10" style="color: var(--text-secondary)">
				<div class="flex items-center justify-between px-3 py-2 border-b border-[var(--border-subtle)]">
					<span class="text-xs font-semibold" style="color: var(--text-primary)">Attachments</span>
					<button type="button" onclick={() => showAttachmentsPanel = false}
						class="flex items-center justify-center w-5 h-5 rounded hover:bg-[var(--surface-hover)] transition-colors cursor-pointer"
						aria-label="Close attachments panel">
						<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
					</button>
				</div>
				<div class="flex flex-col gap-0.5 p-1.5 max-h-32 overflow-y-auto">
					{#if sessionAttachments.length === 0}
						<div class="px-2 py-3 text-xs text-center opacity-40">No attachments</div>
					{:else}
						{#each sessionAttachments as att}
							<div class="flex items-center gap-2 px-2 py-1.5 rounded-md text-xs group hover:bg-[var(--surface-hover)] transition-colors">
								<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="shrink-0 opacity-50"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/></svg>
								<span class="flex-1 truncate">{att.file_name}</span>
								<button type="button" onclick={() => removeAttachment(att.file_path)}
									aria-label="Remove attachment"
									class="flex items-center justify-center w-5 h-5 rounded opacity-0 group-hover:opacity-100 hover:bg-[var(--bg-primary)] transition-all cursor-pointer"
									style="color: var(--status-error)">
									<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
								</button>
							</div>
						{/each}
					{/if}
				</div>
				<button type="button" onclick={addAttachment}
					class="flex items-center gap-1.5 w-full px-3 py-2 text-xs border-t border-[var(--border-subtle)] hover:bg-[var(--surface-hover)] transition-colors cursor-pointer"
					style="color: var(--accent-primary)">
					<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
					Add file
				</button>
			</div>
		{/if}
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
				<div class="relative">
					<button type="button" onclick={toggleAttachmentsPanel} title="Attach file"
						class="flex items-center justify-center w-8 h-8 rounded-lg transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
						style="color: {sessionAttachments.length > 0 ? 'var(--accent-primary)' : 'var(--text-muted)'}">
						<Icon name="attachment" size={15} />
					</button>
					{#if sessionAttachments.length > 0}
						<span class="absolute -top-0.5 -right-0.5 flex items-center justify-center w-3.5 h-3.5 rounded-full text-[8px] font-bold"
							style="background: var(--accent-primary); color: var(--accent-primary-fg)">
							{sessionAttachments.length}
						</span>
					{/if}
				</div>
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
					<DropdownMenu>
					{#snippet trigger({ toggle })}
						<button type="button" onclick={(e) => { e.stopPropagation(); toggle(); }}
							class="flex items-center gap-1.5 h-7 px-3 rounded-md text-[11px] font-semibold tracking-wider cursor-pointer border transition-colors hover:brightness-125 active:scale-[0.98]"
							style="color: {currentModeColor.color}; border-color: {currentModeColor.border}; background: {currentModeColor.color}08">
							<Icon name={modeIcons[selectedMode]} size={13} />
							{selectedMode}
							<Icon name="chevron-down" size={11} style="opacity: 0.5" />
						</button>
					{/snippet}
					{#snippet content({ close })}
						<DropdownContent align="end" class="bottom-full top-auto mb-1.5 mt-0 w-[220px] p-1.5">
							{#each availableModes as mode}
								{@const mc = modeColors[mode]}
								<button type="button" onclick={() => { selectedMode = mode; saveSessionConfig(); close(); }}
									class="flex items-center gap-2.5 w-full px-2.5 py-2 rounded-md text-left transition-all cursor-pointer"
									style="background: {mode === selectedMode ? `${mc.color}12` : 'transparent'}; color: {mode === selectedMode ? mc.color : 'var(--text-secondary)'};"
									onmouseenter={(e) => { if (mode !== selectedMode) (e.currentTarget as HTMLElement).style.background = 'var(--surface-hover)'; }}
									onmouseleave={(e) => { if (mode !== selectedMode) (e.currentTarget as HTMLElement).style.background = 'transparent'; }}>
									<div class="flex items-center justify-center w-7 h-7 rounded-md shrink-0" style="background: {mc.color}15; color: {mc.color}">
										<Icon name={modeIcons[mode]} size={14} />
									</div>
									<div class="flex flex-col min-w-0 flex-1 leading-tight">
										<div class="flex items-center gap-1.5">
											<span class="text-[12px] font-bold tracking-wider">{mode}</span>
											{#if mode === selectedMode}
												<Icon name="check" size={12} style="color: {mc.color}" />
											{/if}
										</div>
										<span class="text-[10px] opacity-60 truncate">{modeDescriptions[mode]}</span>
									</div>
								</button>
								{#if mode === 'PLAN' || mode === 'FULL'}
									<div class="mx-2 my-0.5" style="border-top: 1px solid var(--border-primary); opacity: 0.3"></div>
								{/if}
							{/each}
						</DropdownContent>
					{/snippet}
				</DropdownMenu>
				<DropdownMenu>
					{#snippet trigger({ toggle })}
						<button type="button" onclick={async (e) => { e.stopPropagation(); selectedProviderFilter = selectedModel.providerName || 'ALL'; modelSearchQuery = ''; await providersStore.refresh(); toggle(); }}
							class="flex items-center gap-1.5 h-7 px-3 rounded-md text-[11px] font-mono cursor-pointer border transition-colors border-[var(--border-primary)] hover:bg-[var(--surface-hover)] hover:border-[var(--border-hover)]"
							style="color: {isNonAdaWorker ? 'var(--accent-primary)' : 'var(--text-secondary)'}">
							<span class="w-2 h-2 rounded-full inline-block shrink-0 shadow-sm border border-black/20" style="background-color: hsl({(selectedModel.health / 100) * 120}, 85%, 45%);"></span>
							{selectedModel.name}
							{#if currentWorkerType === 'cli'}
								<span class="text-[8px] font-bold px-1 py-0.5 rounded uppercase tracking-wider" style="background: var(--accent-primary); color: var(--accent-primary-fg)">CLI</span>
							{/if}
							{#if currentWorkerType === 'url'}
								<span class="text-[8px] font-bold px-1 py-0.5 rounded uppercase tracking-wider" style="background: var(--accent-primary); color: var(--accent-primary-fg)">API</span>
							{/if}
							{#if currentWorkerType === 'opencode_serve'}
								<span class="text-[8px] font-bold px-1 py-0.5 rounded uppercase tracking-wider" style="background: var(--accent-primary); color: var(--accent-primary-fg)">OC</span>
							{/if}
							<Icon name="chevron-down" size={11} />
						</button>
					{/snippet}
					{#snippet content({ close })}
						<DropdownContent align="end" class="bottom-full top-auto mb-1.5 mt-0 w-[320px] h-[300px] p-2.5 bg-zinc-900 border border-zinc-800 rounded-lg shadow-xl flex flex-col gap-2.5">
							<div class="flex flex-col gap-1">
								<div class="flex items-center justify-between">
									<span class="text-[9px] uppercase font-bold tracking-wider text-zinc-500 font-sans">Provider</span>
									{#if currentWorkerType === 'cli'}
										<span class="text-[8px] font-bold px-1 py-0.5 rounded uppercase tracking-wider bg-zinc-700 text-zinc-300">CLI Model</span>
									{/if}
									{#if currentWorkerType === 'url'}
										<span class="text-[8px] font-bold px-1 py-0.5 rounded uppercase tracking-wider bg-zinc-700 text-zinc-300">API Model</span>
									{/if}
									{#if currentWorkerType === 'opencode_serve'}
										<span class="text-[8px] font-bold px-1 py-0.5 rounded uppercase tracking-wider bg-zinc-700 text-zinc-300">OpenCode Model</span>
									{/if}
								</div>
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
												class={cn('relative flex w-full items-center justify-between rounded px-2.5 py-1.5 text-xs cursor-pointer font-mono transition-all border', model.id === selectedModel.id ? 'border-[var(--accent-primary)] bg-[var(--accent-primary)]/10 text-[var(--accent-primary)] font-semibold' : 'border-transparent hover:bg-zinc-800/40 text-zinc-300')}>
												<div class="flex items-center gap-2 min-w-0">
													{#if model.id === selectedModel.id}
														<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="shrink-0"><polyline points="20 6 9 17 4 12"/></svg>
													{:else}
														<span class="w-2.5 h-2.5 rounded-full inline-block shrink-0 shadow-sm border border-black/20" style="background-color: hsl({hue}, 85%, 45%);"></span>
													{/if}
													<span class="truncate">{model.name}</span>
												</div>
												<span class={cn('text-[9px] font-sans tracking-wide uppercase px-1 rounded shrink-0', model.id === selectedModel.id ? 'text-[var(--accent-primary)]/60 bg-[var(--accent-primary)]/10' : 'opacity-40 bg-zinc-800 text-zinc-400')}>{model.providerName}</span>
											</button>
										{/each}
									{/if}
								</div>
							</div>
						</DropdownContent>
					{/snippet}
				</DropdownMenu>
				{#if isLoading}
					<button type="button" onclick={() => StopGeneration(sessionID)}
						class="flex items-center justify-center w-8 h-8 rounded-lg cursor-pointer transition-all duration-150 hover:brightness-125 active:scale-95 animate-pulse"
						style="background-color: var(--status-error); color: white" title="Stop generation">
						<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M12 17v4m-4 0h8"/><rect width="20" height="14" x="2" y="3" rx="2"/><rect width="6" height="6" x="9" y="7" rx="1"/></g></svg>
					</button>
				{:else}
					<button type="button" onclick={handleSend}
						class="flex items-center justify-center w-8 h-8 rounded-lg cursor-pointer transition-all duration-150 hover:brightness-125 active:scale-95"
						style="background-color: var(--accent-primary); color: var(--accent-primary-fg)" title="Send message">
						<Icon name="arrowUp" size={16} />
					</button>
				{/if}
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
	action={permRequest?.action || ''}
	riskLevel={permRequest?.riskLevel || ''}
	onDecision={handlePermissionDecision}
/>
