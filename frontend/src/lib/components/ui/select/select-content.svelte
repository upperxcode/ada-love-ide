<script lang="ts">
	import { Select as SelectPrimitive } from "bits-ui";
	import { cn } from "$lib/utils.js";

	let { class: className, children, ...restProps }: SelectPrimitive.ContentProps = $props();
</script>

<SelectPrimitive.Content
	class={cn(
		"z-50 max-h-[300px] min-w-[var(--bits-select-trigger-width)] overflow-hidden rounded-xl border border-[var(--border-primary)] bg-[var(--bg-tertiary)] p-1 shadow-2xl outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95",
		className
	)}
	sideOffset={4}
	align="start"
	{...restProps}
>
	<SelectPrimitive.Viewport class="p-1.5">
		{@render children?.()}
	</SelectPrimitive.Viewport>
</SelectPrimitive.Content>
