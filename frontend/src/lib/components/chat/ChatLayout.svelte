<script lang="ts">
	import { cn } from '$lib/utils';
	import Sidebar from './Sidebar.svelte';
	import ChatPanel from './ChatPanel.svelte';
	import SettingsPanel from '../settings/SettingsPanel.svelte';
	import LogPanel from './LogPanel.svelte';
	import GitPanel from '../git/GitPanel.svelte';

	interface ChatLayoutProps {
		class?: string;
	}

	let { class: className }: ChatLayoutProps = $props();

	let sidebarOpen = $state(true);
	let settingsOpen = $state(false);
	let logsOpen = $state(false);
	let gitOpen = $state(false);
	let activeWorkspace = $state('');
	let activeSessionID = $state('');
	let settingsCategory = $state<string>('general');
	let settingsEntity = $state<Record<string, any> | null>(null);

	function openSettings(category = 'general', entity: Record<string, any> | null = null) {
		settingsCategory = category;
		settingsEntity = entity;
		settingsOpen = true;
		sidebarOpen = false;
	}

	function closeSettings() {
		settingsOpen = false;
		sidebarOpen = true;
		settingsEntity = null;
	}

	function toggleSidebar() {
		if (settingsOpen) {
			closeSettings();
			return;
		}
		sidebarOpen = !sidebarOpen;
	}

	function toggleLogs() {
		logsOpen = !logsOpen;
		gitOpen = false;
	}

	function toggleGit() {
		gitOpen = !gitOpen;
		logsOpen = false;
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
	{#if sidebarOpen}
		<Sidebar
			onOpenSettings={() => openSettings()}
			onNewWorkspace={() => openSettings('workspaces')}
			onEditWorkspace={(ws) => openSettings('workspaces', ws)}
			onOpenLogs={toggleLogs}
			onOpenGit={toggleGit}
			bind:activeWorkspace
			bind:activeSessionID
		/>
	{/if}

	{#if settingsOpen}
		<SettingsPanel
			onClose={closeSettings}
			initialCategory={settingsCategory as any}
			initialEntity={settingsEntity}
		/>
	{/if}

	<ChatPanel
		{sidebarOpen}
		{activeWorkspace}
		bind:activeSessionID
		onToggleSidebar={toggleSidebar}
	/>

	{#if logsOpen}
		<LogPanel onClose={() => logsOpen = false} />
	{/if}

	{#if gitOpen}
		<GitPanel onClose={() => gitOpen = false} {activeWorkspace} />
	{/if}
</div>
