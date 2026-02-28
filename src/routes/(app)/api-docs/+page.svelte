<script lang="ts">
	let activeSection = $state("overview");
	let copiedId = $state("");

	const sections = [
		{ id: "overview", label: "Overview" },
		{ id: "auth", label: "Authentication" },
		{ id: "api-keys", label: "API Keys" },
		{ id: "users", label: "Users" },
		{ id: "teams", label: "Teams" },
		{ id: "projects", label: "Projects" },
		{ id: "boards", label: "Boards & Cards" },
		{ id: "events", label: "Events" },
		{ id: "files", label: "Files" },
		{ id: "webhooks", label: "Webhooks" },
		{ id: "notifications", label: "Notifications" },
	];

	async function copy(text: string, id: string) {
		await navigator.clipboard.writeText(text);
		copiedId = id;
		setTimeout(() => (copiedId = ""), 2000);
	}

	function scrollTo(id: string) {
		activeSection = id;
		document
			.getElementById(id)
			?.scrollIntoView({ behavior: "smooth", block: "start" });
	}
</script>

<svelte:head>
	<title>API Documentation — FPMB</title>
	<meta
		name="description"
		content="Complete REST API reference for FPMB — authentication, API keys, projects, boards, teams, files, and more."
	/>
</svelte:head>

<div class="max-w-7xl mx-auto">
	<!-- Header -->
	<div class="mb-8">
		<div class="flex items-center gap-3 mb-2">
			<div
				class="w-9 h-9 rounded-lg bg-blue-600/20 border border-blue-500/30 flex items-center justify-center"
			>
				<svg
					class="w-5 h-5 text-blue-400"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
					/>
				</svg>
			</div>
			<h1 class="text-3xl font-bold text-white tracking-tight">
				API Documentation
			</h1>
		</div>
		<p class="text-neutral-400 ml-12">
			REST API reference for programmatic access to FPMB. All endpoints are
			prefixed with <code
				class="text-blue-300 bg-neutral-800 px-1.5 py-0.5 rounded text-sm"
				>/api</code
			>.
		</p>
	</div>

	<div class="flex gap-8">
		<!-- Sidebar nav -->
		<aside class="hidden lg:block w-48 shrink-0">
			<nav class="sticky top-6 space-y-0.5">
				{#each sections as s}
					<button
						onclick={() => scrollTo(s.id)}
						class="w-full text-left px-3 py-1.5 text-sm rounded-md transition-colors {activeSection ===
						s.id
							? 'bg-blue-600/20 text-blue-300 font-medium'
							: 'text-neutral-400 hover:text-white hover:bg-neutral-800'}"
					>
						{s.label}
					</button>
				{/each}
			</nav>
		</aside>

		<!-- Content -->
		<div class="flex-1 min-w-0 space-y-12">
			<!-- Overview -->
			<section id="overview">
				<h2
					class="text-xl font-bold text-white mb-4 pb-2 border-b border-neutral-700"
				>
					Overview
				</h2>
				<div class="space-y-4 text-sm text-neutral-300 leading-relaxed">
					<p>
						The FPMB REST API lets you integrate with projects, boards, teams,
						files, and more. All responses are JSON.
					</p>
					<div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
						{#each [{ label: "Base URL", value: "http://localhost:8080/api" }, { label: "Content-Type", value: "application/json" }, { label: "Auth Scheme", value: "Bearer token / API key" }] as item}
							<div
								class="bg-neutral-800 border border-neutral-700 rounded-lg p-3"
							>
								<p class="text-xs text-neutral-500 mb-1">{item.label}</p>
								<p class="font-mono text-xs text-blue-300">{item.value}</p>
							</div>
						{/each}
					</div>
					<div
						class="bg-amber-900/20 border border-amber-700/40 rounded-lg p-4"
					>
						<p class="text-amber-300 text-sm">
							<strong>Note:</strong> Protected endpoints require an
							<code class="bg-neutral-800 px-1 rounded"
								>Authorization: Bearer &lt;token&gt;</code
							> header. Tokens can be JWT access tokens (from login) or personal
							API keys.
						</p>
					</div>
				</div>
			</section>

			<!-- Authentication -->
			<section id="auth">
				<h2
					class="text-xl font-bold text-white mb-4 pb-2 border-b border-neutral-700"
				>
					Authentication
				</h2>
				<div class="space-y-4">
					{#each [{ method: "POST", path: "/auth/register", label: "Register", desc: "Create a new account. Returns access + refresh tokens.", body: `{ "name": "Alice", "email": "alice@example.com", "password": "secret" }`, response: `{ "access_token": "eyJ...", "refresh_token": "eyJ...", "user": { ... } }` }, { method: "POST", path: "/auth/login", label: "Login", desc: "Authenticate with email and password.", body: `{ "email": "alice@example.com", "password": "secret" }`, response: `{ "access_token": "eyJ...", "refresh_token": "eyJ..." }` }, { method: "POST", path: "/auth/refresh", label: "Refresh Token", desc: "Exchange a refresh token for a new access token.", body: `{ "refresh_token": "eyJ..." }`, response: `{ "access_token": "eyJ...", "refresh_token": "eyJ..." }` }, { method: "POST", path: "/auth/logout", label: "Logout", desc: "Invalidate the current session (requires auth).", body: null, response: `{ "message": "Logged out successfully" }` }] as ep}
						<div
							class="bg-neutral-800 border border-neutral-700 rounded-lg overflow-hidden"
						>
							<div
								class="flex items-center gap-3 px-4 py-3 border-b border-neutral-700"
							>
								<span
									class="text-xs font-bold px-2 py-0.5 rounded bg-green-700/60 text-green-300"
									>{ep.method}</span
								>
								<code class="text-sm font-mono text-neutral-200">{ep.path}</code
								>
								<span class="text-sm text-neutral-500 ml-auto">{ep.label}</span>
							</div>
							<div class="px-4 py-3 text-sm text-neutral-400">{ep.desc}</div>
							{#if ep.body}
								<div class="border-t border-neutral-700/50">
									<p class="text-xs text-neutral-500 px-4 pt-2">Request body</p>
									<pre
										class="text-xs font-mono text-neutral-300 px-4 pb-3 overflow-x-auto">{ep.body}</pre>
								</div>
							{/if}
							<div class="border-t border-neutral-700/50">
								<p class="text-xs text-neutral-500 px-4 pt-2">Response</p>
								<pre
									class="text-xs font-mono text-neutral-300 px-4 pb-3 overflow-x-auto">{ep.response}</pre>
							</div>
						</div>
					{/each}
				</div>
			</section>

			<!-- API Keys -->
			<section id="api-keys">
				<h2
					class="text-xl font-bold text-white mb-1 pb-2 border-b border-neutral-700"
				>
					API Keys
				</h2>
				<p class="text-sm text-neutral-400 mb-4">
					Personal API keys can be used instead of JWT tokens. Pass them the
					same way: <code class="bg-neutral-800 px-1 rounded text-blue-300"
						>Authorization: Bearer fpmb_...</code
					>
				</p>

				<!-- Scopes table -->
				<div
					class="mb-6 bg-neutral-800 border border-neutral-700 rounded-lg overflow-hidden"
				>
					<div class="px-4 py-3 border-b border-neutral-700">
						<h3 class="text-sm font-semibold text-white">Available Scopes</h3>
					</div>
					<table class="w-full text-sm">
						<thead>
							<tr class="border-b border-neutral-700 text-left">
								<th
									class="px-4 py-2 text-xs font-semibold text-neutral-400 uppercase tracking-wider"
									>Scope</th
								>
								<th
									class="px-4 py-2 text-xs font-semibold text-neutral-400 uppercase tracking-wider"
									>Description</th
								>
							</tr>
						</thead>
						<tbody class="divide-y divide-neutral-700/60">
							{#each [["read:projects", "List and view projects"], ["write:projects", "Create, update, and delete projects"], ["read:boards", "Read board columns and cards"], ["write:boards", "Create, move, and delete cards and columns"], ["read:teams", "List teams and their members"], ["write:teams", "Create and manage teams"], ["read:files", "Browse and download files"], ["write:files", "Upload, create folders, and delete files"], ["read:notifications", "Read notifications"]] as [scope, desc]}
								<tr>
									<td class="px-4 py-2"
										><code
											class="text-xs font-mono text-blue-300 bg-blue-900/20 px-1.5 py-0.5 rounded"
											>{scope}</code
										></td
									>
									<td class="px-4 py-2 text-neutral-400 text-xs">{desc}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>

				<div class="space-y-4">
					{#each [{ method: "GET", path: "/users/me/api-keys", label: "List keys", auth: true, desc: "Returns all active (non-revoked) API keys for the authenticated user.", body: null, response: `[{ "id": "...", "name": "CI Pipeline", "scopes": ["read:projects"], "prefix": "fpmb_ab12", "created_at": "..." }]` }, { method: "POST", path: "/users/me/api-keys", label: "Create key", auth: true, desc: "Generate a new API key. The raw key is returned only once.", body: `{ "name": "My Key", "scopes": ["read:projects", "read:boards"] }`, response: `{ "id": "...", "name": "My Key", "key": "fpmb_...", "scopes": [...], "prefix": "fpmb_ab12", "created_at": "..." }` }, { method: "DELETE", path: "/users/me/api-keys/:keyId", label: "Revoke key", auth: true, desc: "Permanently revokes an API key. This cannot be undone.", body: null, response: `{ "message": "Key revoked" }` }] as ep}
						<div
							class="bg-neutral-800 border border-neutral-700 rounded-lg overflow-hidden"
						>
							<div
								class="flex items-center gap-3 px-4 py-3 border-b border-neutral-700"
							>
								<span
									class="text-xs font-bold px-2 py-0.5 rounded {ep.method ===
									'GET'
										? 'bg-blue-700/60 text-blue-300'
										: ep.method === 'POST'
											? 'bg-green-700/60 text-green-300'
											: 'bg-red-700/60 text-red-300'}">{ep.method}</span
								>
								<code class="text-sm font-mono text-neutral-200">{ep.path}</code
								>
								{#if ep.auth}<span
										class="text-xs bg-yellow-700/40 text-yellow-300 border border-yellow-700/40 px-1.5 py-0.5 rounded ml-1"
										>🔒 auth</span
									>{/if}
								<span class="text-sm text-neutral-500 ml-auto">{ep.label}</span>
							</div>
							<div class="px-4 py-3 text-sm text-neutral-400">{ep.desc}</div>
							{#if ep.body}
								<div class="border-t border-neutral-700/50">
									<p class="text-xs text-neutral-500 px-4 pt-2">Request body</p>
									<pre
										class="text-xs font-mono text-neutral-300 px-4 pb-3 overflow-x-auto">{ep.body}</pre>
								</div>
							{/if}
							<div class="border-t border-neutral-700/50">
								<p class="text-xs text-neutral-500 px-4 pt-2">Response</p>
								<pre
									class="text-xs font-mono text-neutral-300 px-4 pb-3 overflow-x-auto">{ep.response}</pre>
							</div>
						</div>
					{/each}
				</div>
			</section>

			<!-- Users -->
			<section id="users">
				<h2
					class="text-xl font-bold text-white mb-4 pb-2 border-b border-neutral-700"
				>
					Users
				</h2>
				<div class="space-y-4">
					{#each [{ method: "GET", path: "/users/me", label: "Get current user", desc: "Returns the authenticated user's profile." }, { method: "PUT", path: "/users/me", label: "Update profile", desc: "Update name, email, or avatar_url.", body: '{ "name": "Alice", "email": "alice@example.com" }' }, { method: "PUT", path: "/users/me/password", label: "Change password", desc: "Change account password.", body: '{ "current_password": "old", "new_password": "new" }' }, { method: "POST", path: "/users/me/avatar", label: "Upload avatar", desc: "Multipart/form-data. Field: file." }, { method: "GET", path: "/users/me/avatar", label: "Get avatar", desc: "Returns the avatar image file." }, { method: "GET", path: "/users/search?q=", label: "Search users", desc: "Search by name or email. Returns up to 10 results." }] as ep}
						<div
							class="bg-neutral-800 border border-neutral-700 rounded-lg overflow-hidden"
						>
							<div class="flex items-center gap-3 px-4 py-3">
								<span
									class="text-xs font-bold px-2 py-0.5 rounded {ep.method ===
									'GET'
										? 'bg-blue-700/60 text-blue-300'
										: ep.method === 'POST'
											? 'bg-green-700/60 text-green-300'
											: ep.method === 'DELETE'
												? 'bg-red-700/60 text-red-300'
												: 'bg-yellow-700/60 text-yellow-300'}">{ep.method}</span
								>
								<code class="text-sm font-mono text-neutral-200">{ep.path}</code
								>
								<span
									class="text-xs bg-yellow-700/40 text-yellow-300 border border-yellow-700/40 px-1.5 py-0.5 rounded ml-1"
									>🔒 auth</span
								>
								<span class="text-sm text-neutral-500 ml-auto">{ep.label}</span>
							</div>
							{#if ep.desc || ep.body}
								<div
									class="border-t border-neutral-700/50 px-4 py-3 text-sm text-neutral-400"
								>
									{ep.desc}
									{#if ep.body}
										<pre
											class="mt-2 text-xs font-mono text-neutral-300 bg-neutral-900 rounded p-2 overflow-x-auto">{ep.body}</pre>
									{/if}
								</div>
							{/if}
						</div>
					{/each}
				</div>
			</section>

			<!-- Teams -->
			<section id="teams">
				<h2
					class="text-xl font-bold text-white mb-4 pb-2 border-b border-neutral-700"
				>
					Teams
				</h2>
				<div class="space-y-2">
					{#each [{ method: "GET", path: "/teams", label: "List teams" }, { method: "POST", path: "/teams", label: "Create team", body: '{ "name": "Engineering" }' }, { method: "GET", path: "/teams/:teamId", label: "Get team" }, { method: "PUT", path: "/teams/:teamId", label: "Update team", body: '{ "name": "New Name" }' }, { method: "DELETE", path: "/teams/:teamId", label: "Delete team" }, { method: "GET", path: "/teams/:teamId/members", label: "List members" }, { method: "POST", path: "/teams/:teamId/members/invite", label: "Invite member", body: '{ "email": "bob@example.com", "role_flags": 1 }' }, { method: "PUT", path: "/teams/:teamId/members/:userId", label: "Update member role", body: '{ "role_flags": 2 }' }, { method: "DELETE", path: "/teams/:teamId/members/:userId", label: "Remove member" }, { method: "GET", path: "/teams/:teamId/projects", label: "List team projects" }, { method: "POST", path: "/teams/:teamId/projects", label: "Create project", body: '{ "name": "Sprint 1", "description": "..." }' }, { method: "GET", path: "/teams/:teamId/events", label: "List team events" }, { method: "POST", path: "/teams/:teamId/events", label: "Create event" }, { method: "GET", path: "/teams/:teamId/docs", label: "List docs" }, { method: "POST", path: "/teams/:teamId/docs", label: "Create doc", body: '{ "title": "RFC", "content": "..." }' }, { method: "GET", path: "/teams/:teamId/files", label: "List files" }, { method: "POST", path: "/teams/:teamId/files/folder", label: "Create folder" }, { method: "POST", path: "/teams/:teamId/files/upload", label: "Upload file (multipart)" }, { method: "POST", path: "/teams/:teamId/avatar", label: "Upload avatar" }, { method: "POST", path: "/teams/:teamId/banner", label: "Upload banner" }] as ep}
						<div class="bg-neutral-800 border border-neutral-700 rounded-md">
							<div class="flex items-center gap-3 px-4 py-2.5">
								<span
									class="text-xs font-bold px-2 py-0.5 rounded w-14 text-center {ep.method ===
									'GET'
										? 'bg-blue-700/60 text-blue-300'
										: ep.method === 'POST'
											? 'bg-green-700/60 text-green-300'
											: ep.method === 'DELETE'
												? 'bg-red-700/60 text-red-300'
												: 'bg-yellow-700/60 text-yellow-300'}">{ep.method}</span
								>
								<code class="text-sm font-mono text-neutral-200 flex-1"
									>{ep.path}</code
								>
								<span class="text-xs text-neutral-500">{ep.label}</span>
							</div>
							{#if ep.body}
								<pre
									class="text-xs font-mono text-neutral-400 px-4 pb-2.5 overflow-x-auto">{ep.body}</pre>
							{/if}
						</div>
					{/each}
				</div>
			</section>

			<!-- Projects -->
			<section id="projects">
				<h2
					class="text-xl font-bold text-white mb-4 pb-2 border-b border-neutral-700"
				>
					Projects
				</h2>
				<div class="space-y-2">
					{#each [{ method: "GET", path: "/projects", label: "List all projects" }, { method: "POST", path: "/projects", label: "Create personal project", body: '{ "name": "My Project", "description": "..." }' }, { method: "GET", path: "/projects/:projectId", label: "Get project" }, { method: "PUT", path: "/projects/:projectId", label: "Update project", body: '{ "name": "New Name", "description": "..." }' }, { method: "PUT", path: "/projects/:projectId/archive", label: "Archive project" }, { method: "DELETE", path: "/projects/:projectId", label: "Delete project" }, { method: "GET", path: "/projects/:projectId/members", label: "List members" }, { method: "POST", path: "/projects/:projectId/members", label: "Add member", body: '{ "user_id": "...", "role_flags": 1 }' }, { method: "PUT", path: "/projects/:projectId/members/:userId", label: "Update member role" }, { method: "DELETE", path: "/projects/:projectId/members/:userId", label: "Remove member" }, { method: "GET", path: "/projects/:projectId/events", label: "List events" }, { method: "POST", path: "/projects/:projectId/events", label: "Create event" }, { method: "GET", path: "/projects/:projectId/files", label: "List files" }, { method: "POST", path: "/projects/:projectId/files/folder", label: "Create folder" }, { method: "POST", path: "/projects/:projectId/files/upload", label: "Upload file (multipart)" }, { method: "GET", path: "/projects/:projectId/webhooks", label: "List webhooks" }, { method: "POST", path: "/projects/:projectId/webhooks", label: "Create webhook" }, { method: "GET", path: "/projects/:projectId/whiteboard", label: "Get whiteboard data" }, { method: "PUT", path: "/projects/:projectId/whiteboard", label: "Save whiteboard", body: '{ "data": "<canvas JSON>" }' }] as ep}
						<div class="bg-neutral-800 border border-neutral-700 rounded-md">
							<div class="flex items-center gap-3 px-4 py-2.5">
								<span
									class="text-xs font-bold px-2 py-0.5 rounded w-14 text-center {ep.method ===
									'GET'
										? 'bg-blue-700/60 text-blue-300'
										: ep.method === 'POST'
											? 'bg-green-700/60 text-green-300'
											: ep.method === 'DELETE'
												? 'bg-red-700/60 text-red-300'
												: 'bg-yellow-700/60 text-yellow-300'}">{ep.method}</span
								>
								<code class="text-sm font-mono text-neutral-200 flex-1"
									>{ep.path}</code
								>
								<span class="text-xs text-neutral-500">{ep.label}</span>
							</div>
							{#if ep.body}
								<pre
									class="text-xs font-mono text-neutral-400 px-4 pb-2.5 overflow-x-auto">{ep.body}</pre>
							{/if}
						</div>
					{/each}
				</div>
			</section>

			<!-- Boards & Cards -->
			<section id="boards">
				<h2
					class="text-xl font-bold text-white mb-4 pb-2 border-b border-neutral-700"
				>
					Boards & Cards
				</h2>
				<div class="space-y-2">
					{#each [{ method: "GET", path: "/projects/:projectId/board", label: "Get board (columns + cards)" }, { method: "POST", path: "/projects/:projectId/columns", label: "Create column", body: '{ "title": "To Do" }' }, { method: "PUT", path: "/projects/:projectId/columns/:columnId", label: "Rename column", body: '{ "title": "In Progress" }' }, { method: "PUT", path: "/projects/:projectId/columns/:columnId/position", label: "Reorder column", body: '{ "position": 2 }' }, { method: "DELETE", path: "/projects/:projectId/columns/:columnId", label: "Delete column" }, { method: "POST", path: "/projects/:projectId/columns/:columnId/cards", label: "Create card", body: '{ "title": "Task", "priority": "high", "assignees": ["email@x.com"] }' }, { method: "PUT", path: "/cards/:cardId", label: "Update card" }, { method: "PUT", path: "/cards/:cardId/move", label: "Move card", body: '{ "column_id": "...", "position": 0 }' }, { method: "DELETE", path: "/cards/:cardId", label: "Delete card" }] as ep}
						<div class="bg-neutral-800 border border-neutral-700 rounded-md">
							<div class="flex items-center gap-3 px-4 py-2.5">
								<span
									class="text-xs font-bold px-2 py-0.5 rounded w-14 text-center {ep.method ===
									'GET'
										? 'bg-blue-700/60 text-blue-300'
										: ep.method === 'POST'
											? 'bg-green-700/60 text-green-300'
											: ep.method === 'DELETE'
												? 'bg-red-700/60 text-red-300'
												: 'bg-yellow-700/60 text-yellow-300'}">{ep.method}</span
								>
								<code class="text-sm font-mono text-neutral-200 flex-1"
									>{ep.path}</code
								>
								<span class="text-xs text-neutral-500">{ep.label}</span>
							</div>
							{#if ep.body}
								<pre
									class="text-xs font-mono text-neutral-400 px-4 pb-2.5 overflow-x-auto">{ep.body}</pre>
							{/if}
						</div>
					{/each}
				</div>
			</section>

			<!-- Events -->
			<section id="events">
				<h2
					class="text-xl font-bold text-white mb-4 pb-2 border-b border-neutral-700"
				>
					Events
				</h2>
				<div class="space-y-2">
					{#each [{ method: "PUT", path: "/events/:eventId", label: "Update event", body: '{ "title": "Sprint Review", "date": "2025-03-01", "time": "14:00", "color": "#6366f1" }' }, { method: "DELETE", path: "/events/:eventId", label: "Delete event" }] as ep}
						<div class="bg-neutral-800 border border-neutral-700 rounded-md">
							<div class="flex items-center gap-3 px-4 py-2.5">
								<span
									class="text-xs font-bold px-2 py-0.5 rounded w-14 text-center {ep.method ===
									'DELETE'
										? 'bg-red-700/60 text-red-300'
										: 'bg-yellow-700/60 text-yellow-300'}">{ep.method}</span
								>
								<code class="text-sm font-mono text-neutral-200 flex-1"
									>{ep.path}</code
								>
								<span class="text-xs text-neutral-500">{ep.label}</span>
							</div>
							{#if ep.body}
								<pre
									class="text-xs font-mono text-neutral-400 px-4 pb-2.5 overflow-x-auto">{ep.body}</pre>
							{/if}
						</div>
					{/each}
				</div>
			</section>

			<!-- Files -->
			<section id="files">
				<h2
					class="text-xl font-bold text-white mb-4 pb-2 border-b border-neutral-700"
				>
					Files
				</h2>
				<div class="space-y-2">
					{#each [{ method: "GET", path: "/users/me/files", label: "List personal files", note: "?parent_id= for folder navigation" }, { method: "POST", path: "/users/me/files/folder", label: "Create folder", body: '{ "name": "Designs", "parent_id": "" }' }, { method: "POST", path: "/users/me/files/upload", label: "Upload file", note: "Multipart: file field + optional parent_id" }, { method: "GET", path: "/files/:fileId/download", label: "Download file" }, { method: "DELETE", path: "/files/:fileId", label: "Delete file" }] as ep}
						<div class="bg-neutral-800 border border-neutral-700 rounded-md">
							<div class="flex items-center gap-3 px-4 py-2.5">
								<span
									class="text-xs font-bold px-2 py-0.5 rounded w-14 text-center {ep.method ===
									'GET'
										? 'bg-blue-700/60 text-blue-300'
										: ep.method === 'POST'
											? 'bg-green-700/60 text-green-300'
											: 'bg-red-700/60 text-red-300'}">{ep.method}</span
								>
								<code class="text-sm font-mono text-neutral-200 flex-1"
									>{ep.path}</code
								>
								<span class="text-xs text-neutral-500">{ep.label}</span>
							</div>
							{#if ep.body || ep.note}
								<p class="text-xs font-mono text-neutral-400 px-4 pb-2.5">
									{ep.body ?? ep.note}
								</p>
							{/if}
						</div>
					{/each}
				</div>
			</section>

			<!-- Webhooks -->
			<section id="webhooks">
				<h2
					class="text-xl font-bold text-white mb-4 pb-2 border-b border-neutral-700"
				>
					Webhooks
				</h2>
				<p class="text-sm text-neutral-400 mb-4">
					Webhook types: <code class="bg-neutral-800 px-1 rounded text-blue-300"
						>discord</code
					>
					·
					<code class="bg-neutral-800 px-1 rounded text-blue-300">github</code>
					· <code class="bg-neutral-800 px-1 rounded text-blue-300">gitea</code>
					· <code class="bg-neutral-800 px-1 rounded text-blue-300">slack</code>
					·
					<code class="bg-neutral-800 px-1 rounded text-blue-300">custom</code>
				</p>
				<div class="space-y-2">
					{#each [{ method: "GET", path: "/projects/:projectId/webhooks", label: "List webhooks" }, { method: "POST", path: "/projects/:projectId/webhooks", label: "Create webhook", body: '{ "name": "Deploy notify", "url": "https://...", "type": "discord" }' }, { method: "PUT", path: "/webhooks/:webhookId", label: "Update webhook" }, { method: "PUT", path: "/webhooks/:webhookId/toggle", label: "Enable / disable" }, { method: "DELETE", path: "/webhooks/:webhookId", label: "Delete webhook" }] as ep}
						<div class="bg-neutral-800 border border-neutral-700 rounded-md">
							<div class="flex items-center gap-3 px-4 py-2.5">
								<span
									class="text-xs font-bold px-2 py-0.5 rounded w-14 text-center {ep.method ===
									'GET'
										? 'bg-blue-700/60 text-blue-300'
										: ep.method === 'POST'
											? 'bg-green-700/60 text-green-300'
											: ep.method === 'DELETE'
												? 'bg-red-700/60 text-red-300'
												: 'bg-yellow-700/60 text-yellow-300'}">{ep.method}</span
								>
								<code class="text-sm font-mono text-neutral-200 flex-1"
									>{ep.path}</code
								>
								<span class="text-xs text-neutral-500">{ep.label}</span>
							</div>
							{#if ep.body}
								<pre
									class="text-xs font-mono text-neutral-400 px-4 pb-2.5 overflow-x-auto">{ep.body}</pre>
							{/if}
						</div>
					{/each}
				</div>
			</section>

			<!-- Notifications -->
			<section id="notifications">
				<h2
					class="text-xl font-bold text-white mb-4 pb-2 border-b border-neutral-700"
				>
					Notifications
				</h2>
				<div class="space-y-2">
					{#each [{ method: "GET", path: "/notifications", label: "List notifications" }, { method: "PUT", path: "/notifications/read-all", label: "Mark all as read" }, { method: "PUT", path: "/notifications/:notifId/read", label: "Mark one as read" }, { method: "DELETE", path: "/notifications/:notifId", label: "Delete notification" }] as ep}
						<div class="bg-neutral-800 border border-neutral-700 rounded-md">
							<div class="flex items-center gap-3 px-4 py-2.5">
								<span
									class="text-xs font-bold px-2 py-0.5 rounded w-14 text-center {ep.method ===
									'GET'
										? 'bg-blue-700/60 text-blue-300'
										: ep.method === 'DELETE'
											? 'bg-red-700/60 text-red-300'
											: 'bg-yellow-700/60 text-yellow-300'}">{ep.method}</span
								>
								<code class="text-sm font-mono text-neutral-200 flex-1"
									>{ep.path}</code
								>
								<span class="text-xs text-neutral-500">{ep.label}</span>
							</div>
						</div>
					{/each}
				</div>
			</section>

			<!-- Quick example -->
			<section
				class="bg-neutral-800/60 border border-neutral-700 rounded-xl p-6 mb-4"
			>
				<h3
					class="text-base font-semibold text-white mb-3 flex items-center gap-2"
				>
					<svg
						class="w-4 h-4 text-blue-400"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
						><path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
						/></svg
					>
					Quick Example — curl
				</h3>
				<div class="relative">
					<pre
						class="text-xs font-mono text-neutral-300 leading-relaxed overflow-x-auto bg-neutral-900 rounded-lg p-4">curl -X GET https://your-fpmb-instance/api/projects \
  -H "Authorization: Bearer fpmb_your_api_key_here" \
  -H "Content-Type: application/json"</pre>
					<button
						onclick={() =>
							copy(
								`curl -X GET https://your-fpmb-instance/api/projects \\\n  -H "Authorization: Bearer fpmb_your_api_key_here" \\\n  -H "Content-Type: application/json"`,
								"curl-example",
							)}
						class="absolute top-3 right-3 text-xs flex items-center gap-1 px-2.5 py-1.5 rounded border transition-colors {copiedId ===
						'curl-example'
							? 'bg-green-700 border-green-600 text-white'
							: 'bg-neutral-700 border-neutral-600 text-neutral-400 hover:text-white hover:bg-neutral-600'}"
					>
						{#if copiedId === "curl-example"}
							<svg
								class="w-3 h-3"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
								><path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M5 13l4 4L19 7"
								/></svg
							>
							Copied
						{:else}
							<svg
								class="w-3 h-3"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
								><path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
								/></svg
							>
							Copy
						{/if}
					</button>
				</div>
				<p class="mt-3 text-sm text-neutral-500">
					Generate an API key in <a
						href="/settings/user"
						class="text-blue-400 hover:underline">User Settings → API Keys</a
					>.
				</p>
			</section>
		</div>
	</div>
</div>
