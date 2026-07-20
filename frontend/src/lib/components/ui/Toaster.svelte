<script lang="ts">
	import * as Alert from "$lib/components/ui/alert";
	import { toastStore } from "$lib/stores/toast.svelte";
	import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
	import AlertCircle from "lucide-svelte/icons/alert-circle";
	import Info from "lucide-svelte/icons/info";
	import AlertTriangle from "lucide-svelte/icons/alert-triangle";
	import { flip } from "svelte/animate";
	import { fade, fly } from "svelte/transition";

	const iconMap = {
		success: CheckCircle2,
		error: AlertCircle,
		warning: AlertTriangle,
		info: Info,
	};

	const variantMap = {
		success: "success",
		error: "destructive",
		warning: "warning",
		info: "info",
	} as const;
</script>

<div class="fixed top-4 right-4 z-[100] flex flex-col gap-2 w-full max-w-[400px] pointer-events-none">
	{#each toastStore.toasts as toast (toast.id)}
		<div
			animate:flip={{ duration: 200 }}
			in:fly={{ x: 100, opacity: 0, duration: 300 }}
			out:fade={{ duration: 200 }}
			class="pointer-events-auto shadow-2xl"
		>
			<Alert.Root variant={variantMap[toast.type]}>
				{@const IconComp = iconMap[toast.type]}
				<IconComp class="h-4 w-4" />
				<Alert.Title>{toast.title}</Alert.Title>
				{#if toast.description}
					<Alert.Description>{toast.description}</Alert.Description>
				{/if}
			</Alert.Root>
		</div>
	{/each}
</div>
