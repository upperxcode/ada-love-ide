<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import type { ActionLog } from './chat-types';
	import { toolIcon } from './chat-types';

	interface Props {
		log: ActionLog;
		onToggle?: (id: string) => void;
		class?: string;
	}

	let { log, onToggle, class: className = '' }: Props = $props();

	let iconName = $derived(toolIcon(log.tool));
	let expanded = $derived(log.status === 'expanded');

	let toolLabel = $derived.by(() => {
		switch (log.tool) {
			case 'search': return 'Explore';
			case 'write': return log.diffStats ? 'Edited' : 'Writing';
			case 'read': return 'Read';
			case 'exec': return 'Ran';
			case 'plan': return 'Plan';
			default: return '';
		}
	});

	let diffAdd = $derived.by(() => {
		if (!log.diffStats) return '';
		const m = log.diffStats.match(/(\+\d+)/);
		return m ? m[1] : '';
	});

	let diffDel = $derived.by(() => {
		if (!log.diffStats) return '';
		const m = log.diffStats.match(/-\d+/);
		return m ? m[0] : '';
	});
</script>

<div
	class={cn('group/action rounded-md overflow-hidden transition-colors', className)}
	style="background: color-mix(in srgb, var(--bg-tertiary), transparent); border: 1px solid var(--border-subtle)"
>
	<button
		type="button"
		onclick={() => onToggle?.(log.id)}
		class="flex w-full items-center gap-2 px-3 py-2 text-[11px] font-mono cursor-pointer transition-colors"
		style="color: var(--text-muted); background: transparent; border: none; text-align: left"
	>
		<!-- Tool icon -->
		{#if log.status === 'pending'}
			<Icon name="loader" size={12} class="shrink-0 animate-spin" style="color: var(--text-faint)" />
		{:else if log.status === 'error'}
			<Icon name={iconName} size={12} class="shrink-0" style="color: var(--status-error)" />
		{:else}
			<Icon name={iconName} size={12} class="shrink-0 opacity-70" />
		{/if}

		<!-- Label line -->
		<span class="shrink-0 text-[10px] uppercase tracking-wider opacity-50 font-sans">
			{toolLabel}
		</span>

		<!-- Main content (truncated) -->
		<span class="flex-1 truncate min-w-0">{log.label}</span>

		<!-- Diff stats -->
		{#if log.diffStats}
			<span class="shrink-0" style="color: var(--status-success)">{diffAdd()}</span>
			<span class="shrink-0 ml-0.5" style="color: var(--status-error)">{diffDel()}</span>
		{/if}

		<!-- Result count for search -->
		{#if log.resultCount !== undefined && log.resultCount > 0}
			<span class="shrink-0 opacity-50">{log.resultCount} results</span>
		{/if}

		<!-- Success/error indicator for exec -->
		{#if log.tool === 'exec' && log.status === 'done'}
			<span class="shrink-0 text-[10px]" style="color: var(--status-success)">✓</span>
		{:else if log.tool === 'exec' && log.status === 'error'}
			<span class="shrink-0 text-[10px]" style="color: var(--status-error)">✗</span>
		{/if}

		<!-- Expand chevron (only if detail exists) -->
		{#if log.detail}
			<svg
				xmlns="http://www.w3.org/2000/svg"
				width="11"
				height="11"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				class="shrink-0 transition-transform opacity-50 {expanded ? 'rotate-180' : ''}"
			>
				<path d="m6 9 6 6 6-6" />
			</svg>
		{/if}
	</button>

	<!-- Expanded detail -->
	{#if expanded && log.detail}
		<pre
			class="m-0 px-3 py-2 text-[10px] leading-[1.5] overflow-x-auto max-h-60 overflow-y-auto"
			style="background: var(--bg-secondary); color: var(--text-faint); border-top: 1px solid var(--border-subtle)"
		>{log.detail}</pre>
	{/if}
</div>