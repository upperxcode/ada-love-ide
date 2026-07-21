<script lang="ts">
	import { marked } from 'marked';

	let { content = '' }: { content?: string } = $props();

	let html = $derived(marked.parse(content, { async: false }) as string);
</script>

<div class="prose prose-sm prose-invert max-w-none p-4 rounded-xl border border-[var(--border-primary)] bg-[var(--surface-input)] min-h-[120px]">
	{@html html}
</div>

<style>
	:global(.prose h1, .prose h2, .prose h3, .prose h4) {
		color: var(--text-primary);
		margin-top: 1.5em;
		margin-bottom: 0.5em;
	}
	:global(.prose p) {
		color: var(--text-secondary);
		line-height: 1.6;
		margin: 0.5em 0;
	}
	:global(.prose code) {
		background: var(--bg-tertiary);
		padding: 0.2em 0.4em;
		border-radius: 4px;
		font-size: 0.875em;
		color: var(--accent-primary);
	}
	:global(.prose pre) {
		background: var(--bg-tertiary);
		border: 1px solid var(--border-primary);
		border-radius: 8px;
		padding: 1em;
		overflow-x: auto;
	}
	:global(.prose pre code) {
		background: none;
		padding: 0;
	}
	:global(.prose ul, .prose ol) {
		color: var(--text-secondary);
		padding-left: 1.5em;
		margin: 0.5em 0;
	}
	:global(.prose li) {
		margin: 0.25em 0;
	}
	:global(.prose strong) {
		color: var(--text-primary);
	}
	:global(.prose a) {
		color: var(--accent-primary);
		text-decoration: underline;
	}
	:global(.prose blockquote) {
		border-left: 3px solid var(--accent-primary);
		padding-left: 1em;
		color: var(--text-muted);
		margin: 0.5em 0;
	}
</style>