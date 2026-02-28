<script lang="ts">
	import { page } from "$app/stores";
	import { goto } from "$app/navigation";
	import { onMount } from "svelte";
	import { authStore } from "$lib/stores/auth.svelte";
	import { notifications as notifApi } from "$lib/api";

	let { children } = $props();

	let currentPath = $derived($page.url.pathname);
	let userInitial = $derived(
		authStore.user?.name?.charAt(0).toUpperCase() ?? "U",
	);
	let unreadCount = $state(0);

	onMount(async () => {
		await authStore.init();
		if (!authStore.user) {
			goto("/login");
			return;
		}
		try {
			const all = await notifApi.list();
			unreadCount = all.filter((n) => !n.read).length;
		} catch {
			unreadCount = 0;
		}
	});

	async function logout() {
		await authStore.logout();
		goto("/login");
	}
</script>

<div
	class="h-screen w-screen bg-neutral-900 text-neutral-50 flex flex-col overflow-hidden"
>
	<!-- Top Navbar -->
	<header
		class="h-16 shrink-0 bg-neutral-800 border-b border-neutral-700 flex items-center justify-between px-6 z-20 relative"
	>
		<div class="flex items-center space-x-8">
			<a
				href="/"
				class="text-xl font-bold tracking-tight text-white hover:text-blue-400 transition-colors"
				>FPMB</a
			>

			<nav class="hidden md:flex space-x-2">
				<a
					href="/"
					class="px-3 py-2 text-sm font-medium rounded-md transition-colors {currentPath ===
					'/'
						? 'bg-blue-600 text-white'
						: 'text-neutral-300 hover:bg-neutral-700 hover:text-white'}"
				>
					Dashboard
				</a>
				<a
					href="/projects"
					class="px-3 py-2 text-sm font-medium rounded-md transition-colors {currentPath.startsWith(
						'/projects',
					)
						? 'bg-blue-600 text-white'
						: 'text-neutral-300 hover:bg-neutral-700 hover:text-white'}"
				>
					Projects
				</a>
				<a
					href="/calendar"
					class="px-3 py-2 text-sm font-medium rounded-md transition-colors {currentPath.startsWith(
						'/calendar',
					)
						? 'bg-blue-600 text-white'
						: 'text-neutral-300 hover:bg-neutral-700 hover:text-white'}"
				>
					Calendar
				</a>
				<a
					href="/files"
					class="px-3 py-2 text-sm font-medium rounded-md transition-colors {currentPath.startsWith(
						'/files',
					)
						? 'bg-blue-600 text-white'
						: 'text-neutral-300 hover:bg-neutral-700 hover:text-white'}"
				>
					Files
				</a>
				<a
					href="/docs"
					class="px-3 py-2 text-sm font-medium rounded-md transition-colors flex items-center gap-1.5 {currentPath ===
					'/docs'
						? 'bg-blue-600 text-white'
						: 'text-neutral-300 hover:bg-neutral-700 hover:text-white'}"
				>
					<svg
						class="w-3.5 h-3.5"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
						><path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
						/></svg
					>
					Docs
				</a>
				<a
					href="/api-docs"
					class="px-3 py-2 text-sm font-medium rounded-md transition-colors flex items-center gap-1.5 {currentPath.startsWith(
						'/api-docs',
					)
						? 'bg-blue-600 text-white'
						: 'text-neutral-300 hover:bg-neutral-700 hover:text-white'}"
				>
					<svg
						class="w-3.5 h-3.5"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
						><path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
						/></svg
					>
					API Docs
				</a>
			</nav>
		</div>

		<div class="flex items-center space-x-4">
			<a
				href="/notifications"
				class="text-neutral-400 hover:text-white transition-colors p-2 rounded-full hover:bg-neutral-700 relative"
			>
				<svg
					class="h-5 w-5"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
					/>
				</svg>
				{#if unreadCount > 0}
					<span
						class="absolute top-1 right-1 min-w-[1.1rem] h-[1.1rem] bg-red-500 rounded-full border border-neutral-800 flex items-center justify-center text-[10px] font-bold text-white leading-none px-0.5"
					>
						{unreadCount > 99 ? "99+" : unreadCount}
					</span>
				{/if}
			</a>

			<!-- Mobile menu button -->
			<button
				class="md:hidden text-neutral-400 hover:text-white transition-colors p-2"
				aria-label="Open menu"
			>
				<svg
					class="h-6 w-6"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 6h16M4 12h16M4 18h16"
					/>
				</svg>
			</button>

			<a
				href="/settings/user"
				class="hidden md:flex items-center space-x-3 p-1 rounded-full hover:bg-neutral-700 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-neutral-800 focus:ring-blue-500"
			>
				<div
					class="h-8 w-8 rounded-full bg-blue-600 flex items-center justify-center text-sm font-medium text-white shadow-sm"
				>
					{userInitial}
				</div>
			</a>

			<button
				onclick={logout}
				class="hidden md:flex items-center p-2 rounded-md text-neutral-400 hover:text-white hover:bg-neutral-700 transition-colors"
				aria-label="Log out"
			>
				<svg
					class="h-5 w-5"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
					/>
				</svg>
			</button>
		</div>
	</header>

	<!-- Main Content -->
	<main class="flex-1 flex flex-col min-w-0 overflow-hidden relative">
		<div class="flex-1 overflow-auto">
			<div class="flex flex-col min-h-full">
				<div class="flex-1 p-6 lg:p-8">
					{@render children()}
				</div>

				<!-- Footer -->
				<footer class="shrink-0 px-6 lg:px-8 pb-6 pt-6">
					<div class="border-t border-neutral-800 pt-5">
						<div
							class="flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-neutral-600"
						>
							<div class="flex items-center gap-1.5">
								<span class="font-semibold text-neutral-500">FPMB</span>
								<span>&middot;</span>
								<span>Free Project Management Boards</span>
							</div>
							<div class="flex items-center gap-4">
								<a
									href="/api-docs"
									class="hover:text-neutral-400 transition-colors">API Docs</a
								>
								<a
									href="/settings/user"
									class="hover:text-neutral-400 transition-colors">Settings</a
								>
								<span>v0.1.0</span>
							</div>
						</div>
					</div>
				</footer>
			</div>
		</div>
	</main>
</div>
