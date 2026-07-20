<script lang="ts">
	import {
		Dialog,
		DialogContent,
		DialogDescription,
		DialogFooter,
		DialogHeader,
		DialogTitle,
		DialogOverlay,
		DialogPortal
	} from "$lib/components/ui/dialog";
	import { Button } from "$lib/components/ui/button";
	import AlertTriangle from "lucide-svelte/icons/alert-triangle";
	import { cn } from "$lib/utils";

	interface ConfirmDeleteDialogProps {
		open: boolean;
		onOpenChange: (open: boolean) => void;
		title?: string;
		description?: string;
		onConfirm: () => void;
		loading?: boolean;
	}

	let {
		open = $bindable(),
		onOpenChange,
		title = "Are you sure?",
		description = "This action cannot be undone. This will permanently delete the item.",
		onConfirm,
		loading = false
	}: ConfirmDeleteDialogProps = $props();

	function handleConfirm() {
		onConfirm();
		onOpenChange(false);
	}
</script>

<Dialog bind:open onOpenChange={onOpenChange}>
	<DialogPortal>
		<DialogOverlay />
		<DialogContent class="sm:max-w-[400px]">
			<DialogHeader>
				<div class="flex items-center gap-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-full bg-orange-500/10 text-orange-500">
						<AlertTriangle class="h-5 w-5" />
					</div>
					<DialogTitle>{title}</DialogTitle>
				</div>
				<DialogDescription class="pt-2">
					{description}
				</DialogDescription>
			</DialogHeader>
			<DialogFooter class="mt-4 gap-2 sm:gap-0">
				<Button variant="ghost" onclick={() => onOpenChange(false)} disabled={loading}>
					Cancel
				</Button>
				<Button 
					variant="destructive" 
					onclick={handleConfirm} 
					disabled={loading}
					class="bg-orange-600 hover:bg-orange-700 text-white border-none"
				>
					{loading ? "Deleting..." : "Delete"}
				</Button>
			</DialogFooter>
		</DialogContent>
	</DialogPortal>
</Dialog>
