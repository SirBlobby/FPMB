<script lang="ts">
	let {
		isOpen = $bindable(false),
		title,
		children,
		maxWidth = "max-w-2xl",
		onClose = () => {},
	} = $props();

	function close() {
		isOpen = false;
		if (onClose) onClose();
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-neutral-900/80 backdrop-blur-sm overflow-y-auto"
	>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="fixed inset-0" onclick={close}></div>

		<div
			class="relative bg-neutral-800 rounded-lg shadow-xl border border-neutral-700 w-full {maxWidth} max-h-[90vh] flex flex-col"
		>
			<div
				class="flex items-center justify-between p-4 border-b border-neutral-700 shrink-0"
			>
				<h2 class="text-xl font-semibold text-white">{title}</h2>
				<button
					onclick={close}
					aria-label="Close"
					class="text-neutral-400 hover:text-white transition-colors p-1 rounded-md hover:bg-neutral-700"
				>
					<svg
						class="w-5 h-5"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
						><path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						></path></svg
					>
				</button>
			</div>

			<div class="p-6 overflow-y-auto flex-1">
				{@render children()}
			</div>
		</div>
	</div>
{/if}
