<script lang="ts">
	import { Label } from '$lib/components/ui/label/index.js';
	import { cn } from '$lib/utils';

	interface ExpandableTextareaProps {
		id?: string;
		label?: string;
		value?: string;
		placeholder?: string;
		class?: string;
		textareaClass?: string;
		minHeight?: number;
		disabled?: boolean;
		oninput?: (value: string) => void;
	}

	let {
		id,
		label,
		value = $bindable(''),
		placeholder,
		class: className,
		textareaClass,
		minHeight = 48,
		disabled = false,
		oninput,
	}: ExpandableTextareaProps = $props();

	// ── Auto-resize action (grows with content, user can still manually resize) ──
	function autoResize(node: HTMLTextAreaElement) {
		let isManualResize = false;
		let rafId: number | null = null;

		// Schedule a resize on the next animation frame so the style write never
		// happens synchronously inside the ResizeObserver callback. Writing
		// height synchronously re-fires the observer and produces the
		// "ResizeObserver loop completed with undelivered notifications" error.
		function scheduleResize() {
			if (rafId !== null) return;
			rafId = requestAnimationFrame(() => {
				rafId = null;
				resize();
			});
		}

		function resize() {
			if (isManualResize) return;
			const target = `${Math.max(node.scrollHeight, minHeight)}px`;
			// Read-then-compare guard: only write when the height actually
			// changes, breaking the observer→write→observer feedback loop.
			if (node.style.height === target) return;
			node.style.height = 'auto';
			node.style.height = target;
		}
		resize();

		// Detect manual resize: stop auto-resize until next input
		function onPointerUp() {
			// Check if the user actually dragged the resize handle
			// We mark manual so auto-resize pauses until next input resets it
			isManualResize = true;
		}

		function onInput() {
			isManualResize = false;
			resize();
		}

		node.addEventListener('pointerup', onPointerUp);
		node.addEventListener('input', onInput);

		// Use rAF-batched resize so the observer's notification is fully
		// delivered before we touch layout, avoiding the loop error.
		const observer = new ResizeObserver(() => {
			// Only auto-resize if not manually resized
			if (!isManualResize) scheduleResize();
		});
		observer.observe(node);

		return {
			update() { if (!isManualResize) scheduleResize(); },
			destroy() {
				if (rafId !== null) cancelAnimationFrame(rafId);
				observer.disconnect();
				node.removeEventListener('pointerup', onPointerUp);
				node.removeEventListener('input', onInput);
			},
		};
	}

	function handleInput(e: Event) {
		const val = (e.target as HTMLTextAreaElement).value;
		value = val;
		oninput?.(val);
	}

	const textareaBase = 'flex min-h-[60px] w-full rounded-lg border border-[var(--border-primary)] bg-[var(--surface-input)] px-3 py-2 text-xs outline-none transition-colors placeholder:text-[var(--text-faint)] focus-visible:ring-1 focus-visible:ring-[var(--accent-primary)]/30 focus-visible:border-[var(--accent-primary)] disabled:cursor-not-allowed disabled:opacity-50';
</script>

<div class={cn('grid w-full gap-1.5', className)}>
	{#if label}
		<Label for={id}>{label}</Label>
	{/if}
	<textarea
		{id}
		use:autoResize
		bind:value
		oninput={handleInput}
		{placeholder}
		{disabled}
		rows={1}
		class={cn(textareaBase, 'resize-y font-mono leading-relaxed', textareaClass)}
		style="color: var(--text-primary); min-height: {minHeight}px"
	></textarea>
</div>
