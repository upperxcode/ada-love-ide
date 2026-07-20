<script lang="ts">
	import { cn } from '$lib/utils';
	import Sidebar from './Sidebar.svelte';
	import ChatPanel from './ChatPanel.svelte';
	import SettingsPanel from '../settings/SettingsPanel.svelte';

	interface ChatLayoutProps {
		class?: string;
	}

	let { class: className }: ChatLayoutProps = $props();

	let sidebarOpen = $state(true);
	let settingsOpen = $state(false);

	function openSettings() {
		settingsOpen = true;
		sidebarOpen = false;
	}

	function closeSettings() {
		settingsOpen = false;
		sidebarOpen = true;
	}

	function toggleSidebar() {
		// If settings is open, close it first and go back to sidebar
		if (settingsOpen) {
			closeSettings();
			return;
		}
		sidebarOpen = !sidebarOpen;
	}
</script>

<div
	class={cn(
		'flex h-dvh w-screen overflow-hidden',
		'bg-[var(--bg-primary)] text-[var(--text-primary)]',
		'font-[family-name:var(--font-sans)]',
		className
	)}
>
	<!-- Sidebar — mutually exclusive with Settings, both on the left -->
	{#if sidebarOpen}
		<Sidebar onOpenSettings={openSettings} />
	{/if}

	<!-- Settings — replaces sidebar on the left -->
	{#if settingsOpen}
		<SettingsPanel onClose={closeSettings} />
	{/if}

	<!-- Chat Panel — fills remaining space -->
	<ChatPanel
		{sidebarOpen}
		onToggleSidebar={toggleSidebar}
	/>
</div>
