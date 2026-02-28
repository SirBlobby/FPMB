<script lang="ts">
	import { goto } from "$app/navigation";
	import { authStore } from "$lib/stores/auth.svelte";

	let name = $state("");
	let email = $state("");
	let password = $state("");
	let confirmPassword = $state("");
	let isLoading = $state(false);
	let error = $state("");

	async function handleSubmit(event: Event) {
		event.preventDefault();
		if (password !== confirmPassword) {
			error = "Passwords do not match";
			return;
		}
		isLoading = true;
		error = "";
		try {
			await authStore.register(name, email, password);
			goto("/");
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : "Registration failed";
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Create Account — FPMB</title>
	<meta
		name="description"
		content="Create a new FPMB account to start managing projects and collaborating with your team."
	/>
</svelte:head>

<div
	class="w-full max-w-md p-8 bg-neutral-800 rounded-lg shadow-xl border border-neutral-700"
>
	<div class="text-center mb-8">
		<h1 class="text-3xl font-bold text-white tracking-tight">FPMB</h1>
		<p class="text-neutral-400 mt-2">Create a new account</p>
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
			<label for="name" class="block text-sm font-medium text-neutral-300"
				>Full name</label
			>
			<div class="mt-1">
				<input
					id="name"
					name="name"
					type="text"
					autocomplete="name"
					required
					bind:value={name}
					class="appearance-none block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
			</div>
		</div>
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
					autocomplete="new-password"
					required
					bind:value={password}
					class="appearance-none block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
			</div>
		</div>

		<div>
			<label
				for="confirmPassword"
				class="block text-sm font-medium text-neutral-300"
				>Confirm Password</label
			>
			<div class="mt-1">
				<input
					id="confirmPassword"
					name="confirmPassword"
					type="password"
					autocomplete="new-password"
					required
					bind:value={confirmPassword}
					class="appearance-none block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
			</div>
		</div>

		<div>
			<button
				type="submit"
				disabled={isLoading}
				class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 focus:ring-offset-neutral-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
			>
				{#if isLoading}
					Creating account...
				{:else}
					Create account
				{/if}
			</button>
		</div>
	</form>

	<div class="mt-6 text-center text-sm">
		<span class="text-neutral-400">Already have an account?</span>
		<a href="/login" class="font-medium text-blue-500 hover:text-blue-400 ml-1"
			>Sign in</a
		>
	</div>
</div>
