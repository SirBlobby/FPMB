<script lang="ts">
	import { page } from "$app/stores";
	import Icon from "@iconify/svelte";
	import Modal from "$lib/components/Modal/Modal.svelte";
	import Markdown from "$lib/components/Markdown/Markdown.svelte";
	import { onMount } from "svelte";
	import {
		board as boardApi,
		cards as cardsApi,
		projects as projectsApi,
		users as usersApi,
		teams as teamsApi,
	} from "$lib/api";
	import type {
		Column,
		Card,
		Subtask,
		ProjectMember,
		FileItem,
	} from "$lib/types/api";

	let boardId = $derived($page.params.id ?? "");

	interface LocalCard extends Omit<Card, "subtasks"> {
		subtasks: { id: number; text: string; done: boolean }[];
	}

	interface LocalColumn {
		id: string;
		title: string;
		cards: LocalCard[];
	}

	let columns = $state<LocalColumn[]>([]);
	let projectName = $state("");
	let projectVisibility = $state("private");
	let projectTeamId = $state("");
	let projectFiles = $state<FileItem[]>([]);
	let isArchived = $state(false);
	let loading = $state(true);
	let currentView = $state<"kanban" | "table" | "gantt" | "roadmap">("kanban");

	const boardViews = [
		{ id: "kanban", label: "Kanban", icon: "lucide:columns-3" },
		{ id: "table", label: "Task Board", icon: "lucide:table" },
		{ id: "gantt", label: "Gantt Chart", icon: "lucide:gantt-chart" },
		{ id: "roadmap", label: "Roadmap", icon: "lucide:milestone" },
	];

	let allCards = $derived(
		columns.flatMap((col) =>
			col.cards.map((card) => ({
				...card,
				columnTitle: col.title,
				columnId: col.id,
			})),
		),
	);

	onMount(async () => {
		try {
			const [data, project] = await Promise.all([
				boardApi.get(boardId),
				projectsApi.get(boardId),
			]);
			projectName = project.name;
			projectVisibility =
				project.visibility || (project.is_public ? "public" : "private");
			projectTeamId = project.team_id || "";
			isArchived = project.is_archived ?? false;
			projectFiles = projectTeamId
				? await teamsApi.listFiles(projectTeamId).catch(() => [])
				: await usersApi.listFiles().catch(() => []);
			columns = [...data.columns]
				.sort((a, b) => a.position - b.position)
				.map((col) => ({
					id: col.id,
					title: col.title,
					cards: [...(col.cards ?? [])]
						.sort((a, b) => a.position - b.position)
						.map((card) => ({
							...card,
							subtasks: (card.subtasks ?? []).map((st) => ({
								id: st.id,
								text: st.text,
								done: st.done,
							})),
						})),
				}));
		} finally {
			loading = false;
		}
	});

	let draggedCardId = $state<string | null>(null);
	let sourceColumnId = $state<string | null>(null);

	function handleDragStart(cardId: string, columnId: string, e: DragEvent) {
		draggedCardId = cardId;
		sourceColumnId = columnId;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = "move";
			e.dataTransfer.setData("text/plain", cardId);
		}
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		if (e.dataTransfer) e.dataTransfer.dropEffect = "move";
	}

	async function handleDrop(targetColumnId: string, e: DragEvent) {
		e.preventDefault();
		if (
			isArchived ||
			!draggedCardId ||
			!sourceColumnId ||
			sourceColumnId === targetColumnId
		)
			return;

		const srcCol = columns.find((c) => c.id === sourceColumnId);
		const targetCol = columns.find((c) => c.id === targetColumnId);

		if (srcCol && targetCol) {
			const cardIndex = srcCol.cards.findIndex((c) => c.id === draggedCardId);
			if (cardIndex !== -1) {
				const [card] = srcCol.cards.splice(cardIndex, 1);
				targetCol.cards.push(card);
				columns = [...columns];
				const newPosition = targetCol.cards.length - 1;
				cardsApi.move(card.id, targetColumnId, newPosition).catch(() => {});
			}
		}

		draggedCardId = null;
		sourceColumnId = null;
	}

	let isModalOpen = $state(false);
	let activeColumnIdForNewTask = $state<string | null>(null);
	let editingCardId = $state<string | null>(null);

	let newTask = $state({
		title: "",
		description: "",
		priority: "Medium",
		color: "neutral",
		dueDate: "",
		assignees: [] as string[],
		subtasks: [] as { id: number; text: string; done: boolean }[],
	});

	let previewMarkdown = $state(false);
	let newSubtaskText = $state("");

	let assigneeInput = $state("");
	let userSearchResults = $state<{ id: string; name: string; email: string }[]>(
		[],
	);
	let showUserDropdown = $state(false);
	let searchTimeout: ReturnType<typeof setTimeout> | null = null;

	function handleAssigneeInput(e: Event) {
		const val = (e.target as HTMLInputElement).value;
		assigneeInput = val;
		const atIdx = val.lastIndexOf("@");
		if (atIdx !== -1) {
			const query = val.slice(atIdx + 1);
			if (searchTimeout) clearTimeout(searchTimeout);
			if (query.length > 0) {
				searchTimeout = setTimeout(async () => {
					userSearchResults = await usersApi.search(query);
					showUserDropdown = userSearchResults.length > 0;
				}, 200);
			} else {
				userSearchResults = [];
				showUserDropdown = false;
			}
		} else {
			userSearchResults = [];
			showUserDropdown = false;
		}
	}

	function selectUser(user: { id: string; name: string; email: string }) {
		if (!newTask.assignees.includes(user.email)) {
			newTask.assignees = [...newTask.assignees, user.email];
		}
		const atIdx = assigneeInput.lastIndexOf("@");
		assigneeInput =
			atIdx !== -1 ? assigneeInput.slice(0, atIdx) : assigneeInput;
		showUserDropdown = false;
		userSearchResults = [];
	}

	function removeAssignee(email: string) {
		newTask.assignees = newTask.assignees.filter((a) => a !== email);
	}

	function openCreateTaskModal(columnId: string) {
		activeColumnIdForNewTask = columnId;
		editingCardId = null;
		newTask = {
			title: "",
			description: "",
			priority: "Medium",
			color: "neutral",
			dueDate: "",
			assignees: [],
			subtasks: [],
		};
		assigneeInput = "";
		previewMarkdown = false;
		isModalOpen = true;
	}

	function openEditTaskModal(columnId: string, card: LocalCard) {
		activeColumnIdForNewTask = columnId;
		editingCardId = card.id;
		newTask = {
			title: card.title,
			description: card.description || "",
			priority: card.priority || "Medium",
			color: card.color || "neutral",
			dueDate: card.due_date ? card.due_date.split("T")[0] : "",
			assignees: [...(card.assignees || [])],
			subtasks: card.subtasks.map((st) => ({ ...st })),
		};
		assigneeInput = "";
		previewMarkdown = true;
		isModalOpen = true;
	}

	function addSubtask(e: Event) {
		e.preventDefault();
		if (newSubtaskText.trim()) {
			newTask.subtasks = [
				...newTask.subtasks,
				{ id: Date.now(), text: newSubtaskText, done: false },
			];
			newSubtaskText = "";
		}
	}

	async function saveNewTask() {
		if (!newTask.title.trim() || !activeColumnIdForNewTask) return;

		const targetCol = columns.find((c) => c.id === activeColumnIdForNewTask);
		if (!targetCol) return;

		if (editingCardId) {
			const updated = await cardsApi.update(editingCardId, {
				title: newTask.title,
				description: newTask.description,
				priority: newTask.priority,
				color: newTask.color,
				due_date: newTask.dueDate,
				assignees: newTask.assignees,
				subtasks: newTask.subtasks.map((st) => ({
					id: st.id,
					text: st.text,
					done: st.done,
				})),
			});
			targetCol.cards = targetCol.cards.map((card) =>
				card.id === editingCardId
					? {
							...card,
							...updated,
							subtasks: (updated.subtasks ?? []).map((st) => ({
								id: st.id,
								text: st.text,
								done: st.done,
							})),
						}
					: card,
			);
		} else {
			const created = await boardApi.createCard(
				boardId,
				activeColumnIdForNewTask,
				{
					title: newTask.title,
					description: newTask.description,
					priority: newTask.priority,
					color: newTask.color,
					due_date: newTask.dueDate || "",
					assignees: newTask.assignees,
				},
			);
			targetCol.cards = [...targetCol.cards, { ...created, subtasks: [] }];
		}
		columns = [...columns];
		isModalOpen = false;
		editingCardId = null;
	}

	async function addColumn() {
		const name = prompt("Column name:");
		if (!name?.trim()) return;
		const col = await boardApi.createColumn(boardId, name.trim());
		columns = [...columns, { id: col.id, title: col.title, cards: [] }];
	}

	const colorClasses: Record<string, string> = {
		red: "bg-red-500",
		blue: "bg-blue-500",
		green: "bg-green-500",
		purple: "bg-purple-500",
		yellow: "bg-yellow-500",
		neutral: "bg-neutral-600",
	};

	const priorityColors: Record<string, string> = {
		Low: "bg-neutral-500/10 text-neutral-400 border-neutral-500/20",
		Medium: "bg-blue-500/10 text-blue-400 border-blue-500/20",
		High: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20",
		Urgent: "bg-red-500/10 text-red-400 border-red-500/20",
	};

	const priorityIcons: Record<string, string> = {
		Low: "lucide:arrow-down",
		Medium: "lucide:minus",
		High: "lucide:arrow-up",
		Urgent: "lucide:alert-circle",
	};

	let isShareModalOpen = $state(false);
	let selectedVisibility = $state("private");
	let shareUrl = $derived(
		typeof window !== "undefined"
			? window.location.origin + "/board/" + boardId
			: "",
	);
	let copied = $state(false);
	let savingVisibility = $state(false);

	let shareMembers = $state<ProjectMember[]>([]);
	let memberSearchQuery = $state("");
	let memberSearchResults = $state<
		{ id: string; name: string; email: string }[]
	>([]);
	let showMemberDropdown = $state(false);
	let memberSearchTimeout: ReturnType<typeof setTimeout> | null = null;
	let newMemberRole = $state(1);
	let addingMember = $state(false);

	async function openShareModal() {
		selectedVisibility = projectVisibility;
		isShareModalOpen = true;
		if (selectedVisibility === "unlisted") {
			shareMembers = await projectsApi.listMembers(boardId).catch(() => []);
		}
	}

	$effect(() => {
		if (isShareModalOpen && selectedVisibility === "unlisted") {
			projectsApi
				.listMembers(boardId)
				.then((m) => {
					shareMembers = m;
				})
				.catch(() => {});
		}
	});

	function handleMemberSearch(e: Event) {
		const val = (e.target as HTMLInputElement).value;
		memberSearchQuery = val;
		if (memberSearchTimeout) clearTimeout(memberSearchTimeout);
		if (val.length > 0) {
			memberSearchTimeout = setTimeout(async () => {
				memberSearchResults = await usersApi.search(val);
				showMemberDropdown = memberSearchResults.length > 0;
			}, 200);
		} else {
			memberSearchResults = [];
			showMemberDropdown = false;
		}
	}

	async function addMember(user: { id: string; name: string; email: string }) {
		addingMember = true;
		showMemberDropdown = false;
		memberSearchQuery = "";
		memberSearchResults = [];
		try {
			const m = await projectsApi.addMember(boardId, user.id, newMemberRole);
			shareMembers = [...shareMembers, m];
		} finally {
			addingMember = false;
		}
	}

	async function updateMemberRole(userId: string, roleFlags: number) {
		const m = await projectsApi.updateMemberRole(boardId, userId, roleFlags);
		shareMembers = shareMembers.map((sm) =>
			sm.user_id === userId
				? { ...sm, role_flags: m.role_flags, role_name: m.role_name }
				: sm,
		);
	}

	async function removeMember(userId: string) {
		await projectsApi.removeMember(boardId, userId);
		shareMembers = shareMembers.filter((sm) => sm.user_id !== userId);
	}

	async function saveVisibility() {
		savingVisibility = true;
		try {
			await projectsApi.update(boardId, { visibility: selectedVisibility });
			projectVisibility = selectedVisibility;
			isShareModalOpen = false;
		} finally {
			savingVisibility = false;
		}
	}

	function copyLink() {
		navigator.clipboard.writeText(shareUrl).then(() => {
			copied = true;
			setTimeout(() => (copied = false), 2000);
		});
	}

	const visibilityOptions = [
		{
			value: "private",
			icon: "lucide:lock",
			label: "Private",
			description: "Only project members can access this board.",
		},
		{
			value: "unlisted",
			icon: "lucide:users",
			label: "Unlisted",
			description: "Invite specific users with a role. No public link.",
		},
		{
			value: "public",
			icon: "lucide:globe",
			label: "Public",
			description: "Visible to everyone. Listed publicly.",
		},
	];

	const visibilityIcon: Record<string, string> = {
		private: "lucide:lock",
		unlisted: "lucide:users",
		public: "lucide:globe",
	};
</script>

<svelte:head>
	<title>{projectName ? `${projectName} — FPMB` : "Board — FPMB"}</title>
	<meta
		name="description"
		content={projectName
			? `Kanban board for ${projectName}. Manage tasks, columns, and team members.`
			: "Kanban board — manage tasks and track project progress in FPMB."}
	/>
</svelte:head>

<div class="h-full flex flex-col -m-6 p-6 overflow-hidden">
	<header
		class="flex items-center justify-between mb-6 pb-6 border-b border-neutral-700 shrink-0"
	>
		<div class="flex items-center space-x-4">
			<h1 class="text-2xl font-bold text-white">
				{projectName || `Board #${boardId}`}
			</h1>
		</div>
		<div class="flex items-center space-x-3">
			<a
				href="/whiteboard/{boardId}"
				class="text-neutral-400 hover:text-white transition-colors p-2 rounded-md hover:bg-neutral-800 border border-transparent"
				title="Open Whiteboard"
			>
				<Icon icon="lucide:pen-tool" class="w-5 h-5" />
			</a>
			<a
				href="/projects/{boardId}/settings"
				class="text-neutral-400 hover:text-white transition-colors p-2 rounded-md hover:bg-neutral-800 border border-transparent"
				title="Project Settings"
			>
				<Icon icon="lucide:settings" class="w-5 h-5" />
			</a>
			<button
				onclick={openShareModal}
				class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-transparent transition-colors flex items-center space-x-2"
			>
				<Icon
					icon={visibilityIcon[projectVisibility] || "lucide:share-2"}
					class="w-4 h-4"
				/>
				<span>Share</span>
			</button>
		</div>
	</header>

	{#if loading}
		<p class="text-neutral-500 text-sm">Loading board...</p>
	{:else}
		<!-- View Switcher -->
		<div
			class="flex items-center gap-1 mb-4 shrink-0 bg-neutral-800/50 p-1 rounded-lg w-fit border border-neutral-700/50"
		>
			{#each boardViews as v}
				<button
					onclick={() => (currentView = v.id as typeof currentView)}
					class="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium transition-all {currentView ===
					v.id
						? 'bg-blue-600 text-white shadow-sm'
						: 'text-neutral-400 hover:text-white hover:bg-neutral-700'}"
				>
					<Icon icon={v.icon} class="w-3.5 h-3.5" />
					{v.label}
				</button>
			{/each}
		</div>

		{#if isArchived}
			<div
				class="mb-4 shrink-0 flex items-center gap-3 px-4 py-3 rounded-lg bg-yellow-500/10 border border-yellow-500/30 text-yellow-300 text-sm"
			>
				<Icon icon="lucide:archive" class="w-4 h-4 shrink-0" />
				<span>This project is archived and is in read-only mode.</span>
			</div>
		{/if}

		{#if currentView === "kanban"}
			<div
				class="flex-1 overflow-x-auto overflow-y-hidden pb-4 custom-scrollbar"
			>
				<div class="flex h-full items-start space-x-6 min-w-max">
					{#each columns as column (column.id)}
						<div
							class="w-80 h-full flex flex-col bg-neutral-800 rounded-lg border border-neutral-700 shrink-0"
							ondragover={handleDragOver}
							ondrop={(e) => handleDrop(column.id, e)}
							role="list"
						>
							<div
								class="p-4 border-b border-neutral-700 flex justify-between items-center bg-neutral-800 rounded-t-lg shrink-0"
							>
								<h3 class="font-semibold text-white">{column.title}</h3>
								<span
									class="text-xs font-medium bg-neutral-700 text-neutral-300 py-1 px-2 rounded-full"
									>{column.cards.length}</span
								>
							</div>

							<div
								class="flex-1 p-3 overflow-y-auto space-y-3 custom-scrollbar"
							>
								{#each column.cards as card (card.id)}
									<div
										draggable={!isArchived}
										ondragstart={(e) => handleDragStart(card.id, column.id, e)}
										class="bg-neutral-750 p-4 rounded-md border border-neutral-600 shadow-sm {isArchived
											? 'cursor-default'
											: 'cursor-grab active:cursor-grabbing'} hover:border-neutral-500 transition-colors flex flex-col gap-2 group relative overflow-hidden"
										role="listitem"
										onclick={() =>
											!isArchived && openEditTaskModal(column.id, card)}
										onkeydown={(e) =>
											e.key === "Enter" &&
											!isArchived &&
											openEditTaskModal(column.id, card)}
										tabindex="0"
									>
										{#if card.color && card.color !== "neutral"}
											<div
												class="absolute top-0 left-0 w-1 h-full {colorClasses[
													card.color
												]}"
											></div>
										{/if}

										<div class="flex items-start justify-between">
											<div class="flex items-center gap-2 flex-wrap">
												<span
													class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded border {priorityColors[
														card.priority
													] ||
														priorityColors[
															'Medium'
														]} text-[10px] uppercase font-bold tracking-wider"
												>
													<Icon
														icon={priorityIcons[card.priority] ||
															priorityIcons["Medium"]}
														class="w-3 h-3"
													/>
													{card.priority}
												</span>
											</div>
											<button
												class="p-1 rounded-md text-neutral-500 hover:text-white hover:bg-neutral-600 transition-colors opacity-0 group-hover:opacity-100 {isArchived
													? 'hidden'
													: ''}"
												onclick={(e) => {
													e.stopPropagation();
													openEditTaskModal(column.id, card);
												}}
												aria-label="Edit Task"
											>
												<Icon icon="lucide:more-horizontal" class="w-4 h-4" />
											</button>
										</div>

										<h4 class="text-sm font-medium text-white">{card.title}</h4>

										{#if card.subtasks && card.subtasks.length > 0}
											<div
												class="flex items-center gap-1 text-xs text-neutral-400 mt-1"
											>
												<Icon icon="lucide:check-square" class="w-3.5 h-3.5" />
												<span
													>{card.subtasks.filter((st) => st.done).length}/{card
														.subtasks.length}</span
												>
											</div>
										{/if}

										<div class="flex items-center justify-between mt-2">
											<div class="flex items-center gap-2">
												{#if card.due_date}
													<div
														class="flex items-center text-xs font-medium text-neutral-400 bg-neutral-700/50 px-2 py-1 rounded"
													>
														<Icon icon="lucide:calendar" class="w-3 h-3 mr-1" />
														{new Date(card.due_date).toLocaleDateString(
															"en-US",
															{
																month: "2-digit",
																day: "2-digit",
																year: "numeric",
															},
														)}
													</div>
												{/if}
											</div>
											{#if card.assignees && card.assignees.length > 0}
												<div class="flex -space-x-1 overflow-hidden">
													{#each card.assignees as assignee}
														<div
															class="h-5 w-5 rounded-full ring-2 ring-neutral-800 bg-blue-600 flex items-center justify-center text-[10px] font-bold text-white uppercase shrink-0"
														>
															{assignee.charAt(0)}
														</div>
													{/each}
												</div>
											{/if}
										</div>
									</div>
								{/each}
							</div>

							<div class="p-3 border-t border-neutral-700 shrink-0">
								{#if !isArchived}
									<button
										onclick={() => openCreateTaskModal(column.id)}
										class="w-full flex items-center justify-center py-2 text-sm font-medium text-neutral-400 hover:text-white hover:bg-neutral-700 rounded-md transition-colors"
									>
										<Icon icon="lucide:plus" class="w-4 h-4 mr-1" />
										Add Card
									</button>
								{/if}
							</div>
						</div>
					{/each}

					{#if !isArchived}
						<button
							onclick={addColumn}
							class="w-80 shrink-0 h-12 flex items-center justify-center text-neutral-400 border-2 border-dashed border-neutral-700 hover:border-neutral-500 hover:text-white rounded-lg transition-colors bg-neutral-800/50 font-medium"
						>
							<Icon icon="lucide:plus" class="w-5 h-5 mr-2" />
							Add Column
						</button>
					{/if}
				</div>
			</div>
		{:else if currentView === "table"}
			<!-- Task Board View -->
			<div class="flex-1 overflow-auto rounded-lg border border-neutral-700">
				<table class="w-full text-sm">
					<thead class="bg-neutral-800 sticky top-0 z-10">
						<tr
							class="text-left text-neutral-400 text-xs uppercase tracking-wider"
						>
							<th class="px-4 py-3 font-medium">Task</th>
							<th class="px-4 py-3 font-medium">Status</th>
							<th class="px-4 py-3 font-medium">Priority</th>
							<th class="px-4 py-3 font-medium">Due Date</th>
							<th class="px-4 py-3 font-medium">Assignees</th>
							<th class="px-4 py-3 font-medium">Subtasks</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-neutral-800">
						{#each allCards as card (card.id)}
							<tr
								class="hover:bg-neutral-800/50 transition-colors cursor-pointer"
								onclick={() =>
									!isArchived && openEditTaskModal(card.columnId, card)}
							>
								<td class="px-4 py-3">
									<div class="flex items-center gap-2">
										{#if card.color && card.color !== "neutral"}
											<div
												class="w-1 h-6 rounded-full {colorClasses[card.color]}"
											></div>
										{/if}
										<span class="text-white font-medium">{card.title}</span>
									</div>
								</td>
								<td class="px-4 py-3">
									<span
										class="px-2 py-0.5 rounded text-xs font-medium bg-neutral-700 text-neutral-300"
										>{card.columnTitle}</span
									>
								</td>
								<td class="px-4 py-3">
									<span
										class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded border {priorityColors[
											card.priority
										] ||
											priorityColors[
												'Medium'
											]} text-[10px] uppercase font-bold tracking-wider"
									>
										<Icon
											icon={priorityIcons[card.priority] ||
												priorityIcons["Medium"]}
											class="w-3 h-3"
										/>
										{card.priority}
									</span>
								</td>
								<td class="px-4 py-3 text-neutral-400 text-xs">
									{#if card.due_date}
										{new Date(card.due_date).toLocaleDateString("en-US", {
											month: "short",
											day: "numeric",
											year: "numeric",
										})}
									{:else}
										<span class="text-neutral-600">—</span>
									{/if}
								</td>
								<td class="px-4 py-3">
									{#if card.assignees?.length}
										<div class="flex -space-x-1">
											{#each card.assignees as a}
												<div
													class="h-5 w-5 rounded-full ring-2 ring-neutral-900 bg-blue-600 flex items-center justify-center text-[9px] font-bold text-white uppercase"
												>
													{a.charAt(0)}
												</div>
											{/each}
										</div>
									{:else}
										<span class="text-neutral-600 text-xs">—</span>
									{/if}
								</td>
								<td class="px-4 py-3 text-neutral-400 text-xs">
									{#if card.subtasks?.length}
										{card.subtasks.filter((s) => s.done).length}/{card.subtasks
											.length}
									{:else}
										<span class="text-neutral-600">—</span>
									{/if}
								</td>
							</tr>
						{:else}
							<tr
								><td colspan="6" class="px-4 py-8 text-center text-neutral-500"
									>No tasks yet.</td
								></tr
							>
						{/each}
					</tbody>
				</table>
			</div>
		{:else if currentView === "gantt"}
			<!-- Gantt Chart View -->
			{@const now = new Date()}
			{@const datedCards = allCards.filter((c) => c.due_date)}
			{@const ganttStart = datedCards.length
				? new Date(
						Math.min(
							...datedCards.map((c) =>
								new Date(c.created_at || c.due_date).getTime(),
							),
							now.getTime() - 7 * 86400000,
						),
					)
				: new Date(now.getTime() - 14 * 86400000)}
			{@const ganttEnd = datedCards.length
				? new Date(
						Math.max(
							...datedCards.map((c) => new Date(c.due_date).getTime()),
							now.getTime() + 7 * 86400000,
						),
					)
				: new Date(now.getTime() + 30 * 86400000)}
			{@const totalDays = Math.max(
				Math.ceil((ganttEnd.getTime() - ganttStart.getTime()) / 86400000),
				14,
			)}
			<div class="flex-1 overflow-auto rounded-lg border border-neutral-700">
				<div class="min-w-[800px]">
					<!-- Timeline header -->
					<div
						class="flex bg-neutral-800 border-b border-neutral-700 sticky top-0 z-10"
					>
						<div
							class="w-56 shrink-0 px-4 py-2 text-xs font-medium text-neutral-400 uppercase tracking-wider border-r border-neutral-700"
						>
							Task
						</div>
						<div class="flex-1 flex">
							{#each Array(totalDays) as _, i}
								{@const d = new Date(ganttStart.getTime() + i * 86400000)}
								{@const isToday = d.toDateString() === now.toDateString()}
								<div
									class="flex-1 min-w-[40px] text-center py-2 text-[10px] border-r border-neutral-800 {isToday
										? 'bg-blue-600/10 text-blue-400 font-bold'
										: 'text-neutral-500'}"
								>
									{d.getDate()}/{d.getMonth() + 1}
								</div>
							{/each}
						</div>
					</div>
					<!-- Rows -->
					{#if datedCards.length === 0}
						<div class="px-8 py-12 text-center text-neutral-500 text-sm">
							<Icon
								icon="lucide:calendar-x"
								class="w-8 h-8 mx-auto mb-2 text-neutral-600"
							/>
							No tasks with due dates. Add due dates to see them on the Gantt chart.
						</div>
					{:else}
						{#each datedCards as card (card.id)}
							{@const created = new Date(card.created_at || card.due_date)}
							{@const due = new Date(card.due_date)}
							{@const startOffset = Math.max(
								0,
								((created.getTime() - ganttStart.getTime()) /
									(ganttEnd.getTime() - ganttStart.getTime())) *
									100,
							)}
							{@const barWidth = Math.max(
								2,
								((due.getTime() - created.getTime()) /
									(ganttEnd.getTime() - ganttStart.getTime())) *
									100,
							)}
							<div
								class="flex border-b border-neutral-800 hover:bg-neutral-800/30 transition-colors group"
							>
								<div
									class="w-56 shrink-0 px-4 py-3 border-r border-neutral-700 flex items-center gap-2 cursor-pointer"
									onclick={() =>
										!isArchived && openEditTaskModal(card.columnId, card)}
								>
									{#if card.color && card.color !== "neutral"}<div
											class="w-1 h-5 rounded-full {colorClasses[card.color]}"
										></div>{/if}
									<span class="text-xs text-white font-medium truncate"
										>{card.title}</span
									>
								</div>
								<div class="flex-1 relative py-2 px-1">
									<div
										class="absolute h-6 rounded-md top-1/2 -translate-y-1/2 {card.color &&
										card.color !== 'neutral'
											? colorClasses[card.color]
											: 'bg-blue-600'} opacity-80 group-hover:opacity-100 transition-opacity flex items-center justify-center"
										style="left: {startOffset}%; width: {barWidth}%; min-width: 24px;"
									>
										<span
											class="text-[9px] text-white font-medium truncate px-1"
											>{card.columnTitle}</span
										>
									</div>
								</div>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		{:else if currentView === "roadmap"}
			<!-- Roadmap View -->
			<div class="flex-1 overflow-auto">
				<div class="max-w-3xl mx-auto space-y-0">
					{#each columns as column, colIdx (column.id)}
						<div class="relative pl-8">
							<!-- Timeline line -->
							{#if colIdx < columns.length - 1}
								<div
									class="absolute left-[15px] top-10 bottom-0 w-0.5 bg-neutral-700"
								></div>
							{/if}
							<!-- Milestone dot -->
							<div
								class="absolute left-[8px] top-3 w-4 h-4 rounded-full border-2 {column
									.cards.length > 0 &&
								column.cards.every(
									(c) =>
										c.subtasks?.length > 0 && c.subtasks.every((s) => s.done),
								)
									? 'bg-green-500 border-green-400'
									: 'bg-neutral-600 border-neutral-500'}"
							></div>
							<!-- Column header -->
							<div class="pb-2 pt-1">
								<h3
									class="text-base font-semibold text-white flex items-center gap-2"
								>
									{column.title}
									<span class="text-xs font-normal text-neutral-500"
										>{column.cards.length} tasks</span
									>
								</h3>
							</div>
							<!-- Cards -->
							<div class="space-y-2 pb-8">
								{#each column.cards as card (card.id)}
									<div
										class="bg-neutral-800 border border-neutral-700 rounded-lg p-3 hover:border-neutral-600 transition-colors cursor-pointer group"
										onclick={() =>
											!isArchived && openEditTaskModal(column.id, card)}
									>
										<div class="flex items-start justify-between gap-2">
											<div class="flex items-center gap-2 min-w-0">
												{#if card.color && card.color !== "neutral"}<div
														class="w-1 h-5 rounded-full shrink-0 {colorClasses[
															card.color
														]}"
													></div>{/if}
												<span class="text-sm font-medium text-white truncate"
													>{card.title}</span
												>
											</div>
											<span
												class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded border {priorityColors[
													card.priority
												] ||
													priorityColors[
														'Medium'
													]} text-[9px] uppercase font-bold tracking-wider shrink-0"
											>
												<Icon
													icon={priorityIcons[card.priority] ||
														priorityIcons["Medium"]}
													class="w-2.5 h-2.5"
												/>
												{card.priority}
											</span>
										</div>
										{#if card.due_date || card.subtasks?.length > 0}
											<div
												class="flex items-center gap-3 mt-2 text-xs text-neutral-500"
											>
												{#if card.due_date}
													<span class="flex items-center gap-1">
														<Icon icon="lucide:calendar" class="w-3 h-3" />
														{new Date(card.due_date).toLocaleDateString(
															"en-US",
															{ month: "short", day: "numeric" },
														)}
													</span>
												{/if}
												{#if card.subtasks?.length > 0}
													<span class="flex items-center gap-1">
														<Icon icon="lucide:check-square" class="w-3 h-3" />
														{card.subtasks.filter((s) => s.done).length}/{card
															.subtasks.length}
													</span>
												{/if}
											</div>
										{/if}
										{#if card.subtasks?.length > 0}
											{@const done = card.subtasks.filter((s) => s.done).length}
											{@const pct = Math.round(
												(done / card.subtasks.length) * 100,
											)}
											<div
												class="mt-2 h-1 bg-neutral-700 rounded-full overflow-hidden"
											>
												<div
													class="h-full rounded-full transition-all {pct === 100
														? 'bg-green-500'
														: 'bg-blue-500'}"
													style="width: {pct}%"
												></div>
											</div>
										{/if}
									</div>
								{:else}
									<p class="text-xs text-neutral-600 italic pl-1">
										No tasks in this stage
									</p>
								{/each}
							</div>
						</div>
					{:else}
						<div class="text-center py-12 text-neutral-500 text-sm">
							No columns. Add columns in Kanban view first.
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>

<Modal
	bind:isOpen={isModalOpen}
	title={editingCardId ? "Edit Task" : "Create New Task"}
>
	<div class="space-y-6">
		<div class="grid grid-cols-1 md:grid-cols-12 gap-4">
			<div class="md:col-span-8">
				<label class="block text-sm font-medium text-neutral-300 mb-1"
					>Title</label
				>
				<input
					type="text"
					bind:value={newTask.title}
					placeholder="Task title"
					class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-500 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
			</div>
			<div class="md:col-span-4 grid grid-cols-2 gap-2">
				<div>
					<label class="block text-sm font-medium text-neutral-300 mb-1"
						>Priority</label
					>
					<select
						bind:value={newTask.priority}
						class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
					>
						<option value="Low">Low</option>
						<option value="Medium">Medium</option>
						<option value="High">High</option>
						<option value="Urgent">Urgent</option>
					</select>
				</div>
				<div>
					<label class="block text-sm font-medium text-neutral-300 mb-1"
						>Color</label
					>
					<select
						bind:value={newTask.color}
						class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
					>
						<option value="neutral">None</option>
						<option value="red">Red</option>
						<option value="blue">Blue</option>
						<option value="green">Green</option>
						<option value="yellow">Yellow</option>
						<option value="purple">Purple</option>
					</select>
				</div>
			</div>
		</div>

		<div>
			<div class="flex items-center justify-between mb-1">
				<label class="block text-sm font-medium text-neutral-300"
					>Description</label
				>
				<div class="flex items-center space-x-2">
					<button
						class="text-xs font-medium px-2 py-1 rounded {previewMarkdown
							? 'text-neutral-400 hover:bg-neutral-700'
							: 'bg-neutral-700 text-white'}"
						onclick={() => (previewMarkdown = false)}>Write</button
					>
					<button
						class="text-xs font-medium px-2 py-1 rounded {previewMarkdown
							? 'bg-neutral-700 text-white'
							: 'text-neutral-400 hover:bg-neutral-700'}"
						onclick={() => (previewMarkdown = true)}>Preview</button
					>
				</div>
			</div>
			<div
				class="border border-neutral-600 rounded-md bg-neutral-700 min-h-[120px]"
			>
				{#if previewMarkdown}
					<div class="p-4 h-full">
						{#if newTask.description}
							<Markdown content={newTask.description} files={projectFiles} />
						{:else}
							<p class="text-neutral-500 italic text-sm">
								No description provided.
							</p>
						{/if}
					</div>
				{:else}
					<textarea
						bind:value={newTask.description}
						placeholder="Supports Markdown format..."
						class="block w-full h-full min-h-[120px] p-3 border-0 bg-transparent text-white placeholder-neutral-500 focus:ring-0 sm:text-sm resize-y"
					></textarea>
				{/if}
			</div>
		</div>

		<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
			<div>
				<label class="block text-sm font-medium text-neutral-300 mb-1"
					>Due Date</label
				>
				<input
					type="date"
					bind:value={newTask.dueDate}
					class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
			</div>
			<div class="relative">
				<label class="block text-sm font-medium text-neutral-300 mb-1"
					>Assignees</label
				>
				{#if newTask.assignees.length > 0}
					<div class="flex flex-wrap gap-1 mb-2">
						{#each newTask.assignees as email}
							<span
								class="inline-flex items-center gap-1 bg-blue-600/20 text-blue-300 border border-blue-500/30 text-xs px-2 py-0.5 rounded-full"
							>
								{email}
								<button
									onclick={() => removeAssignee(email)}
									class="hover:text-white ml-0.5"
								>
									<Icon icon="lucide:x" class="w-3 h-3" />
								</button>
							</span>
						{/each}
					</div>
				{/if}
				<input
					type="text"
					value={assigneeInput}
					oninput={handleAssigneeInput}
					placeholder="Type @ to search users..."
					class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-500 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
				{#if showUserDropdown && userSearchResults.length > 0}
					<div
						class="absolute z-50 w-full mt-1 bg-neutral-800 border border-neutral-600 rounded-md shadow-lg overflow-hidden"
					>
						{#each userSearchResults as user}
							<button
								type="button"
								onclick={() => selectUser(user)}
								class="w-full flex items-center gap-3 px-3 py-2 hover:bg-neutral-700 transition-colors text-left"
							>
								<div
									class="w-7 h-7 rounded-full bg-blue-600 flex items-center justify-center text-xs font-bold text-white uppercase shrink-0"
								>
									{user.name.charAt(0)}
								</div>
								<div class="min-w-0">
									<div class="text-sm font-medium text-white truncate">
										{user.name}
									</div>
									<div class="text-xs text-neutral-400 truncate">
										{user.email}
									</div>
								</div>
							</button>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<div>
			<label class="block text-sm font-medium text-neutral-300 mb-2"
				>Subtasks</label
			>
			<ul class="space-y-2 mb-3">
				{#each newTask.subtasks as subtask, i}
					<li class="flex items-center gap-2 text-sm text-neutral-200">
						<input
							type="checkbox"
							bind:checked={newTask.subtasks[i].done}
							class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-neutral-600 rounded bg-neutral-700"
						/>
						<span class={subtask.done ? "line-through text-neutral-500" : ""}
							>{subtask.text}</span
						>
						<button
							class="ml-auto text-neutral-500 hover:text-red-400"
							onclick={() =>
								(newTask.subtasks = newTask.subtasks.filter(
									(st) => st.id !== subtask.id,
								))}
						>
							<Icon icon="lucide:x" class="w-4 h-4" />
						</button>
					</li>
				{/each}
			</ul>
			<form onsubmit={addSubtask} class="flex gap-2">
				<input
					type="text"
					bind:value={newSubtaskText}
					placeholder="Add a subtask..."
					class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-500 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
				<button
					type="submit"
					class="bg-neutral-600 hover:bg-neutral-500 text-white px-3 py-2 rounded-md transition-colors"
				>
					Add
				</button>
			</form>
		</div>

		<div class="flex justify-end pt-4 border-t border-neutral-700 gap-3">
			<button
				onclick={() => (isModalOpen = false)}
				class="bg-transparent hover:bg-neutral-700 text-neutral-300 font-medium py-2 px-4 rounded-md border border-neutral-600 transition-colors text-sm"
			>
				Cancel
			</button>
			<button
				onclick={saveNewTask}
				disabled={!newTask.title.trim()}
				class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md shadow-sm border border-transparent transition-colors text-sm disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{editingCardId ? "Save Changes" : "Create Task"}
			</button>
		</div>
	</div>
</Modal>

<Modal bind:isOpen={isShareModalOpen} title="Share & Visibility">
	<div class="space-y-5">
		<div class="space-y-2">
			<p class="text-sm font-medium text-neutral-300">Visibility</p>
			<div class="space-y-2">
				{#each visibilityOptions as opt}
					<button
						type="button"
						onclick={() => (selectedVisibility = opt.value)}
						class="w-full flex items-start gap-3 p-3 rounded-lg border transition-all text-left {selectedVisibility ===
						opt.value
							? 'border-blue-500 bg-blue-500/10'
							: 'border-neutral-600 hover:border-neutral-500 bg-neutral-700/30'}"
					>
						<div
							class="mt-0.5 p-1.5 rounded-md {selectedVisibility === opt.value
								? 'bg-blue-500/20 text-blue-400'
								: 'bg-neutral-700 text-neutral-400'}"
						>
							<Icon icon={opt.icon} class="w-4 h-4" />
						</div>
						<div class="flex-1 min-w-0">
							<div
								class="text-sm font-medium {selectedVisibility === opt.value
									? 'text-white'
									: 'text-neutral-300'}"
							>
								{opt.label}
							</div>
							<div class="text-xs text-neutral-500 mt-0.5">
								{opt.description}
							</div>
						</div>
						{#if selectedVisibility === opt.value}
							<Icon
								icon="lucide:check-circle"
								class="w-4 h-4 text-blue-400 mt-0.5 shrink-0"
							/>
						{/if}
					</button>
				{/each}
			</div>
		</div>

		{#if selectedVisibility === "public"}
			<div>
				<p class="text-sm font-medium text-neutral-300 mb-2">Board Link</p>
				<div class="flex gap-2">
					<input
						type="text"
						value={shareUrl}
						readonly
						class="flex-1 px-3 py-2 border border-neutral-600 rounded-md bg-neutral-900 text-neutral-300 text-sm focus:outline-none select-all"
					/>
					<button
						onclick={copyLink}
						class="px-3 py-2 rounded-md border transition-colors text-sm font-medium {copied
							? 'bg-green-600/20 border-green-500/40 text-green-400'
							: 'bg-neutral-700 border-neutral-600 text-neutral-300 hover:text-white hover:bg-neutral-600'}"
					>
						{#if copied}
							<Icon icon="lucide:check" class="w-4 h-4" />
						{:else}
							<Icon icon="lucide:copy" class="w-4 h-4" />
						{/if}
					</button>
				</div>
			</div>
		{/if}

		{#if selectedVisibility === "unlisted"}
			<div class="space-y-3">
				<p class="text-sm font-medium text-neutral-300">Invite Members</p>
				<p class="text-xs text-neutral-500">
					Only people you invite can access this board. No public link is
					generated.
				</p>
				<div class="relative">
					<div class="flex gap-2">
						<div class="flex-1 relative">
							<input
								type="text"
								value={memberSearchQuery}
								oninput={handleMemberSearch}
								placeholder="Search by name or email..."
								class="block w-full px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white placeholder-neutral-500 text-sm focus:outline-none focus:border-blue-500"
							/>
							{#if showMemberDropdown && memberSearchResults.length > 0}
								<div
									class="absolute z-50 w-full mt-1 bg-neutral-800 border border-neutral-600 rounded-md shadow-lg overflow-hidden"
								>
									{#each memberSearchResults as user}
										<button
											type="button"
											onclick={() => addMember(user)}
											class="w-full flex items-center gap-3 px-3 py-2 hover:bg-neutral-700 transition-colors text-left"
										>
											<div
												class="w-7 h-7 rounded-full bg-blue-600 flex items-center justify-center text-xs font-bold text-white uppercase shrink-0"
											>
												{user.name.charAt(0)}
											</div>
											<div class="min-w-0">
												<div class="text-sm font-medium text-white truncate">
													{user.name}
												</div>
												<div class="text-xs text-neutral-400 truncate">
													{user.email}
												</div>
											</div>
										</button>
									{/each}
								</div>
							{/if}
						</div>
						<select
							bind:value={newMemberRole}
							class="px-2 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white text-sm focus:outline-none focus:border-blue-500"
						>
							<option value={1}>Viewer</option>
							<option value={2}>Editor</option>
							<option value={4}>Admin</option>
						</select>
					</div>
				</div>

				{#if shareMembers.length > 0}
					<div class="space-y-2 mt-2 max-h-48 overflow-y-auto">
						{#each shareMembers as member (member.user_id)}
							<div
								class="flex items-center gap-3 py-1.5 px-2 rounded-lg bg-neutral-700/40"
							>
								<div
									class="w-7 h-7 rounded-full bg-blue-600 flex items-center justify-center text-xs font-bold text-white uppercase shrink-0"
								>
									{(member.name || member.email || "?").charAt(0)}
								</div>
								<div class="flex-1 min-w-0">
									<div class="text-sm font-medium text-white truncate">
										{member.name || member.email}
									</div>
									<div class="text-xs text-neutral-400 truncate">
										{member.email}
									</div>
								</div>
								{#if member.role_flags >= 8}
									<span class="text-xs text-neutral-400 px-2">Owner</span>
								{:else}
									<select
										value={member.role_flags}
										onchange={(e) =>
											updateMemberRole(
												member.user_id,
												Number((e.target as HTMLSelectElement).value),
											)}
										class="text-xs bg-neutral-700 border border-neutral-600 text-white rounded px-1 py-1 focus:outline-none focus:border-blue-500"
									>
										<option value={1}>Viewer</option>
										<option value={2}>Editor</option>
										<option value={4}>Admin</option>
									</select>
									<button
										onclick={() => removeMember(member.user_id)}
										class="text-neutral-500 hover:text-red-400 transition-colors p-1 rounded"
										title="Remove member"
									>
										<Icon icon="lucide:x" class="w-4 h-4" />
									</button>
								{/if}
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-xs text-neutral-500 italic">No members invited yet.</p>
				{/if}
			</div>
		{/if}

		<div class="flex justify-end pt-3 border-t border-neutral-700 gap-3">
			<button
				onclick={() => (isShareModalOpen = false)}
				class="bg-transparent hover:bg-neutral-700 text-neutral-300 font-medium py-2 px-4 rounded-md border border-neutral-600 transition-colors text-sm"
			>
				Cancel
			</button>
			<button
				onclick={saveVisibility}
				disabled={savingVisibility}
				class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-5 rounded-md shadow-sm transition-colors text-sm disabled:opacity-50"
			>
				{savingVisibility ? "Saving..." : "Save"}
			</button>
		</div>
	</div>
</Modal>

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
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background-color: #737373;
	}
</style>
