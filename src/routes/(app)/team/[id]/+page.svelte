<script lang="ts">
	import { page } from "$app/stores";
	import Icon from "@iconify/svelte";
	import { onMount } from "svelte";
	import { teams as teamsApi, board as boardApi } from "$lib/api";
	import { RoleFlag, hasPermission } from "$lib/types/roles";
	import type { Team, Project, TeamMember, Event, Doc } from "$lib/types/api";

	let teamId = $derived($page.params.id ?? "");

	let team = $state<Team | null>(null);
	let members = $state<TeamMember[]>([]);
	let recentProjects = $state<Project[]>([]);
	let upcomingEvents = $state<Event[]>([]);
	let cardEvents = $state<
		{
			id: string;
			date: string;
			title: string;
			projectName: string;
			projectId: string;
		}[]
	>([]);
	let recentDocs = $state<Doc[]>([]);
	let myRole = $state(0);
	let loading = $state(true);

	let showCreateProject = $state(false);
	let newProjectName = $state("");
	let newProjectDesc = $state("");
	let creating = $state(false);
	let createError = $state("");

	let teamRoleName = $derived.by(() => {
		if (hasPermission(myRole, RoleFlag.Owner)) return "Owner";
		if (hasPermission(myRole, RoleFlag.Admin)) return "Admin";
		if (hasPermission(myRole, RoleFlag.Editor)) return "Editor";
		return "Viewer";
	});

	let canCreate = $derived(hasPermission(myRole, RoleFlag.Editor));

	onMount(async () => {
		try {
			const [teamData, memberData, projectData, eventData, docData] =
				await Promise.all([
					teamsApi.get(teamId),
					teamsApi.listMembers(teamId),
					teamsApi.listProjects(teamId),
					teamsApi.listEvents(teamId),
					teamsApi.listDocs(teamId),
				]);
			team = teamData;
			members = memberData;
			recentProjects = projectData;
			upcomingEvents = eventData;
			recentDocs = docData;

			const boards = await Promise.all(
				projectData.map((p: Project) => boardApi.get(p.id).catch(() => null)),
			);
			cardEvents = boards.flatMap((b, i) => {
				if (!b) return [];
				return b.columns.flatMap((col) =>
					(col.cards ?? [])
						.filter((c) => c.due_date)
						.map((c) => ({
							id: c.id,
							date: c.due_date!.split("T")[0],
							title: c.title,
							projectName: projectData[i].name,
							projectId: projectData[i].id,
						})),
				);
			});

			const stored =
				typeof localStorage !== "undefined"
					? localStorage.getItem("user_id")
					: null;
			const me = memberData.find((m) => m.user_id === stored);
			if (me) myRole = me.role_flags;
		} finally {
			loading = false;
		}
	});

	function openCreateProject() {
		newProjectName = "";
		newProjectDesc = "";
		createError = "";
		showCreateProject = true;
	}

	function closeCreateProject() {
		showCreateProject = false;
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString("en-US", {
			month: "2-digit",
			day: "2-digit",
			year: "numeric",
		});
	}

	async function submitCreateProject() {
		if (!newProjectName.trim()) {
			createError = "Project name is required.";
			return;
		}
		creating = true;
		createError = "";
		try {
			const project = await teamsApi.createProject(
				teamId,
				newProjectName.trim(),
				newProjectDesc.trim(),
			);
			recentProjects = [project, ...recentProjects];
			showCreateProject = false;
		} catch (e: unknown) {
			createError =
				e instanceof Error ? e.message : "Failed to create project.";
		} finally {
			creating = false;
		}
	}
</script>

<svelte:head>
	<title>{team ? `${team.name} — FPMB` : "Team — FPMB"}</title>
	<meta
		name="description"
		content={team
			? `Overview of the ${team.name} team — projects, calendar, docs, and members.`
			: "Team overview in FPMB."}
	/>
</svelte:head>

<div class="max-w-7xl mx-auto space-y-8 h-full flex flex-col pb-8">
	{#if loading}
		<p class="text-neutral-500 text-sm">Loading...</p>
	{:else if team}
		<!-- Team Header -->
		<header
			class="bg-neutral-800 rounded-xl border border-neutral-700 shadow-sm overflow-hidden shrink-0 relative"
		>
			<div
				class="h-32 bg-linear-to-r from-blue-900/40 to-purple-900/40 relative"
			>
				{#if team.banner_url}
					<img
						src={team.banner_url}
						alt="Team banner"
						class="absolute inset-0 w-full h-full object-cover"
					/>
				{/if}
				<div
					class="absolute inset-0 bg-neutral-900 opacity-80"
					style="background-image: radial-gradient(#333 1px, transparent 1px); background-size: 20px 20px;"
				></div>
			</div>

			<div
				class="px-8 pb-8 pt-4 relative flex flex-col md:flex-row md:items-end justify-between gap-6 -mt-16"
			>
				<div class="flex items-end gap-6">
					<div
						class="w-24 h-24 rounded-xl bg-neutral-800 border-4 border-neutral-900 flex items-center justify-center shadow-lg overflow-hidden shrink-0 relative z-10"
					>
						{#if team.avatar_url}
							<img
								src={team.avatar_url}
								alt="Team avatar"
								class="w-full h-full object-cover"
							/>
						{:else}
							<div
								class="w-full h-full bg-blue-600 flex items-center justify-center text-4xl font-bold text-white shadow-inner"
							>
								{team.name.charAt(0)}
							</div>
						{/if}
					</div>
					<div class="pb-2">
						<div class="flex items-center gap-3">
							<h1 class="text-3xl font-bold text-white tracking-tight">
								{team.name}
							</h1>
							{#if myRole > 0}
								<span
									class="inline-flex items-center px-2.5 py-1 rounded text-xs font-medium bg-neutral-700 text-neutral-300 border border-neutral-600"
								>
									{teamRoleName}
								</span>
							{/if}
						</div>
						<p class="text-neutral-400 mt-1 flex items-center gap-2">
							<Icon icon="lucide:users" class="w-4 h-4" />
							{members.length} Members
						</p>
					</div>
				</div>

				<div class="flex flex-wrap items-center gap-3 pb-2 mt-4 md:mt-0">
					<a
						href="/team/{teamId}/calendar"
						class="bg-neutral-700 hover:bg-neutral-600 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-neutral-600 transition-colors flex items-center gap-2 text-sm"
					>
						<Icon icon="lucide:calendar-days" class="w-4 h-4" />
						Calendar
					</a>
					<a
						href="/team/{teamId}/docs"
						class="bg-neutral-700 hover:bg-neutral-600 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-neutral-600 transition-colors flex items-center gap-2 text-sm"
					>
						<Icon icon="lucide:book-open" class="w-4 h-4" />
						Docs
					</a>
					<a
						href="/team/{teamId}/files"
						class="bg-neutral-700 hover:bg-neutral-600 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-neutral-600 transition-colors flex items-center gap-2 text-sm"
					>
						<Icon icon="lucide:paperclip" class="w-4 h-4" />
						Files
					</a>
					<a
						href="/team/{teamId}/chat"
						class="bg-blue-600/20 hover:bg-blue-600/30 text-blue-300 font-medium py-2 px-4 rounded-md shadow-sm border border-blue-500/30 transition-colors flex items-center gap-2 text-sm"
					>
						<Icon icon="lucide:message-circle" class="w-4 h-4" />
						Chat
					</a>
					{#if hasPermission(myRole, RoleFlag.Admin) || hasPermission(myRole, RoleFlag.Owner)}
						<a
							href="/team/{teamId}/settings"
							class="bg-neutral-700 hover:bg-neutral-600 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-neutral-600 transition-colors flex items-center gap-2 text-sm"
						>
							<Icon icon="lucide:settings" class="w-4 h-4" />
							Settings
						</a>
					{/if}
				</div>
			</div>
		</header>

		<!-- Widgets Grid -->
		<div
			class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 flex-1 items-start"
		>
			<!-- Projects Widget -->
			<section
				class="bg-neutral-800 rounded-lg border border-neutral-700 shadow-sm flex flex-col h-full col-span-1 md:col-span-2 lg:col-span-2"
			>
				<div
					class="p-5 border-b border-neutral-700 flex items-center justify-between"
				>
					<h2 class="text-lg font-semibold text-white flex items-center gap-2">
						<Icon icon="lucide:folder-open" class="w-5 h-5 text-blue-400" />
						Active Projects
					</h2>
					<a
						href="/projects"
						class="text-sm font-medium text-blue-500 hover:text-blue-400 flex items-center gap-1 transition-colors"
					>
						View all <Icon icon="lucide:arrow-right" class="w-4 h-4" />
					</a>
				</div>
				<div class="p-5 grid grid-cols-1 sm:grid-cols-2 gap-4">
					{#each recentProjects.slice(0, 4) as project (project.id)}
						<a
							href="/board/{project.id}"
							class="block bg-neutral-750 p-4 rounded-md border border-neutral-600 hover:border-blue-500 hover:bg-neutral-700 transition-all group shadow-sm"
						>
							<div class="flex items-start justify-between mb-3">
								<div
									class="w-10 h-10 rounded bg-neutral-700 flex items-center justify-center text-blue-400 group-hover:scale-110 transition-transform"
								>
									<Icon icon="lucide:kanban-square" class="w-5 h-5" />
								</div>
							</div>
							<h3
								class="font-semibold text-white group-hover:text-blue-400 transition-colors"
							>
								{project.name}
							</h3>
							<p class="text-xs text-neutral-400 mt-1 flex items-center gap-1">
								<Icon icon="lucide:clock" class="w-3 h-3" />
								{formatDate(project.updated_at)}
							</p>
						</a>
					{/each}

					{#if canCreate}
						<button
							onclick={openCreateProject}
							class="flex flex-col items-center justify-center p-4 rounded-md border-2 border-dashed border-neutral-700 text-neutral-500 hover:text-white hover:border-neutral-500 transition-colors bg-neutral-800/50 group h-full min-h-[140px]"
						>
							<Icon
								icon="lucide:plus-circle"
								class="w-8 h-8 mb-2 group-hover:scale-110 transition-transform"
							/>
							<span class="font-medium text-sm">New Project</span>
						</button>
					{/if}
				</div>
			</section>

			<!-- Upcoming Events Widget -->
			<section
				class="bg-neutral-800 rounded-lg border border-neutral-700 shadow-sm flex flex-col h-full lg:col-span-1"
			>
				<div
					class="p-5 border-b border-neutral-700 flex items-center justify-between"
				>
					<h2 class="text-lg font-semibold text-white flex items-center gap-2">
						<Icon icon="lucide:calendar-days" class="w-5 h-5 text-green-400" />
						Calendar
					</h2>
					<a
						href="/team/{teamId}/calendar"
						class="p-1.5 rounded-md text-neutral-400 hover:text-white hover:bg-neutral-700 transition-colors"
						title="View Calendar"
					>
						<Icon icon="lucide:external-link" class="w-4 h-4" />
					</a>
				</div>
				<div class="p-0 flex-1">
					<ul class="divide-y divide-neutral-700">
						{#each upcomingEvents.slice(0, 3) as event (event.id)}
							<li
								class="p-4 hover:bg-neutral-750 transition-colors cursor-pointer group"
							>
								<div class="flex items-start gap-3">
									<div class="w-2 h-2 rounded-full mt-1.5 bg-blue-500"></div>
									<div class="flex-1">
										<h3
											class="text-sm font-semibold text-white group-hover:text-blue-400 transition-colors"
										>
											{event.title}
										</h3>
										<p class="text-xs text-neutral-400 mt-0.5">{event.date}</p>
									</div>
									<Icon
										icon="lucide:chevron-right"
										class="w-4 h-4 text-neutral-600 group-hover:text-neutral-400 mt-1"
									/>
								</div>
							</li>
						{/each}
						{#each cardEvents.slice(0, 3 - Math.min(upcomingEvents.length, 3)) as card (card.id)}
							<li
								class="p-4 hover:bg-neutral-750 transition-colors cursor-pointer group"
							>
								<a
									href="/board/{card.projectId}"
									class="flex items-start gap-3"
								>
									<div class="w-2 h-2 rounded-full mt-1.5 bg-yellow-500"></div>
									<div class="flex-1">
										<h3
											class="text-sm font-semibold text-white group-hover:text-blue-400 transition-colors"
										>
											{card.title}
										</h3>
										<p class="text-xs text-neutral-400 mt-0.5">
											{formatDate(card.date)} · {card.projectName}
										</p>
									</div>
									<Icon
										icon="lucide:chevron-right"
										class="w-4 h-4 text-neutral-600 group-hover:text-neutral-400 mt-1"
									/>
								</a>
							</li>
						{/each}
						{#if upcomingEvents.length === 0 && cardEvents.length === 0}
							<li class="p-8 text-center text-neutral-500 text-sm">
								No upcoming events.
							</li>
						{/if}
					</ul>
				</div>
				<div class="p-4 border-t border-neutral-700 mt-auto">
					<a
						href="/team/{teamId}/calendar"
						class="block w-full text-center py-2 bg-neutral-700 hover:bg-neutral-600 text-white rounded-md text-sm font-medium transition-colors border border-neutral-600"
					>
						Open Full Calendar
					</a>
				</div>
			</section>

			<!-- Recent Docs Widget -->
			<section
				class="bg-neutral-800 rounded-lg border border-neutral-700 shadow-sm flex flex-col h-full md:col-span-2 lg:col-span-3"
			>
				<div
					class="p-5 border-b border-neutral-700 flex items-center justify-between"
				>
					<h2 class="text-lg font-semibold text-white flex items-center gap-2">
						<Icon icon="lucide:book-open" class="w-5 h-5 text-purple-400" />
						Team Knowledge Base
					</h2>
					<a
						href="/team/{teamId}/docs"
						class="p-1.5 rounded-md text-neutral-400 hover:text-white hover:bg-neutral-700 transition-colors"
						title="View All Docs"
					>
						<Icon icon="lucide:external-link" class="w-4 h-4" />
					</a>
				</div>
				<div class="p-5">
					{#if recentDocs.length === 0}
						<p class="text-neutral-500 text-sm">No docs yet.</p>
					{:else}
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each recentDocs.slice(0, 6) as doc (doc.id)}
								<a
									href="/team/{teamId}/docs"
									class="flex items-center gap-3 p-3 rounded-md bg-neutral-750 border border-neutral-600 hover:border-blue-500 hover:bg-neutral-700 transition-all group shadow-sm"
								>
									<div
										class="w-10 h-10 rounded bg-neutral-800 flex items-center justify-center text-purple-400 group-hover:scale-110 transition-transform shadow-inner"
									>
										<Icon icon="lucide:file-text" class="w-5 h-5" />
									</div>
									<div class="flex-1 min-w-0">
										<h3
											class="text-sm font-semibold text-white truncate group-hover:text-blue-400 transition-colors"
										>
											{doc.title}
										</h3>
										<p
											class="text-xs text-neutral-400 flex items-center gap-1 mt-0.5"
										>
											<Icon icon="lucide:clock" class="w-3 h-3" />
											{formatDate(doc.updated_at)}
										</p>
									</div>
								</a>
							{/each}
						</div>
					{/if}
				</div>
			</section>
		</div>
	{/if}
</div>

{#if showCreateProject}
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
					onclick={closeCreateProject}
					class="text-neutral-400 hover:text-white transition-colors p-1 rounded"
				>
					<Icon icon="lucide:x" class="w-5 h-5" />
				</button>
			</div>

			<div class="space-y-4">
				<div>
					<label
						for="project-name"
						class="block text-sm font-medium text-neutral-300 mb-1.5"
						>Project name</label
					>
					<input
						id="project-name"
						type="text"
						bind:value={newProjectName}
						placeholder="e.g. Website Redesign"
						class="w-full bg-neutral-900 border border-neutral-600 rounded-md px-3 py-2 text-white placeholder-neutral-500 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
					/>
				</div>
				<div>
					<label
						for="project-desc"
						class="block text-sm font-medium text-neutral-300 mb-1.5"
						>Description <span class="text-neutral-500 font-normal"
							>(optional)</span
						></label
					>
					<textarea
						id="project-desc"
						bind:value={newProjectDesc}
						placeholder="What is this project about?"
						rows="3"
						class="w-full bg-neutral-900 border border-neutral-600 rounded-md px-3 py-2 text-white placeholder-neutral-500 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
					></textarea>
				</div>

				{#if createError}
					<p class="text-red-400 text-sm">{createError}</p>
				{/if}
			</div>

			<div class="flex justify-end gap-3 mt-6">
				<button
					onclick={closeCreateProject}
					class="px-4 py-2 text-sm font-medium text-neutral-300 bg-neutral-700 hover:bg-neutral-600 border border-neutral-600 rounded-md transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={submitCreateProject}
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
