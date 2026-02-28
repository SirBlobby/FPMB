<script lang="ts">
	import { goto } from "$app/navigation";
	import { authStore } from "$lib/stores/auth.svelte";

	let email = $state("");
	let password = $state("");
	let isLoading = $state(false);
	let error = $state("");

	async function handleSubmit(event: Event) {
		event.preventDefault();
		isLoading = true;
		error = "";
		try {
			await authStore.login(email, password);
			goto("/");
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : "Login failed";
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Sign In — FPMB</title>
	<meta
		name="description"
		content="Sign in to your FPMB account to manage your projects, teams, and tasks."
	/>
</svelte:head>

<div
	class="w-full max-w-md p-8 bg-neutral-800 rounded-lg shadow-xl border border-neutral-700"
>
	<div class="text-center mb-8">
		<h1 class="text-3xl font-bold text-white tracking-tight">FPMB</h1>
		<p class="text-neutral-400 mt-2">Sign in to your account</p>
	</div>

	<form onsubmit={handleSubmit} class="space-y-6">
		{#if error}
			<div
				class="rounded-md bg-red-900/50 border border-red-700 p-3 text-sm text-red-300"
			>
				{error}
			</div>
		{/if}
		<div>
			<label for="email" class="block text-sm font-medium text-neutral-300"
				>Email address</label
			>
			<div class="mt-1">
				<input
					id="email"
					name="email"
					type="email"
					autocomplete="email"
					required
					bind:value={email}
					class="appearance-none block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
			</div>
		</div>

		<div>
			<label for="password" class="block text-sm font-medium text-neutral-300"
				>Password</label
			>
			<div class="mt-1">
				<input
					id="password"
					name="password"
					type="password"
					autocomplete="current-password"
					required
					bind:value={password}
					class="appearance-none block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
			</div>
		</div>

		<div class="flex items-center justify-between">
			<div class="flex items-center">
				<input
					id="remember-me"
					name="remember-me"
					type="checkbox"
					class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-neutral-600 rounded bg-neutral-700"
				/>
				<label for="remember-me" class="ml-2 block text-sm text-neutral-300"
					>Remember me</label
				>
			</div>

			<div class="text-sm">
				<a
					href="/forgot-password"
					class="font-medium text-blue-500 hover:text-blue-400"
					>Forgot your password?</a
				>
			</div>
		</div>

		<div>
			<button
				type="submit"
				disabled={isLoading}
				class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 focus:ring-offset-neutral-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
			>
				{#if isLoading}
					Signing in...
				{:else}
					Sign in
				{/if}
			</button>
		</div>
	</form>

	<div class="mt-6 text-center text-sm">
		<span class="text-neutral-400">Don't have an account?</span>
		<a
			href="/register"
			class="font-medium text-blue-500 hover:text-blue-400 ml-1">Sign up</a
		>
	</div>
</div>
