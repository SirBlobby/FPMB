<script lang="ts">
	import { onMount } from "svelte";
	import { users as usersApi, apiKeys as apiKeysApi } from "$lib/api";
	import { authStore } from "$lib/stores/auth.svelte";
	import type { ApiKey, ApiKeyCreated } from "$lib/types/api";

	// --- Profile state ---
	let name = $state("");
	let email = $state("");
	let currentPassword = $state("");
	let newPassword = $state("");
	let confirmPassword = $state("");
	let profileLoading = $state(false);
	let passwordLoading = $state(false);
	let avatarLoading = $state(false);
	let profileError = $state("");
	let profileSuccess = $state("");
	let passwordError = $state("");
	let passwordSuccess = $state("");
	let avatarError = $state("");
	let avatarSuccess = $state("");
	let avatarCacheBust = $state("");

	// --- API Keys state ---
	const ALL_SCOPES = [
		{ id: "read:projects", label: "Read Projects", group: "Projects" },
		{ id: "write:projects", label: "Write Projects", group: "Projects" },
		{ id: "read:boards", label: "Read Boards", group: "Boards" },
		{ id: "write:boards", label: "Write Boards", group: "Boards" },
		{ id: "read:teams", label: "Read Teams", group: "Teams" },
		{ id: "write:teams", label: "Write Teams", group: "Teams" },
		{ id: "read:files", label: "Read Files", group: "Files" },
		{ id: "write:files", label: "Write Files", group: "Files" },
		{
			id: "read:notifications",
			label: "Read Notifications",
			group: "Notifications",
		},
	];

	let apiKeyList = $state<ApiKey[]>([]);
	let apiKeysLoading = $state(true);
	let newKeyName = $state("");
	let newKeyScopes = $state<Record<string, boolean>>({});
	let creatingKey = $state(false);
	let newKeyError = $state("");
	let createdKey = $state<ApiKeyCreated | null>(null);
	let copiedKey = $state(false);
	let showCreateForm = $state(false);

	onMount(async () => {
		if (authStore.user) {
			name = authStore.user.name;
			email = authStore.user.email;
		}
		try {
			apiKeyList = await apiKeysApi.list();
		} catch {
			apiKeyList = [];
		} finally {
			apiKeysLoading = false;
		}
	});

	let userInitial = $derived(name?.charAt(0).toUpperCase() ?? "U");

	// --- Profile handlers ---
	async function uploadAvatar(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		avatarLoading = true;
		avatarError = "";
		avatarSuccess = "";
		try {
			const updated = await usersApi.uploadAvatar(file);
			authStore.setUser(updated);
			avatarCacheBust = "?t=" + Date.now();
			avatarSuccess = "Avatar updated successfully.";
		} catch (err: unknown) {
			avatarError =
				err instanceof Error ? err.message : "Failed to upload avatar";
		} finally {
			avatarLoading = false;
			input.value = "";
		}
	}

	async function saveProfile(e: Event) {
		e.preventDefault();
		profileLoading = true;
		profileError = "";
		profileSuccess = "";
		try {
			const updated = await usersApi.updateMe({ name, email });
			authStore.setUser(updated);
			profileSuccess = "Profile updated successfully.";
		} catch (err: unknown) {
			profileError = err instanceof Error ? err.message : "Failed to save";
		} finally {
			profileLoading = false;
		}
	}

	async function savePassword(e: Event) {
		e.preventDefault();
		if (newPassword !== confirmPassword) {
			passwordError = "Passwords do not match";
			return;
		}
		passwordLoading = true;
		passwordError = "";
		passwordSuccess = "";
		try {
			await usersApi.changePassword(currentPassword, newPassword);
			currentPassword = "";
			newPassword = "";
			confirmPassword = "";
			passwordSuccess = "Password updated successfully.";
		} catch (err: unknown) {
			passwordError =
				err instanceof Error ? err.message : "Failed to update password";
		} finally {
			passwordLoading = false;
		}
	}

	// --- API Key handlers ---
	function openCreateForm() {
		newKeyName = "";
		newKeyScopes = {};
		newKeyError = "";
		createdKey = null;
		showCreateForm = true;
	}

	function closeCreateForm() {
		showCreateForm = false;
	}

	let selectedScopeCount = $derived(
		Object.values(newKeyScopes).filter(Boolean).length,
	);

	async function createKey(e: Event) {
		e.preventDefault();
		const scopes = ALL_SCOPES.filter((s) => newKeyScopes[s.id]).map(
			(s) => s.id,
		);
		if (!newKeyName.trim()) {
			newKeyError = "Name is required.";
			return;
		}
		if (scopes.length === 0) {
			newKeyError = "Select at least one scope.";
			return;
		}
		creatingKey = true;
		newKeyError = "";
		try {
			const result = await apiKeysApi.create(newKeyName.trim(), scopes);
			createdKey = result;
			apiKeyList = await apiKeysApi.list();
			newKeyName = "";
			newKeyScopes = {};
			showCreateForm = false;
		} catch (err: unknown) {
			newKeyError =
				err instanceof Error ? err.message : "Failed to create key.";
		} finally {
			creatingKey = false;
		}
	}

	async function revokeKey(keyId: string) {
		if (!confirm("Revoke this API key? Any apps using it will lose access."))
			return;
		try {
			await apiKeysApi.revoke(keyId);
			apiKeyList = apiKeyList.filter((k) => k.id !== keyId);
			if (createdKey?.id === keyId) createdKey = null;
		} catch {
			/* ignore */
		}
	}

	async function copyKey() {
		if (!createdKey) return;
		await navigator.clipboard.writeText(createdKey.key);
		copiedKey = true;
		setTimeout(() => (copiedKey = false), 2500);
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleDateString("en-US", {
			year: "numeric",
			month: "short",
			day: "numeric",
		});
	}

	const scopeGroups = ALL_SCOPES.reduce<Record<string, typeof ALL_SCOPES>>(
		(acc, s) => {
			(acc[s.group] ??= []).push(s);
			return acc;
		},
		{},
	);
</script>

<svelte:head>
	<title>User Settings — FPMB</title>
	<meta
		name="description"
		content="Manage your FPMB profile, avatar, account password, and API keys."
	/>
</svelte:head>

<div class="max-w-4xl mx-auto space-y-10">
	<div>
		<h1 class="text-3xl font-bold text-white tracking-tight mb-2">
			User Settings
		</h1>
		<p class="text-neutral-400">Manage your profile and account preferences.</p>
	</div>

	<!-- Profile Section -->
	<section
		class="bg-neutral-800 rounded-lg shadow-sm border border-neutral-700 overflow-hidden"
	>
		<div class="p-6 border-b border-neutral-700">
			<h2 class="text-xl font-semibold text-white mb-1">Profile Information</h2>
			<p class="text-sm text-neutral-400">
				Update your account's profile information and email address.
			</p>
		</div>

		<form onsubmit={saveProfile} class="p-6 space-y-6">
			{#if profileError}
				<div
					class="rounded-md bg-red-900/50 border border-red-700 p-3 text-sm text-red-300"
				>
					{profileError}
				</div>
			{/if}
			{#if profileSuccess}
				<div
					class="rounded-md bg-green-900/50 border border-green-700 p-3 text-sm text-green-300"
				>
					{profileSuccess}
				</div>
			{/if}

			<div class="flex items-center space-x-6">
				<div class="shrink-0">
					{#if authStore.user?.avatar_url}
						<img
							src="{authStore.user.avatar_url}{avatarCacheBust}"
							alt="Avatar"
							class="h-16 w-16 rounded-full object-cover shadow-inner"
						/>
					{:else}
						<div
							class="h-16 w-16 rounded-full bg-blue-600 flex items-center justify-center text-xl font-medium text-white shadow-inner"
						>
							{userInitial}
						</div>
					{/if}
				</div>
				<div class="space-y-2">
					{#if avatarError}<p class="text-sm text-red-400">
							{avatarError}
						</p>{/if}
					{#if avatarSuccess}<p class="text-sm text-green-400">
							{avatarSuccess}
						</p>{/if}
					<label
						class="cursor-pointer inline-flex items-center gap-2 bg-neutral-700 hover:bg-neutral-600 text-white text-sm font-medium py-1.5 px-4 rounded-md border border-neutral-600 transition-colors {avatarLoading
							? 'opacity-50 pointer-events-none'
							: ''}"
					>
						{avatarLoading ? "Uploading..." : "Upload Avatar"}
						<input
							type="file"
							accept=".jpg,.jpeg,.png,.gif,.webp"
							class="hidden"
							onchange={uploadAvatar}
							disabled={avatarLoading}
						/>
					</label>
				</div>
			</div>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<div>
					<label for="name" class="block text-sm font-medium text-neutral-300"
						>Full Name</label
					>
					<input
						type="text"
						id="name"
						bind:value={name}
						class="mt-1 block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
					/>
				</div>
				<div>
					<label for="email" class="block text-sm font-medium text-neutral-300"
						>Email Address</label
					>
					<input
						type="email"
						id="email"
						bind:value={email}
						class="mt-1 block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
					/>
				</div>
			</div>

			<div class="flex justify-end pt-4 border-t border-neutral-700 mt-6">
				<button
					type="submit"
					disabled={profileLoading}
					class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md shadow-sm border border-transparent transition-colors text-sm disabled:opacity-50"
				>
					{profileLoading ? "Saving..." : "Save Changes"}
				</button>
			</div>
		</form>
	</section>

	<!-- Security Section -->
	<section
		class="bg-neutral-800 rounded-lg shadow-sm border border-neutral-700 overflow-hidden"
	>
		<div class="p-6 border-b border-neutral-700">
			<h2 class="text-xl font-semibold text-white mb-1">Update Password</h2>
			<p class="text-sm text-neutral-400">
				Ensure your account is using a long, random password to stay secure.
			</p>
		</div>

		<form onsubmit={savePassword} class="p-6 space-y-6">
			{#if passwordError}
				<div
					class="rounded-md bg-red-900/50 border border-red-700 p-3 text-sm text-red-300"
				>
					{passwordError}
				</div>
			{/if}
			{#if passwordSuccess}
				<div
					class="rounded-md bg-green-900/50 border border-green-700 p-3 text-sm text-green-300"
				>
					{passwordSuccess}
				</div>
			{/if}

			<div class="max-w-md space-y-4">
				<div>
					<label
						for="current_password"
						class="block text-sm font-medium text-neutral-300"
						>Current Password</label
					>
					<input
						type="password"
						id="current_password"
						bind:value={currentPassword}
						class="mt-1 block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
					/>
				</div>
				<div>
					<label
						for="new_password"
						class="block text-sm font-medium text-neutral-300"
						>New Password</label
					>
					<input
						type="password"
						id="new_password"
						bind:value={newPassword}
						class="mt-1 block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
					/>
				</div>
				<div>
					<label
						for="confirm_password"
						class="block text-sm font-medium text-neutral-300"
						>Confirm Password</label
					>
					<input
						type="password"
						id="confirm_password"
						bind:value={confirmPassword}
						class="mt-1 block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
					/>
				</div>
			</div>

			<div class="flex justify-end pt-4 border-t border-neutral-700 mt-6">
				<button
					type="submit"
					disabled={passwordLoading}
					class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md shadow-sm border border-transparent transition-colors text-sm disabled:opacity-50"
				>
					{passwordLoading ? "Updating..." : "Update Password"}
				</button>
			</div>
		</form>
	</section>

	<!-- API Keys Section -->
	<section
		class="bg-neutral-800 rounded-lg shadow-sm border border-neutral-700 overflow-hidden"
	>
		<div
			class="p-6 border-b border-neutral-700 flex items-center justify-between gap-4"
		>
			<div>
				<h2 class="text-xl font-semibold text-white mb-1">API Keys</h2>
				<p class="text-sm text-neutral-400">
					Generate personal API keys with granular scopes for programmatic
					access.
				</p>
			</div>
			{#if !showCreateForm}
				<button
					onclick={openCreateForm}
					class="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md text-sm transition-colors border border-transparent shrink-0"
				>
					<svg
						class="w-4 h-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
						><path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 4v16m8-8H4"
						/></svg
					>
					New API Key
				</button>
			{/if}
		</div>

		<!-- Newly-created key banner (shown once) -->
		{#if createdKey}
			<div
				class="mx-6 mt-6 rounded-lg border border-green-600/40 bg-green-900/20 p-4"
			>
				<div class="flex items-start gap-3">
					<svg
						class="w-5 h-5 text-green-400 mt-0.5 shrink-0"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<div class="flex-1 min-w-0">
						<p class="text-sm font-semibold text-green-300 mb-2">
							Key created — copy it now, it won't be shown again.
						</p>
						<div class="flex items-center gap-2">
							<code
								class="flex-1 text-xs font-mono bg-neutral-900 border border-neutral-600 rounded px-3 py-2 text-green-300 truncate select-all"
								>{createdKey.key}</code
							>
							<button
								onclick={copyKey}
								class="shrink-0 flex items-center gap-1.5 text-xs font-medium px-3 py-2 rounded-md border transition-colors {copiedKey
									? 'bg-green-700 border-green-600 text-white'
									: 'bg-neutral-700 border-neutral-600 text-neutral-200 hover:bg-neutral-600'}"
							>
								{#if copiedKey}
									<svg
										class="w-3.5 h-3.5"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
										><path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M5 13l4 4L19 7"
										/></svg
									>
									Copied!
								{:else}
									<svg
										class="w-3.5 h-3.5"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
										><path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
										/></svg
									>
									Copy
								{/if}
							</button>
						</div>
					</div>
					<button
						onclick={() => (createdKey = null)}
						class="text-neutral-500 hover:text-white transition-colors p-0.5 rounded"
						aria-label="Dismiss banner"
					>
						<svg
							class="w-4 h-4"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
							><path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M6 18L18 6M6 6l12 12"
							/></svg
						>
					</button>
				</div>
			</div>
		{/if}

		<!-- Create form -->
		{#if showCreateForm}
			<form
				onsubmit={createKey}
				class="p-6 space-y-6 border-b border-neutral-700"
			>
				<div>
					<label
						for="api-key-name"
						class="block text-sm font-medium text-neutral-300 mb-1.5"
						>Key Name</label
					>
					<input
						id="api-key-name"
						type="text"
						bind:value={newKeyName}
						placeholder="e.g. CI / CD Pipeline"
						class="w-full max-w-sm px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>

				<div>
					<p class="text-sm font-medium text-neutral-300 mb-3">
						Scopes <span class="text-neutral-500 font-normal"
							>({selectedScopeCount} selected)</span
						>
					</p>
					<div
						class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-x-8 gap-y-5"
					>
						{#each Object.entries(scopeGroups) as [group, scopes]}
							<div>
								<p
									class="text-xs font-semibold text-neutral-400 uppercase tracking-wider mb-2"
								>
									{group}
								</p>
								<div class="space-y-2">
									{#each scopes as scope}
										<label
											class="flex items-center gap-2.5 cursor-pointer group"
											for="scope-{scope.id}"
										>
											<input
												type="checkbox"
												id="scope-{scope.id}"
												bind:checked={newKeyScopes[scope.id]}
												class="w-4 h-4 rounded border-neutral-600 bg-neutral-700 text-blue-500 focus:ring-blue-500 focus:ring-offset-neutral-800 cursor-pointer"
											/>
											<span
												class="text-sm text-neutral-300 group-hover:text-white transition-colors"
												>{scope.label}</span
											>
										</label>
									{/each}
								</div>
							</div>
						{/each}
					</div>
				</div>

				{#if newKeyError}
					<p class="text-sm text-red-400">{newKeyError}</p>
				{/if}

				<div class="flex items-center gap-3 pt-2">
					<button
						type="submit"
						disabled={creatingKey}
						class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-5 rounded-md text-sm transition-colors disabled:opacity-50 flex items-center gap-2"
					>
						{#if creatingKey}
							<svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
								<circle
									class="opacity-25"
									cx="12"
									cy="12"
									r="10"
									stroke="currentColor"
									stroke-width="4"
								></circle>
								<path
									class="opacity-75"
									fill="currentColor"
									d="M4 12a8 8 0 018-8v8z"
								></path>
							</svg>
						{/if}
						Generate Key
					</button>
					<button
						type="button"
						onclick={closeCreateForm}
						class="text-sm font-medium text-neutral-400 hover:text-white transition-colors px-3 py-2 rounded hover:bg-neutral-700"
					>
						Cancel
					</button>
				</div>
			</form>
		{/if}

		<!-- Key list -->
		<div class="divide-y divide-neutral-700">
			{#if apiKeysLoading}
				<div class="p-8 text-center text-neutral-500 text-sm">
					Loading keys...
				</div>
			{:else if apiKeyList.length === 0 && !createdKey}
				<div
					class="p-10 flex flex-col items-center justify-center text-neutral-500"
				>
					<svg
						class="w-10 h-10 mb-3 opacity-40"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="1.5"
							d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
						/>
					</svg>
					<p class="text-sm">No API keys yet. Create one above.</p>
				</div>
			{:else}
				{#each apiKeyList as key (key.id)}
					<div
						class="px-6 py-4 flex items-start justify-between gap-4 group hover:bg-white/2 transition-colors"
					>
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-3 mb-1.5">
								<span class="text-sm font-semibold text-white">{key.name}</span>
								<code
									class="text-xs font-mono bg-neutral-900 border border-neutral-700 text-neutral-400 px-2 py-0.5 rounded"
									>{key.prefix}…</code
								>
							</div>
							<div class="flex flex-wrap gap-1.5 mb-2">
								{#each key.scopes as scope}
									<span
										class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-900/40 text-blue-300 border border-blue-700/40"
										>{scope}</span
									>
								{/each}
							</div>
							<p class="text-xs text-neutral-500">
								Created {formatDate(key.created_at)}{#if key.last_used}
									· Last used {formatDate(key.last_used)}{/if}
							</p>
						</div>
						<button
							onclick={() => revokeKey(key.id)}
							class="shrink-0 flex items-center gap-1.5 text-xs font-medium text-neutral-500 hover:text-red-400 transition-colors px-2 py-1.5 rounded hover:bg-red-900/20 opacity-0 group-hover:opacity-100"
							title="Revoke this key"
						>
							<svg
								class="w-3.5 h-3.5"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
								/>
							</svg>
							Revoke
						</button>
					</div>
				{/each}
			{/if}
		</div>
	</section>
</div>
