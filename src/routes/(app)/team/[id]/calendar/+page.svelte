<script lang="ts">
	import Icon from "@iconify/svelte";
	import { onMount } from "svelte";
	import { page } from "$app/stores";
	import Calendar from "$lib/components/Calendar/Calendar.svelte";
	import { teams as teamsApi, board as boardApi } from "$lib/api";
	import type { Event, Project } from "$lib/types/api";

	let teamId = $derived($page.params.id ?? "");

	let teamEvents = $state<Event[]>([]);
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
	let loading = $state(true);
	let isModalOpen = $state(false);
	let saving = $state(false);
	let error = $state("");

	let newEvent = $state({
		title: "",
		date: "",
		time: "",
		color: "blue",
		description: "",
	});

	const priorityColor: Record<string, string> = {
		Low: "neutral",
		Medium: "blue",
		High: "yellow",
		Urgent: "red",
	};

	let calendarEvents = $derived([
		...teamEvents.map((e) => ({
			id: e.id,
			date: e.date,
			title: e.title,
			time: e.time,
			color: e.color,
			description: e.description,
		})),
		...cardEvents,
	]);

	onMount(async () => {
		try {
			const [events, projects] = await Promise.all([
				teamsApi.listEvents(teamId),
				teamsApi.listProjects(teamId),
			]);
			teamEvents = events;
			const boards = await Promise.all(
				(projects as Project[]).map((p) =>
					boardApi.get(p.id).catch(() => null),
				),
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
							description: `${(projects as Project[])[i].name} — ${c.priority}`,
						})),
				);
			});
		} finally {
			loading = false;
		}
	});

	async function addEvent(ev: SubmitEvent) {
		ev.preventDefault();
		if (!newEvent.title.trim() || !newEvent.date) return;
		saving = true;
		error = "";
		try {
			const created = await teamsApi.createEvent(teamId, {
				title: newEvent.title,
				date: newEvent.date,
				time: newEvent.time,
				color: newEvent.color,
				description: newEvent.description,
			});
			teamEvents = [...teamEvents, created];
			isModalOpen = false;
			newEvent = {
				title: "",
				date: "",
				time: "",
				color: "blue",
				description: "",
			};
		} catch {
			error = "Failed to create event.";
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Team Calendar — FPMB</title>
	<meta
		name="description"
		content="View and manage team events and card due dates on this team's calendar in FPMB."
	/>
</svelte:head>

<div class="flex flex-col -m-6 p-6">
	<header
		class="flex flex-col md:flex-row md:items-center justify-between mb-6 pb-6 border-b border-neutral-700 shrink-0 gap-4"
	>
		<div class="flex items-center space-x-4">
			<a
				href="/team/{teamId}"
				class="text-neutral-400 hover:text-white transition-colors p-2 rounded-md hover:bg-neutral-800 border border-transparent"
			>
				<Icon icon="lucide:arrow-left" class="w-5 h-5" />
			</a>
			<div>
				<h1 class="text-2xl font-bold text-white flex items-center gap-2">
					Team Calendar
				</h1>
				<p class="text-sm text-neutral-400 mt-1">
					Team events and task due dates
				</p>
			</div>
		</div>
		<div class="flex items-center space-x-3">
			<button
				onclick={() => (isModalOpen = true)}
				class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-transparent transition-colors text-sm flex items-center gap-2"
			>
				<Icon icon="lucide:plus" class="w-4 h-4" />
				Add Event
			</button>
		</div>
	</header>

	{#if loading}
		<div class="flex-1 flex items-center justify-center text-neutral-400">
			Loading events...
		</div>
	{:else}
		<Calendar events={calendarEvents} />
	{/if}
</div>

{#if isModalOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-neutral-900/80 backdrop-blur-sm"
	>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="fixed inset-0" onclick={() => (isModalOpen = false)}></div>
		<div
			class="relative bg-neutral-800 rounded-lg shadow-xl border border-neutral-700 w-full max-w-md"
		>
			<div
				class="flex items-center justify-between p-4 border-b border-neutral-700"
			>
				<h2 class="text-lg font-semibold text-white">Add Event</h2>
				<button
					onclick={() => (isModalOpen = false)}
					class="text-neutral-400 hover:text-white p-1 rounded-md hover:bg-neutral-700"
				>
					<Icon icon="lucide:x" class="w-5 h-5" />
				</button>
			</div>
			<form onsubmit={addEvent} class="p-6 space-y-4">
				{#if error}
					<p class="text-sm text-red-400">{error}</p>
				{/if}
				<div>
					<label class="block text-sm font-medium text-neutral-300 mb-1"
						>Title</label
					>
					<input
						type="text"
						bind:value={newEvent.title}
						required
						placeholder="Event title"
						class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-neutral-300 mb-1"
						>Description</label
					>
					<textarea
						bind:value={newEvent.description}
						rows="2"
						placeholder="Optional description"
						class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 focus:ring-blue-500 focus:border-blue-500 sm:text-sm resize-none"
					></textarea>
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
							class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
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
							class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
						/>
					</div>
				</div>
				<div>
					<label class="block text-sm font-medium text-neutral-300 mb-1"
						>Color</label
					>
					<select
						bind:value={newEvent.color}
						class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
					>
						<option value="blue">Blue</option>
						<option value="green">Green</option>
						<option value="red">Red</option>
						<option value="yellow">Yellow</option>
						<option value="purple">Purple</option>
					</select>
				</div>
				<div class="flex justify-end pt-2 gap-3">
					<button
						type="button"
						onclick={() => (isModalOpen = false)}
						class="px-4 py-2 text-sm font-medium text-neutral-300 hover:text-white hover:bg-neutral-700 rounded-md transition-colors"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={saving}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md shadow-sm transition-colors disabled:opacity-50"
					>
						{saving ? "Saving..." : "Save Event"}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
