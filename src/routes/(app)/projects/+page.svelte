<script lang="ts">
	import { onMount } from "svelte";
	import Icon from "@iconify/svelte";
	import { projects as projectsApi, teams as teamsApi } from "$lib/api";
	import type { Project, Team } from "$lib/types/api";

	let projects = $state<Project[]>([]);
	let teams = $state<Team[]>([]);
	let loading = $state(true);
	let error = $state("");

	let showModal = $state(false);
	let newName = $state("");
	let newDesc = $state("");
	let selectedTeamId = $state("");
	let creating = $state(false);
	let createError = $state("");

	onMount(async () => {
		try {
			const [p, t] = await Promise.all([projectsApi.list(), teamsApi.list()]);
			projects = p;
			teams = t;
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : "Failed to load projects";
		} finally {
			loading = false;
		}
	});

	function openModal() {
		newName = "";
		newDesc = "";
		selectedTeamId = "";
		createError = "";
		showModal = true;
	}

	function closeModal() {
		showModal = false;
	}

	async function submitCreate() {
		if (!newName.trim()) {
			createError = "Project name is required.";
			return;
		}
		creating = true;
		createError = "";
		try {
			let project: Project;
			if (selectedTeamId) {
				project = await teamsApi.createProject(
					selectedTeamId,
					newName.trim(),
					newDesc.trim(),
				);
			} else {
				project = await projectsApi.createPersonal(
					newName.trim(),
					newDesc.trim(),
				);
			}
			projects = [project, ...projects];
			showModal = false;
		} catch (e: unknown) {
			createError =
				e instanceof Error ? e.message : "Failed to create project.";
		} finally {
			creating = false;
		}
	}

	function statusLabel(p: Project): string {
		return p.is_archived ? "Archived" : "Active";
	}

	function statusClass(p: Project): string {
		return p.is_archived
			? "bg-neutral-700 text-neutral-400"
			: "bg-blue-900/50 text-blue-300";
	}
</script>

<svelte:head>
	<title>Projects — FPMB</title>
	<meta
		name="description"
		content="View and manage all your personal and team projects in FPMB."
	/>
</svelte:head>

<div class="max-w-7xl mx-auto">
	<div class="flex items-center justify-between mb-8">
		<h1 class="text-3xl font-bold text-white">Projects</h1>
		<button
			onclick={openModal}
			class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium rounded-md transition-colors"
		>
			<Icon icon="lucide:plus" class="w-4 h-4" />
			New Project
		</button>
	</div>

	{#if loading}
		<p class="text-neutral-500 text-sm">Loading...</p>
	{:else if error}
		<p class="text-red-400 text-sm">{error}</p>
	{:else if projects.length === 0}
		<div class="flex flex-col items-center justify-center py-24 text-center">
			<div
				class="w-16 h-16 rounded-full bg-neutral-800 flex items-center justify-center mb-4 border border-neutral-700"
			>
				<Icon icon="lucide:folder-open" class="w-8 h-8 text-neutral-500" />
			</div>
			<h2 class="text-white font-semibold text-lg mb-1">No projects yet</h2>
			<p class="text-neutral-500 text-sm mb-6">
				Create your first project to get started.
			</p>
			<button
				onclick={openModal}
				class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium rounded-md transition-colors"
			>
				<Icon icon="lucide:plus" class="w-4 h-4" />
				New Project
			</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each projects as project (project.id)}
				<div
					class="bg-neutral-800 rounded-lg border border-neutral-700 p-6 hover:border-blue-500 transition-colors shadow-sm group flex flex-col"
				>
					<div class="flex justify-between items-start mb-3">
						<a
							href="/board/{project.id}"
							class="text-xl font-semibold text-white group-hover:text-blue-400 transition-colors leading-tight"
							>{project.name}</a
						>
						<span
							class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ml-2 shrink-0 {statusClass(
								project,
							)}"
						>
							{statusLabel(project)}
						</span>
					</div>
					{#if project.team_name}
						<p class="text-xs text-neutral-500 mb-2 flex items-center gap-1">
							<Icon icon="lucide:users" class="w-3 h-3" />
							{project.team_name}
						</p>
					{/if}
					<p class="text-neutral-400 text-sm mb-6 line-clamp-2 flex-1">
						{project.description || "No description"}
					</p>
					<div class="flex items-center justify-between mt-auto">
						<div class="text-xs text-neutral-500">
							Updated {new Date(project.updated_at).toLocaleDateString(
								"en-US",
								{ month: "2-digit", day: "2-digit", year: "numeric" },
							)}
						</div>
						<div class="flex items-center gap-2">
							<a
								href="/projects/{project.id}/calendar"
								class="text-neutral-400 hover:text-white text-xs flex items-center gap-1"
								title="Calendar"
							>
								<Icon icon="lucide:calendar" class="w-4 h-4" />
							</a>
							{#if !project.team_name}
								<a
									href="/projects/{project.id}/files"
									class="text-neutral-400 hover:text-white text-xs flex items-center gap-1"
									title="Files"
								>
									<Icon icon="lucide:paperclip" class="w-4 h-4" />
								</a>
							{/if}
							<a
								href="/projects/{project.id}/settings"
								class="text-neutral-400 hover:text-white text-xs flex items-center gap-1"
								title="Settings"
							>
								<Icon icon="lucide:settings" class="w-4 h-4" />
							</a>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

{#if showModal}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-neutral-800 border border-neutral-700 rounded-xl shadow-2xl w-full max-w-md mx-4 p-6"
		>
			<div class="flex items-center justify-between mb-5">
				<h2 class="text-lg font-semibold text-white">New Project</h2>
				<button
					onclick={closeModal}
					class="text-neutral-400 hover:text-white transition-colors p-1 rounded"
				>
					<Icon icon="lucide:x" class="w-5 h-5" />
				</button>
			</div>

			<div class="space-y-4">
				<div>
					<label
						for="proj-name"
						class="block text-sm font-medium text-neutral-300 mb-1.5"
						>Project name</label
					>
					<input
						id="proj-name"
						type="text"
						bind:value={newName}
						placeholder="e.g. Website Redesign"
						class="w-full bg-neutral-900 border border-neutral-600 rounded-md px-3 py-2 text-white placeholder-neutral-500 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
					/>
				</div>
				<div>
					<label
						for="proj-desc"
						class="block text-sm font-medium text-neutral-300 mb-1.5"
						>Description <span class="text-neutral-500 font-normal"
							>(optional)</span
						></label
					>
					<textarea
						id="proj-desc"
						bind:value={newDesc}
						placeholder="What is this project about?"
						rows="3"
						class="w-full bg-neutral-900 border border-neutral-600 rounded-md px-3 py-2 text-white placeholder-neutral-500 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
					></textarea>
				</div>
				<div>
					<label
						for="proj-team"
						class="block text-sm font-medium text-neutral-300 mb-1.5"
						>Team <span class="text-neutral-500 font-normal">(optional)</span
						></label
					>
					<select
						id="proj-team"
						bind:value={selectedTeamId}
						class="w-full bg-neutral-900 border border-neutral-600 rounded-md px-3 py-2 text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
					>
						<option value="">Personal (no team)</option>
						{#each teams as team (team.id)}
							<option value={team.id}>{team.name}</option>
						{/each}
					</select>
					<p class="text-xs text-neutral-500 mt-1">
						Personal projects are only visible to you.
					</p>
				</div>

				{#if createError}
					<p class="text-red-400 text-sm">{createError}</p>
				{/if}
			</div>

			<div class="flex justify-end gap-3 mt-6">
				<button
					onclick={closeModal}
					class="px-4 py-2 text-sm font-medium text-neutral-300 bg-neutral-700 hover:bg-neutral-600 border border-neutral-600 rounded-md transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={submitCreate}
					disabled={creating}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed rounded-md transition-colors flex items-center gap-2"
				>
					{#if creating}
						<svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24"
							><circle
								class="opacity-25"
								cx="12"
								cy="12"
								r="10"
								stroke="currentColor"
								stroke-width="4"
							></circle><path
								class="opacity-75"
								fill="currentColor"
								d="M4 12a8 8 0 018-8v8z"
							></path></svg
						>
					{/if}
					Create Project
				</button>
			</div>
		</div>
	</div>
{/if}
