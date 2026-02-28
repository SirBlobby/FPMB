<script lang="ts">
	import { page } from "$app/stores";
	import { onMount } from "svelte";
	import { teams as teamsApi } from "$lib/api";
	import type { Team, TeamMember } from "$lib/types/api";
	import { RoleFlag } from "$lib/types/roles";

	let teamId = $derived($page.params.id ?? "");

	let team = $state<Team | null>(null);
	let members = $state<TeamMember[]>([]);
	let teamName = $state("");
	let inviteEmail = $state("");
	let inviteRole = $state(RoleFlag.Editor);
	let saving = $state(false);
	let error = $state("");

	let avatarFile = $state<File | null>(null);
	let bannerFile = $state<File | null>(null);
	let avatarUploading = $state(false);
	let bannerUploading = $state(false);
	let avatarPreview = $state("");
	let bannerPreview = $state("");

	onMount(async () => {
		const [teamData, memberData] = await Promise.all([
			teamsApi.get(teamId),
			teamsApi.listMembers(teamId),
		]);
		team = teamData;
		members = memberData;
		teamName = teamData.name;
		avatarPreview = teamData.avatar_url ?? "";
		bannerPreview = teamData.banner_url ?? "";
	});

	async function saveGeneral(e: Event) {
		e.preventDefault();
		saving = true;
		try {
			const updated = await teamsApi.update(teamId, { name: teamName });
			team = updated;
		} finally {
			saving = false;
		}
	}

	function handleAvatarChange(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		avatarFile = file;
		avatarPreview = URL.createObjectURL(file);
	}

	function handleBannerChange(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		bannerFile = file;
		bannerPreview = URL.createObjectURL(file);
	}

	async function uploadAvatar() {
		if (!avatarFile) return;
		avatarUploading = true;
		try {
			const updated = await teamsApi.uploadAvatar(teamId, avatarFile);
			team = updated;
			avatarFile = null;
		} finally {
			avatarUploading = false;
		}
	}

	async function uploadBanner() {
		if (!bannerFile) return;
		bannerUploading = true;
		try {
			const updated = await teamsApi.uploadBanner(teamId, bannerFile);
			team = updated;
			bannerFile = null;
		} finally {
			bannerUploading = false;
		}
	}

	async function handleInvite(e: Event) {
		e.preventDefault();
		error = "";
		if (!inviteEmail) return;
		try {
			const res = await teamsApi.invite(teamId, inviteEmail, inviteRole);
			members = [...members, res.member];
			inviteEmail = "";
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : "Failed to invite";
		}
	}

	async function removeMember(userId: string) {
		await teamsApi.removeMember(teamId, userId);
		members = members.filter((m) => m.user_id !== userId);
	}
</script>

<svelte:head>
	<title
		>{teamName ? `${teamName} Settings — FPMB` : "Team Settings — FPMB"}</title
	>
	<meta
		name="description"
		content="Manage team name, avatar, banner, and member roles in FPMB."
	/>
</svelte:head>

<div class="max-w-5xl mx-auto space-y-10">
	<div class="flex justify-between items-end">
		<div>
			<h1 class="text-3xl font-bold text-white tracking-tight mb-2">
				Team Settings
			</h1>
			<p class="text-neutral-400">Manage your team members and roles.</p>
		</div>
	</div>

	<!-- General Section -->
	<section
		class="bg-neutral-800 rounded-lg shadow-sm border border-neutral-700 overflow-hidden"
	>
		<div class="p-6 border-b border-neutral-700">
			<h2 class="text-xl font-semibold text-white mb-1">General</h2>
			<p class="text-sm text-neutral-400">
				Update your team's name and description.
			</p>
		</div>
		<form onsubmit={saveGeneral} class="p-6 space-y-4">
			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<div>
					<label
						for="team-name"
						class="block text-sm font-medium text-neutral-300 mb-1"
						>Team Name</label
					>
					<input
						id="team-name"
						type="text"
						bind:value={teamName}
						class="block w-full px-3 py-2 border border-neutral-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
					/>
				</div>
			</div>
			<div class="flex justify-end pt-2">
				<button
					type="submit"
					disabled={saving}
					class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md shadow-sm border border-transparent transition-colors text-sm disabled:opacity-50"
				>
					{saving ? "Saving..." : "Save Changes"}
				</button>
			</div>
		</form>
	</section>

	<!-- Avatar Section -->
	<section
		class="bg-neutral-800 rounded-lg shadow-sm border border-neutral-700 overflow-hidden"
	>
		<div class="p-6 border-b border-neutral-700">
			<h2 class="text-xl font-semibold text-white mb-1">Team Avatar</h2>
			<p class="text-sm text-neutral-400">
				Upload a square image to represent your team.
			</p>
		</div>
		<div class="p-6 flex items-start gap-6">
			<div
				class="w-20 h-20 rounded-xl border border-neutral-600 overflow-hidden shrink-0 bg-neutral-700 flex items-center justify-center"
			>
				{#if avatarPreview}
					<img
						src={avatarPreview}
						alt="Team avatar"
						class="w-full h-full object-cover"
					/>
				{:else}
					<span class="text-3xl font-bold text-white"
						>{team?.name.charAt(0) ?? ""}</span
					>
				{/if}
			</div>
			<div class="flex flex-col gap-3">
				<label
					for="team-avatar"
					class="block text-sm font-medium text-neutral-300"
				>
					Image file <span class="text-neutral-500 font-normal"
						>(jpg, png, gif, webp)</span
					>
				</label>
				<input
					id="team-avatar"
					type="file"
					accept=".jpg,.jpeg,.png,.gif,.webp"
					onchange={handleAvatarChange}
					class="block text-sm text-neutral-400 file:mr-3 file:py-1.5 file:px-4 file:rounded file:border-0 file:text-sm file:font-medium file:bg-neutral-700 file:text-white hover:file:bg-neutral-600 cursor-pointer"
				/>
				<button
					type="button"
					onclick={uploadAvatar}
					disabled={!avatarFile || avatarUploading}
					class="self-start bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-5 rounded-md text-sm disabled:opacity-50 transition-colors"
				>
					{avatarUploading ? "Uploading..." : "Upload Avatar"}
				</button>
			</div>
		</div>
	</section>

	<!-- Banner Section -->
	<section
		class="bg-neutral-800 rounded-lg shadow-sm border border-neutral-700 overflow-hidden"
	>
		<div class="p-6 border-b border-neutral-700">
			<h2 class="text-xl font-semibold text-white mb-1">Team Banner</h2>
			<p class="text-sm text-neutral-400">
				Upload a wide image shown at the top of your team page.
			</p>
		</div>
		<div class="p-6 flex flex-col gap-4">
			<div
				class="w-full h-28 rounded-lg border border-neutral-600 overflow-hidden bg-neutral-700 flex items-center justify-center"
			>
				{#if bannerPreview}
					<img
						src={bannerPreview}
						alt="Team banner"
						class="w-full h-full object-cover"
					/>
				{:else}
					<span class="text-sm text-neutral-500">No banner set</span>
				{/if}
			</div>
			<div class="flex flex-col gap-3">
				<label
					for="team-banner"
					class="block text-sm font-medium text-neutral-300"
				>
					Image file <span class="text-neutral-500 font-normal"
						>(jpg, png, gif, webp)</span
					>
				</label>
				<input
					id="team-banner"
					type="file"
					accept=".jpg,.jpeg,.png,.gif,.webp"
					onchange={handleBannerChange}
					class="block text-sm text-neutral-400 file:mr-3 file:py-1.5 file:px-4 file:rounded file:border-0 file:text-sm file:font-medium file:bg-neutral-700 file:text-white hover:file:bg-neutral-600 cursor-pointer"
				/>
				<button
					type="button"
					onclick={uploadBanner}
					disabled={!bannerFile || bannerUploading}
					class="self-start bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-5 rounded-md text-sm disabled:opacity-50 transition-colors"
				>
					{bannerUploading ? "Uploading..." : "Upload Banner"}
				</button>
			</div>
		</div>
	</section>

	<!-- Members Section -->
	<section
		class="bg-neutral-800 rounded-lg shadow-sm border border-neutral-700 overflow-hidden"
	>
		<div
			class="p-6 border-b border-neutral-700 flex flex-col md:flex-row md:items-center justify-between gap-4"
		>
			<div>
				<h2 class="text-xl font-semibold text-white mb-1">Members</h2>
				<p class="text-sm text-neutral-400">
					Invite new members and manage roles.
				</p>
			</div>

			<form onsubmit={handleInvite} class="flex space-x-2 w-full md:w-auto">
				<input
					type="email"
					placeholder="name@example.com"
					bind:value={inviteEmail}
					required
					class="block w-full min-w-48 px-3 py-2 border border-neutral-600 rounded-md shadow-sm placeholder-neutral-500 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-neutral-700 text-white"
				/>
				<select
					bind:value={inviteRole}
					class="px-3 py-2 border border-neutral-600 rounded-md bg-neutral-700 text-white text-sm focus:ring-blue-500 focus:border-blue-500"
				>
					<option value={RoleFlag.Viewer}>Viewer</option>
					<option value={RoleFlag.Editor}>Editor</option>
					<option value={RoleFlag.Admin}>Admin</option>
				</select>
				<button
					type="submit"
					class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md shadow-sm border border-transparent transition-colors text-sm whitespace-nowrap"
				>
					Invite
				</button>
			</form>
		</div>

		{#if error}
			<div class="px-6 py-3 text-red-400 text-sm border-b border-neutral-700">
				{error}
			</div>
		{/if}

		<div class="overflow-x-auto">
			<table class="w-full text-left border-collapse">
				<thead>
					<tr class="bg-neutral-850 border-b border-neutral-700">
						<th
							class="px-6 py-4 text-xs font-semibold text-neutral-400 uppercase tracking-wider"
							>Member</th
						>
						<th
							class="px-6 py-4 text-xs font-semibold text-neutral-400 uppercase tracking-wider"
							>Role</th
						>
						<th
							class="px-6 py-4 text-xs font-semibold text-neutral-400 uppercase tracking-wider text-right"
							>Actions</th
						>
					</tr>
				</thead>
				<tbody class="divide-y divide-neutral-700">
					{#each members as member (member.user_id)}
						<tr class="hover:bg-neutral-750 transition-colors">
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="flex items-center">
									<div
										class="h-8 w-8 rounded-full bg-blue-900 text-blue-300 flex items-center justify-center font-bold text-xs shadow-inner"
									>
										{member.name.charAt(0).toUpperCase()}
									</div>
									<div class="ml-4">
										<div class="text-sm font-medium text-white">
											{member.name}
										</div>
										<div class="text-xs text-neutral-500">{member.email}</div>
									</div>
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<span
									class="px-2.5 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-neutral-700 text-neutral-300 border border-neutral-600"
								>
									{member.role_name}
								</span>
							</td>
							<td
								class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium"
							>
								{#if member.role_flags < RoleFlag.Owner}
									<button
										onclick={() => removeMember(member.user_id)}
										class="text-red-500 hover:text-red-400">Remove</button
									>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</section>
</div>
