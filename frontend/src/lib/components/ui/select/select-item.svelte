<script lang="ts">
	import { Select as SelectPrimitive } from "bits-ui";
	import { cn } from "$lib/utils.js";

	let { class: className, children, value, label, ...restProps }: SelectPrimitive.ItemProps = $props();
</script>

<SelectPrimitive.Item
	{value}
	{label}
	{...restProps}
>
	{#snippet child({ props, selected, highlighted })}
		<div
			{...props}
			class={cn(
				"relative flex w-full cursor-pointer select-none items-center rounded-lg py-2.5 px-4 text-[14px] outline-none transition-all data-[highlighted]:bg-[var(--accent-primary)]/10 data-[highlighted]:text-[var(--text-primary)] data-[disabled]:pointer-events-none data-[disabled]:opacity-50 text-[var(--text-secondary)]",
				className
			)}
		>
			{@render children?.({ selected, highlighted })}
		</div>
	{/snippet}
</SelectPrimitive.Item>
