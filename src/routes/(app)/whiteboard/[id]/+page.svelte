<script lang="ts">
	import { onMount } from "svelte";
	import { page } from "$app/stores";
	import Icon from "@iconify/svelte";
	import { projects as projectsApi } from "$lib/api";
	import { getAccessToken } from "$lib/api/client";
	import { authStore } from "$lib/stores/auth.svelte";

	let boardId = $derived($page.params.id ?? "");
	let canvas: HTMLCanvasElement;
	let ctx: CanvasRenderingContext2D | null = null;
	let canvasContainer: HTMLDivElement;

	type DrawObject =
		| {
				type: "pen";
				points: { x: number; y: number }[];
				color: string;
				width: number;
		  }
		| {
				type: "rect";
				x: number;
				y: number;
				w: number;
				h: number;
				color: string;
				width: number;
		  }
		| {
				type: "circle";
				cx: number;
				cy: number;
				r: number;
				color: string;
				width: number;
		  }
		| {
				type: "text";
				x: number;
				y: number;
				text: string;
				color: string;
				fontSize: number;
		  };

	let objects = $state<DrawObject[]>([]);
	let undoStack = $state<DrawObject[][]>([]);
	let redoStack = $state<DrawObject[][]>([]);

	let currentTool = $state<
		"pen" | "rect" | "circle" | "eraser" | "text" | "select"
	>("pen");
	let strokeColor = $state("#ffffff");
	let lineWidth = $state(3);
	let fontSize = $state(20);
	let saving = $state(false);

	let isDrawing = $state(false);
	let startX = $state(0);
	let startY = $state(0);
	let currentPenPoints = $state<{ x: number; y: number }[]>([]);
	let previewShape = $state<DrawObject | null>(null);

	let showTextInput = $state(false);
	let textInputValue = $state("");
	let textX = $state(0);
	let textY = $state(0);

	let selectedIndex = $state<number | null>(null);
	let editingIndex = $state<number | null>(null);
	let editValue = $state("");
	let isDragging = $state(false);
	let dragOffsetX = $state(0);
	let dragOffsetY = $state(0);

	let saveTimer: ReturnType<typeof setTimeout> | null = null;

	let ws: WebSocket | null = null;
	let wsConnected = $state(false);
	let remoteUsers = $state<{ user_id: string; name: string }[]>([]);
	let remoteCursors = $state<
		Record<string, { x: number; y: number; name: string; color: string }>
	>({});
	let cursorSendTimer: ReturnType<typeof setTimeout> | null = null;
	let destroyed = false;
	const CURSOR_COLORS = [
		"#f87171",
		"#fb923c",
		"#a3e635",
		"#34d399",
		"#22d3ee",
		"#818cf8",
		"#e879f9",
		"#f472b6",
	];

	function getCursorColor(userId: string) {
		let hash = 0;
		for (let i = 0; i < userId.length; i++)
			hash = ((hash << 5) - hash + userId.charCodeAt(i)) | 0;
		return CURSOR_COLORS[Math.abs(hash) % CURSOR_COLORS.length];
	}

	function connectWS() {
		const token = getAccessToken();
		if (!token) return;
		const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
		const host = window.location.hostname;
		const port =
			window.location.port ||
			(window.location.protocol === "https:" ? "443" : "80");
		const userName = authStore.user?.name ?? "Anonymous";
		const url = `${proto}//${host}:${port}/ws/whiteboard/${boardId}?token=${encodeURIComponent(token)}&name=${encodeURIComponent(userName)}`;

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

		if (type === "users" || type === "join" || type === "leave") {
			if (Array.isArray(msg.users)) {
				const myId = authStore.user?.id;
				remoteUsers = (msg.users as { user_id: string; name: string }[]).filter(
					(u) => u.user_id !== myId,
				);
			}
			if (type === "leave" && typeof msg.user_id === "string") {
				const copy = { ...remoteCursors };
				delete copy[msg.user_id as string];
				remoteCursors = copy;
			}
		}

		if (type === "update" && Array.isArray(msg.objects)) {
			objects = msg.objects as DrawObject[];
			render();
		}

		if (type === "cursor" && typeof msg.user_id === "string") {
			remoteCursors = {
				...remoteCursors,
				[msg.user_id as string]: {
					x: msg.x as number,
					y: msg.y as number,
					name: (msg.name as string) || "?",
					color: getCursorColor(msg.user_id as string),
				},
			};
		}
	}

	function wsSend(data: Record<string, unknown>) {
		if (ws && ws.readyState === WebSocket.OPEN) {
			ws.send(JSON.stringify(data));
		}
	}

	function broadcastObjects() {
		wsSend({ type: "update", objects: $state.snapshot(objects) });
	}

	function sendCursor(x: number, y: number) {
		if (cursorSendTimer) return;
		cursorSendTimer = setTimeout(() => {
			wsSend({ type: "cursor", x, y });
			cursorSendTimer = null;
		}, 50);
	}

	function pushUndo() {
		undoStack = [...undoStack, $state.snapshot(objects) as DrawObject[]];
		redoStack = [];
		if (undoStack.length > 100) undoStack = undoStack.slice(-100);
	}

	function undo() {
		if (undoStack.length === 0) return;
		redoStack = [...redoStack, $state.snapshot(objects) as DrawObject[]];
		objects = undoStack[undoStack.length - 1];
		undoStack = undoStack.slice(0, -1);
		selectedIndex = null;
		editingIndex = null;
		render();
		scheduleSave();
		broadcastObjects();
	}

	function redo() {
		if (redoStack.length === 0) return;
		undoStack = [...undoStack, $state.snapshot(objects) as DrawObject[]];
		objects = redoStack[redoStack.length - 1];
		redoStack = redoStack.slice(0, -1);
		selectedIndex = null;
		editingIndex = null;
		render();
		scheduleSave();
		broadcastObjects();
	}

	let canUndo = $derived(undoStack.length > 0);
	let canRedo = $derived(redoStack.length > 0);

	function scheduleSave() {
		if (saveTimer) clearTimeout(saveTimer);
		saveTimer = setTimeout(() => {
			saving = true;
			const data = JSON.stringify(objects);
			projectsApi.saveWhiteboard(boardId, data).finally(() => {
				saving = false;
			});
		}, 1500);
	}

	function render() {
		if (!ctx || !canvas) return;
		ctx.fillStyle = "#171717";
		ctx.fillRect(0, 0, canvas.width, canvas.height);
		for (let i = 0; i < objects.length; i++) {
			drawObject(objects[i], i === selectedIndex);
		}
		if (previewShape) drawObject(previewShape, false);
		drawRemoteCursors();
	}

	function drawRemoteCursors() {
		if (!ctx) return;
		for (const [, cursor] of Object.entries(remoteCursors)) {
			ctx.save();
			ctx.fillStyle = cursor.color;
			ctx.beginPath();
			ctx.moveTo(cursor.x, cursor.y);
			ctx.lineTo(cursor.x, cursor.y + 14);
			ctx.lineTo(cursor.x + 5, cursor.y + 11);
			ctx.lineTo(cursor.x + 10, cursor.y + 11);
			ctx.closePath();
			ctx.fill();
			ctx.font = "11px sans-serif";
			ctx.fillStyle = cursor.color;
			const textWidth = ctx.measureText(cursor.name).width;
			ctx.fillStyle = cursor.color;
			const bx = cursor.x + 12;
			const by = cursor.y + 8;
			ctx.beginPath();
			ctx.roundRect(bx - 3, by - 10, textWidth + 6, 14, 3);
			ctx.fill();
			ctx.fillStyle = "#fff";
			ctx.fillText(cursor.name, bx, by);
			ctx.restore();
		}
	}

	function drawObject(obj: DrawObject, selected: boolean) {
		if (!ctx) return;
		ctx.lineCap = "round";
		ctx.lineJoin = "round";

		if (obj.type === "pen") {
			if (obj.points.length < 2) return;
			ctx.strokeStyle = obj.color;
			ctx.lineWidth = obj.width;
			ctx.beginPath();
			ctx.moveTo(obj.points[0].x, obj.points[0].y);
			for (let i = 1; i < obj.points.length; i++) {
				ctx.lineTo(obj.points[i].x, obj.points[i].y);
			}
			ctx.stroke();
		} else if (obj.type === "rect") {
			ctx.strokeStyle = obj.color;
			ctx.lineWidth = obj.width;
			ctx.beginPath();
			ctx.rect(obj.x, obj.y, obj.w, obj.h);
			ctx.stroke();
		} else if (obj.type === "circle") {
			ctx.strokeStyle = obj.color;
			ctx.lineWidth = obj.width;
			ctx.beginPath();
			ctx.arc(obj.cx, obj.cy, obj.r, 0, 2 * Math.PI);
			ctx.stroke();
		} else if (obj.type === "text") {
			ctx.font = `${obj.fontSize}px sans-serif`;
			ctx.fillStyle = obj.color;
			ctx.fillText(obj.text, obj.x, obj.y);
		}

		if (selected) {
			const bb = getBoundingBox(obj);
			if (bb) {
				ctx.save();
				ctx.strokeStyle = "#3b82f6";
				ctx.lineWidth = 1.5;
				ctx.setLineDash([4, 4]);
				ctx.strokeRect(bb.x - 6, bb.y - 6, bb.w + 12, bb.h + 12);
				ctx.setLineDash([]);
				ctx.restore();
			}
		}
	}

	function getBoundingBox(
		obj: DrawObject,
	): { x: number; y: number; w: number; h: number } | null {
		if (obj.type === "pen") {
			if (obj.points.length === 0) return null;
			let minX = Infinity,
				minY = Infinity,
				maxX = -Infinity,
				maxY = -Infinity;
			for (const p of obj.points) {
				if (p.x < minX) minX = p.x;
				if (p.y < minY) minY = p.y;
				if (p.x > maxX) maxX = p.x;
				if (p.y > maxY) maxY = p.y;
			}
			return { x: minX, y: minY, w: maxX - minX, h: maxY - minY };
		} else if (obj.type === "rect") {
			const x = obj.w >= 0 ? obj.x : obj.x + obj.w;
			const y = obj.h >= 0 ? obj.y : obj.y + obj.h;
			return { x, y, w: Math.abs(obj.w), h: Math.abs(obj.h) };
		} else if (obj.type === "circle") {
			return {
				x: obj.cx - obj.r,
				y: obj.cy - obj.r,
				w: obj.r * 2,
				h: obj.r * 2,
			};
		} else if (obj.type === "text") {
			if (!ctx) return null;
			ctx.font = `${obj.fontSize}px sans-serif`;
			const m = ctx.measureText(obj.text);
			return {
				x: obj.x,
				y: obj.y - obj.fontSize,
				w: m.width,
				h: obj.fontSize * 1.2,
			};
		}
		return null;
	}

	function hitTest(x: number, y: number, obj: DrawObject): boolean {
		const bb = getBoundingBox(obj);
		if (!bb) return false;
		const pad = 8;
		return (
			x >= bb.x - pad &&
			x <= bb.x + bb.w + pad &&
			y >= bb.y - pad &&
			y <= bb.y + bb.h + pad
		);
	}

	function resizeCanvas() {
		if (!canvas || !canvas.parentElement) return;
		const rect = canvas.parentElement.getBoundingClientRect();
		canvas.width = rect.width;
		canvas.height = rect.height;
		render();
	}

	onMount(() => {
		ctx = canvas.getContext("2d", { willReadFrequently: true });
		if (!ctx) return;
		resizeCanvas();
		window.addEventListener("resize", resizeCanvas);

		projectsApi
			.getWhiteboard(boardId)
			.then((wb) => {
				if (wb.data) {
					try {
						const parsed = JSON.parse(wb.data);
						if (Array.isArray(parsed)) {
							objects = parsed;
							render();
							return;
						}
					} catch {}
					const img = new Image();
					img.onload = () => {
						ctx!.drawImage(img, 0, 0);
					};
					img.src = wb.data;
				}
			})
			.catch(() => {});

		connectWS();

		function handleKeydown(e: KeyboardEvent) {
			if (showTextInput || editingIndex !== null) return;
			if ((e.ctrlKey || e.metaKey) && e.key === "z") {
				e.preventDefault();
				if (e.shiftKey) redo();
				else undo();
			}
			if ((e.ctrlKey || e.metaKey) && e.key === "y") {
				e.preventDefault();
				redo();
			}
			if (e.key === "Delete" || e.key === "Backspace") {
				if (selectedIndex !== null) {
					e.preventDefault();
					pushUndo();
					objects = objects.filter((_, i) => i !== selectedIndex);
					selectedIndex = null;
					render();
					scheduleSave();
					broadcastObjects();
				}
			}
			if (e.key === "Escape") {
				selectedIndex = null;
				render();
			}
		}

		window.addEventListener("keydown", handleKeydown);
		return () => {
			destroyed = true;
			window.removeEventListener("resize", resizeCanvas);
			window.removeEventListener("keydown", handleKeydown);
			if (saveTimer) clearTimeout(saveTimer);
			if (cursorSendTimer) clearTimeout(cursorSendTimer);
			if (ws) {
				ws.onclose = null;
				ws.close();
			}
		};
	});

	function getMousePos(e: MouseEvent) {
		const rect = canvas.getBoundingClientRect();
		return {
			x: (e.clientX - rect.left) * (canvas.width / rect.width),
			y: (e.clientY - rect.top) * (canvas.height / rect.height),
		};
	}

	function startPosition(e: MouseEvent) {
		const pos = getMousePos(e);

		if (currentTool === "text") {
			textX = pos.x;
			textY = pos.y;
			textInputValue = "";
			showTextInput = true;
			selectedIndex = null;
			return;
		}

		if (currentTool === "select") {
			for (let i = objects.length - 1; i >= 0; i--) {
				if (hitTest(pos.x, pos.y, objects[i])) {
					selectedIndex = i;
					isDragging = true;
					pushUndo();
					const bb = getBoundingBox(objects[i]);
					if (bb) {
						dragOffsetX = pos.x - bb.x;
						dragOffsetY = pos.y - bb.y;
					}
					render();
					return;
				}
			}
			selectedIndex = null;
			render();
			return;
		}

		if (currentTool === "eraser") {
			for (let i = objects.length - 1; i >= 0; i--) {
				if (hitTest(pos.x, pos.y, objects[i])) {
					pushUndo();
					objects = objects.filter((_, j) => j !== i);
					if (selectedIndex === i) selectedIndex = null;
					else if (selectedIndex !== null && selectedIndex > i) selectedIndex--;
					render();
					scheduleSave();
					broadcastObjects();
					return;
				}
			}
			return;
		}

		selectedIndex = null;
		isDrawing = true;
		startX = pos.x;
		startY = pos.y;

		if (currentTool === "pen") {
			currentPenPoints = [{ x: pos.x, y: pos.y }];
		}
	}

	function draw(e: MouseEvent) {
		const pos = getMousePos(e);
		sendCursor(pos.x, pos.y);

		if (!isDrawing) {
			if (isDragging && selectedIndex !== null) {
				moveObject(selectedIndex, pos.x - dragOffsetX, pos.y - dragOffsetY);
				render();
			}
			return;
		}

		if (currentTool === "pen") {
			currentPenPoints = [...currentPenPoints, { x: pos.x, y: pos.y }];
			previewShape = {
				type: "pen",
				points: currentPenPoints,
				color: strokeColor,
				width: lineWidth,
			};
			render();
		} else if (currentTool === "rect") {
			previewShape = {
				type: "rect",
				x: startX,
				y: startY,
				w: pos.x - startX,
				h: pos.y - startY,
				color: strokeColor,
				width: lineWidth,
			};
			render();
		} else if (currentTool === "circle") {
			const r = Math.sqrt(
				Math.pow(pos.x - startX, 2) + Math.pow(pos.y - startY, 2),
			);
			previewShape = {
				type: "circle",
				cx: startX,
				cy: startY,
				r,
				color: strokeColor,
				width: lineWidth,
			};
			render();
		}
	}

	function endPosition() {
		if (isDragging && selectedIndex !== null) {
			isDragging = false;
			scheduleSave();
			broadcastObjects();
			return;
		}

		if (!isDrawing) return;
		isDrawing = false;

		if (previewShape) {
			if (previewShape.type === "pen" && previewShape.points.length < 2) {
				previewShape = null;
				return;
			}
			pushUndo();
			objects = [...objects, previewShape];
			previewShape = null;
			render();
			scheduleSave();
			broadcastObjects();
		}
	}

	function moveObject(idx: number, newX: number, newY: number) {
		const obj = objects[idx];
		if (!obj) return;
		const bb = getBoundingBox(obj);
		if (!bb) return;
		const dx = newX - bb.x;
		const dy = newY - bb.y;

		if (obj.type === "pen") {
			obj.points = obj.points.map((p) => ({ x: p.x + dx, y: p.y + dy }));
		} else if (obj.type === "rect") {
			obj.x += dx;
			obj.y += dy;
		} else if (obj.type === "circle") {
			obj.cx += dx;
			obj.cy += dy;
		} else if (obj.type === "text") {
			obj.x += dx;
			obj.y += dy;
		}
		objects = [...objects];
	}

	function commitText() {
		if (!textInputValue.trim()) {
			showTextInput = false;
			textInputValue = "";
			return;
		}
		pushUndo();
		objects = [
			...objects,
			{
				type: "text",
				x: textX,
				y: textY,
				text: textInputValue.trim(),
				color: strokeColor,
				fontSize,
			},
		];
		showTextInput = false;
		textInputValue = "";
		render();
		scheduleSave();
		broadcastObjects();
	}

	function handleTextKeydown(e: KeyboardEvent) {
		if (e.key === "Enter") {
			e.preventDefault();
			commitText();
		} else if (e.key === "Escape") {
			showTextInput = false;
			textInputValue = "";
		}
	}

	function startEdit() {
		if (selectedIndex === null) return;
		const obj = objects[selectedIndex];
		if (obj.type === "text") {
			editingIndex = selectedIndex;
			editValue = obj.text;
		}
	}

	function commitEdit() {
		if (editingIndex === null) return;
		const obj = objects[editingIndex];
		if (obj.type === "text") {
			if (editValue.trim()) {
				pushUndo();
				obj.text = editValue.trim();
				objects = [...objects];
			} else {
				pushUndo();
				objects = objects.filter((_, i) => i !== editingIndex);
				selectedIndex = null;
			}
		}
		editingIndex = null;
		editValue = "";
		render();
		scheduleSave();
		broadcastObjects();
	}

	function handleEditKeydown(e: KeyboardEvent) {
		if (e.key === "Enter") {
			e.preventDefault();
			commitEdit();
		} else if (e.key === "Escape") {
			editingIndex = null;
			editValue = "";
		}
	}

	function handleDblClick(e: MouseEvent) {
		const pos = getMousePos(e);
		for (let i = objects.length - 1; i >= 0; i--) {
			if (hitTest(pos.x, pos.y, objects[i])) {
				selectedIndex = i;
				const obj = objects[i];
				if (obj.type === "text") {
					editingIndex = i;
					editValue = obj.text;
				} else if (
					obj.type === "rect" ||
					obj.type === "circle" ||
					obj.type === "pen"
				) {
					editingIndex = i;
					editValue = obj.color;
				}
				render();
				return;
			}
		}
	}

	function updateObjectColor(color: string) {
		if (editingIndex === null) return;
		const obj = objects[editingIndex];
		if (obj.type !== "text") {
			pushUndo();
			obj.color = color;
			objects = [...objects];
			render();
			scheduleSave();
			broadcastObjects();
		}
	}

	function clearCanvas() {
		if (objects.length === 0) return;
		pushUndo();
		objects = [];
		selectedIndex = null;
		editingIndex = null;
		render();
		scheduleSave();
		broadcastObjects();
	}

	function deleteSelected() {
		if (selectedIndex === null) return;
		pushUndo();
		objects = objects.filter((_, i) => i !== selectedIndex);
		selectedIndex = null;
		editingIndex = null;
		render();
		scheduleSave();
		broadcastObjects();
	}

	function exportPng() {
		if (!canvas) return;
		const tempCanvas = document.createElement("canvas");
		tempCanvas.width = canvas.width;
		tempCanvas.height = canvas.height;
		const tempCtx = tempCanvas.getContext("2d");
		if (!tempCtx) return;

		tempCtx.fillStyle = "#171717";
		tempCtx.fillRect(0, 0, tempCanvas.width, tempCanvas.height);

		const savedSel = selectedIndex;
		selectedIndex = null;

		for (const obj of objects) {
			tempCtx.lineCap = "round";
			tempCtx.lineJoin = "round";
			if (obj.type === "pen") {
				if (obj.points.length < 2) continue;
				tempCtx.strokeStyle = obj.color;
				tempCtx.lineWidth = obj.width;
				tempCtx.beginPath();
				tempCtx.moveTo(obj.points[0].x, obj.points[0].y);
				for (let i = 1; i < obj.points.length; i++)
					tempCtx.lineTo(obj.points[i].x, obj.points[i].y);
				tempCtx.stroke();
			} else if (obj.type === "rect") {
				tempCtx.strokeStyle = obj.color;
				tempCtx.lineWidth = obj.width;
				tempCtx.beginPath();
				tempCtx.rect(obj.x, obj.y, obj.w, obj.h);
				tempCtx.stroke();
			} else if (obj.type === "circle") {
				tempCtx.strokeStyle = obj.color;
				tempCtx.lineWidth = obj.width;
				tempCtx.beginPath();
				tempCtx.arc(obj.cx, obj.cy, obj.r, 0, 2 * Math.PI);
				tempCtx.stroke();
			} else if (obj.type === "text") {
				tempCtx.font = `${obj.fontSize}px sans-serif`;
				tempCtx.fillStyle = obj.color;
				tempCtx.fillText(obj.text, obj.x, obj.y);
			}
		}

		selectedIndex = savedSel;

		const link = document.createElement("a");
		link.download = `whiteboard-${boardId}.png`;
		link.href = tempCanvas.toDataURL("image/png");
		link.click();
	}

	let selectedObj = $derived(
		selectedIndex !== null ? objects[selectedIndex] : null,
	);
	let otherUserCount = $derived(remoteUsers.length);
</script>

<svelte:head>
	<title>Whiteboard — FPMB</title>
	<meta
		name="description"
		content="Freehand canvas whiteboard for your project — draw, annotate, and brainstorm in FPMB."
	/>
</svelte:head>

<div class="absolute inset-0 z-10 bg-neutral-900 flex flex-col overflow-hidden">
	<div class="absolute top-4 left-4 z-20">
		<a
			href="/board/{boardId}"
			class="flex items-center space-x-2 bg-neutral-800 hover:bg-neutral-700 text-neutral-300 hover:text-white px-3 py-2 rounded-lg shadow-md border border-neutral-700 transition-colors"
		>
			<Icon icon="lucide:arrow-left" class="w-4 h-4" />
			<span class="text-sm font-medium">Back to Board</span>
		</a>
	</div>

	<div
		class="absolute top-4 left-1/2 transform -translate-x-1/2 bg-neutral-800 p-1.5 rounded-lg shadow-lg border border-neutral-600 flex items-center gap-1 z-20"
	>
		<button
			class="p-2 rounded transition-colors {currentTool === 'select'
				? 'bg-blue-600 text-white'
				: 'text-neutral-400 hover:bg-neutral-700 hover:text-white'}"
			onclick={() => {
				currentTool = "select";
				editingIndex = null;
			}}
			title="Select & Move"
		>
			<Icon icon="lucide:mouse-pointer-2" class="w-5 h-5" />
		</button>

		<div class="w-px h-7 bg-neutral-700"></div>

		<button
			class="p-2 rounded transition-colors {currentTool === 'pen'
				? 'bg-blue-600 text-white'
				: 'text-neutral-400 hover:bg-neutral-700 hover:text-white'}"
			onclick={() => (currentTool = "pen")}
			title="Pen"
		>
			<Icon icon="lucide:pen-tool" class="w-5 h-5" />
		</button>
		<button
			class="p-2 rounded transition-colors {currentTool === 'rect'
				? 'bg-blue-600 text-white'
				: 'text-neutral-400 hover:bg-neutral-700 hover:text-white'}"
			onclick={() => (currentTool = "rect")}
			title="Rectangle"
		>
			<Icon icon="lucide:square" class="w-5 h-5" />
		</button>
		<button
			class="p-2 rounded transition-colors {currentTool === 'circle'
				? 'bg-blue-600 text-white'
				: 'text-neutral-400 hover:bg-neutral-700 hover:text-white'}"
			onclick={() => (currentTool = "circle")}
			title="Circle"
		>
			<Icon icon="lucide:circle" class="w-5 h-5" />
		</button>
		<button
			class="p-2 rounded transition-colors {currentTool === 'text'
				? 'bg-blue-600 text-white'
				: 'text-neutral-400 hover:bg-neutral-700 hover:text-white'}"
			onclick={() => (currentTool = "text")}
			title="Text"
		>
			<Icon icon="lucide:type" class="w-5 h-5" />
		</button>

		<div class="w-px h-7 bg-neutral-700"></div>

		<button
			class="p-2 rounded transition-colors {currentTool === 'eraser'
				? 'bg-red-600 text-white'
				: 'text-neutral-400 hover:bg-neutral-700 hover:text-white'}"
			onclick={() => (currentTool = "eraser")}
			title="Eraser (click to delete)"
		>
			<Icon icon="lucide:eraser" class="w-5 h-5" />
		</button>

		<div class="w-px h-7 bg-neutral-700"></div>

		<input
			type="color"
			bind:value={strokeColor}
			class="w-7 h-7 rounded border-0 bg-transparent p-0 cursor-pointer"
			title="Color"
		/>

		{#if currentTool === "text"}
			<div class="w-px h-7 bg-neutral-700"></div>
			<input
				type="number"
				bind:value={fontSize}
				min="8"
				max="120"
				title="Font Size"
				class="w-14 px-1 py-1 text-xs bg-neutral-700 border border-neutral-600 rounded text-white text-center focus:outline-none focus:border-blue-500"
			/>
		{/if}

		<div class="w-px h-7 bg-neutral-700"></div>

		<button
			disabled={!canUndo}
			class="p-2 rounded transition-colors {canUndo
				? 'text-neutral-400 hover:bg-neutral-700 hover:text-white'
				: 'text-neutral-600 cursor-not-allowed'}"
			onclick={undo}
			title="Undo (Ctrl+Z)"
		>
			<Icon icon="lucide:undo-2" class="w-5 h-5" />
		</button>
		<button
			disabled={!canRedo}
			class="p-2 rounded transition-colors {canRedo
				? 'text-neutral-400 hover:bg-neutral-700 hover:text-white'
				: 'text-neutral-600 cursor-not-allowed'}"
			onclick={redo}
			title="Redo (Ctrl+Shift+Z)"
		>
			<Icon icon="lucide:redo-2" class="w-5 h-5" />
		</button>

		<div class="w-px h-7 bg-neutral-700"></div>

		<button
			class="p-2 text-neutral-400 hover:text-blue-400 hover:bg-neutral-700 rounded transition-colors"
			onclick={exportPng}
			title="Export as PNG"
		>
			<Icon icon="lucide:download" class="w-5 h-5" />
		</button>

		<button
			class="p-2 text-neutral-400 hover:text-red-400 hover:bg-neutral-700 rounded transition-colors"
			onclick={clearCanvas}
			title="Clear Canvas"
		>
			<Icon icon="lucide:trash-2" class="w-5 h-5" />
		</button>
	</div>

	<div class="absolute top-4 right-4 z-20 flex items-center gap-2">
		{#if otherUserCount > 0}
			<div
				class="bg-neutral-800 px-3 py-2 rounded-lg shadow-md border border-neutral-700 flex items-center gap-2"
			>
				<div class="flex -space-x-1.5">
					{#each remoteUsers.slice(0, 5) as user}
						<div
							class="w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold text-white border-2 border-neutral-800"
							style="background-color: {getCursorColor(user.user_id)}"
							title={user.name}
						>
							{user.name.charAt(0).toUpperCase()}
						</div>
					{/each}
				</div>
				<span class="text-xs text-neutral-400">{otherUserCount} online</span>
			</div>
		{/if}
		<div
			class="bg-neutral-800 px-4 py-2 rounded-lg shadow-md border border-neutral-700 pointer-events-none flex items-center gap-2"
		>
			{#if saving}
				<Icon
					icon="lucide:loader-2"
					class="w-3.5 h-3.5 text-neutral-400 animate-spin"
				/>
			{/if}
			<div
				class="w-2 h-2 rounded-full {wsConnected
					? 'bg-green-500'
					: 'bg-red-500'}"
			></div>
			<h2 class="text-white font-semibold text-sm">Whiteboard</h2>
		</div>
	</div>

	{#if selectedObj && editingIndex !== null}
		<div
			class="absolute bottom-6 left-1/2 -translate-x-1/2 z-20 bg-neutral-800 border border-neutral-600 rounded-lg shadow-xl p-3 flex items-center gap-3"
		>
			{#if selectedObj.type === "text"}
				<label for="edit-text-input" class="text-xs text-neutral-400"
					>Text:</label
				>
				<input
					id="edit-text-input"
					type="text"
					bind:value={editValue}
					onkeydown={handleEditKeydown}
					onblur={commitEdit}
					class="px-2 py-1 bg-neutral-700 border border-neutral-600 rounded text-sm text-white focus:outline-none focus:border-blue-500 min-w-[200px]"
				/>
			{:else}
				<label for="edit-color-input" class="text-xs text-neutral-400"
					>Color:</label
				>
				<input
					id="edit-color-input"
					type="color"
					bind:value={editValue}
					oninput={(e) =>
						updateObjectColor((e.currentTarget as HTMLInputElement).value)}
					class="w-8 h-8 rounded border-0 bg-transparent p-0 cursor-pointer"
				/>
			{/if}
			<button
				onclick={deleteSelected}
				class="p-1.5 text-neutral-400 hover:text-red-400 hover:bg-red-900/20 rounded transition-colors"
				title="Delete"
			>
				<Icon icon="lucide:trash-2" class="w-4 h-4" />
			</button>
			<button
				onclick={() => {
					editingIndex = null;
					selectedIndex = null;
					render();
				}}
				class="p-1.5 text-neutral-400 hover:text-white hover:bg-neutral-700 rounded transition-colors"
				title="Done"
			>
				<Icon icon="lucide:check" class="w-4 h-4" />
			</button>
		</div>
	{:else if selectedObj}
		<div
			class="absolute bottom-6 left-1/2 -translate-x-1/2 z-20 bg-neutral-800 border border-neutral-600 rounded-lg shadow-xl p-2 flex items-center gap-2"
		>
			<button
				onclick={startEdit}
				class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-neutral-300 hover:text-white hover:bg-neutral-700 rounded transition-colors"
			>
				<Icon icon="lucide:pencil" class="w-3.5 h-3.5" /> Edit
			</button>
			<button
				onclick={deleteSelected}
				class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-neutral-300 hover:text-red-400 hover:bg-red-900/20 rounded transition-colors"
			>
				<Icon icon="lucide:trash-2" class="w-3.5 h-3.5" /> Delete
			</button>
		</div>
	{/if}

	<div
		bind:this={canvasContainer}
		class="flex-1 w-full relative touch-none bg-neutral-900"
	>
		<canvas
			bind:this={canvas}
			onmousedown={startPosition}
			onmouseup={endPosition}
			onmousemove={draw}
			onmouseout={endPosition}
			ondblclick={handleDblClick}
			class="absolute inset-0 w-full h-full touch-none block {currentTool ===
			'text'
				? 'cursor-text'
				: currentTool === 'select'
					? 'cursor-default'
					: currentTool === 'eraser'
						? 'cursor-pointer'
						: 'cursor-crosshair'}"
		></canvas>
		{#if showTextInput}
			{@const rect = canvasContainer
				? canvasContainer.getBoundingClientRect()
				: { width: 0, height: 0 }}
			{@const scaleX = canvas ? canvas.width / (rect.width || 1) : 1}
			{@const scaleY = canvas ? canvas.height / (rect.height || 1) : 1}
			<input
				type="text"
				bind:value={textInputValue}
				onkeydown={handleTextKeydown}
				onblur={commitText}
				autofocus
				style="position:absolute; left:{textX / scaleX}px; top:{(textY -
					fontSize) /
					scaleY}px; font-size:{fontSize}px; font-family:sans-serif; color:{strokeColor}; background:transparent; border:1px dashed #555; outline:none; min-width:100px; padding:2px 4px; z-index:30;"
			/>
		{/if}
	</div>
</div>
