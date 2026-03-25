<script lang="ts">
	import { Button } from "$components/ui/button";
	import type { TeamMembershipAttributes } from "$lib/api";
	import { useOrganizationSettingsViewController } from "../organizationSettingsViewController.svelte";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";

	const view = useOrganizationSettingsViewController();

	let teamName = $state("");
	let selectedAddUserId = $state("");
	let selectedAddRole = $state<TeamMembershipAttributes["role"]>("member");

	const onCreateTeam = () => {
		view.createTeam(teamName);
		teamName = "";
	};

	const onRenameTeam = (teamId: string, currentName: string) => {
		const nextName = prompt("Rename team", currentName)?.trim();
		if (!nextName || nextName === currentName) return;
		view.renameTeam(teamId, nextName);
	};

	const onArchiveTeam = (teamId: string, teamName: string) => {
		if (!confirm(`Archive team '${teamName}'?`)) return;
		view.archiveTeam(teamId);
	};

	const onAddMembership = () => {
		if (!selectedAddUserId) return;
		view.addMembership(selectedAddUserId, selectedAddRole);
		selectedAddUserId = "";
		selectedAddRole = "member";
	};

	const teamsErr = $derived(view.sectionErrors.get("teams"));
	const membershipsErr = $derived(view.sectionErrors.get("memberships"));
</script>

<div class="grid grid-cols-1 gap-3 xl:grid-cols-2">
	<section class="rounded border p-3">
		<div class="mb-2 flex items-center justify-between gap-2">
			<h3 class="text-lg font-semibold">Teams</h3>
		</div>

		{#if teamsErr}
		<InlineAlert
			error={teamsErr}
			onDismiss={() => view.dismissSectionErrorAlert("teams")}
		/>
		{/if}

		{#if view.teamsQuery.isLoading}
			<span>Loading teams...</span>
		{:else if view.teamsQuery.isError}
			<div class="mb-2 rounded border border-danger/40 bg-danger/10 p-3">
				<span class="text-danger">Failed to load teams.</span>
			</div>
		{:else}
			<div class="flex flex-col gap-2">
				{#each view.teams as team (team.id)}
					<div class="rounded border p-2" class:bg-surface-100={team.id === view.selectedTeamId}>
						<div class="flex items-center justify-between gap-2">
							<button
								type="button"
								class="text-left font-medium"
								onclick={() => view.setSelectedTeamId(team.id)}
							>
								{team.attributes.name}
							</button>
							{#if view.isOrgAdmin}
								<div class="flex gap-2">
									<Button onclick={() => onRenameTeam(team.id, team.attributes.name)}>Rename</Button>
									<Button onclick={() => onArchiveTeam(team.id, team.attributes.name)}>Archive</Button>
								</div>
							{/if}
						</div>
					</div>
				{:else}
					<span class="text-surface-content/80">No teams found.</span>
				{/each}
			</div>
		{/if}

		{#if view.isOrgAdmin}
			<div class="mt-3 flex gap-2">
				<input
					class="w-full rounded border bg-surface-200 p-2"
					placeholder="New team name"
					bind:value={teamName}
				/>
				<Button disabled={!teamName.trim()} onclick={onCreateTeam}>Create Team</Button>
			</div>
		{/if}
	</section>

	<section class="rounded border p-3">
		<div class="mb-2 flex items-center justify-between gap-2">
			<h3 class="text-lg font-semibold">Team Memberships</h3>
			{#if view.selectedTeam}
				<span class="text-sm text-surface-content/80">{view.selectedTeam.attributes.name}</span>
			{/if}
		</div>

		{#if membershipsErr}
		<InlineAlert
			error={membershipsErr}
			onDismiss={() => view.dismissSectionErrorAlert("memberships")}
		/>
		{/if}

		{#if !view.selectedTeamId}
			<span class="text-surface-content/80">Select a team to manage memberships.</span>
		{:else if view.membershipsQuery.isLoading}
			<span>Loading memberships...</span>
		{:else if view.membershipsQuery.isError}
			<div class="mb-2 rounded border border-danger/40 bg-danger/10 p-3">
				<span class="text-danger">Failed to load memberships.</span>
			</div>
		{:else}
			<div class="flex flex-col gap-2">
				{#each view.memberships as membership (membership.id)}
					<div class="rounded border p-2">
						<div class="flex items-center justify-between gap-2">
							<div class="flex flex-col">
								<span class="font-medium">{membership.attributes.user?.attributes.name || membership.attributes.userId}</span>
								<span class="text-sm text-surface-content/80">{membership.attributes.user?.attributes.email}</span>
							</div>
							<div class="flex items-center gap-2">
								{#if view.isOrgAdmin}
									<select
										class="rounded border bg-surface-200 p-2"
										value={membership.attributes.role}
										onchange={(e) =>
											view.updateMembershipRole(
												membership.id,
												(e.currentTarget as HTMLSelectElement).value as "admin" | "member"
											)}
									>
										<option value="member">member</option>
										<option value="admin">admin</option>
									</select>
									<Button onclick={() => view.removeMembership(membership.id)}>Remove</Button>
								{:else}
									<span class="rounded border px-2 py-1 text-sm">{membership.attributes.role}</span>
								{/if}
							</div>
						</div>
					</div>
				{:else}
					<span class="text-surface-content/80">No members assigned yet.</span>
				{/each}
			</div>

			{#if view.isOrgAdmin}
				<div class="mt-3 grid grid-cols-1 gap-2 md:grid-cols-[1fr_130px_auto]">
					<select
						class="rounded border bg-surface-200 p-2"
						bind:value={selectedAddUserId}
					>
						<option value="">Select user...</option>
						{#each view.availableUsersForSelectedTeam as user (user.id)}
							<option value={user.id}>{user.attributes.name} ({user.attributes.email})</option>
						{/each}
					</select>
					<select class="rounded border bg-surface-200 p-2" bind:value={selectedAddRole}>
						<option value="member">member</option>
						<option value="admin">admin</option>
					</select>
					<Button disabled={!selectedAddUserId} onclick={onAddMembership}>Add Member</Button>
				</div>
			{/if}
		{/if}
	</section>
</div>
