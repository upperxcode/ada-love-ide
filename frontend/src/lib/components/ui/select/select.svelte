<script lang="ts">
	import { Select as SelectPrimitive } from "bits-ui";

	let {
		children,
		type = "single",
		value = $bindable(),
		onValueChange,
		...restProps
	}: SelectPrimitive.RootProps = $props();
</script>

<SelectPrimitive.Root 
	type={type as any} 
	bind:value={value as any} 
	onValueChange={onValueChange as any} 
	{...restProps}
>
	{@render children?.()}
</SelectPrimitive.Root>
