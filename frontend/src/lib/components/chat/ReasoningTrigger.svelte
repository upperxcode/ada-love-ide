<script lang="ts">
	import { cn } from '$lib/utils';
	import CollapsibleTrigger from '$lib/components/ui/collapsible/collapsible-trigger.svelte';
	import { getReasoningContext } from './reasoning-context.svelte';
	import { Icon } from '$lib/components/icon';

	interface Props {
		class?: string;
		children?: import('svelte').Snippet;
	}

	let { class: className = '', children, ...props }: Props = $props();

	let reasoningContext = getReasoningContext();

	let thinkingMessage = $derived.by(() => {
		const { isStreaming, duration } = reasoningContext;
		if (isStreaming || duration === 0) return 'Thinking...';
		if (duration === undefined) return 'Thought for a few seconds';
		return `Thought for ${duration}s`;
	});
</script>

<CollapsibleTrigger
	class={cn(
		'flex w-full items-center gap-2 text-xs font-medium transition-colors cursor-pointer select-none',
		'hover:text-[var(--text-primary)]',
		className
	)}
	style="color: var(--text-muted)"
	{...props}
>
	{#if children}
		{@render children()}
	{:else}
		{#if reasoningContext.isStreaming}
			<Icon name="loader" size={13} class="animate-spin" />
		{:else}
			<Icon name="brain" size={13} />
		{/if}
		<p>{thinkingMessage}</p>
		<svg
			xmlns="http://www.w3.org/2000/svg"
			width="13"
			height="13"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
			class="transition-transform {reasoningContext.isOpen ? 'rotate-180' : 'rotate-0'}"
		>
			<path d="m6 9 6 6 6-6" />
		</svg>
	{/if}
</CollapsibleTrigger>