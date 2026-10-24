<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";
	import * as Avatar from "$components/ui/avatar";
	import EntityAvatar from "$components/common/entity-avatar/EntityAvatar.svelte";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import * as Tooltip from "$components/ui/tooltip";
	import MemberPicker from "$components/common/member-picker/MemberPicker.svelte";
	import RiCloseLine from "remixicon-svelte/icons/close-line";
	import RiExternalLinkLine from "remixicon-svelte/icons/external-link-line";
	import { resolve } from "$app/paths";
	import { useIncidentResponse } from "../response.svelte";
	import { useIncidentOverviewController } from "./controller.svelte";

	const overview = useIncidentOverviewController();
	const response = useIncidentResponse();
	const review = $derived(overview.review);
</script>

<section aria-labelledby="review-title" class="flex flex-col gap-3">
	<SectionHeading id="review-title" title="Review" level={3} />
	{#if review.kind === "loading"}
		<Skeleton class="h-5 w-24" />
		<Skeleton class="h-4 w-full" />
	{:else if review.kind === "error"}
		<div role="alert" class="flex flex-wrap items-center gap-3 text-sm">
			<span>Review status could not be loaded.</span>
			<Button variant="outline" size="sm" onclick={review.retry}>Retry</Button>
		</div>
	{:else}
		<div>
			<StatusBadge status={review.status} />
		</div>
		<p class="text-sm text-muted-foreground">{review.explanation}</p>
		{#if review.ownerName}
			<p class="text-sm">Owner: {review.ownerName}</p>
		{/if}
		{#if review.reportHref}
			<div>
				<Button
					size="sm"
					variant={review.primaryAction ? "default" : "outline"}
					href={review.reportHref}
				>
					Open report
				</Button>
			</div>
		{/if}
	{/if}
</section>

<section aria-labelledby="roles-title" class="flex flex-col gap-3">
	<SectionHeading id="roles-title" title="Roles" level={3} />
	{#if response.hasRoles}
		{#each response.roleGroups as group (group.roleId)}
			{@const holderIds = group.holders.map((holder) => holder.userId)}
			<div class="flex flex-col gap-2">
				<div class="flex items-center justify-between gap-2">
					<span class="text-sm font-medium">
						{group.name}
						<span class="font-normal text-muted-foreground tabular-nums">{group.countLabel}</span>
						{#if group.archived}
							<span class="font-normal text-muted-foreground">· archived</span>
						{/if}
					</span>
					{#if response.enabled && !group.archived}
						{#if group.full}
							<Tooltip.Root>
								<Tooltip.Trigger>
									{#snippet child({ props })}
										<span {...props}>
											<Button variant="link" size="inline" disabled>Assign</Button>
										</span>
									{/snippet}
								</Tooltip.Trigger>
								<Tooltip.Content class="max-w-64">{group.fullReason}</Tooltip.Content>
							</Tooltip.Root>
						{:else}
							<MemberPicker
								excludeIds={holderIds}
								onSelect={(user) => response.assignRole(group.roleId, user.id)}
							>
								{#snippet trigger(props)}
									<Button {...props} variant="link" size="inline">Assign</Button>
								{/snippet}
							</MemberPicker>
						{/if}
					{/if}
				</div>
				{#if group.holders.length}
					<ul class="flex flex-col gap-2">
						{#each group.holders as holder (holder.assignmentId)}
							<li class="group flex items-center gap-3">
								<Avatar.Root class="size-7 shrink-0 [&_svg]:size-7!" loadingStatus="loaded">
									<EntityAvatar kind="user" id={holder.userId} size={28} />
								</Avatar.Root>
								<span class="min-w-0 flex-1 truncate text-sm">{holder.name}</span>
								{#if response.enabled}
									<Button
										variant="ghost"
										size="icon-sm"
										class="opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
										aria-label="Remove {holder.name} as {group.name}"
										onclick={() => response.unassignRole(holder.assignmentId)}
									>
										<RiCloseLine />
									</Button>
								{/if}
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		{/each}
	{:else}
		<p class="text-sm text-muted-foreground">
			No roles configured
			{#if response.isAdmin}
				·
				<a class="underline underline-offset-4" href={resolve("/settings/incidents")}>
					Add them in settings
				</a>
			{/if}
		</p>
	{/if}
</section>

{#if overview.teams.length}
	<section aria-labelledby="teams-title" class="flex flex-col gap-2">
		<SectionHeading id="teams-title" title="Teams" level={3} />
		<ul class="flex flex-col gap-1 text-sm">
			{#each overview.teams as team (team.id)}
				<li>{team.name}</li>
			{/each}
		</ul>
	</section>
{/if}

{#if overview.ticket}
	<section aria-labelledby="links-title" class="flex flex-col gap-2">
		<SectionHeading id="links-title" title="Links" level={3} />
		<Button
			variant="link"
			size="inline"
			class="self-start"
			href={overview.ticket.url}
			target="_blank"
			rel="noreferrer"
		>
			{overview.ticket.label}
			<RiExternalLinkLine aria-hidden="true" />
		</Button>
	</section>
{/if}
