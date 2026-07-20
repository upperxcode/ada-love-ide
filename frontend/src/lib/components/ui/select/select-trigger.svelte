<script lang="ts">
	import { Select as SelectPrimitive } from "bits-ui";
	import { cn } from "$lib/utils.js";

	let { class: className, children, ...restProps }: SelectPrimitive.TriggerProps = $props();
</script>

	<SelectPrimitive.Trigger
		class={cn(
			"flex h-11 w-full items-center justify-between rounded-lg border border-[var(--border-primary)] bg-[var(--surface-input)] px-4 py-3 text-[14px] transition-all hover:border-[var(--border-hover)] focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)] outline-none disabled:cursor-not-allowed disabled:opacity-50",
			className
		)}
		{...restProps}
	>
		{@render children?.()}
		<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="opacity-50 shrink-0 ml-2 pointer-events-none"><path d="m6 9 6 6 6-6"/></svg>
	</SelectPrimitive.Trigger>
