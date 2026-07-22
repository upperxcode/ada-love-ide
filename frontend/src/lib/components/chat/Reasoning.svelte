<script lang="ts">
	import { cn } from '$lib/utils';
	import Collapsible from '$lib/components/ui/collapsible/collapsible.svelte';
	import { ReasoningContext, setReasoningContext } from './reasoning-context.svelte';

	interface Props {
		class?: string;
		isStreaming?: boolean;
		open?: boolean;
		defaultOpen?: boolean;
		duration?: number;
		children?: import('svelte').Snippet;
	}

	let {
		class: className = '',
		isStreaming = false,
		open = $bindable(),
		defaultOpen = false,
		onOpenChange,
		duration = $bindable(),
		children,
		...props
	}: Props & { onOpenChange?: (open: boolean) => void } = $props();

	const AUTO_CLOSE_DELAY = 1000;
	const MS_IN_S = 1000;

	let reasoningContext = new ReasoningContext({
		isStreaming,
		isOpen: open ?? defaultOpen,
		duration: duration ?? 0
	});

	let isOpen = $state(open ?? defaultOpen);
	let currentDuration = $state(duration ?? 0);
	let hasAutoClosed = $state(false);
	let startTime = $state<number | null>(null);

	$effect(() => {
		reasoningContext.isStreaming = isStreaming;
	});

	$effect(() => {
		if (open !== undefined) {
			isOpen = open;
			reasoningContext.isOpen = open;
		}
	});

	$effect(() => {
		if (duration !== undefined) {
			currentDuration = duration;
			reasoningContext.duration = duration;
		}
	});

	// Track duration when streaming starts and ends
	$effect(() => {
		if (isStreaming) {
			if (startTime === null) {
				startTime = Date.now();
			}
		} else if (startTime !== null) {
			const newDuration = Math.ceil((Date.now() - startTime) / MS_IN_S);
			currentDuration = newDuration;
			reasoningContext.duration = newDuration;
			if (duration !== undefined) {
				duration = newDuration;
			}
			startTime = null;
		}
	});

	// Auto-close when streaming ends (immediately, once only)
	$effect(() => {
		if (!isStreaming && isOpen && !hasAutoClosed) {
			handleOpenChange(false);
			hasAutoClosed = true;
		}
	});

	function handleOpenChange(newOpen: boolean) {
		isOpen = newOpen;
		reasoningContext.setIsOpen(newOpen);
		if (open !== undefined) {
			open = newOpen;
		}
		onOpenChange?.(newOpen);
	}

	setReasoningContext(reasoningContext);
</script>

<Collapsible
	class={cn('not-prose mb-2', className)}
	bind:open={isOpen}
	onOpenChange={handleOpenChange}
	{...props}
>
	{@render children?.()}
</Collapsible>