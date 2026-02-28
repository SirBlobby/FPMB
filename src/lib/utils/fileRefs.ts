import type { FileItem } from '../types/api';

export function resolveFileRefs(content: string, files: FileItem[]): string {
	return content.replace(/\$file:([^\s\])"']+)/g, (_match, name: string) => {
		const file = files.find(f => f.name === name && f.type !== 'folder');
		if (file) {
			return `[${name}](#file-dl:${file.id}:${encodeURIComponent(name)})`;
		}
		return `\`unknown file: ${name}\``;
	});
}
