<script lang="ts">
	import { cn } from '$lib/utils';
	import CollapsibleContent from '$lib/components/ui/collapsible/collapsible-content.svelte';
	import ThinkingFormatter from './ThinkingFormatter.svelte';
	import ThinkingShimmer from './ThinkingShimmer.svelte';
	import { getReasoningContext } from './reasoning-context.svelte';
	import type { ThinkingSection } from './chat-types';

	interface Props {
		class?: string;
		content?: string;
		sections?: ThinkingSection[];
		children?: import('svelte').Snippet;
	}

	let { class: className = '', content = '', sections = [], children, ...props }: Props = $props();

	let reasoningContext = getReasoningContext();
</script>

<CollapsibleContent
	class={cn(
		'mt-2 text-[11px] leading-relaxed',
		'data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:slide-out-to-top-2',
		'data-[state=open]:animate-in data-[state=open]:slide-in-from-top-2',
		'outline-none',
		className
	)}
	{...props}
>
	{#if sections.length > 0 || content}
		<div class="mt-1 rounded-lg p-3" style="background: #151d2b">
			<ThinkingFormatter {sections} {content} />
		</div>
	{:else if !children && reasoningContext.isStreaming}
		<ThinkingShimmer />
	{/if}
	{#if children}
		{@render children()}
	{/if}
</CollapsibleContent>
