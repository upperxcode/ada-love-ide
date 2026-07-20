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
		'text-zinc-400 outline-none transition-colors',
		'hover:bg-zinc-800 hover:text-zinc-200',
		'focus:bg-zinc-800 focus:text-zinc-200',
		className
	)}
	onclick={onclick}
>
	{#if children}
		{@render children()}
	{/if}
</button>
