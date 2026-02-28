<script lang="ts">
	import Icon from "@iconify/svelte";
	import Calendar from "$lib/components/Calendar/Calendar.svelte";
	import Modal from "$lib/components/Modal/Modal.svelte";
	import { onMount } from "svelte";
	import {
		teams as teamsApi,
		projects as projectsApi,
		board as boardApi,
	} from "$lib/api";
	import type { Team, Event, Project } from "$lib/types/api";

	let allEvents = $state<Event[]>([]);
	let cardEvents = $state<
		{
			id: string;
			date: string;
			title: string;
			time: string;
			color: string;
			description: string;
		}[]
	>([]);
	let myTeams = $state<Team[]>([]);
	let loading = $state(true);

	let isModalOpen = $state(false);
	let newEvent = $state({
		title: "",
		date: "",
		time: "",
		color: "blue",
		description: "",
		teamId: "",
	});
	let saving = $state(false);

	const priorityColor: Record<string, string> = {
		Low: "neutral",
		Medium: "blue",
		High: "yellow",
		Urgent: "red",
	};

	onMount(async () => {
		try {
			const [teams, allProjects] = await Promise.all([
				teamsApi.list(),
				projectsApi.list(),
			]);
			myTeams = teams;
			if (teams.length > 0) newEvent.teamId = teams[0].id;
			const perTeam = await Promise.all(
				teams.map((t) => teamsApi.listEvents(t.id)),
			);
			allEvents = perTeam.flat();
			const boards = await Promise.all(
				allProjects.map((p: Project) => boardApi.get(p.id).catch(() => null)),
			);
			cardEvents = boards.flatMap((b, i) => {
				if (!b) return [];
				return b.columns.flatMap((col) =>
					(col.cards ?? [])
						.filter((c) => c.due_date)
						.map((c) => ({
							id: c.id,
							date: c.due_date.split("T")[0],
							title: c.title,
							time: "",
							color: priorityColor[c.priority] ?? "blue",
							description: `${allProjects[i].name} — ${c.priority}`,
						})),
				);
			});
		} finally {
			loading = false;
		}
	});

	let calendarEvents = $derived([
		...allEvents.map((e) => ({
			id: e.id,
			date: e.date,
			title: e.title,
			time: e.time,
			color: e.color,
			description: e.description,
		})),
		...cardEvents,
	]);

	async function handleAddEvent(ev: SubmitEvent) {
		ev.preventDefault();
		if (!newEvent.title.trim() || !newEvent.date || !newEvent.teamId) return;
		saving = true;
		try {
			const created = await teamsApi.createEvent(newEvent.teamId, {
				title: newEvent.title,
				date: newEvent.date,
				time: newEvent.time,
				color: newEvent.color,
				description: newEvent.description,
			});
			allEvents = [...allEvents, created];
			isModalOpen = false;
			newEvent = {
				title: "",
				date: "",
				time: "",
				color: "blue",
				description: "",
				teamId: myTeams[0]?.id ?? "",
			};
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Calendar — FPMB</title>
	<meta
		name="description"
		content="View all team events, milestones, and card due dates across your projects in one calendar."
	/>
</svelte:head>

<div class="max-w-7xl mx-auto flex flex-col">
	<div class="flex items-center justify-between mb-6 shrink-0">
		<div>
			<h1 class="text-3xl font-bold text-white tracking-tight">
				Organization Calendar
			</h1>
			<p class="text-neutral-400 mt-1">
				Overview of all team events and milestones.
			</p>
		</div>
		<div class="flex items-center space-x-4">
			<button
				onclick={() => (isModalOpen = true)}
				class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-transparent transition-colors flex items-center space-x-2 text-sm"
			>
				<Icon icon="lucide:plus" class="w-4 h-4" />
				<span>Add Event</span>
			</button>
		</div>
	</div>

	{#if loading}
		<p class="text-neutral-500 text-sm">Loading events...</p>
	{:else}
		<Calendar events={calendarEvents} />
	{/if}
</div>

<Modal bind:isOpen={isModalOpen} title="Add Event">
	<form onsubmit={handleAddEvent} class="space-y-4">
		<div>
			<label class="block text-sm font-medium text-neutral-300 mb-1"
				>Title</label
			>
			<input
				type="text"
				bind:value={newEvent.title}
				required
				placeholder="Event title"
				class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-500 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
			/>
		</div>

		<div class="grid grid-cols-2 gap-4">
			<div>
				<label class="block text-sm font-medium text-neutral-300 mb-1"
					>Date</label
				>
				<input
					type="date"
					bind:value={newEvent.date}
					required
					class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium text-neutral-300 mb-1"
					>Time</label
				>
				<input
					type="text"
					bind:value={newEvent.time}
					placeholder="e.g. 10:00 AM"
					class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-500 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
			</div>
		</div>

		<div class="grid grid-cols-2 gap-4">
			<div>
				<label class="block text-sm font-medium text-neutral-300 mb-1"
					>Color</label
				>
				<select
					bind:value={newEvent.color}
					class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				>
					<option value="blue">Blue</option>
					<option value="green">Green</option>
					<option value="red">Red</option>
					<option value="yellow">Yellow</option>
					<option value="purple">Purple</option>
				</select>
			</div>
			<div>
				<label class="block text-sm font-medium text-neutral-300 mb-1"
					>Team</label
				>
				<select
					bind:value={newEvent.teamId}
					class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				>
					{#each myTeams as team (team.id)}
						<option value={team.id}>{team.name}</option>
					{/each}
				</select>
			</div>
		</div>

		<div>
			<label class="block text-sm font-medium text-neutral-300 mb-1"
				>Description</label
			>
			<textarea
				bind:value={newEvent.description}
				placeholder="Optional description"
				rows="2"
				class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-500 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white resize-none"
			></textarea>
		</div>

		<div class="flex justify-end pt-2 border-t border-neutral-700 gap-3">
			<button
				type="button"
				onclick={() => (isModalOpen = false)}
				class="bg-transparent hover:bg-neutral-700 text-neutral-300 font-medium py-2 px-4 rounded-md border border-neutral-600 transition-colors text-sm"
			>
				Cancel
			</button>
			<button
				type="submit"
				disabled={saving || !newEvent.title.trim() || !newEvent.date}
				class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md shadow-sm border border-transparent transition-colors text-sm disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{saving ? "Saving..." : "Add Event"}
			</button>
		</div>
	</form>
</Modal>
