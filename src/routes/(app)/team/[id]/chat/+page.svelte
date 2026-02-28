<script lang="ts">
	import { page } from "$app/stores";
	import Icon from "@iconify/svelte";
	import { onMount } from "svelte";
	import { teams as teamsApi } from "$lib/api";
	import { getAccessToken } from "$lib/api/client";
	import { authStore } from "$lib/stores/auth.svelte";
	import type { Team, ChatMessage } from "$lib/types/api";

	let teamId = $derived($page.params.id ?? "");
	let team = $state<Team | null>(null);
	let messages = $state<ChatMessage[]>([]);
	let newMessage = $state("");
	let loading = $state(true);
	let loadingMore = $state(false);
	let hasMore = $state(true);
	let sending = $state(false);

	let ws: WebSocket | null = null;
	let wsConnected = $state(false);
	let destroyed = false;
	let onlineUsers = $state<{ user_id: string; name: string }[]>([]);
	let typingUsers = $state<
		Record<string, { name: string; timeout: ReturnType<typeof setTimeout> }>
	>({});

	let messagesContainer: HTMLDivElement;
	let shouldAutoScroll = $state(true);
	let inputEl: HTMLTextAreaElement;

	let typingNames = $derived(Object.values(typingUsers).map((t) => t.name));
	let myId = $derived(authStore.user?.id ?? "");

	function connectWS() {
		const token = getAccessToken();
		if (!token || destroyed) return;
		const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
		const host = window.location.hostname;
		const port =
			window.location.port ||
			(window.location.protocol === "https:" ? "443" : "80");
		const userName = authStore.user?.name ?? "Anonymous";
		const url = `${proto}//${host}:${port}/ws/team/${teamId}/chat?token=${encodeURIComponent(token)}&name=${encodeURIComponent(userName)}`;

		ws = new WebSocket(url);

		ws.onopen = () => {
			wsConnected = true;
		};

		ws.onclose = () => {
			wsConnected = false;
			if (!destroyed) {
				setTimeout(() => {
					if (!destroyed && (!ws || ws.readyState === WebSocket.CLOSED))
						connectWS();
				}, 3000);
			}
		};

		ws.onerror = () => {
			wsConnected = false;
		};

		ws.onmessage = (event) => {
			try {
				const msg = JSON.parse(event.data);
				handleWSMessage(msg);
			} catch {}
		};
	}

	function handleWSMessage(msg: Record<string, unknown>) {
		const type = msg.type as string;

		if (type === "presence" && Array.isArray(msg.users)) {
			onlineUsers = (msg.users as { user_id: string; name: string }[]).filter(
				(u) => u.user_id !== myId,
			);
		}

		if (type === "message" && msg.message) {
			const chatMsg = msg.message as ChatMessage;
			messages = [...messages, chatMsg];
			if (shouldAutoScroll) {
				requestAnimationFrame(scrollToBottom);
			}
		}

		if (
			type === "typing" &&
			typeof msg.user_id === "string" &&
			msg.user_id !== myId
		) {
			const uid = msg.user_id as string;
			const name = (msg.name as string) || "?";
			if (typingUsers[uid]) clearTimeout(typingUsers[uid].timeout);
			const timeout = setTimeout(() => {
				const copy = { ...typingUsers };
				delete copy[uid];
				typingUsers = copy;
			}, 3000);
			typingUsers = { ...typingUsers, [uid]: { name, timeout } };
		}
	}

	function scrollToBottom() {
		if (messagesContainer) {
			messagesContainer.scrollTop = messagesContainer.scrollHeight;
		}
	}

	function handleScroll() {
		if (!messagesContainer) return;
		const { scrollTop, scrollHeight, clientHeight } = messagesContainer;
		shouldAutoScroll = scrollHeight - scrollTop - clientHeight < 60;

		if (scrollTop < 80 && hasMore && !loadingMore) {
			loadMore();
		}
	}

	async function loadMore() {
		if (messages.length === 0 || !hasMore) return;
		loadingMore = true;
		const oldHeight = messagesContainer?.scrollHeight ?? 0;
		try {
			const older = await teamsApi.listChatMessages(teamId, messages[0].id);
			if (older.length < 50) hasMore = false;
			if (older.length > 0) {
				messages = [...older, ...messages];
				requestAnimationFrame(() => {
					if (messagesContainer) {
						messagesContainer.scrollTop =
							messagesContainer.scrollHeight - oldHeight;
					}
				});
			}
		} catch {}
		loadingMore = false;
	}

	let typeSendTimer: ReturnType<typeof setTimeout> | null = null;

	function sendTyping() {
		if (typeSendTimer) return;
		if (ws && ws.readyState === WebSocket.OPEN) {
			ws.send(JSON.stringify({ type: "typing" }));
		}
		typeSendTimer = setTimeout(() => {
			typeSendTimer = null;
		}, 2000);
	}

	function sendMessage() {
		const content = newMessage.trim();
		if (!content || !ws || ws.readyState !== WebSocket.OPEN) return;
		sending = true;
		ws.send(JSON.stringify({ type: "message", content }));
		newMessage = "";
		sending = false;
		shouldAutoScroll = true;
		requestAnimationFrame(() => inputEl?.focus());
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === "Enter" && !e.shiftKey) {
			e.preventDefault();
			sendMessage();
		} else {
			sendTyping();
		}
	}

	function formatTime(dateStr: string) {
		const d = new Date(dateStr);
		const now = new Date();
		const isToday = d.toDateString() === now.toDateString();
		const yesterday = new Date(now);
		yesterday.setDate(yesterday.getDate() - 1);
		const isYesterday = d.toDateString() === yesterday.toDateString();

		const time = d.toLocaleTimeString([], {
			hour: "2-digit",
			minute: "2-digit",
		});
		if (isToday) return time;
		if (isYesterday) return `Yesterday ${time}`;
		return `${d.toLocaleDateString([], { month: "short", day: "numeric" })} ${time}`;
	}

	const AVATAR_COLORS = [
		"#ef4444",
		"#f97316",
		"#eab308",
		"#22c55e",
		"#06b6d4",
		"#3b82f6",
		"#8b5cf6",
		"#ec4899",
	];

	function getAvatarColor(name: string) {
		let hash = 0;
		for (let i = 0; i < name.length; i++)
			hash = ((hash << 5) - hash + name.charCodeAt(i)) | 0;
		return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length];
	}

	function shouldShowHeader(idx: number) {
		if (idx === 0) return true;
		const prev = messages[idx - 1];
		const curr = messages[idx];
		if (prev.user_id !== curr.user_id) return true;
		const diff =
			new Date(curr.created_at).getTime() - new Date(prev.created_at).getTime();
		return diff > 5 * 60 * 1000;
	}

	function shouldShowDate(idx: number) {
		if (idx === 0) return true;
		const prev = new Date(messages[idx - 1].created_at).toDateString();
		const curr = new Date(messages[idx].created_at).toDateString();
		return prev !== curr;
	}

	function formatDateSeparator(dateStr: string) {
		const d = new Date(dateStr);
		const now = new Date();
		if (d.toDateString() === now.toDateString()) return "Today";
		const yesterday = new Date(now);
		yesterday.setDate(yesterday.getDate() - 1);
		if (d.toDateString() === yesterday.toDateString()) return "Yesterday";
		return d.toLocaleDateString([], {
			weekday: "long",
			month: "long",
			day: "numeric",
		});
	}

	onMount(() => {
		const init = async () => {
			try {
				const [teamData, chatData] = await Promise.all([
					teamsApi.get(teamId),
					teamsApi.listChatMessages(teamId),
				]);
				team = teamData;
				messages = chatData;
				if (chatData.length < 50) hasMore = false;
			} catch {}
			loading = false;
			requestAnimationFrame(scrollToBottom);
			connectWS();
		};
		init();

		return () => {
			destroyed = true;
			if (typeSendTimer) clearTimeout(typeSendTimer);
			for (const t of Object.values(typingUsers)) clearTimeout(t.timeout);
			if (ws) {
				ws.onclose = null;
				ws.close();
			}
		};
	});
</script>

<svelte:head>
	<title>{team ? `Chat — ${team.name}` : "Team Chat"} — FPMB</title>
	<meta name="description" content="Real-time team chat in FPMB." />
</svelte:head>

<div class="h-[calc(100vh-5.5rem)] flex flex-col -m-6 lg:-m-8">
	<!-- Header -->
	<div
		class="shrink-0 bg-neutral-800/60 backdrop-blur-sm border-b border-neutral-700/60 px-5 py-3"
	>
		<div class="flex items-center justify-between max-w-4xl mx-auto">
			<div class="flex items-center gap-3">
				<a
					href="/team/{teamId}"
					class="p-1.5 rounded-lg text-neutral-400 hover:text-white hover:bg-neutral-700 transition-colors"
					title="Back to team"
				>
					<Icon icon="lucide:arrow-left" class="w-4 h-4" />
				</a>
				<div
					class="w-9 h-9 rounded-xl bg-blue-600/20 border border-blue-500/30 flex items-center justify-center"
				>
					<Icon
						icon="lucide:message-circle"
						class="w-4.5 h-4.5 text-blue-400"
					/>
				</div>
				<div>
					<h1 class="text-sm font-semibold text-white leading-tight">
						{team?.name ?? "Team"} Chat
					</h1>
					<p class="text-[11px] text-neutral-500 leading-tight mt-0.5">
						{#if onlineUsers.length > 0}
							{onlineUsers.length + 1} members online
						{:else}
							Just you
						{/if}
					</p>
				</div>
			</div>
			<div class="flex items-center gap-3">
				{#if onlineUsers.length > 0}
					<div class="flex -space-x-2">
						{#each onlineUsers.slice(0, 5) as user}
							<div
								class="w-7 h-7 rounded-full flex items-center justify-center text-[10px] font-bold text-white border-2 border-neutral-800 ring-1 ring-neutral-700 transition-transform hover:scale-110 hover:z-10"
								style="background-color: {getAvatarColor(user.name)}"
								title={user.name}
							>
								{user.name.charAt(0).toUpperCase()}
							</div>
						{/each}
						{#if onlineUsers.length > 5}
							<div
								class="w-7 h-7 rounded-full flex items-center justify-center text-[10px] font-medium text-neutral-300 bg-neutral-700 border-2 border-neutral-800"
							>
								+{onlineUsers.length - 5}
							</div>
						{/if}
					</div>
				{/if}
				<div
					class="flex items-center gap-1.5 px-2 py-1 rounded-md bg-neutral-800 border border-neutral-700"
				>
					<div
						class="w-1.5 h-1.5 rounded-full {wsConnected
							? 'bg-emerald-400 shadow-[0_0_6px_theme(colors.emerald.400)]'
							: 'bg-red-400 shadow-[0_0_6px_theme(colors.red.400)]'}"
					></div>
					<span
						class="text-[10px] font-medium {wsConnected
							? 'text-emerald-400'
							: 'text-red-400'}"
					>
						{wsConnected ? "Live" : "Offline"}
					</span>
				</div>
			</div>
		</div>
	</div>

	<!-- Messages area -->
	<div
		bind:this={messagesContainer}
		onscroll={handleScroll}
		class="flex-1 overflow-y-auto"
	>
		<div class="max-w-4xl mx-auto px-5 py-4">
			{#if loading}
				<div class="flex items-center justify-center h-64">
					<div class="flex flex-col items-center gap-3">
						<Icon
							icon="lucide:loader-2"
							class="w-6 h-6 text-blue-400 animate-spin"
						/>
						<span class="text-xs text-neutral-500">Loading messages…</span>
					</div>
				</div>
			{:else if messages.length === 0}
				<div class="flex flex-col items-center justify-center h-64 text-center">
					<div
						class="w-20 h-20 rounded-2xl bg-gradient-to-br from-blue-600/20 to-purple-600/20 border border-blue-500/20 flex items-center justify-center mb-5"
					>
						<Icon icon="lucide:message-circle" class="w-9 h-9 text-blue-400" />
					</div>
					<h3 class="text-lg font-semibold text-white mb-1.5">
						Start the conversation
					</h3>
					<p class="text-sm text-neutral-500 max-w-xs leading-relaxed">
						Messages are visible to all team members. Say hello!
					</p>
				</div>
			{:else}
				{#if loadingMore}
					<div class="flex justify-center py-4">
						<Icon
							icon="lucide:loader-2"
							class="w-4 h-4 text-neutral-500 animate-spin"
						/>
					</div>
				{/if}
				{#if !hasMore}
					<div class="flex items-center gap-3 py-4 mb-2">
						<div class="flex-1 h-px bg-neutral-800"></div>
						<span
							class="text-[10px] font-medium text-neutral-600 uppercase tracking-wider"
							>Beginning of conversation</span
						>
						<div class="flex-1 h-px bg-neutral-800"></div>
					</div>
				{/if}

				{#each messages as msg, idx}
					{@const isMe = msg.user_id === myId}
					{@const showHeader = shouldShowHeader(idx)}
					{@const showDate = shouldShowDate(idx)}

					{#if showDate}
						<div class="flex items-center gap-3 py-3 my-2">
							<div class="flex-1 h-px bg-neutral-800"></div>
							<span
								class="text-[10px] font-medium text-neutral-500 uppercase tracking-wider"
								>{formatDateSeparator(msg.created_at)}</span
							>
							<div class="flex-1 h-px bg-neutral-800"></div>
						</div>
					{/if}

					<div
						class="group relative {showHeader
							? 'mt-5'
							: 'mt-0.5'} rounded-lg hover:bg-white/[0.02] px-2 py-0.5 -mx-2 transition-colors"
					>
						{#if showHeader}
							<div class="flex items-center gap-2.5 mb-1">
								<div
									class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold text-white shrink-0 shadow-md"
									style="background-color: {getAvatarColor(msg.user_name)}"
								>
									{msg.user_name.charAt(0).toUpperCase()}
								</div>
								<span
									class="text-[13px] font-semibold {isMe
										? 'text-blue-300'
										: 'text-neutral-100'}">{isMe ? "You" : msg.user_name}</span
								>
								<span class="text-[11px] text-neutral-600"
									>{formatTime(msg.created_at)}</span
								>
							</div>
						{/if}
						<div class={showHeader ? "pl-[42px]" : "pl-[42px]"}>
							<p
								class="text-[13.5px] text-neutral-300 leading-[1.55] whitespace-pre-wrap wrap-break-word"
							>
								{msg.content}
							</p>
						</div>
						<span
							class="absolute right-2 top-1 text-[10px] text-neutral-600 opacity-0 group-hover:opacity-100 transition-opacity"
							>{formatTime(msg.created_at)}</span
						>
					</div>
				{/each}
			{/if}
		</div>
	</div>

	<!-- Typing indicator -->
	{#if typingNames.length > 0}
		<div class="shrink-0 px-5">
			<div
				class="max-w-4xl mx-auto py-1.5 flex items-center gap-2 text-xs text-neutral-500"
			>
				<span class="flex gap-[3px]">
					<span
						class="w-[5px] h-[5px] bg-blue-400/60 rounded-full animate-bounce"
						style="animation-delay: 0ms"
					></span>
					<span
						class="w-[5px] h-[5px] bg-blue-400/60 rounded-full animate-bounce"
						style="animation-delay: 150ms"
					></span>
					<span
						class="w-[5px] h-[5px] bg-blue-400/60 rounded-full animate-bounce"
						style="animation-delay: 300ms"
					></span>
				</span>
				{#if typingNames.length === 1}
					<span
						><strong class="text-neutral-400">{typingNames[0]}</strong> is typing…</span
					>
				{:else if typingNames.length === 2}
					<span
						><strong class="text-neutral-400">{typingNames[0]}</strong> and
						<strong class="text-neutral-400">{typingNames[1]}</strong> are typing…</span
					>
				{:else}
					<span
						><strong class="text-neutral-400"
							>{typingNames.length} people</strong
						> are typing…</span
					>
				{/if}
			</div>
		</div>
	{/if}

	<!-- Input -->
	<div
		class="shrink-0 border-t border-neutral-700/60 bg-neutral-800/40 backdrop-blur-sm px-5 py-3"
	>
		<div class="max-w-4xl mx-auto">
			<div class="flex items-end gap-2.5">
				<div class="flex-1 relative">
					<textarea
						bind:this={inputEl}
						bind:value={newMessage}
						onkeydown={handleKeydown}
						placeholder={wsConnected ? "Type a message…" : "Connecting…"}
						rows="1"
						class="w-full resize-none bg-neutral-800 border border-neutral-600/80 rounded-xl px-4 py-2.5 text-sm text-white placeholder-neutral-500 focus:outline-none focus:border-blue-500/60 focus:ring-1 focus:ring-blue-500/20 transition-all max-h-32"
						disabled={!wsConnected}
					></textarea>
				</div>
				<button
					onclick={sendMessage}
					disabled={!newMessage.trim() || !wsConnected || sending}
					class="p-2.5 rounded-xl transition-all duration-150 shrink-0 {newMessage.trim() &&
					wsConnected
						? 'bg-blue-600 text-white hover:bg-blue-500 shadow-md shadow-blue-600/20 hover:shadow-blue-500/30 active:scale-95'
						: 'bg-neutral-800 text-neutral-600 border border-neutral-700 cursor-not-allowed'}"
					title="Send (Enter)"
				>
					<Icon icon="lucide:send" class="w-5 h-5" />
				</button>
			</div>
		</div>
	</div>
</div>
