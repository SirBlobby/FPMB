<script lang="ts">
	import { page } from "$app/stores";
	import Icon from "@iconify/svelte";
	import Markdown from "$lib/components/Markdown/Markdown.svelte";
	import { onMount } from "svelte";
	import { teams as teamsApi, docs as docsApi } from "$lib/api";
	import type { Doc, FileItem } from "$lib/types/api";

	let teamId = $derived($page.params.id ?? "");

	let docs = $state<Doc[]>([]);
	let activeDoc = $state<Doc | null>(null);
	let teamFiles = $state<FileItem[]>([]);
	let isEditing = $state(false);
	let editTitle = $state("");
	let editContent = $state("");
	let saving = $state(false);

	onMount(async () => {
		[docs, teamFiles] = await Promise.all([
			teamsApi.listDocs(teamId),
			teamsApi.listFiles(teamId).catch(() => [] as FileItem[]),
		]);
		if (docs.length > 0) {
			activeDoc = await docsApi.get(docs[0].id);
		}
	});

	async function selectDoc(doc: Doc) {
		activeDoc = await docsApi.get(doc.id);
		isEditing = false;
	}

	function startEdit() {
		if (!activeDoc) return;
		editTitle = activeDoc.title;
		editContent = activeDoc.content;
		isEditing = true;
	}

	async function saveDoc() {
		if (!activeDoc) return;
		saving = true;
		try {
			const updated = await docsApi.update(activeDoc.id, {
				title: editTitle,
				content: editContent,
			});
			docs = docs.map((d) => (d.id === updated.id ? updated : d));
			activeDoc = updated;
			isEditing = false;
		} finally {
			saving = false;
		}
	}

	async function createNewDoc() {
		const created = await teamsApi.createDoc(
			teamId,
			"Untitled Document",
			"# Untitled Document\n\nStart typing here...",
		);
		docs = [created, ...docs];
		activeDoc = created;
		editTitle = created.title;
		editContent = created.content;
		isEditing = true;
	}
</script>

<svelte:head>
	<title>Team Docs — FPMB</title>
	<meta
		name="description"
		content="Browse and edit your team's Markdown knowledge base documents in FPMB."
	/>
</svelte:head>

<div class="flex flex-col -m-6 p-6 overflow-hidden h-full">
	<div
		class="flex flex-1 overflow-hidden rounded-lg border border-neutral-700 bg-neutral-800 shadow-sm h-full"
	>
		<!-- Sidebar List -->
		<div
			class="w-80 border-r border-neutral-700 flex flex-col shrink-0 bg-neutral-850"
		>
			<div
				class="p-4 border-b border-neutral-700 flex items-center justify-between"
			>
				<h2 class="text-lg font-semibold text-white">Team Docs</h2>
				<button
					onclick={createNewDoc}
					class="p-1.5 text-neutral-400 hover:text-white hover:bg-neutral-700 rounded-md transition-colors"
					title="New Document"
				>
					<Icon icon="lucide:file-plus" class="w-5 h-5" />
				</button>
			</div>

			<div class="flex-1 overflow-y-auto custom-scrollbar">
				<ul class="divide-y divide-neutral-700">
					{#each docs as doc (doc.id)}
						<li>
							<button
								onclick={() => selectDoc(doc)}
								class="w-full text-left px-4 py-3 hover:bg-neutral-750 transition-colors flex items-start gap-3 {activeDoc?.id ===
								doc.id
									? 'bg-neutral-750 border-l-2 border-blue-500'
									: 'border-l-2 border-transparent'}"
							>
								<Icon
									icon="lucide:file-text"
									class="w-5 h-5 text-neutral-400 mt-0.5 shrink-0"
								/>
								<div class="flex-1 min-w-0">
									<h3
										class="text-sm font-medium text-white truncate {activeDoc?.id ===
										doc.id
											? 'text-blue-400'
											: ''}"
									>
										{doc.title}
									</h3>
									<p class="text-xs text-neutral-500 mt-1">
										{new Date(doc.updated_at).toLocaleDateString("en-US", {
											month: "2-digit",
											day: "2-digit",
											year: "numeric",
										})}
									</p>
								</div>
							</button>
						</li>
					{/each}
				</ul>
			</div>
		</div>

		<!-- Main Content Area -->
		<div class="flex-1 flex flex-col min-w-0 bg-neutral-900 overflow-hidden">
			{#if activeDoc}
				<div
					class="flex items-center justify-between px-8 py-4 border-b border-neutral-700 bg-neutral-850 shrink-0"
				>
					<div class="flex-1 min-w-0 mr-4">
						{#if isEditing}
							<input
								type="text"
								bind:value={editTitle}
								class="block w-full px-3 py-1.5 border border-neutral-600 rounded-md bg-neutral-800 text-white text-lg font-semibold focus:ring-blue-500 focus:border-blue-500"
							/>
						{:else}
							<h1 class="text-xl font-bold text-white truncate">
								{activeDoc.title}
							</h1>
							<p class="text-xs text-neutral-500 mt-0.5">
								Last updated {new Date(activeDoc.updated_at).toLocaleDateString(
									"en-US",
									{ month: "2-digit", day: "2-digit", year: "numeric" },
								)}
							</p>
						{/if}
					</div>
					<div class="flex items-center gap-2 shrink-0">
						{#if isEditing}
							<button
								onclick={saveDoc}
								disabled={saving}
								class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-1.5 rounded-md text-sm font-medium transition-colors disabled:opacity-50"
							>
								{saving ? "Saving..." : "Save"}
							</button>
						{:else}
							<button
								onclick={startEdit}
								class="bg-neutral-700 hover:bg-neutral-600 text-white px-4 py-1.5 rounded-md text-sm font-medium transition-colors flex items-center gap-2"
							>
								<Icon icon="lucide:edit-2" class="w-4 h-4" />
								Edit
							</button>
						{/if}
					</div>
				</div>

				<div class="flex-1 overflow-y-auto custom-scrollbar p-8">
					<div class="max-w-5xl mx-auto">
						{#if isEditing}
							<textarea
								bind:value={editContent}
								class="w-full h-[600px] bg-neutral-800 border border-neutral-700 text-neutral-300 rounded-lg p-4 font-mono text-sm focus:ring-blue-500 focus:border-blue-500 resize-y"
								placeholder="Write your markdown here..."
							></textarea>
						{:else}
							<div
								class="bg-neutral-800 rounded-lg border border-neutral-700 p-8 shadow-sm min-h-full"
							>
								<Markdown content={activeDoc.content} files={teamFiles} />
							</div>
						{/if}
					</div>
				</div>
			{:else}
				<div
					class="flex-1 flex flex-col items-center justify-center text-neutral-500"
				>
					<Icon icon="lucide:file-text" class="w-16 h-16 mb-4 opacity-50" />
					<p class="text-lg">Select a document or create a new one</p>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 8px;
		height: 8px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background-color: #525252;
		border-radius: 20px;
	}
</style>
