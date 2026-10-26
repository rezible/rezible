<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";
	import SearchInput from "$components/forms/search-input/SearchInput.svelte";
	import { Badge } from "$components/ui/badge";
	import * as Card from "$components/ui/card";
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { memberRolePresentation } from "$features/settings/lib/memberRoles";
	import { initMembersSettingsController } from "./controller.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";

	const view = initMembersSettingsController();

	registerPageDescriptor(() => ({
		title: "Members",
		parents: [
			{ label: "Settings", path: resolve("/settings") },
			{ label: "Organization", path: resolve("/settings/organization") },
		],
	}));
</script>

<div class="flex max-w-4xl flex-col gap-4">
	<div class="max-w-sm">
		<SearchInput label="Search members" bind:value={view.search} />
	</div>

	{#if view.loading}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<LoadingIndicator />
			<span>Loading members...</span>
		</div>
	{:else if view.error}
		<InlineAlert error={view.error} />
	{:else}
		<PaginatedQueryListBox {...view.paginatedUsersQuery}>
			<Card.Root class="py-0">
				<Card.Content class="divide-y p-0">
					{#each view.users as user (user.id)}
						{@const role = memberRolePresentation(user.attributes.organizationRole)}
						<div class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-4 px-6 py-3">
							<div class="min-w-0">
								<div class="truncate font-medium">
									{user.attributes.name || user.attributes.email}
								</div>
								<div class="truncate text-sm text-muted-foreground">
									{user.attributes.email}
								</div>
							</div>
							<Badge variant={role.badgeVariant}>{role.label}</Badge>
						</div>
					{:else}
						<div class="px-6 py-8 text-sm text-muted-foreground">No members found.</div>
					{/each}
				</Card.Content>
			</Card.Root>
		</PaginatedQueryListBox>
	{/if}
</div>
