<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLTextareaAttributes } from 'svelte/elements';

	type TextareaProps = HTMLTextareaAttributes & {
		class?: string;
	};

	let {
		class: className,
		...restProps
	}: TextareaProps = $props();
</script>

<textarea
	class={cn(
		'flex min-h-[60px] w-full rounded-lg border border-[var(--border-primary)] bg-[var(--surface-input)] px-3 py-2 text-xs outline-none transition-colors placeholder:text-[var(--text-faint)] focus-visible:ring-1 focus-visible:ring-[var(--accent-primary)]/30 focus-visible:border-[var(--accent-primary)] disabled:cursor-not-allowed disabled:opacity-50',
		className
	)}
	style="color: var(--text-primary)"
	{...restProps}
></textarea>
