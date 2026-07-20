// ── ADA LOVE IDE — Frontend Library Exports ────────────────────────

// Utils
export { cn } from '$lib/utils';

// Theme system
export { theme } from '$lib/stores/theme.svelte';
export * from '$lib/themes/definitions';

// Components — UI primitives
export { Button } from '$lib/components/ui/button';
export { Switch } from '$lib/components/ui/switch';
export { Separator } from '$lib/components/ui/separator';
export {
	DropdownMenu,
	DropdownTrigger,
	DropdownContent,
	DropdownItem,
} from '$lib/components/ui/dropdown';

// Components — Icon adapter
export { Icon } from '$lib/components/icon';
