export interface User {
	id: string;
	name: string;
	email: string;
	avatar_url: string;
	created_at: string;
	updated_at: string;
}

export interface Team {
	id: string;
	name: string;
	workspace_id: string;
	avatar_url?: string;
	banner_url?: string;
	member_count: number;
	role_flags: number;
	role_name: string;
	created_at: string;
	updated_at?: string;
}

export interface TeamMember {
	id: string;
	user_id: string;
	team_id?: string;
	name: string;
	email: string;
	role_flags: number;
	role_name: string;
	joined_at: string;
}

export interface Project {
	id: string;
	team_id: string;
	team_name?: string;
	name: string;
	description: string;
	visibility?: string;
	is_public: boolean;
	is_archived: boolean;
	created_by: string;
	created_at: string;
	updated_at: string;
}

export interface ProjectMember {
	id: string;
	user_id: string;
	project_id: string;
	name: string;
	email: string;
	role_flags: number;
	role_name: string;
	added_at: string;
}

export interface Subtask {
	id: number;
	text: string;
	done: boolean;
}

export interface Card {
	id: string;
	column_id: string;
	project_id: string;
	title: string;
	description: string;
	priority: string;
	color: string;
	due_date?: string;
	assignees: string[];
	estimated_minutes?: number;
	actual_minutes?: number;
	subtasks: Subtask[];
	position: number;
	created_by: string;
	created_at: string;
	updated_at: string;
}

export interface Column {
	id: string;
	project_id: string;
	title: string;
	position: number;
	cards?: Card[];
	created_at: string;
	updated_at: string;
}

export interface BoardData {
	project_id: string;
	columns: Column[];
}

export interface Event {
	id: string;
	title: string;
	date: string;
	time: string;
	color: string;
	description: string;
	scope: string;
	scope_id: string;
	created_by: string;
	created_at: string;
	updated_at: string;
}

export interface Notification {
	id: string;
	user_id: string;
	type: string;
	message: string;
	project_id: string;
	read: boolean;
	created_at: string;
}

export interface Doc {
	id: string;
	team_id: string;
	title: string;
	content: string;
	created_by: string;
	created_at: string;
	updated_at: string;
}

export interface FileItem {
	id: string;
	project_id?: string;
	team_id?: string;
	user_id?: string;
	parent_id?: string;
	name: string;
	type: string;
	size_bytes: number;
	storage_url: string;
	created_by: string;
	created_at: string;
	updated_at: string;
}

export interface Webhook {
	id: string;
	project_id: string;
	name: string;
	type: string;
	url: string;
	status: string;
	active?: boolean;
	last_triggered?: string;
	created_by: string;
	created_at: string;
	updated_at: string;
}

export interface Whiteboard {
	id: string;
	project_id: string;
	data: string;
	updated_at: string;
}

export interface AuthResponse {
	access_token: string;
	refresh_token: string;
	user: User;
}

export interface ApiKey {
	id: string;
	name: string;
	scopes: string[];
	prefix: string;
	last_used?: string;
	created_at: string;
}

/** Returned only once when a key is first created — contains the raw key. */
export interface ApiKeyCreated extends ApiKey {
	key: string;
}

export interface ChatMessage {
	id: string;
	team_id: string;
	user_id: string;
	user_name: string;
	content: string;
	reply_to?: string;
	edited_at?: string;
	deleted?: boolean;
	created_at: string;
}
