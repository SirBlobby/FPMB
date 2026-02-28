<script lang="ts">
	import { onMount } from "svelte";
	import { page } from "$app/stores";
	import { goto } from "$app/navigation";
	import { projects as projectsApi } from "$lib/api";
	import type { Project } from "$lib/types/api";

	let projectId = $derived($page.params.id ?? "");

	let project = $state<Project | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let saveError = $state("");
	let saveSuccess = $state(false);

	let projectName = $state("");
	let projectDescription = $state("");

	onMount(async () => {
		try {
			project = await projectsApi.get(projectId);
			projectName = project.name;
			projectDescription = project.description;
		} catch {
		} finally {
			loading = false;
		}
	});

	async function saveSettings(e: SubmitEvent) {
		e.preventDefault();
		saving = true;
		saveError = "";
		saveSuccess = false;
		try {
			project = await projectsApi.update(projectId, {
				name: projectName,
				description: projectDescription,
			});
			saveSuccess = true;
			setTimeout(() => (saveSuccess = false), 3000);
		} catch (err: unknown) {
			saveError =
				err instanceof Error ? err.message : "Failed to save changes.";
		} finally {
			saving = false;
		}
	}

	async function archiveProject() {
		if (!confirm("Archive this project? It will become read-only.")) return;
		try {
			await projectsApi.archive(projectId);
			goto("/projects");
		} catch {}
	}

	async function deleteProject() {
		if (
			!confirm(
				"Permanently delete this project and all its data? This cannot be undone.",
			)
		)
			return;
		try {
			await projectsApi.delete(projectId);
			goto("/projects");
		} catch {}
	}
</script>

<svelte:head>
	<title
		>{projectName
			? `${projectName} Settings — FPMB`
			: "Project Settings — FPMB"}</title
	>
	<meta
		name="description"
		content="Configure project name, description, archive state, and danger zone settings in FPMB."
	/>
</svelte:head>

<div class="max-w-4xl mx-auto space-y-10">
	<div class="flex items-center space-x-4 mb-2">
		<a
			href="/projects"
			aria-label="Back to projects"
			class="text-neutral-400 hover:text-white transition-colors p-2 rounded-md hover:bg-neutral-800 border border-transparent"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
				><path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M10 19l-7-7m0 0l7-7m-7 7h18"
				></path></svg
			>
		</a>
		<div>
			<h1 class="text-3xl font-bold text-white tracking-tight">
				Project Settings
			</h1>
			<p class="text-neutral-400 mt-1">
				Configure {projectName || "..."} preferences and access.
			</p>
		</div>
	</div>

	<div class="border-b border-neutral-700 mb-8">
		<nav class="-mb-px flex space-x-8" aria-label="Tabs">
			<a
				href="/projects/{projectId}/settings"
				class="border-blue-500 text-blue-400 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm"
			>
				General Settings
			</a>
			<a
				href="/projects/{projectId}/webhooks"
				class="border-transparent text-neutral-400 hover:text-white hover:border-neutral-500 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors"
			>
				Webhooks & Integrations
			</a>
		</nav>
	</div>

	{#if loading}
		<div class="text-neutral-400 py-12 text-center">Loading...</div>
	{:else}
		<section
			class="bg-neutral-800 rounded-lg shadow-sm border border-neutral-700 overflow-hidden"
		>
			<div class="p-6 border-b border-neutral-700">
				<h2 class="text-xl font-semibold text-white mb-1">General Info</h2>
				<p class="text-sm text-neutral-400">
					Update project name and description.
				</p>
			</div>

			<form onsubmit={saveSettings} class="p-6 space-y-6">
				{#if saveError}
					<p class="text-sm text-red-400">{saveError}</p>
				{/if}
				{#if saveSuccess}
					<p class="text-sm text-green-400">Changes saved.</p>
				{/if}

				<div>
					<label
						for="projectName"
						class="block text-sm font-medium text-neutral-300"
						>Project Name</label
					>
					<input
						type="text"
						id="projectName"
						bind:value={projectName}
						required
						class="mt-1 block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
					/>
				</div>

				<div>
					<label
						for="projectDescription"
						class="block text-sm font-medium text-neutral-300"
						>Description</label
					>
					<textarea
						id="projectDescription"
						bind:value={projectDescription}
						rows="3"
						class="mt-1 block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white resize-none"
					></textarea>
				</div>

				<div class="flex justify-end pt-4 border-t border-neutral-700 mt-6">
					<button
						type="submit"
						disabled={saving}
						class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md shadow-sm border border-transparent transition-colors text-sm disabled:opacity-50"
					>
						{saving ? "Saving..." : "Save Changes"}
					</button>
				</div>
			</form>
		</section>

		<section
			class="bg-neutral-800 rounded-lg shadow-sm border border-red-900 overflow-hidden"
		>
			<div class="p-6 border-b border-red-900 bg-red-900/10">
				<h2 class="text-xl font-semibold text-red-500 mb-1">Danger Zone</h2>
				<p class="text-sm text-neutral-400">
					Irreversible destructive actions.
				</p>
			</div>

			<div class="p-6 space-y-6">
				<div class="flex items-center justify-between">
					<div>
						<h3 class="text-sm font-medium text-white">Archive Project</h3>
						<p class="text-xs text-neutral-400 mt-1">
							Mark this project as read-only and hide it from the active lists.
						</p>
					</div>
					<button
						onclick={archiveProject}
						class="bg-neutral-700 hover:bg-neutral-600 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-neutral-600 transition-colors text-sm"
					>
						Archive
					</button>
				</div>

				<div
					class="border-t border-neutral-700 pt-6 flex items-center justify-between"
				>
					<div>
						<h3 class="text-sm font-medium text-red-400">Delete Project</h3>
						<p class="text-xs text-neutral-400 mt-1">
							Permanently remove this project, its boards, files, and all
							associated data.
						</p>
					</div>
					<button
						onclick={deleteProject}
						class="bg-red-600 hover:bg-red-700 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-transparent transition-colors text-sm"
					>
						Delete Permanently
					</button>
				</div>
			</div>
		</section>
	{/if}
</div>
