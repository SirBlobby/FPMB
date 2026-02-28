<script lang="ts">
	import Icon from "@iconify/svelte";

	let activeSection = $state("overview");

	const sections = [
		{ id: "overview", label: "Overview", icon: "lucide:clipboard-list" },
		{ id: "dashboard", label: "Dashboard", icon: "lucide:layout-dashboard" },
		{ id: "teams", label: "Teams", icon: "lucide:users" },
		{ id: "projects", label: "Projects", icon: "lucide:folder-kanban" },
		{ id: "boards", label: "Board Views", icon: "lucide:bar-chart-3" },
		{ id: "cards", label: "Cards & Tasks", icon: "lucide:square-check-big" },
		{ id: "whiteboard", label: "Whiteboard", icon: "lucide:pen-tool" },
		{ id: "chat", label: "Team Chat", icon: "lucide:message-circle" },
		{ id: "calendar", label: "Calendar", icon: "lucide:calendar-days" },
		{ id: "files", label: "Files", icon: "lucide:paperclip" },
		{ id: "docs-feature", label: "Team Docs", icon: "lucide:file-text" },
		{ id: "notifications", label: "Notifications", icon: "lucide:bell" },
		{ id: "settings", label: "Settings", icon: "lucide:settings" },
		{ id: "api-keys", label: "API Keys", icon: "lucide:key-round" },
		{ id: "webhooks", label: "Webhooks", icon: "lucide:webhook" },
		{ id: "shortcuts", label: "Keyboard Shortcuts", icon: "lucide:keyboard" },
	];
</script>

<svelte:head>
	<title>Documentation — FPMB</title>
	<meta
		name="description"
		content="Complete guide to using FPMB — boards, teams, projects, whiteboard, chat, and more."
	/>
</svelte:head>

<div class="max-w-6xl mx-auto flex gap-8">
	<!-- Sidebar -->
	<nav class="w-52 shrink-0 hidden lg:block sticky top-0 self-start pt-2">
		<h2
			class="text-xs font-semibold uppercase tracking-wider text-neutral-500 mb-3 px-2"
		>
			Documentation
		</h2>
		<ul class="space-y-0.5">
			{#each sections as s}
				<li>
					<button
						onclick={() => {
							activeSection = s.id;
							document
								.getElementById(`section-${s.id}`)
								?.scrollIntoView({ behavior: "smooth", block: "start" });
						}}
						class="w-full text-left px-2 py-1.5 rounded-md text-sm transition-colors flex items-center gap-2 {activeSection ===
						s.id
							? 'bg-blue-600/15 text-blue-300 font-medium'
							: 'text-neutral-400 hover:text-white hover:bg-neutral-800'}"
					>
						<Icon icon={s.icon} class="w-3.5 h-3.5 shrink-0" />
						{s.label}
					</button>
				</li>
			{/each}
		</ul>
	</nav>

	<!-- Content -->
	<div class="flex-1 min-w-0 space-y-12">
		<header>
			<h1 class="text-3xl font-bold text-white mb-2">FPMB Documentation</h1>
			<p class="text-neutral-400 text-base leading-relaxed">
				Everything you need to know about Free Project Management Boards —
				features, workflows, and tips.
			</p>
		</header>

		<!-- Overview -->
		<section id="section-overview" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:clipboard-list" class="w-5 h-5 text-blue-400" /> Overview
			</h2>
			<div
				class="bg-neutral-800 rounded-lg border border-neutral-700 p-5 space-y-3"
			>
				<p class="text-neutral-300 text-sm leading-relaxed">
					FPMB is a full-featured project management platform built with <strong
						class="text-white">SvelteKit</strong
					>
					and <strong class="text-white">Go</strong>. It provides Kanban boards,
					Gantt charts, roadmaps, real-time collaboration, whiteboards, team
					chat, file management, and more.
				</p>
				<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2">
					{#each [{ label: "Board Views", value: "4 types", icon: "lucide:columns-3" }, { label: "Real-time", value: "WebSocket", icon: "lucide:radio" }, { label: "File Storage", value: "Unlimited", icon: "lucide:hard-drive" }, { label: "API Access", value: "Full REST", icon: "lucide:terminal" }] as stat}
						<div class="bg-neutral-900/50 rounded-lg p-3 text-center">
							<Icon
								icon={stat.icon}
								class="w-5 h-5 text-blue-400 mx-auto mb-1"
							/>
							<div class="text-lg font-bold text-blue-400">{stat.value}</div>
							<div class="text-xs text-neutral-500">{stat.label}</div>
						</div>
					{/each}
				</div>
			</div>
		</section>

		<!-- Dashboard -->
		<section id="section-dashboard" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:layout-dashboard" class="w-5 h-5 text-blue-400" /> Dashboard
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>The dashboard is your home page after logging in. It shows:</p>
				<ul class="list-disc pl-5 space-y-1">
					<li>
						<strong class="text-white">Your teams</strong> — all teams you belong
						to with quick access
					</li>
					<li>
						<strong class="text-white">Recent projects</strong> — your most recently
						updated projects
					</li>
					<li>
						<strong class="text-white">Upcoming events</strong> — events from your
						calendar across all teams
					</li>
					<li>
						<strong class="text-white">Notifications</strong> — unread notification
						count in the top bar
					</li>
				</ul>
			</div>
		</section>

		<!-- Teams -->
		<section id="section-teams" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:users" class="w-5 h-5 text-blue-400" /> Teams
			</h2>
			<div class="prose-sm text-neutral-300 space-y-3">
				<p>
					Teams are the organizational unit in FPMB. Every project belongs to a
					team (or to you personally).
				</p>
				<div
					class="bg-neutral-800 rounded-lg border border-neutral-700 p-4 space-y-3"
				>
					<h3 class="text-sm font-semibold text-white">Creating a team</h3>
					<p class="text-xs text-neutral-400">
						Go to the Dashboard and click <strong class="text-neutral-200"
							>"Create Team"</strong
						>. You'll be the owner.
					</p>
					<h3 class="text-sm font-semibold text-white">Roles</h3>
					<div class="grid grid-cols-2 gap-2 text-xs">
						{#each [{ role: "Viewer", desc: "Read-only access to projects and boards", flag: "1", icon: "lucide:eye" }, { role: "Editor", desc: "Create and edit cards, columns, docs", flag: "2", icon: "lucide:pencil" }, { role: "Admin", desc: "Manage members, settings, webhooks", flag: "4", icon: "lucide:shield" }, { role: "Owner", desc: "Full control including team deletion", flag: "8", icon: "lucide:crown" }] as r}
							<div class="bg-neutral-900/50 rounded p-2">
								<div class="flex items-center gap-1.5">
									<Icon icon={r.icon} class="w-3 h-3 text-blue-300" />
									<span class="font-semibold text-blue-300">{r.role}</span>
									<span class="text-neutral-500">(flag {r.flag})</span>
								</div>
								<p class="text-neutral-400 mt-0.5">{r.desc}</p>
							</div>
						{/each}
					</div>
					<h3 class="text-sm font-semibold text-white">Inviting members</h3>
					<p class="text-xs text-neutral-400">
						Team page → <strong class="text-neutral-200">Settings</strong> → Invite
						by email. Choose a role before sending the invite.
					</p>
				</div>
			</div>
		</section>

		<!-- Projects -->
		<section id="section-projects" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:folder-kanban" class="w-5 h-5 text-blue-400" /> Projects
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>
					Projects contain boards, whiteboards, calendars, files, and webhooks.
				</p>
				<ul class="list-disc pl-5 space-y-1">
					<li>
						<strong class="text-white">Team projects</strong> — created from a team
						page, visible to all team members
					</li>
					<li>
						<strong class="text-white">Personal projects</strong> — created from
						the Projects page, only you can access
					</li>
					<li>
						<strong class="text-white">Visibility</strong> — Private (members only),
						Unlisted (invite-only), or Public
					</li>
					<li>
						<strong class="text-white">Archiving</strong> — archived projects become
						read-only, preserving all data
					</li>
				</ul>
				<div
					class="bg-neutral-800 rounded-lg border border-neutral-700 p-4 mt-3"
				>
					<h3
						class="text-sm font-semibold text-white mb-2 flex items-center gap-1.5"
					>
						<Icon icon="lucide:settings" class="w-3.5 h-3.5" /> Project settings
					</h3>
					<p class="text-xs text-neutral-400">
						Access from the board page (gear icon) or from <code
							class="bg-neutral-700 px-1 rounded text-xs"
							>/projects/[id]/settings</code
						>. Here you can rename, update the description, archive, or delete
						the project.
					</p>
				</div>
			</div>
		</section>

		<!-- Board Views -->
		<section id="section-boards" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:bar-chart-3" class="w-5 h-5 text-blue-400" /> Board Views
			</h2>
			<div class="prose-sm text-neutral-300 space-y-3">
				<p>
					Every project board supports <strong class="text-white"
						>4 different views</strong
					> of the same data. Switch between them using the view switcher tabs at
					the top of the board.
				</p>
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
					{#each [{ name: "Kanban", icon: "lucide:columns-3", desc: "Classic column-based board. Drag and drop cards between columns. Add new columns and cards. This is the default view." }, { name: "Task Board", icon: "lucide:table", desc: "Spreadsheet-style table showing all tasks with Status, Priority, Due Date, Assignees, and Subtask progress in sortable rows." }, { name: "Gantt Chart", icon: "lucide:gantt-chart", desc: "Timeline view with horizontal bars from creation to due date. Day-level grid with today highlighted. Requires tasks to have due dates." }, { name: "Roadmap", icon: "lucide:milestone", desc: "Vertical milestone timeline. Each column is a stage. Shows progress bars for subtasks. Green milestones when all subtasks are complete." }] as view}
						<div
							class="bg-neutral-800 rounded-lg border border-neutral-700 p-4"
						>
							<div class="flex items-center gap-2 mb-2">
								<Icon icon={view.icon} class="w-4 h-4 text-blue-400" />
								<h3 class="text-sm font-semibold text-white">{view.name}</h3>
							</div>
							<p class="text-xs text-neutral-400 leading-relaxed">
								{view.desc}
							</p>
						</div>
					{/each}
				</div>
			</div>
		</section>

		<!-- Cards & Tasks -->
		<section id="section-cards" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:square-check-big" class="w-5 h-5 text-blue-400" /> Cards
				& Tasks
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>Cards are the core unit of work. Each card has:</p>
				<ul class="list-disc pl-5 space-y-1">
					<li>
						<strong class="text-white">Title</strong> — required, displayed prominently
						on the board
					</li>
					<li>
						<strong class="text-white">Description</strong> — supports
						<strong class="text-blue-300">Markdown</strong> with preview toggle
					</li>
					<li>
						<strong class="text-white">Priority</strong> — Low, Medium, High, or
						Urgent (color-coded badges)
					</li>
					<li>
						<strong class="text-white">Color label</strong> — visual sidebar indicator
						(red, blue, green, purple, yellow)
					</li>
					<li>
						<strong class="text-white">Due date</strong> — used in calendar, Gantt
						chart, and roadmap views
					</li>
					<li>
						<strong class="text-white">Assignees</strong> — mention with
						<code class="bg-neutral-700 px-1 rounded text-xs">@email</code>,
						shown as avatar circles
					</li>
					<li>
						<strong class="text-white">Subtasks</strong> — checklist items with completion
						tracking and progress bars
					</li>
				</ul>
				<div
					class="bg-neutral-800 rounded-lg border border-neutral-700 p-4 mt-3"
				>
					<h3
						class="text-sm font-semibold text-white mb-1 flex items-center gap-1.5"
					>
						<Icon icon="lucide:move" class="w-3.5 h-3.5" /> Moving cards
					</h3>
					<p class="text-xs text-neutral-400">
						In Kanban view, drag and drop cards between columns. The card's
						position and column are updated automatically via the API.
					</p>
				</div>
			</div>
		</section>

		<!-- Whiteboard -->
		<section id="section-whiteboard" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:pen-tool" class="w-5 h-5 text-blue-400" /> Whiteboard
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>
					Each project has a collaborative whiteboard accessible from the board
					header (pen icon).
				</p>
				<div
					class="bg-neutral-800 rounded-lg border border-neutral-700 p-4 space-y-2"
				>
					<h3 class="text-sm font-semibold text-white mb-2">Tools</h3>
					<div class="grid grid-cols-2 gap-1.5 text-xs">
						{#each [{ tool: "Select", desc: "Click to select objects. Drag to move. Double-click to edit.", icon: "lucide:mouse-pointer" }, { tool: "Pen", desc: "Freehand drawing with configurable color and width.", icon: "lucide:pencil" }, { tool: "Rectangle", desc: "Click and drag to draw rectangles.", icon: "lucide:square" }, { tool: "Circle", desc: "Click and drag to draw circles.", icon: "lucide:circle" }, { tool: "Text", desc: "Click to place text. Set font size in toolbar.", icon: "lucide:type" }, { tool: "Eraser", desc: "Click on any object to delete it.", icon: "lucide:eraser" }] as t}
							<div class="bg-neutral-900/50 rounded p-2 flex items-start gap-2">
								<Icon
									icon={t.icon}
									class="w-3.5 h-3.5 text-blue-300 shrink-0 mt-0.5"
								/>
								<div>
									<span class="font-semibold text-white">{t.tool}</span>
									<span class="text-neutral-400 ml-1">— {t.desc}</span>
								</div>
							</div>
						{/each}
					</div>
					<h3 class="text-sm font-semibold text-white mt-3 mb-1">Features</h3>
					<ul class="list-disc pl-5 space-y-0.5 text-xs text-neutral-400">
						<li>
							<strong class="text-neutral-200">Undo/Redo</strong> — Ctrl+Z / Ctrl+Shift+Z
							(up to 100 steps)
						</li>
						<li>
							<strong class="text-neutral-200">Export PNG</strong> — download the
							whiteboard as a clean image
						</li>
						<li>
							<strong class="text-neutral-200">Real-time collaboration</strong> —
							see other users' cursors live via WebSocket
						</li>
						<li>
							<strong class="text-neutral-200">Auto-save</strong> — changes are saved
							as JSON objects
						</li>
					</ul>
				</div>
			</div>
		</section>

		<!-- Team Chat -->
		<section id="section-chat" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:message-circle" class="w-5 h-5 text-blue-400" /> Team
				Chat
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>
					Each team has a real-time chat room accessible from the team page → <strong
						class="text-white">Chat</strong
					> button.
				</p>
				<ul class="list-disc pl-5 space-y-1">
					<li>
						<strong class="text-white">Real-time messaging</strong> — via WebSocket,
						messages appear instantly
					</li>
					<li>
						<strong class="text-white">Persistent history</strong> — all messages
						saved to the database, infinite scroll to load older messages
					</li>
					<li>
						<strong class="text-white">Typing indicators</strong> — see when teammates
						are typing
					</li>
					<li>
						<strong class="text-white">Online presence</strong> — colored avatars
						show who's currently in the chat
					</li>
					<li>
						<strong class="text-white">Message grouping</strong> — consecutive messages
						from the same user within 5 minutes are grouped
					</li>
					<li>
						<strong class="text-white">Multi-line support</strong> — Shift+Enter
						for new lines, Enter to send
					</li>
				</ul>
			</div>
		</section>

		<!-- Calendar -->
		<section id="section-calendar" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:calendar-days" class="w-5 h-5 text-blue-400" /> Calendar
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>FPMB provides calendars at multiple levels:</p>
				<ul class="list-disc pl-5 space-y-1">
					<li>
						<strong class="text-white">Global calendar</strong> — aggregates events
						from all your teams
					</li>
					<li>
						<strong class="text-white">Team calendar</strong> — events scoped to
						a specific team
					</li>
					<li>
						<strong class="text-white">Project calendar</strong> — events + card
						due dates for a specific project
					</li>
				</ul>
				<p>
					Events have a title, description, date, time, and color label. Card
					due dates automatically appear on their project calendar.
				</p>
			</div>
		</section>

		<!-- Files -->
		<section id="section-files" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:paperclip" class="w-5 h-5 text-blue-400" /> Files
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>
					File storage with folder hierarchy is available at multiple levels:
				</p>
				<ul class="list-disc pl-5 space-y-1">
					<li>
						<strong class="text-white">Personal files</strong> — accessible from
						the top nav "Files" link
					</li>
					<li>
						<strong class="text-white">Team files</strong> — shared with all team
						members
					</li>
					<li>
						<strong class="text-white">Project files</strong> — attached to a specific
						project
					</li>
				</ul>
				<p>
					You can create folders, upload files, download, and delete. Files are
					stored in the <code class="bg-neutral-700 px-1 rounded text-xs"
						>data/</code
					> directory on the server.
				</p>
			</div>
		</section>

		<!-- Team Docs -->
		<section id="section-docs-feature" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:file-text" class="w-5 h-5 text-blue-400" /> Team Docs
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>
					Teams have a built-in document editor for shared notes, meeting
					minutes, and documentation.
				</p>
				<ul class="list-disc pl-5 space-y-1">
					<li>Create and edit documents with a title and rich content</li>
					<li>
						Accessible from the team page → <strong class="text-white"
							>Docs</strong
						> button
					</li>
					<li>Documents are stored per-team and visible to all members</li>
				</ul>
			</div>
		</section>

		<!-- Notifications -->
		<section id="section-notifications" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:bell" class="w-5 h-5 text-blue-400" /> Notifications
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>
					Notifications keep you updated on activity across your teams and
					projects.
				</p>
				<ul class="list-disc pl-5 space-y-1">
					<li>Bell icon in the navbar shows unread count</li>
					<li>Mark individual notifications as read or mark all as read</li>
					<li>
						Triggered by team invites, task assignments, due date reminders, and
						more
					</li>
					<li>Due date reminders run automatically every hour on the server</li>
				</ul>
			</div>
		</section>

		<!-- Settings -->
		<section id="section-settings" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:settings" class="w-5 h-5 text-blue-400" /> Settings
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>
					User settings are available at <code
						class="bg-neutral-700 px-1 rounded text-xs">/settings/user</code
					> (click your avatar in the top bar).
				</p>
				<ul class="list-disc pl-5 space-y-1">
					<li>
						<strong class="text-white">Profile</strong> — update your name, email,
						and avatar
					</li>
					<li>
						<strong class="text-white">Password</strong> — change your password
					</li>
					<li>
						<strong class="text-white">API keys</strong> — generate and manage API
						keys (see below)
					</li>
				</ul>
			</div>
		</section>

		<!-- API Keys -->
		<section id="section-api-keys" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:key-round" class="w-5 h-5 text-blue-400" /> API Keys
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>
					Generate personal API keys for programmatic access to the FPMB REST
					API.
				</p>
				<div
					class="bg-neutral-800 rounded-lg border border-neutral-700 p-4 space-y-2"
				>
					<h3 class="text-sm font-semibold text-white mb-1">Creating a key</h3>
					<ol class="list-decimal pl-5 space-y-0.5 text-xs text-neutral-400">
						<li>
							Go to <strong class="text-neutral-200">Settings → API Keys</strong
							>
						</li>
						<li>Enter a name and select scopes (read, write, admin)</li>
						<li>Click <strong class="text-neutral-200">Generate</strong></li>
						<li>Copy the key immediately — it's only shown once!</li>
					</ol>
					<h3 class="text-sm font-semibold text-white mt-3 mb-1">
						Using a key
					</h3>
					<div
						class="bg-neutral-900 rounded p-3 font-mono text-xs text-green-400"
					>
						curl -H "Authorization: Bearer YOUR_API_KEY" \<br
						/>&nbsp;&nbsp;{typeof window !== "undefined"
							? window.location.origin
							: "https://your-domain.com"}/api/projects
					</div>
					<p class="text-xs text-neutral-500 mt-2">
						See the <a href="/api-docs" class="text-blue-400 hover:underline"
							>API Documentation</a
						> for all available endpoints.
					</p>
				</div>
			</div>
		</section>

		<!-- Webhooks -->
		<section id="section-webhooks" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:webhook" class="w-5 h-5 text-blue-400" /> Webhooks
			</h2>
			<div class="prose-sm text-neutral-300 space-y-2">
				<p>
					Set up webhooks to receive HTTP notifications when events occur in
					your projects.
				</p>
				<ul class="list-disc pl-5 space-y-1">
					<li>
						Configure from <strong class="text-white"
							>Project Settings → Webhooks</strong
						>
					</li>
					<li>
						Choose event types: card created, card moved, card deleted, etc.
					</li>
					<li>Provide a URL — FPMB will POST JSON payloads to it</li>
					<li>Toggle webhooks on/off without deleting them</li>
					<li>View last triggered timestamp for debugging</li>
				</ul>
			</div>
		</section>

		<!-- Keyboard Shortcuts -->
		<section id="section-shortcuts" class="scroll-mt-8">
			<h2 class="text-xl font-bold text-white mb-3 flex items-center gap-2">
				<Icon icon="lucide:keyboard" class="w-5 h-5 text-blue-400" /> Keyboard Shortcuts
			</h2>
			<div class="bg-neutral-800 rounded-lg border border-neutral-700 p-4">
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
					{#each [{ keys: "Ctrl + Z", action: "Undo (whiteboard)" }, { keys: "Ctrl + Shift + Z", action: "Redo (whiteboard)" }, { keys: "Ctrl + Y", action: "Redo (whiteboard, alt)" }, { keys: "Delete / Backspace", action: "Delete selected object (whiteboard)" }, { keys: "Escape", action: "Deselect / close editing" }, { keys: "Enter", action: "Send message (chat) / confirm text (whiteboard)" }, { keys: "Shift + Enter", action: "New line in chat" }, { keys: "Double-click", action: "Edit object (whiteboard)" }] as shortcut}
						<div
							class="flex items-center justify-between bg-neutral-900/50 rounded p-2"
						>
							<span class="text-neutral-400">{shortcut.action}</span>
							<kbd
								class="bg-neutral-700 border border-neutral-600 rounded px-1.5 py-0.5 text-[10px] font-mono text-neutral-300"
								>{shortcut.keys}</kbd
							>
						</div>
					{/each}
				</div>
			</div>
		</section>

		<div class="h-8"></div>
	</div>
</div>
