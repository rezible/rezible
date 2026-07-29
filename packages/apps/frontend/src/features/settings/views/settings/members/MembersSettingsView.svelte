<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";
	import { Badge } from "$components/ui/badge";
	import * as Card from "$components/ui/card";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import { initMembersSettingsController } from "./controller.svelte";

	const view = initMembersSettingsController();

	setPageBreadcrumbs(() => [
		{ label: "Settings", path: "/settings" },
		{ label: "Members", path: "/settings/organization/members" },
	]);
</script>

<div class="flex max-w-4xl flex-col gap-4">
	<div class="grid max-w-sm gap-1.5">
		<Label for="member-search">Search members</Label>
		<Input id="member-search" bind:value={view.search} placeholder="Name" />
	</div>

	{#if view.loading}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<LoadingIndicator />
			<span>Loading members...</span>
		</div>
	{:else if view.error}
		<InlineAlert error={view.error} />
	{:else}
		<Card.Root>
			<Card.Header>
				<Card.Title>Members</Card.Title>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="divide-y">
					{#each view.users as user (user.id)}
						<div class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-4 px-6 py-3">
							<div class="min-w-0">
								<div class="truncate font-medium">
									{user.attributes.name || user.attributes.email}
								</div>
								<div class="truncate text-sm text-muted-foreground">
									{user.attributes.email}
								</div>
							</div>
							<Badge
								variant={user.attributes.organizationRole === "admin"
									? "default"
									: "secondary"}
							>
								{user.attributes.organizationRole}
							</Badge>
						</div>
					{:else}
						<div class="px-6 py-8 text-sm text-muted-foreground">No members found.</div>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
