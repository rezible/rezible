<script lang="ts">
	import { Button } from "$components/ui/button";
	import * as Tooltip from "$components/ui/tooltip";
	import { Badge } from "$components/ui/badge";
	import { Skeleton } from "$components/ui/skeleton";
	import PageCanvas from "$components/layout/page-canvas/PageCanvas.svelte";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import Timeline from "$components/common/timeline/Timeline.svelte";
	import RiArrowRightLine from "remixicon-svelte/icons/arrow-right-line";
	import RiServerLine from "remixicon-svelte/icons/server-line";
	import { initIncidentOverviewController } from "./controller.svelte";
	import IncidentOverviewContext from "./IncidentOverviewContext.svelte";

	const overview = initIncidentOverviewController();
</script>

{#snippet context()}
	<IncidentOverviewContext />
{/snippet}

{#snippet servicePill(name: string)}
	<RiServerLine class="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
	{name}
{/snippet}

{#if overview.attributes}
	<PageCanvas {context} contextLabel="Incident context">
		<header class="flex flex-col gap-3">
			<h1 class="text-[28px] leading-9 font-semibold tracking-tight wrap-anywhere">{overview.title}</h1>
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

		<section aria-labelledby="understanding-title" class="flex flex-col gap-3">
			<div class="flex flex-col gap-1">
				<SectionHeading id="understanding-title" title="Current understanding" />
				<p class="text-sm text-muted-foreground">
					Latest conclusions from investigations on linked situations.
				</p>
			</div>
			{#if overview.understanding.length}
				<ul class="divide-y rounded-lg border bg-card">
					{#each overview.understanding as item (item.id)}
						<li class="flex min-h-[60px] flex-col gap-2 px-4 py-3">
							<div class="flex flex-wrap items-center gap-x-2 gap-y-1">
								<a
									class="rounded-sm text-[15px] leading-[22px] font-medium wrap-anywhere hover:underline focus-visible:outline-2 focus-visible:outline-ring"
									href={item.href}
								>
									{item.title}
								</a>
								{#if item.status}
									<StatusBadge status={item.status} variant="inline" />
								{/if}
							</div>
							{#if item.conclusion.kind === "loading"}
								<Skeleton class="h-4 w-3/4" />
							{:else if item.conclusion.kind === "text"}
								<p class="max-w-[65ch] text-[15px] leading-6">{item.conclusion.text}</p>
							{:else}
								<p class="text-sm text-muted-foreground">No investigation conclusion yet.</p>
							{/if}
							{#if item.reportHref}
								<Button
									variant="link"
									size="inline"
									class="self-start"
									href={item.reportHref}
								>
									Read report
									<RiArrowRightLine aria-hidden="true" />
								</Button>
							{/if}
						</li>
					{/each}
				</ul>
				{#if overview.moreLinkedSituations > 0}
					<p class="text-sm text-muted-foreground">
						{overview.moreLinkedSituations} more linked
						{overview.moreLinkedSituations === 1 ? "situation" : "situations"}
					</p>
				{/if}
			{:else}
				<p class="text-sm text-muted-foreground">No situations are linked to this incident.</p>
			{/if}
		</section>

		<section aria-labelledby="milestones-title" class="flex flex-col gap-3">
			<SectionHeading id="milestones-title" title="Milestones" count={overview.milestones.length} />
			{#if overview.milestones.length}
				<Timeline entries={overview.milestones} label="Incident milestones" timeFormat="absolute" />
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
