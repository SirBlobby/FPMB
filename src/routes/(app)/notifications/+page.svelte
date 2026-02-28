<script lang="ts">
	import Icon from "@iconify/svelte";
	import { onMount } from "svelte";
	import { notifications as notifApi } from "$lib/api";
	import type { Notification } from "$lib/types/api";

	let notifications = $state<Notification[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			notifications = await notifApi.list();
		} finally {
			loading = false;
		}
	});

	async function markAllRead() {
		await notifApi.markAllRead();
		notifications = notifications.map((n) => ({ ...n, read: true }));
	}

	async function markRead(id: string) {
		await notifApi.markRead(id);
		notifications = notifications.map((n) =>
			n.id === id ? { ...n, read: true } : n,
		);
	}

	async function deleteNotification(id: string) {
		await notifApi.delete(id);
		notifications = notifications.filter((n) => n.id !== id);
	}

	function labelForType(type: string) {
		if (type === "assign") return "Task Assigned";
		if (type === "team_invite") return "Team Invite";
		if (type === "due_soon") return "Due Soon";
		if (type === "mention") return "Mention";
		return "Notification";
	}

	function iconForType(type: string) {
		if (type === "assign") return "lucide:user-plus";
		if (type === "team_invite") return "lucide:users";
		if (type === "due_soon") return "lucide:clock";
		if (type === "mention") return "lucide:at-sign";
		return "lucide:bell";
	}

	function colorForType(type: string) {
		if (type === "assign") return "text-green-400";
		if (type === "team_invite") return "text-purple-400";
		if (type === "due_soon") return "text-orange-400";
		if (type === "mention") return "text-blue-400";
		return "text-yellow-400";
	}

	function relativeTime(dateStr: string) {
		const diff = Date.now() - new Date(dateStr).getTime();
		const minutes = Math.floor(diff / 60000);
		if (minutes < 1) return "just now";
		if (minutes < 60) return `${minutes}m ago`;
		const hours = Math.floor(minutes / 60);
		if (hours < 24) return `${hours}h ago`;
		const days = Math.floor(hours / 24);
		if (days < 7) return `${days}d ago`;
		return new Date(dateStr).toLocaleDateString("en-US", {
			month: "2-digit",
			day: "2-digit",
			year: "numeric",
		});
	}
</script>

<svelte:head>
	<title>Notifications — FPMB</title>
	<meta
		name="description"
		content="View and manage your FPMB notifications for project updates, task assignments, and team activity."
	/>
</svelte:head>

<div class="max-w-4xl mx-auto space-y-6">
	<div class="flex items-end justify-between border-b border-neutral-700 pb-4">
		<div>
			<h1 class="text-3xl font-bold text-white tracking-tight">
				Notifications
			</h1>
			<p class="text-neutral-400 mt-1">
				Stay updated on your projects and tasks.
			</p>
		</div>
		<button
			onclick={markAllRead}
			class="text-sm font-medium text-blue-500 hover:text-blue-400 transition-colors"
		>
			Mark all as read
		</button>
	</div>

	{#if loading}
		<p class="text-neutral-500 text-sm">Loading...</p>
	{:else if notifications.length === 0}
		<div
			class="text-center py-12 bg-neutral-800 rounded-lg border border-neutral-700"
		>
			<Icon
				icon="lucide:bell-off"
				class="w-12 h-12 text-neutral-600 mx-auto mb-3"
			/>
			<h3 class="text-lg font-medium text-white mb-1">All caught up!</h3>
			<p class="text-neutral-400 text-sm">You have no new notifications.</p>
		</div>
	{:else}
		<div
			class="bg-neutral-800 rounded-lg border border-neutral-700 shadow-sm overflow-hidden"
		>
			<ul class="divide-y divide-neutral-700">
				{#each notifications as notification (notification.id)}
					<li
						class="p-4 hover:bg-neutral-750 transition-colors {notification.read
							? 'opacity-60'
							: ''}"
					>
						<div class="flex items-start gap-4">
							<div class="shrink-0 mt-1">
								<div
									class="w-10 h-10 rounded-full bg-neutral-700 border border-neutral-600 flex items-center justify-center"
								>
									<Icon
										icon={iconForType(notification.type)}
										class="w-5 h-5 {colorForType(notification.type)}"
									/>
								</div>
							</div>
							<div class="flex-1 min-w-0">
								<div class="flex items-center justify-between mb-1">
									<p class="text-sm font-semibold text-white truncate pr-4">
										{labelForType(notification.type)}
									</p>
									<span class="text-xs text-neutral-500 whitespace-nowrap"
										>{relativeTime(notification.created_at)}</span
									>
								</div>
								<p class="text-sm text-neutral-300">
									{notification.message}
								</p>
							</div>
							<div class="shrink-0 mt-1 flex flex-col items-center gap-2">
								{#if !notification.read}
									<div class="w-2.5 h-2.5 bg-blue-500 rounded-full"></div>
									<button
										onclick={() => markRead(notification.id)}
										class="text-xs text-neutral-400 hover:text-white"
										title="Mark as read"
									>
										<Icon icon="lucide:check" class="w-4 h-4" />
									</button>
								{/if}
								<button
									onclick={() => deleteNotification(notification.id)}
									class="text-xs text-neutral-500 hover:text-red-400 transition-colors"
									title="Delete"
								>
									<Icon icon="lucide:trash-2" class="w-4 h-4" />
								</button>
							</div>
						</div>
					</li>
				{/each}
			</ul>
		</div>
	{/if}
</div>
