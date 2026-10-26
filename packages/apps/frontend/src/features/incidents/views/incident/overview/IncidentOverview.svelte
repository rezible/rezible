<script lang="ts">
	import { Button } from "$components/ui/button";
	import * as Tooltip from "$components/ui/tooltip";
	import { Badge } from "$components/ui/badge";
	import { Skeleton } from "$components/ui/skeleton";
	import PageCanvas from "$components/layout/page-canvas/PageCanvas.svelte";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import * as Popover from "$components/ui/popover";
	import RiAddLine from "remixicon-svelte/icons/add-line";
	import RiArrowRightLine from "remixicon-svelte/icons/arrow-right-line";
	import RiCloseLine from "remixicon-svelte/icons/close-line";
	import RiPencilLine from "remixicon-svelte/icons/pencil-line";
	import RiServerLine from "remixicon-svelte/icons/server-line";
	import SituationCombobox from "$features/situations/components/situation-combobox/SituationCombobox.svelte";
	import { useIncidentResponse } from "../response.svelte";
	import { initIncidentOverviewController, type LinkedSituationView } from "./controller.svelte";
	import IncidentOverviewContext from "./IncidentOverviewContext.svelte";

	const overview = initIncidentOverviewController();
	const response = useIncidentResponse();
</script>

{#snippet context()}
	<IncidentOverviewContext />
{/snippet}

{#snippet unlinkSituation(situation: LinkedSituationView)}
	<Popover.Root>
		<Popover.Trigger>
			{#snippet child({ props })}
				<Button
					{...props}
					variant="ghost"
					size="icon-sm"
					class="ml-auto"
					aria-label="Unlink {situation.title}"
				>
					<RiCloseLine />
				</Button>
			{/snippet}
		</Popover.Trigger>
		<Popover.Content class="flex w-64 flex-col gap-3" align="end">
			<p class="text-sm">Unlink this situation? It will be re-evaluated and may close.</p>
			<Button
				size="sm"
				variant="outline"
				class="self-end"
				onclick={() => response.unlinkSituation(situation.id)}
			>
				Unlink
			</Button>
		</Popover.Content>
	</Popover.Root>
{/snippet}

{#snippet servicePill(name: string)}
	<RiServerLine class="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
	{name}
{/snippet}

{#if overview.attributes}
	<PageCanvas {context} contextLabel="Incident context">
		<header class="flex flex-col gap-3">
			<h1 class="text-[28px] leading-9 font-semibold tracking-tight wrap-anywhere">{overview.title}</h1>
			{#if overview.declaredBy}
				<p class="text-sm text-muted-foreground">
					Declared by {overview.declaredBy} ·
					<Timestamp value={overview.attributes.openedAt} format="absolute" />
				</p>
			{/if}
			{#if overview.summary}
				<p
					class="max-w-[72ch] text-[15px] leading-6 whitespace-pre-wrap text-muted-foreground wrap-anywhere"
				>
					{overview.summary}
				</p>
			{/if}
		</header>

		<dl class="grid grid-cols-2 gap-x-6 gap-y-4 border-y py-4 sm:grid-cols-3">
			{#each overview.facts as fact (fact.key)}
				<div class="flex min-w-0 flex-col gap-1">
					<dt class="text-xs text-muted-foreground">{fact.label}</dt>
					<dd class="text-sm font-medium wrap-anywhere">
						{#if fact.kind === "status"}
							<StatusBadge status={fact.status} />
						{:else if fact.kind === "time"}
							{#if fact.value}
								<Timestamp value={fact.value} format="absolute" />
							{:else}
								<span class="font-normal text-muted-foreground">{fact.fallback}</span>
							{/if}
						{:else}
							{fact.text}
						{/if}
					</dd>
				</div>
			{/each}
		</dl>

		<section aria-labelledby="affected-services-title" class="flex flex-col gap-3">
			<SectionHeading
				id="affected-services-title"
				title="Affected services"
				count={overview.services.length}
			/>
			{#if overview.services.length}
				<ul class="flex flex-wrap gap-2">
					{#each overview.services as service (service.key)}
						<li>
							{#if service.note || service.source}
								<Tooltip.Root>
									<Tooltip.Trigger
										class="inline-flex h-8 items-center gap-2 rounded-full border bg-card px-3 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
									>
										{@render servicePill(service.name)}
									</Tooltip.Trigger>
									<Tooltip.Content class="max-w-72">
										{#if service.note}
											<p>{service.note}</p>
										{/if}
										{#if service.source}
											<p>Source: {service.source}</p>
										{/if}
									</Tooltip.Content>
								</Tooltip.Root>
							{:else}
								<span
									class="inline-flex h-8 items-center gap-2 rounded-full border bg-card px-3 text-sm"
								>
									{@render servicePill(service.name)}
								</span>
							{/if}
						</li>
					{/each}
				</ul>
			{:else}
				<p class="text-sm text-muted-foreground">No affected services recorded.</p>
			{/if}
		</section>

		<section aria-labelledby="linked-situations-title" class="flex flex-col gap-3">
			<SectionHeading
				id="linked-situations-title"
				title="Linked situations"
				count={overview.situationLinks.length}
			>
				{#snippet actions()}
					<SituationCombobox
						linkedIds={response.linkedSituationIds}
						onSelect={response.linkSituation}
					>
						{#snippet trigger(props)}
							<Button {...props} variant="outline" size="sm">
								<RiAddLine />
								Link situation
							</Button>
						{/snippet}
					</SituationCombobox>
				{/snippet}
			</SectionHeading>
			{#if overview.situationLinks.length}
				<ul class="divide-y rounded-lg border bg-card">
					{#each overview.situationLinks as situation (situation.id)}
						<li class="flex min-h-12 flex-col justify-center gap-2 px-4 py-2">
							<div class="flex items-center gap-2">
								<a
									class="min-w-0 truncate text-sm font-medium hover:underline focus-visible:outline-2 focus-visible:outline-ring"
									href={situation.href}
								>
									{situation.title}
								</a>
								<StatusBadge status={situation.status} variant="inline" />
								{#if situation.closedAt}
									<span class="text-xs text-muted-foreground">
										closed <Timestamp value={situation.closedAt} format="absolute" />
									</span>
								{/if}
								{@render unlinkSituation(situation)}
							</div>
							{#if situation.conclusion}
								<div class="flex flex-col gap-1 pb-1">
									<span class="text-xs text-muted-foreground">Current understanding</span>
									<p class="max-w-[65ch] text-[15px] leading-6">{situation.conclusion}</p>
									{#if situation.reportHref}
										<Button
											variant="link"
											size="inline"
											class="self-start"
											href={situation.reportHref}
										>
											Read report
											<RiArrowRightLine aria-hidden="true" />
										</Button>
									{/if}
								</div>
							{/if}
						</li>
					{/each}
				</ul>
			{:else}
				<p class="text-sm text-muted-foreground">No situations are linked to this incident.</p>
			{/if}
		</section>

		<section aria-labelledby="milestones-title" class="flex flex-col gap-3">
			<SectionHeading id="milestones-title" title="Milestones" count={overview.milestones.length}>
				{#snippet actions()}
					{#if response.enabled}
						<Button variant="outline" size="sm" onclick={response.openRecordMilestone}>
							Record milestone
						</Button>
					{/if}
				{/snippet}
			</SectionHeading>
			{#if overview.milestones.length}
				<ol aria-label="Incident milestones" class="flex flex-col">
					{#each overview.milestones as entry (entry.id)}
						<li
							class="group relative grid grid-cols-[12px_minmax(0,1fr)_auto] gap-x-3 pb-5 before:absolute before:top-4 before:bottom-0 before:left-[5.5px] before:w-px before:bg-border last:pb-0 last:before:hidden"
						>
							<span
								class="mt-1.5 size-2 justify-self-center rounded-full {entry.marker}"
								aria-hidden="true"
							></span>
							<div class="flex min-w-0 flex-col gap-0.5">
								<span class="flex items-baseline gap-2 text-sm font-medium">
									{entry.label}
									{#if entry.caption}
										<span class="text-xs font-normal text-muted-foreground">
											{entry.caption}
										</span>
									{/if}
								</span>
								<Timestamp
									value={entry.at}
									format="absolute"
									class="self-start text-xs text-muted-foreground"
								/>
								{#if entry.description}
									<p class="text-sm whitespace-pre-wrap text-muted-foreground">
										{entry.description}
									</p>
								{/if}
								{#if entry.meta}
									<span class="text-xs text-muted-foreground">{entry.meta}</span>
								{/if}
							</div>
							{#if response.enabled}
								<Button
									variant="ghost"
									size="icon-sm"
									class="opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
									aria-label="Edit {entry.label} milestone"
									onclick={() => response.openEditMilestone(entry.milestone)}
								>
									<RiPencilLine />
								</Button>
							{/if}
						</li>
					{/each}
				</ol>
			{:else}
				<p class="text-sm text-muted-foreground">No milestones recorded.</p>
			{/if}
		</section>

		{#if overview.tags.length}
			<section aria-labelledby="tags-title" class="flex flex-col gap-2">
				<h2 id="tags-title" class="region-label">Tags</h2>
				<ul class="flex flex-wrap gap-2" aria-label="Tags">
					{#each overview.tags as tag (tag.id)}
						<li><Badge variant="neutral">{tag.label}</Badge></li>
					{/each}
				</ul>
			</section>
		{/if}
	</PageCanvas>
{:else}
	<PageCanvas>
		<div class="flex flex-col gap-3" aria-label="Loading incident">
			<Skeleton class="h-9 w-2/3" />
			<Skeleton class="h-5 w-full" />
			<Skeleton class="h-5 w-4/5" />
		</div>
	</PageCanvas>
{/if}
