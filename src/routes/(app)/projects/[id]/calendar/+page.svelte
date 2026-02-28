<script lang="ts">
	import Icon from "@iconify/svelte";
	import { onMount } from "svelte";
	import { page } from "$app/stores";
	import Calendar from "$lib/components/Calendar/Calendar.svelte";
	import { projects as projectsApi, events as eventsApi } from "$lib/api";
	import type { Event } from "$lib/types/api";

	let projectId = $derived($page.params.id ?? "");

	let rawEvents = $state<Event[]>([]);
	let loading = $state(true);
	let isModalOpen = $state(false);
	let saving = $state(false);
	let error = $state("");

	let newEvent = $state({ title: "", description: "", date: "", time: "" });

	let calendarEvents = $derived(
		rawEvents.map((e) => {
			const dt = new Date(e.date + "T" + (e.time || "00:00"));
			const hours = dt.getHours();
			const minutes = dt.getMinutes().toString().padStart(2, "0");
			const ampm = hours >= 12 ? "PM" : "AM";
			const h = hours % 12 || 12;
			return {
				id: e.id,
				date: e.date,
				title: e.title,
				time: `${h}:${minutes} ${ampm}`,
				color: "blue",
				description: e.description,
			};
		}),
	);

	onMount(async () => {
		try {
			rawEvents = await projectsApi.listEvents(projectId);
		} catch {
			rawEvents = [];
		} finally {
			loading = false;
		}
	});

	async function addEvent(e: SubmitEvent) {
		e.preventDefault();
		if (!newEvent.title || !newEvent.date) return;
		saving = true;
		error = "";
		try {
			const created = await projectsApi.createEvent(projectId, {
				title: newEvent.title,
				description: newEvent.description,
				date: newEvent.date,
				time: newEvent.time,
				color: "blue", // Hardcoded color since no color selector is present in this form yet
			});
			rawEvents = [...rawEvents, created];
			isModalOpen = false;
			newEvent = { title: "", description: "", date: "", time: "" };
		} catch {
			error = "Failed to create event.";
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Project Calendar — FPMB</title>
	<meta
		name="description"
		content="View and manage events and milestones for this project's calendar in FPMB."
	/>
</svelte:head>

<div class="flex flex-col -m-6 p-6">
	<header
		class="flex flex-col md:flex-row md:items-center justify-between mb-6 pb-6 border-b border-neutral-700 shrink-0 gap-4"
	>
		<div class="flex items-center space-x-4">
			<a
				href="/projects"
				aria-label="Back to projects"
				class="text-neutral-400 hover:text-white transition-colors p-2 rounded-md hover:bg-neutral-800 border border-transparent"
			>
				<Icon icon="lucide:arrow-left" class="w-5 h-5" />
			</a>
			<div>
				<h1 class="text-2xl font-bold text-white flex items-center gap-2">
					Project Calendar
				</h1>
				<div class="text-sm text-neutral-400 flex items-center space-x-2 mt-1">
					<a
						href="/projects/{projectId}/calendar"
						class="hover:text-blue-400 transition-colors">Overview</a
					>
					<span>/</span>
				</div>
			</div>
		</div>
		<div class="flex items-center space-x-3">
			<button
				onclick={() => (isModalOpen = true)}
				class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-transparent transition-colors text-sm flex items-center"
			>
				<Icon icon="lucide:plus" class="w-4 h-4 mr-2" />
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
					<label
						for="event-title"
						class="block text-sm font-medium text-neutral-300 mb-1">Title</label
					>
					<input
						id="event-title"
						type="text"
						bind:value={newEvent.title}
						required
						placeholder="Event title"
						class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
					/>
				</div>
				<div>
					<label
						for="event-desc"
						class="block text-sm font-medium text-neutral-300 mb-1"
						>Description</label
					>
					<textarea
						id="event-desc"
						bind:value={newEvent.description}
						rows="2"
						placeholder="Optional description"
						class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 focus:ring-blue-500 focus:border-blue-500 sm:text-sm resize-none"
					></textarea>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label
							for="event-date"
							class="block text-sm font-medium text-neutral-300 mb-1"
							>Date</label
						>
						<input
							id="event-date"
							type="date"
							bind:value={newEvent.date}
							required
							class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
						/>
					</div>
					<div>
						<label
							for="event-time"
							class="block text-sm font-medium text-neutral-300 mb-1"
							>Time</label
						>
						<input
							id="event-time"
							type="time"
							bind:value={newEvent.time}
							class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
						/>
					</div>
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
