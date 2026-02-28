<script lang="ts">
	import Icon from "@iconify/svelte";
	import { onMount } from "svelte";
	import { page } from "$app/stores";
	import { projects as projectsApi, webhooks as webhooksApi } from "$lib/api";
	import type { Webhook } from "$lib/types/api";

	let projectId = $derived($page.params.id ?? "");

	let webhookList = $state<Webhook[]>([]);
	let loading = $state(true);
	let isModalOpen = $state(false);
	let saving = $state(false);
	let newWebhook = $state({ name: "", type: "discord", url: "", secret: "" });

	onMount(async () => {
		try {
			webhookList = await projectsApi.listWebhooks(projectId);
		} catch {
			webhookList = [];
		} finally {
			loading = false;
		}
	});

	async function addWebhook(e: SubmitEvent) {
		e.preventDefault();
		if (!newWebhook.name || !newWebhook.url) return;
		saving = true;
		try {
			const created = await projectsApi.createWebhook(projectId, {
				name: newWebhook.name,
				type: "discord",
				url: newWebhook.url,
			});
			webhookList = [...webhookList, created];
			isModalOpen = false;
			newWebhook = { name: "", type: "discord", url: "", secret: "" };
		} catch {
		} finally {
			saving = false;
		}
	}

	async function toggleStatus(id: string) {
		try {
			const updated = await webhooksApi.toggle(id);
			webhookList = webhookList.map((w) => (w.id === id ? updated : w));
		} catch {}
	}

	async function deleteWebhook(id: string) {
		try {
			await webhooksApi.delete(id);
			webhookList = webhookList.filter((w) => w.id !== id);
		} catch {}
	}

	function getWebhookType(url: string): string {
		if (url.includes("discord.com")) return "discord";
		if (url.includes("github.com")) return "github";
		if (url.includes("gitea") || url.includes("git.")) return "gitea";
		if (url.includes("slack.com")) return "slack";
		return "custom";
	}

	function getIcon(type: string) {
		switch (type) {
			case "discord":
				return "simple-icons:discord";
			case "github":
				return "simple-icons:github";
			case "gitea":
				return "simple-icons:gitea";
			case "slack":
				return "simple-icons:slack";
			default:
				return "lucide:webhook";
		}
	}

	function getColor(type: string) {
		switch (type) {
			case "discord":
				return "text-[#5865F2]";
			case "github":
				return "text-white";
			case "gitea":
				return "text-[#609926]";
			case "slack":
				return "text-[#E01E5A]";
			default:
				return "text-neutral-400";
		}
	}

	function formatLastTriggered(iso: string): string {
		if (!iso || iso === "0001-01-01T00:00:00Z") return "Never";
		const d = new Date(iso);
		const diff = Date.now() - d.getTime();
		const mins = Math.floor(diff / 60000);
		if (mins < 1) return "Just now";
		if (mins < 60) return `${mins}m ago`;
		const hours = Math.floor(mins / 60);
		if (hours < 24) return `${hours}h ago`;
		return `${Math.floor(hours / 24)}d ago`;
	}
</script>

<svelte:head>
	<title>Webhooks & Integrations — FPMB</title>
	<meta
		name="description"
		content="Configure webhooks and integrations with Discord, GitHub, Gitea, Slack, and custom endpoints for this project."
	/>
</svelte:head>

<div class="max-w-4xl mx-auto space-y-10">
	<div class="flex items-center space-x-4 mb-2">
		<a
			href="/projects/{projectId}/settings"
			class="text-neutral-400 hover:text-white transition-colors p-2 rounded-md hover:bg-neutral-800 border border-transparent"
		>
			<Icon icon="lucide:arrow-left" class="w-5 h-5" />
		</a>
		<div>
			<h1 class="text-3xl font-bold text-white tracking-tight">
				Webhooks & Integrations
			</h1>
			<p class="text-neutral-400 mt-1">
				Connect your project with external tools and services.
			</p>
		</div>
	</div>

	<div class="border-b border-neutral-700 mb-8">
		<nav class="-mb-px flex space-x-8" aria-label="Tabs">
			<a
				href="/projects/{projectId}/settings"
				class="border-transparent text-neutral-400 hover:text-white hover:border-neutral-500 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors"
			>
				General Settings
			</a>
			<a
				href="/projects/{projectId}/webhooks"
				class="border-blue-500 text-blue-400 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm"
			>
				Webhooks & Integrations
			</a>
		</nav>
	</div>

	<section
		class="bg-neutral-800 rounded-lg shadow-sm border border-neutral-700 overflow-hidden"
	>
		<div
			class="p-6 border-b border-neutral-700 flex justify-between items-center"
		>
			<div>
				<h2 class="text-xl font-semibold text-white mb-1">
					Configured Webhooks
				</h2>
				<p class="text-sm text-neutral-400">
					Trigger actions in other apps when events occur in FPMB.
				</p>
			</div>
			<button
				onclick={() => (isModalOpen = true)}
				class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-transparent transition-colors text-sm flex items-center"
			>
				<Icon icon="lucide:plus" class="w-4 h-4 mr-2" />
				Add Webhook
			</button>
		</div>

		{#if loading}
			<div class="p-12 text-center text-neutral-400">Loading webhooks...</div>
		{:else if webhookList.length === 0}
			<div class="p-12 text-center flex flex-col items-center justify-center">
				<Icon icon="lucide:webhook" class="w-12 h-12 text-neutral-600 mb-4" />
				<h3 class="text-lg font-medium text-white mb-1">No Webhooks Yet</h3>
				<p class="text-neutral-400 text-sm">
					Add a webhook to start receiving automated updates in Discord, GitHub,
					and more.
				</p>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left border-collapse">
					<thead>
						<tr class="bg-neutral-850 border-b border-neutral-700">
							<th
								class="px-6 py-4 text-xs font-semibold text-neutral-400 uppercase tracking-wider"
								>Integration</th
							>
							<th
								class="px-6 py-4 text-xs font-semibold text-neutral-400 uppercase tracking-wider hidden md:table-cell"
								>Target URL</th
							>
							<th
								class="px-6 py-4 text-xs font-semibold text-neutral-400 uppercase tracking-wider"
								>Status</th
							>
							<th
								class="px-6 py-4 text-xs font-semibold text-neutral-400 uppercase tracking-wider text-right"
								>Actions</th
							>
						</tr>
					</thead>
					<tbody class="divide-y divide-neutral-700">
						{#each webhookList as webhook (webhook.id)}
							{@const wtype = getWebhookType(webhook.url)}
							<tr class="hover:bg-neutral-750 transition-colors group">
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="flex items-center">
										<div
											class="shrink-0 flex items-center justify-center w-8 h-8 rounded bg-neutral-900 border border-neutral-700"
										>
											<Icon
												icon={getIcon(wtype)}
												class="w-5 h-5 {getColor(wtype)}"
											/>
										</div>
										<div class="ml-4">
											<div
												class="text-sm font-medium text-white group-hover:text-blue-400 transition-colors"
											>
												{webhook.name}
											</div>
											<div class="text-xs text-neutral-500 mt-1 capitalize">
												{wtype} • Last: {formatLastTriggered(
													webhook.last_triggered || "",
												)}
											</div>
										</div>
									</div>
								</td>
								<td
									class="px-6 py-4 whitespace-nowrap text-sm text-neutral-400 hidden md:table-cell max-w-[200px] truncate"
								>
									{webhook.url}
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<button
										onclick={() => toggleStatus(webhook.id)}
										class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border {webhook.active
											? 'bg-green-500/10 text-green-400 border-green-500/20'
											: 'bg-neutral-700 text-neutral-400 border-neutral-600'}"
									>
										{webhook.active ? "Active" : "Inactive"}
									</button>
								</td>
								<td
									class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium"
								>
									<div class="flex items-center justify-end space-x-2">
										<button
											onclick={() => deleteWebhook(webhook.id)}
											class="text-neutral-400 hover:text-red-400 p-2 rounded hover:bg-neutral-700 transition-colors"
											title="Delete"
										>
											<Icon icon="lucide:trash-2" class="w-4 h-4" />
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</section>
</div>

{#if isModalOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-neutral-900/80 backdrop-blur-sm"
	>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="fixed inset-0" onclick={() => (isModalOpen = false)}></div>

		<div
			class="relative bg-neutral-800 rounded-lg shadow-xl border border-neutral-700 w-full max-w-lg"
		>
			<div
				class="flex items-center justify-between p-4 border-b border-neutral-700"
			>
				<h2 class="text-lg font-semibold text-white">Add Webhook</h2>
				<button
					onclick={() => (isModalOpen = false)}
					class="text-neutral-400 hover:text-white p-1 rounded-md hover:bg-neutral-700"
				>
					<Icon icon="lucide:x" class="w-5 h-5" />
				</button>
			</div>

			<form onsubmit={addWebhook} class="p-6 space-y-4">
				<div>
					<label
						for="webhook-type"
						class="block text-sm font-medium text-neutral-300 mb-1"
						>Service Type</label
					>
					<select
						id="webhook-type"
						bind:value={newWebhook.type}
						class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
					>
						<option value="discord">Discord</option>
						<option value="github">GitHub</option>
						<option value="gitea">Gitea</option>
						<option value="slack">Slack</option>
						<option value="custom">Custom Webhook</option>
					</select>
				</div>

				<div>
					<label
						for="webhook-name"
						class="block text-sm font-medium text-neutral-300 mb-1">Name</label
					>
					<input
						id="webhook-name"
						type="text"
						bind:value={newWebhook.name}
						required
						placeholder="e.g. My Team Discord"
						class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
					/>
				</div>

				<div>
					<label
						for="webhook-url"
						class="block text-sm font-medium text-neutral-300 mb-1"
						>Payload URL</label
					>
					<input
						id="webhook-url"
						type="url"
						bind:value={newWebhook.url}
						required
						placeholder="https://..."
						class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
					/>
				</div>

				<div>
					<label
						for="webhook-secret"
						class="block text-sm font-medium text-neutral-300 mb-1"
						>Secret Token (Optional)</label
					>
					<input
						id="webhook-secret"
						type="password"
						bind:value={newWebhook.secret}
						placeholder="Used to sign webhook payloads"
						class="w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
					/>
				</div>

				<div class="flex justify-end pt-4 gap-3">
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
						{saving ? "Saving..." : "Save Webhook"}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
