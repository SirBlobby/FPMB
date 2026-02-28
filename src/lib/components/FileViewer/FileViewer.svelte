<script lang="ts">
	import Markdown from '$lib/components/Markdown/Markdown.svelte';
	import type { FileItem } from '$lib/types/api';
	import { getAccessToken } from '$lib/api/client';
	import { files as filesApi } from '$lib/api';

	let { file = $bindable<FileItem | null>(null), downloadUrl }: {
		file: FileItem | null;
		downloadUrl: (id: string) => string;
	} = $props();

	type ViewerType = 'pdf' | 'image' | 'video' | 'audio' | 'markdown' | 'text' | 'none';

	const IMAGE_EXTS = new Set(['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp', 'ico']);
	const VIDEO_EXTS = new Set(['mp4', 'webm', 'ogv', 'mov']);
	const AUDIO_EXTS = new Set(['mp3', 'wav', 'ogg', 'flac', 'm4a', 'aac']);
	const TEXT_EXTS = new Set([
		'txt', 'json', 'csv', 'xml', 'yaml', 'yml', 'toml', 'ini', 'env',
		'sh', 'bash', 'zsh', 'fish',
		'js', 'ts', 'jsx', 'tsx', 'mjs', 'cjs',
		'py', 'go', 'rs', 'java', 'c', 'cpp', 'h', 'hpp', 'cs', 'rb', 'php',
		'html', 'css', 'scss', 'less', 'svelte', 'vue',
		'sql', 'graphql', 'proto', 'dockerfile', 'makefile',
		'log', 'gitignore', 'gitattributes', 'editorconfig',
	]);

	function ext(filename: string): string {
		const parts = filename.toLowerCase().split('.');
		return parts.length > 1 ? parts[parts.length - 1] : '';
	}

	function viewerType(f: FileItem): ViewerType {
		const e = ext(f.name);
		if (e === 'pdf') return 'pdf';
		if (IMAGE_EXTS.has(e)) return 'image';
		if (VIDEO_EXTS.has(e)) return 'video';
		if (AUDIO_EXTS.has(e)) return 'audio';
		if (e === 'md' || e === 'mdx') return 'markdown';
		if (TEXT_EXTS.has(e)) return 'text';
		return 'none';
	}

	function authFetch(url: string): Promise<Response> {
		const token = getAccessToken();
		const headers: Record<string, string> = {};
		if (token) headers['Authorization'] = `Bearer ${token}`;
		return fetch(url, { headers });
	}

	let textContent = $state('');
	let textLoading = $state(false);
	let textError = $state('');

	let blobUrl = $state('');
	let blobLoading = $state(false);
	let blobError = $state('');

	let activeType = $derived(file ? viewerType(file) : 'none');
	let rawUrl = $derived(file ? downloadUrl(file.id) : '');

	$effect(() => {
		const needsBlob = activeType === 'pdf' || activeType === 'image' || activeType === 'video' || activeType === 'audio';
		if (blobUrl) {
			URL.revokeObjectURL(blobUrl);
			blobUrl = '';
		}
		blobError = '';
		if (!file || !needsBlob) return;
		blobLoading = true;
		authFetch(rawUrl)
			.then((r) => {
				if (!r.ok) throw new Error(`HTTP ${r.status}`);
				return r.blob();
			})
			.then((b) => { blobUrl = URL.createObjectURL(b); })
			.catch((e) => { blobError = e.message; })
			.finally(() => { blobLoading = false; });

		return () => {
			if (blobUrl) URL.revokeObjectURL(blobUrl);
		};
	});

	$effect(() => {
		textContent = '';
		textError = '';
		if (!file || (activeType !== 'text' && activeType !== 'markdown')) return;
		textLoading = true;
		authFetch(rawUrl)
			.then((r) => {
				if (!r.ok) throw new Error(`HTTP ${r.status}`);
				return r.text();
			})
			.then((t) => { textContent = t; })
			.catch((e) => { textError = e.message; })
			.finally(() => { textLoading = false; });
	});

	function close() {
		file = null;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}

	function formatSize(bytes: number): string {
		if (!bytes) return '';
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if file}
	<div class="fixed inset-0 z-50 flex flex-col bg-neutral-950/95 backdrop-blur-sm">
		<div class="flex items-center justify-between px-4 py-3 border-b border-neutral-700 shrink-0 bg-neutral-900">
			<div class="flex items-center gap-3 min-w-0">
				<div class="text-sm font-medium text-white truncate">{file.name}</div>
				{#if file.size_bytes}
					<div class="text-xs text-neutral-500 shrink-0">{formatSize(file.size_bytes)}</div>
				{/if}
			</div>
		<div class="flex items-center gap-2 shrink-0 ml-4">
			<button
				onclick={() => filesApi.download(file.id, file.name)}
				class="flex items-center gap-1.5 text-xs text-neutral-300 hover:text-white bg-neutral-700 hover:bg-neutral-600 px-3 py-1.5 rounded transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path></svg>
				Download
			</button>
				<button
					onclick={close}
					class="text-neutral-400 hover:text-white p-1.5 rounded hover:bg-neutral-700 transition-colors"
					title="Close (Esc)"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
				</button>
			</div>
		</div>

		<div class="flex-1 overflow-hidden flex items-center justify-center">
			{#if activeType === 'pdf'}
				{#if blobLoading}
					<div class="text-neutral-400">Loading…</div>
				{:else if blobError}
					<div class="text-red-400">Failed to load: {blobError}</div>
				{:else}
					<iframe
						src={blobUrl}
						title={file.name}
						class="w-full h-full border-0"
					></iframe>
				{/if}

			{:else if activeType === 'image'}
				<div class="w-full h-full overflow-auto flex items-center justify-center p-4">
					{#if blobLoading}
						<div class="text-neutral-400">Loading…</div>
					{:else if blobError}
						<div class="text-red-400">Failed to load: {blobError}</div>
					{:else}
						<img
							src={blobUrl}
							alt={file.name}
							class="max-w-full max-h-full object-contain rounded shadow-lg"
						/>
					{/if}
				</div>

			{:else if activeType === 'video'}
				<div class="w-full h-full flex items-center justify-center p-4">
					{#if blobLoading}
						<div class="text-neutral-400">Loading…</div>
					{:else if blobError}
						<div class="text-red-400">Failed to load: {blobError}</div>
					{:else}
						<!-- svelte-ignore a11y_media_has_caption -->
						<video
							src={blobUrl}
							controls
							class="max-w-full max-h-full rounded shadow-lg"
						></video>
					{/if}
				</div>

			{:else if activeType === 'audio'}
				<div class="flex flex-col items-center justify-center gap-6 p-8">
					{#if blobLoading}
						<div class="text-neutral-400">Loading…</div>
					{:else if blobError}
						<div class="text-red-400">Failed to load: {blobError}</div>
					{:else}
						<div class="w-24 h-24 rounded-full bg-neutral-800 border border-neutral-600 flex items-center justify-center">
							<svg class="w-10 h-10 text-neutral-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3"></path></svg>
						</div>
						<div class="text-neutral-300 text-sm font-medium">{file.name}</div>
						<audio src={blobUrl} controls class="w-80 max-w-full"></audio>
					{/if}
				</div>

			{:else if activeType === 'markdown'}
				<div class="w-full h-full overflow-auto">
					{#if textLoading}
						<div class="flex items-center justify-center h-full text-neutral-400">Loading…</div>
					{:else if textError}
						<div class="flex items-center justify-center h-full text-red-400">Failed to load: {textError}</div>
					{:else}
						<div class="max-w-3xl mx-auto px-8 py-8">
							<Markdown content={textContent} />
						</div>
					{/if}
				</div>

			{:else if activeType === 'text'}
				<div class="w-full h-full overflow-auto">
					{#if textLoading}
						<div class="flex items-center justify-center h-full text-neutral-400">Loading…</div>
					{:else if textError}
						<div class="flex items-center justify-center h-full text-red-400">Failed to load: {textError}</div>
					{:else}
						<pre class="p-6 text-sm text-neutral-200 font-mono leading-relaxed whitespace-pre-wrap break-words">{textContent}</pre>
					{/if}
				</div>

			{:else}
				<div class="flex flex-col items-center justify-center gap-4 text-neutral-400">
					<svg class="w-16 h-16 text-neutral-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path></svg>
					<p class="text-sm">No preview available for this file type.</p>
				<button
					onclick={() => file && filesApi.download(file.id, file.name)}
						class="flex items-center gap-2 text-sm text-white bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded transition-colors"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path></svg>
						Download {file.name}
					</button>
				</div>
			{/if}
		</div>
	</div>
{/if}
