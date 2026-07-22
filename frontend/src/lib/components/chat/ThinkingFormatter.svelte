<script lang="ts">
	import { Icon } from '$lib/components/icon';
	import type { ThinkingSection } from './chat-types';

	interface Props {
		sections?: ThinkingSection[];
		content?: string;
	}

	let { sections = [], content = '' }: Props = $props();

	let items = $derived.by<{ type: string; label: string; detail: string }[]>(() => {
		function make(text: string, type: string) {
			const firstLine = text.split('\n')[0].trim();
			const label = firstLine.length > 120 ? firstLine.slice(0, 120) + '…' : firstLine;
			const detail = text !== label ? text : '';
			return { type, label, detail };
		}
		if (sections.length > 0) {
			return sections.map(s => make(s.content, s.type));
		}
		if (content) {
			const lines = content.split('\n').map(l => l.trim()).filter(Boolean);
			return lines.map(l => make(l, 'text'));
		}
		return [];
	});

	function renderDetail(text: string) {
		const lines = text.split('\n');
		return lines.map(line => {
			if (line.startsWith('+')) return `<span style="color:#22c55e">${line}</span>`;
			if (line.startsWith('-')) return `<span style="color:#ef4444">${line}</span>`;
			return line;
		}).join('\n');
	}

	let expanded = $state<Set<number>>(new Set());

	function toggle(i: number) {
		const next = new Set(expanded);
		if (next.has(i)) next.delete(i); else next.add(i);
		expanded = next;
	}
</script>

{#each items as item, i}
	<div class="cursor-pointer" onclick={() => toggle(i)} onkeydown={(e) => e.key === 'Enter' && toggle(i)} role="button" tabindex="0">
		<div class="flex items-start gap-1.5 px-0.5 py-0.5 text-[11px] leading-relaxed" style="color: var(--text-secondary)">
			{#if item.type === 'plan'}
				<span class="shrink-0 mt-0.5" style="color: #a78bfa"><Icon name="layers" size={11} /></span>
			{:else if item.type === 'explore'}
				<span class="shrink-0 mt-0.5" style="color: #22c55e"><Icon name="search" size={11} /></span>
			{:else if item.type === 'exec'}
				<span class="shrink-0 mt-0.5" style="color: var(--accent-primary)"><Icon name="terminal" size={11} /></span>
			{:else if item.type === 'read'}
				<span class="shrink-0 mt-0.5" style="color: #60a5fa"><Icon name="eye" size={11} /></span>
			{:else if item.type === 'diff'}
				<span class="shrink-0 mt-0.5" style="color: #f59e0b"><Icon name="pencil" size={11} /></span>
			{/if}
			<div class="flex-1 min-w-0 truncate">
				<span>{item.label || ''}</span>
				{#if item.detail}
					<span class="inline-flex items-center ml-0.5 transition-transform {expanded.has(i) ? 'rotate-180' : ''}" style="color: var(--text-faint)">
						<svg xmlns="http://www.w3.org/2000/svg" width="9" height="9" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<path d="m6 9 6 6 6-6" />
						</svg>
					</span>
				{/if}
			</div>
		</div>
		{#if expanded.has(i) && item.detail}
			<pre class="m-0 pl-5 pr-1 pb-1.5 text-[11px] leading-[1.6] overflow-x-auto whitespace-pre-wrap" style="color: var(--text-faint); font-family: inherit">{@html renderDetail(item.detail)}</pre>
		{/if}
	</div>
{/each}

{#if items.length === 0}
	<div class="text-[11px] opacity-40 italic" style="color: var(--text-faint)">No thinking content</div>
{/if}
