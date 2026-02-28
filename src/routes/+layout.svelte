<script lang="ts">
	import "./layout.css";
	import favicon from "$lib/assets/favicon.svg";

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>FPMB — Free Project Management Boards</title>
	<meta
		name="description"
		content="FPMB is a self-hosted, open-source project management platform with Kanban boards, team collaboration, whiteboards, docs, and more."
	/>
</svelte:head>
{@render children()}
