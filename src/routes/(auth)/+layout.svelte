<script lang="ts">
	import '../layout.css';
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>FPMB</title>
</svelte:head>

<div class="h-screen w-screen overflow-hidden bg-neutral-900 text-neutral-50 flex items-center justify-center">
	{@render children()}
</div>
