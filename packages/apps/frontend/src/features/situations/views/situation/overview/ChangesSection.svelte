<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import type { BriefServiceView, ChangesState, DeploymentView } from "$features/situations/lib/brief";
	import BriefServiceName from "./BriefServiceName.svelte";
	import { useSituationOverviewController } from "./controller.svelte";

	type Props = { state: ChangesState };
	const { state }: Props = $props();

	const controller = useSituationOverviewController();
</script>

{#snippet serviceNames(label: string, services: BriefServiceView[])}
	<div class="flex flex-col gap-1">
		<h3 class="text-xs text-muted-foreground">{label}</h3>
		<ul class="flex flex-col gap-1">
			{#each services as service (service.id)}
				<li>
					<BriefServiceName {service} />
				</li>
			{/each}
		</ul>
	</div>
{/snippet}

{#snippet deploymentLink(deployment: DeploymentView)}
	{#if deployment.link?.kind === "url"}
		<a
			href={deployment.link.href}
			target="_blank"
			rel="noopener noreferrer"
			class="text-primary underline-offset-2 hover:underline"
		>
			{deployment.title}
		</a>
	{:else if deployment.link?.kind === "evidence"}
		{@const evidenceId = deployment.link.evidenceId}
		<Button
			variant="link"
			size="inline"
			onclick={(event) => controller.openDeploymentEvidence(evidenceId, event.currentTarget)}
		>
			{deployment.title}
		</Button>
	{:else}
		<span>{deployment.title}</span>
	{/if}
{/snippet}

{#snippet deploymentItem(deployment: DeploymentView)}
	<li class="flex flex-col gap-1.5 border-l-2 pl-3 text-sm">
		<div class="flex flex-wrap items-center gap-x-2 gap-y-1">
			<span class="font-medium">{@render deploymentLink(deployment)}</span>
			{#if deployment.environment}
				<span class="text-muted-foreground">{deployment.environment}</span>
			{/if}
			{#if deployment.status}
				<Badge variant="outline">{deployment.status}</Badge>
			{/if}
			{#if deployment.afterOpened}
				<Badge variant="secondary">After the situation opened</Badge>
			{/if}
			<Timestamp value={deployment.deployedAt} class="text-xs text-muted-foreground" />
		</div>
		{#if deployment.unlinkedReason}
			<p class="text-xs text-muted-foreground">{deployment.unlinkedReason}</p>
		{:else if deployment.mergedChanges.length === 0}
			<p class="text-xs text-muted-foreground">No merged changes were linked.</p>
		{:else}
			<ul class="flex flex-col gap-1 text-xs">
				{#each deployment.mergedChanges as change (change.id)}
					<li class="flex flex-wrap items-baseline gap-x-2">
						{#if change.href}
							<a
								href={change.href}
								target="_blank"
								rel="noopener noreferrer"
								class="text-primary underline-offset-2 wrap-anywhere hover:underline"
							>
								{change.title}
							</a>
						{:else}
							<span class="wrap-anywhere">{change.title}</span>
						{/if}
						<span class="text-muted-foreground">{change.detail}</span>
						<span class="text-muted-foreground">· {change.linkLabel}</span>
					</li>
				{/each}
			</ul>
			{#if deployment.mergedChangesCut}
				<p class="text-xs text-muted-foreground">More merged changes were linked than are listed.</p>
			{/if}
		{/if}
	</li>
{/snippet}

<section aria-labelledby="changes-title" class="flex flex-col gap-3">
	<SectionHeading id="changes-title" title="Changes" />

	{#if state.kind === "loading"}
		<div class="flex flex-col gap-2" aria-label="Loading changes">
			<Skeleton class="h-5 w-3/5" />
			<Skeleton class="h-5 w-2/5" />
		</div>
	{:else if state.kind === "error"}
		<div role="alert" class="flex flex-wrap items-center gap-3 text-sm">
			<span>Changes are unavailable.</span>
			<Button variant="outline" size="sm" onclick={controller.retryBrief}>Retry</Button>
		</div>
	{:else if state.kind === "no_services"}
		<p class="text-sm text-muted-foreground">
			This situation has no services, so there are no deployments to list.
		</p>
	{:else if state.kind === "not_reported"}
		<p class="text-sm text-muted-foreground">
			No integration reports deployments, so changes cannot be listed.
		</p>
		{@render serviceNames("Services in this situation", state.services)}
	{:else if state.kind === "nothing_found"}
		<p class="text-sm text-muted-foreground">
			No deployments of these services were reported from {state.window.label}.
		</p>
		{@render serviceNames("Services checked", state.services)}
	{:else}
		<p class="text-xs text-muted-foreground">Deployments from {state.window.label}.</p>
		<ul class="flex flex-col gap-4">
			{#each state.services as service (service.id)}
				<li class="flex flex-col gap-2">
					<BriefServiceName {service} />
					{#if service.deployments.length === 0}
						<p class="text-xs text-muted-foreground">No deployments in this window.</p>
					{:else}
						<ul class="flex flex-col gap-3">
							{#each service.deployments as deployment (deployment.id)}
								{@render deploymentItem(deployment)}
							{/each}
						</ul>
					{/if}
					{#each service.limits as limit (limit)}
						<p class="text-xs text-muted-foreground">{limit}</p>
					{/each}
				</li>
			{/each}
		</ul>
	{/if}
	{#if controller.briefRefreshFailed}
		<div role="status" class="flex items-center gap-3 text-xs text-muted-foreground">
			<span>Refresh failed. Showing previously loaded changes.</span>
			<Button variant="ghost" size="sm" onclick={controller.retryBrief}>Retry</Button>
		</div>
	{/if}
</section>
