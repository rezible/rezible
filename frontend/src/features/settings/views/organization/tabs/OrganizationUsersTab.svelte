<script lang="ts">
	import { Button } from "$components/ui/button";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";
	import { useOrganizationSettingsViewController } from "../organizationSettingsViewController.svelte";

	const view = useOrganizationSettingsViewController();

	const setUserSearch = (e: Event) => {
		const target = e.currentTarget as HTMLInputElement;
		view.setUserSearch(target.value);
	};

	const usersError = $derived(view.sectionErrors.get("users"));
	const membershipsError = $derived(view.sectionErrors.get("memberships"));
</script>

<section class="rounded border p-3">
	<div class="mb-2 flex flex-wrap items-center justify-between gap-2">
		<h3 class="text-lg font-semibold">Users</h3>
		<div class="flex items-center gap-2">
			<span class="text-sm text-surface-content/80">Assignment Team</span>
			<select
				class="rounded border bg-surface-200 p-2"
				value={view.selectedTeamId ?? ""}
				onchange={(e) => view.setSelectedTeamId((e.currentTarget as HTMLSelectElement).value || undefined)}
			>
				<option value="">None</option>
				{#each view.teams as team (team.id)}
					<option value={team.id}>{team.attributes.name}</option>
				{/each}
			</select>
		</div>
	</div>

	{#if usersError}
	<InlineAlert
		error={usersError}
		onDismiss={() => {
			view.dismissSectionErrorAlert("users");
			view.dismissSectionErrorAlert("memberships");
		}}
	/>
	{/if}

	<div class="mb-3">
		<input
			class="w-full rounded border bg-surface-200 p-2"
			placeholder="Search users..."
			oninput={setUserSearch}
		/>
	</div>

	{#if view.usersQuery.isLoading}
		<span>Loading users...</span>
	{:else if view.usersQuery.isError}
		<div class="mb-2 rounded border border-danger/40 bg-danger/10 p-3">
			<span class="text-danger">Failed to load users.</span>
		</div>
	{:else}
		<div class="overflow-x-auto">
			<table class="min-w-full border-collapse text-sm">
				<thead>
					<tr class="border-b">
						<th class="px-2 py-2 text-left">Name</th>
						<th class="px-2 py-2 text-left">Email</th>
						<th class="px-2 py-2 text-left">Org Role</th>
						<th class="px-2 py-2 text-left">Team Memberships</th>
						<th class="px-2 py-2 text-left">Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each view.users as user (user.id)}
						{@const memberships = view.membershipsForUser(user.id)}
						{@const selectedTeamMembership = view.teamMembershipForUser(user.id)}
						<tr class="border-b align-top">
							<td class="px-2 py-2">{user.attributes.name}</td>
							<td class="px-2 py-2">{user.attributes.email}</td>
							<td class="px-2 py-2">{user.attributes.isOrgAdmin ? "Org Admin" : "User"}</td>
							<td class="px-2 py-2">
								<div class="flex flex-col gap-1">
									{#each memberships as membership (membership.id)}
										<span class="rounded border px-2 py-1">
											{membership.attributes.team?.attributes.name || membership.attributes.teamId}
											({membership.attributes.role})
										</span>
									{:else}
										<span class="text-surface-content/80">None</span>
									{/each}
								</div>
							</td>
							<td class="px-2 py-2">
								{#if !view.selectedTeamId}
									<span class="text-surface-content/80">Choose assignment team</span>
								{:else if !view.isOrgAdmin}
									<span class="text-surface-content/80">Read only</span>
								{:else if selectedTeamMembership}
									<div class="flex items-center gap-2">
										<select
											class="rounded border bg-surface-200 p-2"
											value={selectedTeamMembership.attributes.role}
											onchange={(e) =>
												view.updateMembershipRole(
													selectedTeamMembership.id,
													(e.currentTarget as HTMLSelectElement).value as "admin" | "member"
												)}
										>
											<option value="member">member</option>
											<option value="admin">admin</option>
										</select>
										<Button onclick={() => view.removeMembership(selectedTeamMembership.id)}>Remove</Button>
									</div>
								{:else}
									<Button onclick={() => view.addMembership(user.id, "member")}>Add as member</Button>
								{/if}
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="5" class="px-2 py-3 text-surface-content/80">No users found.</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
