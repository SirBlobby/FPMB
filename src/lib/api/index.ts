import { apiFetch, apiFetchFormData } from './client';
import type {
	AuthResponse,
	User,
	Team,
	TeamMember,
	Project,
	ProjectMember,
	BoardData,
	Column,
	Card,
	Event,
	Notification,
	Doc,
	FileItem,
	Webhook,
	Whiteboard,
	ApiKey,
	ApiKeyCreated,
	ChatMessage
} from '$lib/types/api';

export const auth = {
	register: (name: string, email: string, password: string) =>
		apiFetch<AuthResponse>('/auth/register', {
			method: 'POST',
			body: JSON.stringify({ name, email, password })
		}),

	login: (email: string, password: string) =>
		apiFetch<AuthResponse>('/auth/login', {
			method: 'POST',
			body: JSON.stringify({ email, password })
		}),

	logout: () => apiFetch<void>('/auth/logout', { method: 'POST' })
};

export const users = {
	me: () => apiFetch<User>('/users/me'),

	updateMe: (data: Partial<Pick<User, 'name' | 'email' | 'avatar_url'>>) =>
		apiFetch<User>('/users/me', { method: 'PUT', body: JSON.stringify(data) }),

	changePassword: (current_password: string, new_password: string) =>
		apiFetch<void>('/users/me/password', {
			method: 'PUT',
			body: JSON.stringify({ current_password, new_password })
		}),

	search: (q: string) =>
		apiFetch<{ id: string; name: string; email: string }[]>(`/users/search?q=${encodeURIComponent(q)}`),

	listFiles: (parentId = '') => {
		const qs = parentId ? `?parent_id=${encodeURIComponent(parentId)}` : '';
		return apiFetch<FileItem[]>(`/users/me/files${qs}`);
	},

	createFolder: (name: string, parent_id = '') =>
		apiFetch<FileItem>('/users/me/files/folder', {
			method: 'POST',
			body: JSON.stringify({ name, parent_id })
		}),

	uploadFile: (file: File, parent_id = '') => {
		const fd = new FormData();
		fd.append('file', file);
		if (parent_id) fd.append('parent_id', parent_id);
		return apiFetchFormData<FileItem>('/users/me/files/upload', fd);
	},

	uploadAvatar: (file: File) => {
		const fd = new FormData();
		fd.append('file', file);
		return apiFetchFormData<User>('/users/me/avatar', fd);
	}
};

export const teams = {
	list: () => apiFetch<Team[]>('/teams'),

	create: (name: string) =>
		apiFetch<Team>('/teams', { method: 'POST', body: JSON.stringify({ name }) }),

	get: (teamId: string) => apiFetch<Team>(`/teams/${teamId}`),

	update: (teamId: string, data: Partial<Pick<Team, 'name'>>) =>
		apiFetch<Team>(`/teams/${teamId}`, { method: 'PUT', body: JSON.stringify(data) }),

	delete: (teamId: string) => apiFetch<void>(`/teams/${teamId}`, { method: 'DELETE' }),

	listMembers: (teamId: string) => apiFetch<TeamMember[]>(`/teams/${teamId}/members`),

	invite: (teamId: string, email: string, role_flags: number) =>
		apiFetch<{ message: string; member: TeamMember }>(`/teams/${teamId}/members/invite`, {
			method: 'POST',
			body: JSON.stringify({ email, role_flags })
		}),

	updateMemberRole: (teamId: string, userId: string, role_flags: number) =>
		apiFetch<{ user_id: string; role_flags: number; role_name: string }>(
			`/teams/${teamId}/members/${userId}`,
			{ method: 'PUT', body: JSON.stringify({ role_flags }) }
		),

	removeMember: (teamId: string, userId: string) =>
		apiFetch<void>(`/teams/${teamId}/members/${userId}`, { method: 'DELETE' }),

	listProjects: (teamId: string) => apiFetch<Project[]>(`/teams/${teamId}/projects`),

	createProject: (teamId: string, name: string, description: string) =>
		apiFetch<Project>(`/teams/${teamId}/projects`, {
			method: 'POST',
			body: JSON.stringify({ name, description })
		}),

	listEvents: (teamId: string) => apiFetch<Event[]>(`/teams/${teamId}/events`),

	createEvent: (
		teamId: string,
		data: Pick<Event, 'title' | 'description' | 'date' | 'time' | 'color'>
	) =>
		apiFetch<Event>(`/teams/${teamId}/events`, { method: 'POST', body: JSON.stringify(data) }),

	listDocs: (teamId: string) => apiFetch<Doc[]>(`/teams/${teamId}/docs`),

	createDoc: (teamId: string, title: string, content: string) =>
		apiFetch<Doc>(`/teams/${teamId}/docs`, {
			method: 'POST',
			body: JSON.stringify({ title, content })
		}),

	listFiles: (teamId: string, parentId = '') => {
		const qs = parentId ? `?parent_id=${encodeURIComponent(parentId)}` : '';
		return apiFetch<FileItem[]>(`/teams/${teamId}/files${qs}`);
	},

	createFolder: (teamId: string, name: string, parent_id = '') =>
		apiFetch<FileItem>(`/teams/${teamId}/files/folder`, {
			method: 'POST',
			body: JSON.stringify({ name, parent_id })
		}),

	uploadFile: (teamId: string, file: File, parent_id = '') => {
		const fd = new FormData();
		fd.append('file', file);
		if (parent_id) fd.append('parent_id', parent_id);
		return apiFetchFormData<FileItem>(`/teams/${teamId}/files/upload`, fd);
	},

	uploadAvatar: (teamId: string, file: File) => {
		const fd = new FormData();
		fd.append('file', file);
		return apiFetchFormData<Team>(`/teams/${teamId}/avatar`, fd);
	},

	uploadBanner: (teamId: string, file: File) => {
		const fd = new FormData();
		fd.append('file', file);
		return apiFetchFormData<Team>(`/teams/${teamId}/banner`, fd);
	},

	listChatMessages: (teamId: string, before?: string) => {
		const qs = before ? `?before=${encodeURIComponent(before)}` : '';
		return apiFetch<ChatMessage[]>(`/teams/${teamId}/chat${qs}`);
	}
};

export const projects = {
	list: () => apiFetch<Project[]>('/projects'),

	createPersonal: (name: string, description: string) =>
		apiFetch<Project>('/projects', {
			method: 'POST',
			body: JSON.stringify({ name, description })
		}),

	get: (projectId: string) => apiFetch<Project>(`/projects/${projectId}`),

	update: (projectId: string, data: Partial<Pick<Project, 'name' | 'description' | 'visibility'>>) =>
		apiFetch<Project>(`/projects/${projectId}`, { method: 'PUT', body: JSON.stringify(data) }),

	archive: (projectId: string) =>
		apiFetch<Project>(`/projects/${projectId}/archive`, { method: 'PUT' }),

	delete: (projectId: string) => apiFetch<void>(`/projects/${projectId}`, { method: 'DELETE' }),

	listMembers: (projectId: string) => apiFetch<ProjectMember[]>(`/projects/${projectId}/members`),

	addMember: (projectId: string, userId: string, role_flags: number) =>
		apiFetch<ProjectMember>(`/projects/${projectId}/members`, {
			method: 'POST',
			body: JSON.stringify({ user_id: userId, role_flags })
		}),

	updateMemberRole: (projectId: string, userId: string, role_flags: number) =>
		apiFetch<ProjectMember>(`/projects/${projectId}/members/${userId}`, {
			method: 'PUT',
			body: JSON.stringify({ role_flags })
		}),

	removeMember: (projectId: string, userId: string) =>
		apiFetch<void>(`/projects/${projectId}/members/${userId}`, { method: 'DELETE' }),

	listEvents: (projectId: string) => apiFetch<Event[]>(`/projects/${projectId}/events`),

	createEvent: (
		projectId: string,
		data: Pick<Event, 'title' | 'description' | 'date' | 'time' | 'color'>
	) =>
		apiFetch<Event>(`/projects/${projectId}/events`, {
			method: 'POST',
			body: JSON.stringify(data)
		}),

	listFiles: (projectId: string, parentId = '') => {
		const qs = parentId ? `?parent_id=${encodeURIComponent(parentId)}` : '';
		return apiFetch<FileItem[]>(`/projects/${projectId}/files${qs}`);
	},

	createFolder: (projectId: string, name: string, parent_id = '') =>
		apiFetch<FileItem>(`/projects/${projectId}/files/folder`, {
			method: 'POST',
			body: JSON.stringify({ name, parent_id })
		}),

	uploadFile: (projectId: string, file: File, parent_id = '') => {
		const fd = new FormData();
		fd.append('file', file);
		if (parent_id) fd.append('parent_id', parent_id);
		return apiFetchFormData<FileItem>(`/projects/${projectId}/files/upload`, fd);
	},

	listWebhooks: (projectId: string) => apiFetch<Webhook[]>(`/projects/${projectId}/webhooks`),

	createWebhook: (projectId: string, data: Pick<Webhook, 'name' | 'url' | 'type'>) =>
		apiFetch<Webhook>(`/projects/${projectId}/webhooks`, {
			method: 'POST',
			body: JSON.stringify(data)
		}),

	getWhiteboard: (projectId: string) =>
		apiFetch<Whiteboard>(`/projects/${projectId}/whiteboard`),

	saveWhiteboard: (projectId: string, data: string) =>
		apiFetch<Whiteboard>(`/projects/${projectId}/whiteboard`, {
			method: 'PUT',
			body: JSON.stringify({ data })
		})
};

export const board = {
	get: (projectId: string) => apiFetch<BoardData>(`/projects/${projectId}/board`),

	createColumn: (projectId: string, title: string) =>
		apiFetch<Column>(`/projects/${projectId}/columns`, {
			method: 'POST',
			body: JSON.stringify({ title })
		}),

	updateColumn: (projectId: string, columnId: string, title: string) =>
		apiFetch<Column>(`/projects/${projectId}/columns/${columnId}`, {
			method: 'PUT',
			body: JSON.stringify({ title })
		}),

	reorderColumn: (projectId: string, columnId: string, position: number) =>
		apiFetch<{ id: string; position: number }>(
			`/projects/${projectId}/columns/${columnId}/position`,
			{ method: 'PUT', body: JSON.stringify({ position }) }
		),

	deleteColumn: (projectId: string, columnId: string) =>
		apiFetch<void>(`/projects/${projectId}/columns/${columnId}`, { method: 'DELETE' }),

	createCard: (
		projectId: string,
		columnId: string,
		data: Pick<Card, 'title' | 'description' | 'priority' | 'color' | 'due_date' | 'assignees' | 'estimated_minutes' | 'actual_minutes'>
	) =>
		apiFetch<Card>(`/projects/${projectId}/columns/${columnId}/cards`, {
			method: 'POST',
			body: JSON.stringify(data)
		})
};

export const cards = {
	update: (
		cardId: string,
		data: Partial<Pick<Card, 'title' | 'description' | 'priority' | 'color' | 'due_date' | 'assignees' | 'subtasks' | 'estimated_minutes' | 'actual_minutes'>>
	) => apiFetch<Card>(`/cards/${cardId}`, { method: 'PUT', body: JSON.stringify(data) }),

	move: (cardId: string, column_id: string, position: number) =>
		apiFetch<Card>(`/cards/${cardId}/move`, {
			method: 'PUT',
			body: JSON.stringify({ column_id, position })
		}),

	delete: (cardId: string) => apiFetch<void>(`/cards/${cardId}`, { method: 'DELETE' })
};

export const events = {
	update: (
		eventId: string,
		data: Partial<Pick<Event, 'title' | 'description' | 'date' | 'time' | 'color'>>
	) => apiFetch<Event>(`/events/${eventId}`, { method: 'PUT', body: JSON.stringify(data) }),

	delete: (eventId: string) => apiFetch<void>(`/events/${eventId}`, { method: 'DELETE' })
};

export const notifications = {
	list: () => apiFetch<Notification[]>('/notifications'),

	markRead: (notifId: string) =>
		apiFetch<Notification>(`/notifications/${notifId}/read`, { method: 'PUT' }),

	markAllRead: () => apiFetch<void>('/notifications/read-all', { method: 'PUT' }),

	delete: (notifId: string) => apiFetch<void>(`/notifications/${notifId}`, { method: 'DELETE' })
};

export const docs = {
	get: (docId: string) => apiFetch<Doc>(`/docs/${docId}`),

	update: (docId: string, data: Partial<Pick<Doc, 'title' | 'content'>>) =>
		apiFetch<Doc>(`/docs/${docId}`, { method: 'PUT', body: JSON.stringify(data) }),

	delete: (docId: string) => apiFetch<void>(`/docs/${docId}`, { method: 'DELETE' })
};

export const files = {
	delete: (fileId: string) => apiFetch<void>(`/files/${fileId}`, { method: 'DELETE' }),
	downloadUrl: (fileId: string) => `/api/files/${fileId}/download`,
	download: async (fileId: string, fileName: string) => {
		const { getAccessToken } = await import('./client');
		const token = getAccessToken();
		const headers: Record<string, string> = {};
		if (token) headers['Authorization'] = `Bearer ${token}`;
		const res = await fetch(`/api/files/${fileId}/download`, { headers });
		if (!res.ok) throw new Error(`HTTP ${res.status}`);
		const blob = await res.blob();
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = fileName;
		a.click();
		URL.revokeObjectURL(url);
	}
};

export const webhooks = {
	update: (webhookId: string, data: Partial<Pick<Webhook, 'name' | 'url' | 'type'>>) =>
		apiFetch<Webhook>(`/webhooks/${webhookId}`, { method: 'PUT', body: JSON.stringify(data) }),

	toggle: (webhookId: string) =>
		apiFetch<Webhook>(`/webhooks/${webhookId}/toggle`, { method: 'PUT' }),

	delete: (webhookId: string) => apiFetch<void>(`/webhooks/${webhookId}`, { method: 'DELETE' })
};

export const apiKeys = {
	list: () => apiFetch<ApiKey[]>('/users/me/api-keys'),

	create: (name: string, scopes: string[]) =>
		apiFetch<ApiKeyCreated>('/users/me/api-keys', {
			method: 'POST',
			body: JSON.stringify({ name, scopes })
		}),

	revoke: (keyId: string) => apiFetch<void>(`/users/me/api-keys/${keyId}`, { method: 'DELETE' })
};
