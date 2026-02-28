<script lang="ts">
	import { users as usersApi, files as filesApi } from "$lib/api";
	import type { FileItem } from "$lib/types/api";
	import FileViewer from "$lib/components/FileViewer/FileViewer.svelte";

	let folderStack = $state<{ id: string; name: string }[]>([]);
	let currentParentId = $derived(
		folderStack.length > 0 ? folderStack[folderStack.length - 1].id : "",
	);

	let fileList = $state<FileItem[]>([]);
	let loading = $state(true);
	let folderName = $state("");
	let showFolderInput = $state(false);
	let savingFolder = $state(false);

	async function loadFiles(parentId: string) {
		loading = true;
		try {
			fileList = await usersApi.listFiles(parentId);
		} catch {
			fileList = [];
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadFiles(currentParentId);
	});

	function formatSize(bytes: number): string {
		if (!bytes) return "--";
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	function formatDate(iso: string): string {
		const d = new Date(iso);
		return d.toLocaleDateString("en-US", {
			month: "2-digit",
			day: "2-digit",
			year: "numeric",
		});
	}

	async function createFolder(e: SubmitEvent) {
		e.preventDefault();
		if (!folderName.trim()) return;
		savingFolder = true;
		try {
			const created = await usersApi.createFolder(
				folderName.trim(),
				currentParentId,
			);
			fileList = [created, ...fileList];
			folderName = "";
			showFolderInput = false;
		} catch {
		} finally {
			savingFolder = false;
		}
	}

	async function handleUpload(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		try {
			const created = await usersApi.uploadFile(file, currentParentId);
			fileList = [...fileList, created];
		} catch {}
		input.value = "";
	}

	async function deleteFile(id: string) {
		if (!confirm("Delete this item?")) return;
		try {
			await filesApi.delete(id);
			fileList = fileList.filter((f) => f.id !== id);
		} catch {}
	}

	function openFolder(folder: FileItem) {
		folderStack = [...folderStack, { id: folder.id, name: folder.name }];
	}

	function navigateToBreadcrumb(index: number) {
		if (index === -1) {
			folderStack = [];
		} else {
			folderStack = folderStack.slice(0, index + 1);
		}
	}

	function getIcon(type: string) {
		if (type === "folder") {
			return `<svg class="w-6 h-6 text-blue-400" fill="currentColor" viewBox="0 0 20 20"><path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"></path></svg>`;
		}
		return `<svg class="w-6 h-6 text-neutral-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path></svg>`;
	}

	let viewingFile = $state<FileItem | null>(null);

	let fileInput: HTMLInputElement;
</script>

<svelte:head>
	<title>My Files — FPMB</title>
	<meta
		name="description"
		content="Browse, upload, and organise your personal files and folders in FPMB."
	/>
</svelte:head>

<div class="h-full flex flex-col -m-6 p-6 overflow-hidden">
	<header
		class="flex flex-col md:flex-row md:items-center justify-between mb-6 pb-6 border-b border-neutral-700 shrink-0 gap-4"
	>
		<div class="flex items-center space-x-4">
			<a
				href="/"
				aria-label="Back to home"
				class="text-neutral-400 hover:text-white transition-colors p-2 rounded-md hover:bg-neutral-800 border border-transparent"
			>
				<svg
					class="w-5 h-5"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
					><path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M10 19l-7-7m0 0l7-7m-7 7h18"
					></path></svg
				>
			</a>
			<div>
				<h1 class="text-2xl font-bold text-white flex items-center gap-2">
					My Files
				</h1>
				<div
					class="text-sm text-neutral-400 flex items-center space-x-2 mt-1 flex-wrap gap-y-1"
				>
					<button
						onclick={() => navigateToBreadcrumb(-1)}
						class="hover:text-blue-400 transition-colors">Root</button
					>
					{#each folderStack as crumb, i}
						<span>/</span>
						<button
							onclick={() => navigateToBreadcrumb(i)}
							class="hover:text-blue-400 transition-colors">{crumb.name}</button
						>
					{/each}
					<span>/</span>
				</div>
			</div>
		</div>
		<div class="flex items-center space-x-3">
			<button
				onclick={() => (showFolderInput = !showFolderInput)}
				class="bg-neutral-800 hover:bg-neutral-700 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-neutral-600 transition-colors text-sm flex items-center"
			>
				<svg
					class="w-4 h-4 mr-2"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
					><path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M9 13h6m-3-3v6m5 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
					></path></svg
				>
				New Folder
			</button>
			<input
				bind:this={fileInput}
				type="file"
				class="hidden"
				onchange={handleUpload}
			/>
			<button
				onclick={() => fileInput.click()}
				class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-transparent transition-colors text-sm flex items-center"
			>
				<svg
					class="w-4 h-4 mr-2"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
					><path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
					></path></svg
				>
				Upload
			</button>
		</div>
	</header>

	{#if showFolderInput}
		<form onsubmit={createFolder} class="mb-4 flex gap-2">
			<input
				type="text"
				bind:value={folderName}
				placeholder="Folder name"
				required
				class="flex-1 px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
			/>
			<button
				type="submit"
				disabled={savingFolder}
				class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors disabled:opacity-50"
			>
				{savingFolder ? "Creating..." : "Create"}
			</button>
			<button
				type="button"
				onclick={() => (showFolderInput = false)}
				class="px-4 py-2 text-sm font-medium text-neutral-300 hover:text-white hover:bg-neutral-700 rounded-md transition-colors"
			>
				Cancel
			</button>
		</form>
	{/if}

	<div
		class="flex-1 overflow-auto bg-neutral-800 rounded-lg shadow-sm border border-neutral-700"
	>
		{#if loading}
			<div class="p-12 text-center text-neutral-400">Loading files...</div>
		{:else}
			<table class="w-full text-left border-collapse">
				<thead>
					<tr class="bg-neutral-850 border-b border-neutral-700">
						<th
							class="px-6 py-4 text-xs font-semibold text-neutral-400 uppercase tracking-wider"
							>Name</th
						>
						<th
							class="px-6 py-4 text-xs font-semibold text-neutral-400 uppercase tracking-wider w-32 hidden sm:table-cell"
							>Size</th
						>
						<th
							class="px-6 py-4 text-xs font-semibold text-neutral-400 uppercase tracking-wider w-40 hidden md:table-cell"
							>Last Modified</th
						>
						<th
							class="px-6 py-4 text-xs font-semibold text-neutral-400 uppercase tracking-wider text-right w-24"
							>Actions</th
						>
					</tr>
				</thead>
				<tbody class="divide-y divide-neutral-700">
					{#each fileList as file (file.id)}
						<tr
							class="hover:bg-neutral-750 transition-colors group cursor-pointer"
							ondblclick={() =>
								file.type === "folder"
									? openFolder(file)
									: (viewingFile = file)}
						>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="flex items-center">
									<div class="shrink-0 flex items-center justify-center">
										{@html getIcon(file.type)}
									</div>
									<div class="ml-4">
										{#if file.type === "folder"}
											<button
												onclick={() => openFolder(file)}
												class="text-sm font-medium text-white group-hover:text-blue-400 transition-colors text-left"
												>{file.name}</button
											>
										{:else}
											<div
												class="text-sm font-medium text-white group-hover:text-blue-400 transition-colors"
											>
												{file.name}
											</div>
										{/if}
										<div class="text-xs text-neutral-500 sm:hidden mt-1">
											{formatSize(file.size_bytes)} • {formatDate(
												file.updated_at,
											)}
										</div>
									</div>
								</div>
							</td>
							<td
								class="px-6 py-4 whitespace-nowrap text-sm text-neutral-400 hidden sm:table-cell"
							>
								{formatSize(file.size_bytes)}
							</td>
							<td
								class="px-6 py-4 whitespace-nowrap text-sm text-neutral-400 hidden md:table-cell"
							>
								{formatDate(file.updated_at)}
							</td>
							<td
								class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium"
							>
								<div
									class="flex items-center justify-end space-x-2 opacity-0 group-hover:opacity-100 transition-opacity"
								>
									{#if file.type === "file" && file.storage_url}
										<button
											onclick={() => filesApi.download(file.id, file.name)}
											class="text-neutral-400 hover:text-white p-1 rounded"
											title="Download"
										>
											<svg
												class="w-5 h-5"
												fill="none"
												stroke="currentColor"
												viewBox="0 0 24 24"
												><path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
												></path></svg
											>
										</button>
									{/if}
									<button
										onclick={() => deleteFile(file.id)}
										class="text-neutral-400 hover:text-red-400 p-1 rounded"
										title="Delete"
									>
										<svg
											class="w-5 h-5"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
											><path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
											></path></svg
										>
									</button>
								</div>
							</td>
						</tr>
					{/each}

					{#if fileList.length === 0}
						<tr>
							<td colspan="4" class="px-6 py-12 text-center text-neutral-400">
								<div class="flex flex-col items-center">
									<svg
										class="w-12 h-12 text-neutral-600 mb-4"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
										><path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
										></path></svg
									>
									<p>This folder is empty.</p>
								</div>
							</td>
						</tr>
					{/if}
				</tbody>
			</table>
		{/if}
	</div>
</div>

<FileViewer bind:file={viewingFile} downloadUrl={filesApi.downloadUrl} />
