<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';
	import type { HTMLLabelAttributes } from 'svelte/elements';

	type LabelProps = HTMLLabelAttributes & {
		class?: string;
		children?: Snippet;
	};

	let {
		class: className,
		children,
		...restProps
	}: LabelProps = $props();
</script>

<label
	class={cn(
		'text-[11px] font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
		className
	)}
	style="color: var(--text-secondary)"
	{...restProps}
>
	{#if children}
		{@render children()}
	{/if}
</label>
