<script lang="ts">
	import { cn } from "$lib/utils.js";
	import type { HTMLAttributes } from "svelte/elements";
	import { cva, type VariantProps } from "class-variance-authority";

	const alertVariants = cva(
		"relative w-full rounded-lg border px-4 py-3 text-sm [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:text-foreground [&>svg~*]:pl-7",
		{
			variants: {
				variant: {
					default: "bg-background text-foreground",
					destructive:
						"border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive",
					warning: "border-yellow-500/50 text-yellow-600 dark:border-yellow-500 [&>svg]:text-yellow-500",
					success: "border-green-500/50 text-green-600 dark:border-green-500 [&>svg]:text-green-500",
					info: "border-blue-500/50 text-blue-600 dark:border-blue-500 [&>svg]:text-blue-500",
				},
			},
			defaultVariants: {
				variant: "default",
			},
		}
	);

	type Props = HTMLAttributes<HTMLDivElement> & VariantProps<typeof alertVariants>;

	let { class: className, variant, children, ...restProps }: Props = $props();
</script>

<div
	role="alert"
	class={cn(alertVariants({ variant }), className)}
	{...restProps}
>
	{@render children?.()}
</div>
