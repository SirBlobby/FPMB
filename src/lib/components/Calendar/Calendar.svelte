<script lang="ts">
	import Icon from '@iconify/svelte';
	import Modal from '$lib/components/Modal/Modal.svelte';

	let { events = [], onEventClick = null as ((event: any) => void) | null } = $props();

	let currentDate = $state(new Date());
	let viewMode = $state<'month' | 'week'>('month');
	let selectedEvent = $state<any | null>(null);
	let isEventModalOpen = $state(false);

	let daysInMonth = $derived(new Date(currentDate.getFullYear(), currentDate.getMonth() + 1, 0).getDate());
	let firstDayOfMonth = $derived(new Date(currentDate.getFullYear(), currentDate.getMonth(), 1).getDay());
	let weekStart = $derived(new Date(currentDate.getFullYear(), currentDate.getMonth(), currentDate.getDate() - currentDate.getDay()));
	let weekDays = $derived(Array.from({length: 7}, (_, i) => new Date(weekStart.getFullYear(), weekStart.getMonth(), weekStart.getDate() + i)));

	const monthNames = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];
	const dayNames = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
	const dayNamesFull = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];

	function prev() {
		if (viewMode === 'month') {
			currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth() - 1, 1);
		} else {
			currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth(), currentDate.getDate() - 7);
		}
	}

	function next() {
		if (viewMode === 'month') {
			currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth() + 1, 1);
		} else {
			currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth(), currentDate.getDate() + 7);
		}
	}

	function isSameDay(d1: Date, d2: Date) {
		return d1.getDate() === d2.getDate() && d1.getMonth() === d2.getMonth() && d1.getFullYear() === d2.getFullYear();
	}

	function getEventsForDate(d: Date) {
		return events.filter((e: any) => {
			if (!e.date) return false;
			const parts = e.date.split('-');
			const eventDate = new Date(parseInt(parts[0]), parseInt(parts[1]) - 1, parseInt(parts[2]));
			return isSameDay(eventDate, d);
		});
	}

	function openEvent(event: any) {
		selectedEvent = event;
		isEventModalOpen = true;
		if (onEventClick) onEventClick(event);
	}

	const colorDot: Record<string, string> = {
		red: 'bg-red-500',
		blue: 'bg-blue-500',
		green: 'bg-green-500',
		purple: 'bg-purple-500',
		yellow: 'bg-yellow-500',
		neutral: 'bg-neutral-500',
		orange: 'bg-orange-500'
	};

	const colorBadge: Record<string, string> = {
		red: 'bg-red-500/15 text-red-300 border-red-500/25',
		blue: 'bg-blue-500/15 text-blue-300 border-blue-500/25',
		green: 'bg-green-500/15 text-green-300 border-green-500/25',
		purple: 'bg-purple-500/15 text-purple-300 border-purple-500/25',
		yellow: 'bg-yellow-500/15 text-yellow-300 border-yellow-500/25',
		neutral: 'bg-neutral-600/30 text-neutral-300 border-neutral-600/50',
		orange: 'bg-orange-500/15 text-orange-300 border-orange-500/25'
	};

	const colorFull: Record<string, string> = {
		red: 'bg-red-500/20 border-red-500/40 text-red-200',
		blue: 'bg-blue-500/20 border-blue-500/40 text-blue-200',
		green: 'bg-green-500/20 border-green-500/40 text-green-200',
		purple: 'bg-purple-500/20 border-purple-500/40 text-purple-200',
		yellow: 'bg-yellow-500/20 border-yellow-500/40 text-yellow-200',
		neutral: 'bg-neutral-600/30 border-neutral-600/50 text-neutral-300',
		orange: 'bg-orange-500/20 border-orange-500/40 text-orange-200'
	};

	function formatDate(dateStr: string) {
		if (!dateStr) return '';
		const parts = dateStr.split('-');
		const d = new Date(parseInt(parts[0]), parseInt(parts[1]) - 1, parseInt(parts[2]));
		return d.toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' });
	}
</script>

<div class="bg-neutral-800 rounded-xl shadow-lg border border-neutral-700 flex flex-col">
	<div class="px-5 py-4 border-b border-neutral-700 flex flex-wrap gap-3 items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="flex items-center gap-1.5">
				<button
					onclick={prev}
					class="w-8 h-8 flex items-center justify-center text-neutral-400 hover:text-white hover:bg-neutral-700 rounded-lg transition-colors"
				>
					<Icon icon="lucide:chevron-left" class="w-4 h-4" />
				</button>
				<button
					onclick={next}
					class="w-8 h-8 flex items-center justify-center text-neutral-400 hover:text-white hover:bg-neutral-700 rounded-lg transition-colors"
				>
					<Icon icon="lucide:chevron-right" class="w-4 h-4" />
				</button>
			</div>
			<h2 class="text-lg font-semibold text-white min-w-[200px]">
				{#if viewMode === 'month'}
					{monthNames[currentDate.getMonth()]} {currentDate.getFullYear()}
				{:else}
					{monthNames[weekStart.getMonth()].slice(0,3)} {weekStart.getDate()} –
					{monthNames[weekDays[6].getMonth()].slice(0,3)} {weekDays[6].getDate()}, {weekDays[6].getFullYear()}
				{/if}
			</h2>
			<button
				onclick={() => currentDate = new Date()}
				class="hidden sm:block text-xs font-medium text-neutral-400 hover:text-white bg-neutral-700/60 hover:bg-neutral-700 px-3 py-1.5 rounded-lg transition-colors border border-neutral-600"
			>
				Today
			</button>
		</div>

		<div class="flex bg-neutral-900 rounded-lg p-0.5 border border-neutral-700">
			<button
				onclick={() => viewMode = 'month'}
				class="px-4 py-1.5 text-xs font-semibold rounded-md transition-all {viewMode === 'month' ? 'bg-neutral-700 text-white shadow-sm' : 'text-neutral-400 hover:text-white'}"
			>
				Month
			</button>
			<button
				onclick={() => viewMode = 'week'}
				class="px-4 py-1.5 text-xs font-semibold rounded-md transition-all {viewMode === 'week' ? 'bg-neutral-700 text-white shadow-sm' : 'text-neutral-400 hover:text-white'}"
			>
				Week
			</button>
		</div>
	</div>

	{#if viewMode === 'month'}
		<div class="grid grid-cols-7 border-b border-neutral-700 shrink-0">
			{#each dayNames as day}
				<div class="py-2.5 text-center text-[11px] font-bold text-neutral-500 uppercase tracking-widest border-r border-neutral-700/50 last:border-r-0">
					{day}
				</div>
			{/each}
		</div>

		<div class="grid grid-cols-7 bg-neutral-700/30 gap-px" style="grid-auto-rows: minmax(110px, auto);">
			{#each Array(firstDayOfMonth) as _}
				<div class="bg-neutral-800/40 p-2"></div>
			{/each}

			{#each Array(daysInMonth) as _, i}
				{@const dayNum = i + 1}
				{@const cellDate = new Date(currentDate.getFullYear(), currentDate.getMonth(), dayNum)}
				{@const dayEvents = getEventsForDate(cellDate)}
				{@const isToday = isSameDay(cellDate, new Date())}
				{@const isWeekend = cellDate.getDay() === 0 || cellDate.getDay() === 6}

				<div class="bg-neutral-800 p-2 flex flex-col group hover:bg-neutral-750 transition-colors {isWeekend ? 'bg-neutral-800/70' : ''}">
					<div class="flex justify-end mb-1.5">
						<span class="text-xs font-semibold w-6 h-6 flex items-center justify-center rounded-full transition-colors
							{isToday ? 'bg-blue-600 text-white' : 'text-neutral-400 group-hover:text-neutral-200'}">
							{dayNum}
						</span>
					</div>

					<div class="flex-1 space-y-0.5 overflow-hidden">
						{#each dayEvents.slice(0, 3) as event}
							<button
								onclick={() => openEvent(event)}
								class="w-full flex items-center gap-1.5 px-1.5 py-0.5 rounded text-[11px] font-medium text-left truncate border transition-colors hover:brightness-110 {colorBadge[event.color || 'neutral']}"
							>
								<span class="w-1.5 h-1.5 rounded-full shrink-0 {colorDot[event.color || 'neutral']}"></span>
								<span class="truncate">{event.title}</span>
							</button>
						{/each}
						{#if dayEvents.length > 3}
							<button
								onclick={() => openEvent(dayEvents[3])}
								class="w-full text-left px-1.5 py-0.5 text-[10px] font-medium text-neutral-500 hover:text-neutral-300 transition-colors"
							>
								+{dayEvents.length - 3} more
							</button>
						{/if}
					</div>
				</div>
			{/each}

			{#each Array((7 - ((firstDayOfMonth + daysInMonth) % 7)) % 7) as _}
				<div class="bg-neutral-800/40 p-2"></div>
			{/each}
		</div>

	{:else}
		<div class="grid grid-cols-7 border-b border-neutral-700 shrink-0">
			{#each weekDays as day}
				{@const isToday = isSameDay(day, new Date())}
				<div class="py-3 text-center border-r border-neutral-700/50 last:border-r-0 {isToday ? 'bg-blue-600/5' : ''}">
					<div class="text-[10px] font-bold uppercase tracking-widest {isToday ? 'text-blue-400' : 'text-neutral-500'}">{dayNames[day.getDay()]}</div>
					<div class="text-xl font-bold mt-0.5 {isToday ? 'text-blue-400' : 'text-neutral-300'}">{day.getDate()}</div>
					{#if isToday}
						<div class="w-1 h-1 rounded-full bg-blue-500 mx-auto mt-1"></div>
					{/if}
				</div>
			{/each}
		</div>

		<div class="grid grid-cols-7 gap-px bg-neutral-700/30" style="min-height: 400px;">
			{#each weekDays as day}
				{@const isToday = isSameDay(day, new Date())}
				{@const dayEvents = getEventsForDate(day)}
				<div class="bg-neutral-800 p-2 overflow-y-auto space-y-1.5 {isToday ? 'bg-blue-600/5' : ''}">
					{#each dayEvents as event}
						<button
							onclick={() => openEvent(event)}
							class="w-full text-left p-2.5 rounded-lg border shadow-sm transition-all hover:brightness-110 hover:shadow-md flex flex-col gap-0.5 {colorFull[event.color || 'neutral']}"
						>
							{#if event.time}
								<div class="text-[10px] font-bold uppercase tracking-wider opacity-70 flex items-center gap-1">
									<Icon icon="lucide:clock" class="w-2.5 h-2.5" />
									{event.time}
								</div>
							{/if}
							<div class="text-xs font-semibold leading-snug">{event.title}</div>
							{#if event.description}
								<div class="text-[10px] opacity-60 line-clamp-2 mt-0.5">{event.description}</div>
							{/if}
						</button>
					{/each}
					{#if dayEvents.length === 0}
						<div class="h-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity pt-4">
							<span class="text-xs text-neutral-600">No events</span>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>

{#if selectedEvent}
	<Modal bind:isOpen={isEventModalOpen} title="Event Details">
		<div class="space-y-4">
			<div class="flex items-start gap-3">
				<span class="w-3 h-3 rounded-full mt-1 shrink-0 {colorDot[selectedEvent.color || 'neutral']}"></span>
				<div class="flex-1 min-w-0">
					<h3 class="text-lg font-semibold text-white leading-snug">{selectedEvent.title}</h3>
					{#if selectedEvent.date}
						<p class="text-sm text-neutral-400 mt-1 flex items-center gap-1.5">
							<Icon icon="lucide:calendar" class="w-3.5 h-3.5" />
							{formatDate(selectedEvent.date)}
						</p>
					{/if}
					{#if selectedEvent.time}
						<p class="text-sm text-neutral-400 mt-0.5 flex items-center gap-1.5">
							<Icon icon="lucide:clock" class="w-3.5 h-3.5" />
							{selectedEvent.time}
						</p>
					{/if}
				</div>
			</div>

			{#if selectedEvent.description}
				<div class="bg-neutral-700/50 rounded-lg p-4 border border-neutral-600/50">
					<p class="text-sm text-neutral-300 leading-relaxed">{selectedEvent.description}</p>
				</div>
			{/if}

			<div class="flex justify-end pt-2 border-t border-neutral-700">
				<button
					onclick={() => isEventModalOpen = false}
					class="bg-neutral-700 hover:bg-neutral-600 text-white font-medium py-2 px-4 rounded-md transition-colors text-sm"
				>
					Close
				</button>
			</div>
		</div>
	</Modal>
{/if}
