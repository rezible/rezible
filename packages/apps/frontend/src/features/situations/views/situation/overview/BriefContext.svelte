<script lang="ts">
	import { resolve } from "$app/paths";
	import { Skeleton } from "$components/ui/skeleton";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timeline from "$components/common/timeline/Timeline.svelte";
	import { incidentResponseStatus, incidentSeverityStatus } from "$features/incidents/lib/status";
	import { useSituationOverviewController } from "./controller.svelte";

	const controller = useSituationOverviewController();
</script>

<section aria-labelledby="linked-incidents-title" class="flex flex-col gap-3">
	<SectionHeading id="linked-incidents-title" title="Linked incidents" level={3} />
	{#if controller.linkedIncidents.length}
		<ul class="flex flex-col gap-3">
			{#each controller.linkedIncidents as incident (incident.id)}
				<li class="flex flex-col gap-1.5">
					<a
						class="text-sm font-medium wrap-anywhere hover:underline focus-visible:outline-2 focus-visible:outline-ring rounded-sm"
						href={resolve("/incidents/[slug]/[[view=incidentView]]", {
							slug: incident.attributes.slug,
						})}
					>
						{incident.attributes.title}
					</a>
					<div class="flex flex-wrap items-center gap-2">
						<StatusBadge
							status={incidentResponseStatus(incident.attributes.responseState)}
							variant="inline"
						/>
						<StatusBadge status={incidentSeverityStatus(incident.attributes.severity)} />
					</div>
				</li>
			{/each}
		</ul>
	{:else if controller.linkedIncidentsLoading}
		<Skeleton class="h-4 w-2/3" />
	{:else}
		<p class="text-sm text-muted-foreground">Not linked to an incident.</p>
	{/if}
</section>

<section aria-labelledby="related-situations-title" class="flex flex-col gap-3">
	<SectionHeading id="related-situations-title" title="Related situations" level={3} />
	{#if controller.relatedSituations.length}
		<ul class="flex max-h-80 flex-col gap-3 overflow-y-auto">
			{#each controller.relatedSituations as related (related.key)}
				<li class="flex min-w-0 flex-col gap-1">
					<span class="text-xs text-muted-foreground">{related.kindLabel}</span>
					<a
						class="rounded-sm text-sm font-medium wrap-anywhere hover:underline focus-visible:outline-2 focus-visible:outline-ring"
						href={related.href}
					>
						{related.title}
					</a>
					<div>
						<StatusBadge status={related.status} variant="inline" />
					</div>
				</li>
			{/each}
		</ul>
	{:else}
		<p class="text-sm text-muted-foreground">No related situations.</p>
	{/if}
</section>

<section aria-labelledby="brief-timeline-title" class="flex flex-col gap-3">
	<SectionHeading id="brief-timeline-title" title="Timeline" level={3} />
	{#if controller.timeline.length}
		<Timeline entries={controller.timeline} label="Situation timeline" />
	{:else}
		<p class="text-sm text-muted-foreground">No timeline events.</p>
	{/if}
</section>
