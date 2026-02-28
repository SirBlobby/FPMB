<script lang="ts">
	import Icon from "@iconify/svelte";
	import { onMount } from "svelte";
	import { teams as teamsApi, projects as projectsApi } from "$lib/api";
	import type { Team, Project } from "$lib/types/api";

	let myTeams = $state<Team[]>([]);
	let recentProjects = $state<Project[]>([]);
	let loading = $state(true);

	let showNewTeam = $state(false);
	let newTeamName = $state("");
	let savingTeam = $state(false);

	onMount(async () => {
		try {
			[myTeams, recentProjects] = await Promise.all([
				teamsApi.list(),
				projectsApi.list(),
			]);
		} finally {
			loading = false;
		}
	});

	async function createTeam(e: SubmitEvent) {
		e.preventDefault();
		if (!newTeamName.trim()) return;
		savingTeam = true;
		try {
			const team = await teamsApi.create(newTeamName.trim());
			myTeams = [...myTeams, team];
			newTeamName = "";
			showNewTeam = false;
		} catch {
		} finally {
			savingTeam = false;
		}
	}
</script>

<svelte:head>
	<title>Dashboard — FPMB</title>
	<meta
		name="description"
		content="Your FPMB dashboard — an overview of your teams and active projects."
	/>
</svelte:head>

<div class="max-w-7xl mx-auto space-y-12 h-full flex flex-col">
	<div>
		<h1 class="text-3xl font-bold text-white tracking-tight">Dashboard</h1>
		<p class="text-neutral-400 mt-1">
			Welcome back. Here's an overview of your teams and active projects.
		</p>
	</div>

	<!-- My Teams Section -->
	<section>
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-xl font-semibold text-white flex items-center gap-2">
				<Icon icon="lucide:users" class="w-5 h-5 text-neutral-400" />
				My Teams
			</h2>
			<button
				onclick={() => (showNewTeam = true)}
				class="text-sm font-medium text-blue-500 hover:text-blue-400 transition-colors flex items-center gap-1"
			>
				<Icon icon="lucide:plus" class="w-4 h-4" />
				Create Team
			</button>
		</div>

		{#if loading}
			<p class="text-neutral-500 text-sm">Loading...</p>
		{:else if myTeams.length === 0}
			<p class="text-neutral-500 text-sm">
				You're not a member of any teams yet.
			</p>
		{:else}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				{#each myTeams as team (team.id)}
					<a
						href="/team/{team.id}"
						class="block bg-neutral-800 rounded-lg border border-neutral-700 p-6 hover:border-blue-500 hover:shadow-md transition-all shadow-sm group"
					>
						<div class="flex items-start justify-between mb-4">
							<div class="flex items-center gap-3">
								<div
									class="w-10 h-10 rounded-lg bg-blue-900/50 border border-blue-500/30 flex items-center justify-center text-blue-400 font-bold text-lg"
								>
									{team.name.charAt(0)}
								</div>
								<div>
									<h3
										class="text-lg font-semibold text-white group-hover:text-blue-400 transition-colors"
									>
										{team.name}
									</h3>
								</div>
							</div>
						</div>
						<div
							class="mt-4 pt-4 border-t border-neutral-700 flex items-center justify-end"
						>
							<span
								class="text-sm font-medium text-blue-500 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
							>
								Go to Team <Icon icon="lucide:arrow-right" class="w-4 h-4" />
							</span>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	</section>

	<!-- Recent Projects Section -->
	<section class="flex-1">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-xl font-semibold text-white flex items-center gap-2">
				<Icon icon="lucide:folder" class="w-5 h-5 text-neutral-400" />
				Recent Projects
			</h2>
			<a
				href="/projects"
				class="text-sm font-medium text-blue-500 hover:text-blue-400 transition-colors"
			>
				View All
			</a>
		</div>

		{#if loading}
			<p class="text-neutral-500 text-sm">Loading...</p>
		{:else if recentProjects.length === 0}
			<p class="text-neutral-500 text-sm">No projects yet.</p>
		{:else}
			<div
				class="bg-neutral-800 rounded-lg border border-neutral-700 overflow-hidden shadow-sm"
			>
				<ul class="divide-y divide-neutral-700">
					{#each recentProjects.slice(0, 5) as project (project.id)}
						<li class="hover:bg-neutral-750 transition-colors">
							<a
								href="/board/{project.id}"
								class="px-6 py-4 flex items-center justify-between group"
							>
								<div class="flex items-center gap-4">
									<div
										class="w-8 h-8 rounded bg-neutral-700 flex items-center justify-center text-neutral-400 group-hover:text-blue-400 transition-colors"
									>
										<Icon icon="lucide:folder" class="w-4 h-4" />
									</div>
									<div>
										<p
											class="text-sm font-semibold text-white group-hover:text-blue-400 transition-colors"
										>
											{project.name}
										</p>
										<p class="text-xs text-neutral-500 mt-0.5">
											Updated {new Date(project.updated_at).toLocaleDateString(
												"en-US",
												{ month: "2-digit", day: "2-digit", year: "numeric" },
											)}
										</p>
									</div>
								</div>
								<Icon
									icon="lucide:chevron-right"
									class="w-5 h-5 text-neutral-500 group-hover:text-white transition-colors"
								/>
							</a>
						</li>
					{/each}
				</ul>
			</div>
		{/if}
	</section>
</div>

{#if showNewTeam}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
		onclick={() => (showNewTeam = false)}
		onkeydown={(e) => e.key === "Escape" && (showNewTeam = false)}
		role="dialog"
		aria-label="Create Team dialog"
		tabindex="-1"
	>
		<div
			class="bg-neutral-800 border border-neutral-700 rounded-lg shadow-xl w-full max-w-md mx-4"
		>
			<div
				class="flex items-center justify-between p-4 border-b border-neutral-700"
			>
				<h2 class="text-lg font-semibold text-white">Create Team</h2>
				<button
					onclick={() => (showNewTeam = false)}
					class="text-neutral-400 hover:text-white p-1 rounded hover:bg-neutral-700 transition-colors"
					title="Close"
				>
					<Icon icon="lucide:x" class="w-5 h-5" />
				</button>
			</div>
			<form onsubmit={createTeam} class="p-4 space-y-4">
				<div>
					<label
						for="team-name"
						class="block text-sm font-medium text-neutral-300 mb-1"
						>Team Name</label
					>
					<input
						id="team-name"
						type="text"
						bind:value={newTeamName}
						placeholder="e.g. Engineering"
						required
						class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
					/>
				</div>
				<div class="flex justify-end gap-3">
					<button
						type="button"
						onclick={() => (showNewTeam = false)}
						class="px-4 py-2 text-sm font-medium text-neutral-300 hover:text-white hover:bg-neutral-700 rounded-md transition-colors"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={savingTeam}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors disabled:opacity-50"
					>
						{savingTeam ? "Creating..." : "Create"}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
