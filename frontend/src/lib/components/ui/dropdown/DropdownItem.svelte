<script lang="ts">
	import { cn } from '$lib/utils';

	interface DropdownItemProps {
		class?: string;
		onclick?: (e: MouseEvent) => void;
		children?: import('svelte').Snippet;
	}

	let { class: className, onclick, children }: DropdownItemProps = $props();
</script>

<button
	type="button"
	class={cn(
		'relative flex w-full cursor-pointer select-none items-center rounded-sm px-2 py-1.5 text-sm',
		'text-[var(--text-secondary)] outline-none transition-colors',
		'hover:bg-[var(--surface-hover)] hover:text-[var(--text-primary)]',
		'focus:bg-[var(--surface-hover)] focus:text-[var(--text-primary)]',
		className
	)}
	onclick={onclick}
>
	{#if children}
		{@render children()}
	{/if}
</button>
