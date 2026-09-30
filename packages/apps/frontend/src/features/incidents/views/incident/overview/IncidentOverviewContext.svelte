<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";
	import { Spinner } from "$components/ui/spinner";
	import * as Avatar from "$components/ui/avatar";
	import EntityAvatar from "$components/common/entity-avatar/EntityAvatar.svelte";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import RiExternalLinkLine from "remixicon-svelte/icons/external-link-line";
	import { useIncidentOverviewController } from "./controller.svelte";

	const overview = useIncidentOverviewController();
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
		{:else if review.canStart}
			<div>
				<Button
					size="sm"
					disabled={overview.startRetrospectivePending}
					onclick={overview.startRetrospective}
				>
					{#if overview.startRetrospectivePending}
						<Spinner data-icon="inline-start" />
						Starting…
					{:else}
						Start retrospective
					{/if}
				</Button>
			</div>
		{/if}
	{/if}
</section>

{#if overview.roles.length}
	<section aria-labelledby="roles-title" class="flex flex-col gap-3">
		<SectionHeading id="roles-title" title="Roles" level={3} />
		<ul class="flex flex-col gap-3">
			{#each overview.roles as role (role.id)}
				<li class="flex items-center gap-3">
					{#if role.userId}
						<Avatar.Root class="size-7 shrink-0 [&_svg]:size-7!" loadingStatus="loaded">
							<EntityAvatar kind="user" id={role.userId} size={28} />
						</Avatar.Root>
					{/if}
					<div class="flex min-w-0 flex-col">
						<span class="truncate text-sm font-medium">{role.userName}</span>
						<span class="text-xs text-muted-foreground">{role.roleName}</span>
					</div>
				</li>
			{/each}
		</ul>
	</section>
{/if}

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
