<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { EventsOn } from '../../../../wailsjs/runtime/runtime';
	import { Icon } from '$lib/components/icon';

	interface LogEntry {
		time: string;
		event: string;
		summary: string;
		type: 'info' | 'warn' | 'error' | 'perm' | 'action' | 'thinking';
		detail: string;
	}

	interface Props { onClose?: () => void; }

	let { onClose }: Props = $props();

	let logs = $state<LogEntry[]>([]);
	let autoScroll = $state(true);
	let logContainer: HTMLDivElement | null = $state(null);
	let cleanups: (() => void)[] = [];

	$effect(() => {
		if (autoScroll && logContainer) {
			logContainer.scrollTop = logContainer.scrollHeight;
		}
	});

	function add(type: LogEntry['type'], event: string, summary: string, detail = '') {
		logs = [...logs, {
			time: new Date().toLocaleTimeString(),
			event, summary, type, detail
		}];
		if (logs.length > 500) {
			logs = logs.slice(-300);
		}
	}

	function fmt(v: any): string {
		if (!v) return '';
		if (typeof v === 'string') return v.slice(0, 500);
		try { return JSON.stringify(v).slice(0, 500); }
		catch { return String(v).slice(0, 500); }
	}

	onMount(() => {
		let lastContent = '';
		let lastReasoning = '';
		cleanups = [
			EventsOn('stream:chunk', (data: any) => {
				if (!data) return;
				const t = data.type || '?';
				if (t === 'content') return;
				const p = fmt(data.payload);
				add('info', `stream:chunk [${t}]`, p.slice(0, 120), p);
			}),
			EventsOn('chat:thinking', (data: any) => {
				if (!data?.content) return;
				const c = fmt(data.content);
				if (c === lastReasoning) return;
				lastReasoning = c;
				const t = data.type || 'text';
				add('thinking', `thinking [${t}]`, c.slice(0, 120), c);
			}),
			EventsOn('chat:error', (data: any) => {
				add('error', 'chat:error', fmt(data?.error || 'unknown error'), fmt(data));
			}),
			EventsOn('chat:permission-request', (data: any) => {
				if (!data) return;
				add('perm', 'permission', `${data.tool_name} — ${data.reason}`, fmt(data));
			}),
			EventsOn('chat:status', (data: any) => {
				add('info', 'status', fmt(data?.stage || ''), fmt(data));
			}),
		];
	});

	onDestroy(() => {
		cleanups.forEach(fn => fn());
	});

	function clearLogs() {
		logs = [];
	}

	function toggleAutoScroll() {
		autoScroll = !autoScroll;
	}
</script>

<div class="flex flex-col h-full w-[340px] shrink-0 border-l" style="background: var(--bg-primary); border-color: var(--border-primary)">
	<div class="flex items-center justify-between px-4 py-2.5 border-b" style="border-color: var(--border-primary)">
		<h3 class="text-[11px] font-bold uppercase tracking-widest" style="color: var(--text-faint)">Event Logs</h3>
		<div class="flex items-center gap-1">
			<button type="button" onclick={toggleAutoScroll}
				class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
				style="color: {autoScroll ? 'var(--accent-primary)' : 'var(--text-faint)'}"
				title="Auto-scroll">
				<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
			</button>
			<button type="button" onclick={clearLogs}
				class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
				style="color: var(--status-error)" title="Clear logs">
				<Icon name="deleteIcon" size={12} />
			</button>
			<button type="button" onclick={onClose}
				class="flex items-center justify-center w-6 h-6 rounded transition-colors cursor-pointer hover:bg-[var(--surface-hover)]"
				style="color: var(--text-faint)" title="Close panel">
				<Icon name="x" size={12} />
			</button>
		</div>
	</div>

	<div bind:this={logContainer} class="flex-1 overflow-y-auto px-2 py-1.5 font-mono text-[10px] leading-[1.5]" style="color: var(--text-secondary)">
		{#if logs.length === 0}
			<div class="text-center py-8 opacity-40 italic" style="color: var(--text-faint)">No events yet</div>
		{/if}
		{#each logs as entry}
			<div class="group/log py-0.5 px-1.5 rounded-sm transition-colors hover:bg-[var(--surface-hover)] cursor-pointer" onclick={() => {}}>
				<div class="flex items-start gap-1.5">
					<span class="shrink-0 mt-px text-[9px] opacity-50" style="color: var(--text-faint)">{entry.time}</span>
					<span class="shrink-0 text-[9px] uppercase font-bold tracking-wider mt-px"
						style="color: {entry.type === 'error' ? 'var(--status-error)' : entry.type === 'perm' ? 'var(--status-warning)' : entry.type === 'thinking' ? 'var(--accent-primary)' : 'var(--text-faint)'}">
						{entry.type}
					</span>
					<div class="flex-1 min-w-0 truncate text-[10px]">{entry.summary}</div>
				</div>
				{#if entry.detail && entry.detail !== entry.summary}
					<div class="hidden group-hover/log:block mt-0.5 px-1 py-0.5 rounded text-[9px] whitespace-pre-wrap break-all" style="background: var(--bg-tertiary); color: var(--text-faint)">{entry.detail}</div>
				{/if}
			</div>
		{/each}
	</div>

	<div class="px-3 py-1.5 border-t text-[9px] opacity-30" style="color: var(--text-faint); border-color: var(--border-primary)">
		{logs.length} events
	</div>
</div>
