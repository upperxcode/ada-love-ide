<script lang="ts">
	import { cn } from '$lib/utils';

	interface DropdownContentProps {
		align?: 'start' | 'end' | 'center';
		class?: string;
		children?: import('svelte').Snippet;
	}

	let { align = 'start', class: className, children }: DropdownContentProps = $props();

	const alignClasses: Record<NonNullable<DropdownContentProps['align']>, string> = {
		start: 'left-0',
		end: 'right-0',
		center: 'left-1/2 -translate-x-1/2'
	};
</script>

<div
	class={cn(
		'absolute top-full z-50 mt-1 min-w-[8rem] overflow-hidden rounded-lg border border-[var(--border-primary)]',
		'bg-[var(--bg-tertiary)] p-1 text-[var(--text-secondary)] shadow-md',
		'animate-in fade-in-0 zoom-in-95',
		alignClasses[align],
		className
	)}
>
	{#if children}
		{@render children()}
	{/if}
</div>
