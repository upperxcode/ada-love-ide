<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import { theme } from '$lib/stores/theme.svelte';
	import { onMount } from 'svelte';

	import Toaster from '$lib/components/ui/Toaster.svelte';
	import '../app.css';

	let { children } = $props();

	onMount(() => {
		// Initialize theme store — applies CSS variables to <html>
		theme.init();
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css" />
</svelte:head>

	{@render children()}
	<Toaster />
