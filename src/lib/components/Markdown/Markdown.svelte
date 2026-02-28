<script lang="ts">
	import { marked } from 'marked';
	import DOMPurify from 'dompurify';
	import { browser } from '$app/environment';
	import type { FileItem } from '$lib/types/api';
	import { resolveFileRefs } from '$lib/utils/fileRefs';
	import { files as filesApi } from '$lib/api';

	let { content = '', files = [] as FileItem[] } = $props();

	let htmlContent = $derived.by(() => {
		if (!content) return '';
		const resolved = files.length > 0 ? resolveFileRefs(content, files) : content;
		const parsed = marked.parse(resolved);
		if (browser) {
			return DOMPurify.sanitize(parsed as string);
		}
		return parsed as string;
	});

	function handleClick(e: MouseEvent) {
		const target = e.target as HTMLElement;
		const anchor = target.closest('a') as HTMLAnchorElement | null;
		if (!anchor) return;
		const href = anchor.getAttribute('href') ?? '';
		if (!href.startsWith('#file-dl:')) return;
		e.preventDefault();
		const rest = href.slice('#file-dl:'.length);
		const colon = rest.indexOf(':');
		if (colon === -1) return;
		const id = rest.slice(0, colon);
		const name = decodeURIComponent(rest.slice(colon + 1));
		filesApi.download(id, name).catch(() => {});
	}
</script>

<div
	class="prose prose-invert max-w-none prose-sm sm:prose-base prose-neutral"
	onclick={handleClick}
	role="presentation"
>
	{@html htmlContent}
</div>
